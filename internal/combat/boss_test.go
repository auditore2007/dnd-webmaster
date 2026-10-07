package combat

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"heroesbook/internal/bestiary"
	"heroesbook/internal/model"
	"heroesbook/internal/rules"
)

func monster(t *testing.T, m map[string]*model.Character, id, kind string) *model.Character {
	t.Helper()
	n := 0
	c, err := bestiary.Make(kind, 1, func() string { n++; return fmt.Sprintf("%s%d", id, n) })
	if err != nil {
		t.Fatal(err)
	}
	c.ID = id
	m[id] = c
	return c
}

// encounterOf собирает бой с заданным порядком ходов (без бросков инициативы).
func encounterOf(order ...string) *Encounter {
	e := &Encounter{Order: order, Init: map[string]int{}, Round: 1, React: map[string]bool{}, Auto: true, Morale: true}
	for i, id := range order {
		e.Init[id] = 20 - i
	}
	e.Left = 1
	return e
}

func hasLog(e *Encounter, part string) bool {
	return slices.ContainsFunc(e.Log, func(l string) bool { return strings.Contains(l, part) })
}

func TestMonsterTurnAttacksAndPassesTurn(t *testing.T) {
	m, get := arena(t)
	monster(t, m, "g", "goblin")
	e := encounterOf("g", "a")
	r := &seq{v: []int{15}}
	res, names, err := e.MonsterTurn(r, get)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) == 0 || names[0] != "a" {
		t.Fatalf("гоблин должен атаковать героя: %v %v", res, names)
	}
	if e.Current() != "a" {
		t.Errorf("после хода существа ходит %s, ожидался герой", e.Current())
	}
	if _, _, err := e.MonsterTurn(r, get); err == nil {
		t.Error("в ход героя автоход существа должен вернуть ошибку")
	}
}

func TestSmartMonstersFocusTheWeakest(t *testing.T) {
	m, get := arena(t)
	g := monster(t, m, "g", "goblin")
	g.Abilities["int"] = 10
	m["b"].HP = 1
	foes := []*model.Character{m["a"], m["b"], m["c"]}
	if got := pickTarget(&seq{v: []int{0}}, g, foes); got.ID != "b" {
		t.Errorf("разумное существо должно бить самого раненого, выбрало %s", got.ID)
	}
	_ = get
}

func TestMoraleSurrenderEndsBattle(t *testing.T) {
	m, get := arena(t)
	g := monster(t, m, "g", "goblin")
	g.HP = 1
	e := encounterOf("g", "a")
	r := &seq{v: []int{0}} // d20 = 1 – мораль проваливается
	e.seek(r, get)
	if slices.Contains(e.Order, "g") || !slices.Contains(e.Gone, "g") {
		t.Fatalf("раненый гоблин должен покинуть бой: порядок %v, ушли %v", e.Order, e.Gone)
	}
	if !hasLog(e, "сдаётся") && !hasLog(e, "бежит") {
		t.Errorf("в журнале нет бегства: %v", e.Log)
	}
	if w := e.Winner(get); w != "heroes" {
		t.Errorf("все враги ушли – победа героев, получено %q", w)
	}
}

func TestMoraleOffAndFearlessUndead(t *testing.T) {
	m, get := arena(t)
	z := monster(t, m, "z", "zombie")
	z.HP = 1
	e := encounterOf("z", "a")
	e.seek(&seq{v: []int{0}}, get)
	if !slices.Contains(e.Order, "z") {
		t.Error("нежить не знает страха")
	}
	g := monster(t, m, "g", "goblin")
	g.HP = 1
	e = encounterOf("g", "a")
	e.Morale = false
	e.seek(&seq{v: []int{0}}, get)
	if !slices.Contains(e.Order, "g") {
		t.Error("с выключенной моралью враги не бегут")
	}
}

func TestBossPhaseTriggersOnce(t *testing.T) {
	m, get := arena(t)
	d := monster(t, m, "d", "adult-red-dragon")
	atk := d.Stat.Attack
	e := encounterOf("a", "d")
	d.HP = d.Stat.MaxHP * 2 / 5
	e.settle(&seq{v: []int{10}}, get)
	e.settle(&seq{v: []int{10}}, get)
	if d.Stat.Attack != atk+2 || !hasLog(e, "вторая фаза") {
		t.Errorf("вторая фаза: атака %d → %d, журнал %v", atk, d.Stat.Attack, e.Log)
	}
}

// tough даёт героям запас здоровья, чтобы удары босса не заканчивали тест раньше времени.
func tough(m map[string]*model.Character) {
	for _, c := range m {
		if c.Stat == nil {
			c.HP = 500
		}
	}
}

func TestLegendaryActionsSpendAndReset(t *testing.T) {
	m, get := arena(t)
	tough(m)
	d := monster(t, m, "d", "adult-red-dragon")
	e := encounterOf("a", "d")
	r := &seq{v: []int{10}}
	if _, err := e.Legendary(r, get, "d", "tail", "a"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Legendary(r, get, "d", "wing", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Legendary(r, get, "d", "tail", "a"); err == nil {
		t.Error("за раунд только 3 очка легендарных действий")
	}
	rs, _ := rules.Get(d.Ruleset)
	rs.TurnStart(r, d)
	if rules.LegendaryLeft(d) != 3 {
		t.Errorf("в начале своего хода очки восстанавливаются, осталось %d", rules.LegendaryLeft(d))
	}
	e.Turn = 1
	if _, err := e.Legendary(r, get, "d", "tail", "a"); err == nil {
		t.Error("в свой ход легендарные действия недоступны")
	}
}

func TestAutoLegendaryAfterOtherTurns(t *testing.T) {
	m, get := arena(t)
	tough(m)
	d := monster(t, m, "d", "adult-red-dragon")
	e := encounterOf("a", "b", "d")
	e.Next(&seq{v: []int{10}}, get)
	if rules.LegendaryLeft(d) == d.Stat.Legendary {
		t.Errorf("в автоматическом режиме босс тратит легендарное действие после чужого хода: %v", e.Log)
	}
}

func TestLairActsOncePerRound(t *testing.T) {
	m, get := arena(t)
	monster(t, m, "d", "adult-red-dragon")
	e := encounterOf("a", "d")
	e.Lair = true
	r := &seq{v: []int{10}}
	e.lairTurn(r, get)
	n := len(e.Log)
	e.lairTurn(r, get)
	if !hasLog(e, "Логово") || len(e.Log) != n {
		t.Errorf("логово действует один раз за раунд: %v", e.Log)
	}
}

// Двое существ сбегают подряд: следующий участник всё равно получает полноценное начало хода.
func TestTwoFleeInARowStillStartsNextTurn(t *testing.T) {
	m, get := arena(t)
	for _, id := range []string{"g0", "g1", "g2"} {
		monster(t, m, id, "goblin")
	}
	m["g1"].HP, m["g2"].HP = 1, 1
	e := encounterOf("g0", "g1", "g2", "a")
	e.Acted, e.Left = true, 0 // g0 уже отходил
	e.Next(&seq{v: []int{0}}, get)
	if e.Current() != "a" || e.Acted || e.Left < 1 {
		t.Fatalf("после бегства двоих ходит %s, acted=%v left=%d", e.Current(), e.Acted, e.Left)
	}
}

// Последний участник не может сбежать: очередь не пустеет, бой не падает.
func TestLastCombatantCannotFlee(t *testing.T) {
	m, get := arena(t)
	monster(t, m, "g1", "goblin").HP = 1
	monster(t, m, "g2", "goblin").HP = 1
	e := encounterOf("g1", "g2")
	for i := 0; i < 4; i++ {
		e.Next(&seq{v: []int{0}}, get)
	}
	if len(e.Order) == 0 {
		t.Fatal("очередь опустела")
	}
	_ = e.Current()
}

// Конец боя снимает бонусы второй фазы и возвращает легендарные сопротивления.
func TestCleanupResetsBossState(t *testing.T) {
	m, get := arena(t)
	d := monster(t, m, "d", "adult-red-dragon")
	atk, ac, multi := d.Stat.Attack, d.ACBonus, len(d.Stat.Multi)
	e := encounterOf("a", "d")
	d.HP = d.Stat.MaxHP / 3
	e.settle(&seq{v: []int{10}}, get)
	d.Count("legres", 2)
	e.Cleanup(get)
	if d.Stat.Attack != atk || d.ACBonus != ac || len(d.Stat.Multi) != multi || d.Counters["legres"] != 0 {
		t.Errorf("после боя: атака %d (было %d), КД+ %d, мультиатака %d, сопротивлений потрачено %d", d.Stat.Attack, atk, d.ACBonus, len(d.Stat.Multi), d.Counters["legres"])
	}
}
