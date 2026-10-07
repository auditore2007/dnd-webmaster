package combat

import (
	"testing"

	"heroesbook/internal/bestiary"
	"heroesbook/internal/dice"
	"heroesbook/internal/model"
	"heroesbook/internal/rules"
)

type seq struct {
	v []int
	i int
}

func (s *seq) Intn(n int) int { x := s.v[s.i%len(s.v)]; s.i++; return x % n }

func hero(t *testing.T, id, ruleset, race, class string) *model.Character {
	t.Helper()
	rs, _ := rules.Get(ruleset)
	c := &model.Character{ID: id, Name: id, Ruleset: ruleset, Race: race, Class: class}
	if err := rs.Build(c); err != nil {
		t.Fatal(err)
	}
	return c
}

func arena(t *testing.T) (map[string]*model.Character, Lookup) {
	m := map[string]*model.Character{
		"a": hero(t, "a", "dnd5e", "human", "fighter"),
		"b": hero(t, "b", "dnd5e", "human", "wizard"),
		"c": hero(t, "c", "dnd5e", "human", "rogue"),
	}
	return m, func(id string) *model.Character { return m[id] }
}

func TestStartValidation(t *testing.T) {
	m, get := arena(t)
	m["x"] = hero(t, "x", "dnd5e", "human", "fighter")
	m["x"].Ruleset = "other" // чужая система правил
	r := &seq{v: []int{10}}
	for name, ids := range map[string][]string{"один": {"a"}, "дубликаты": {"a", "a"}, "неизвестный": {"a", "zzz"}} {
		if _, err := Start(r, ids, get); err == nil {
			t.Errorf("%s: ожидалась ошибка", name)
		}
	}
}

func TestTurnsSkipDownedAndWrap(t *testing.T) {
	m, get := arena(t)
	r := &seq{v: []int{5, 10, 15}}
	e, err := Start(r, []string{"a", "b", "c"}, get) // порядок: c, b, a
	if err != nil || e.Order[0] != "c" {
		t.Fatalf("%v %v", e, err)
	}
	m["b"].HP, m["b"].Stable = 0, true
	e.Next(r, get)
	if e.Current() != "a" {
		t.Errorf("упавший должен пропускаться, ходит %s", e.Current())
	}
	e.Next(r, get)
	if e.Current() != "c" || e.Round != 2 {
		t.Errorf("круг: ходит %s, раунд %d", e.Current(), e.Round)
	}
	for _, c := range m {
		c.HP, c.Stable = 0, true
	}
	e.Next(r, get) // не должно зависнуть, когда упали все
}

func TestDyingHeroRollsDeathSaveOnTurn(t *testing.T) {
	m, get := arena(t)
	r := &seq{v: []int{5, 10, 15}}
	e, _ := Start(r, []string{"a", "b", "c"}, get)
	m["b"].HP = 0
	r.v = []int{14} // 15 на кубике – успех
	e.Next(r, get)  // c → b: b должен бросить спасбросок и пропустить ход
	if m["b"].DeathOk != 1 || e.Current() != "a" {
		t.Errorf("спасбросок от смерти: ok=%d, ходит %s, лог %v", m["b"].DeathOk, e.Current(), e.Log)
	}
}

func TestAttackRules(t *testing.T) {
	m, get := arena(t)
	r := &seq{v: []int{15, 5, 0}}
	e, _ := Start(r, []string{"a", "b", "c"}, get)
	cur := e.Current()
	target := "b"
	if cur == "b" {
		target = "a"
	}
	w := model.Weapon{Name: "Меч", Dice: "1d8", Ability: "str"}
	if _, err := e.Attack(&seq{v: []int{19, 3}}, get, cur, w, dice.Normal); err == nil {
		t.Error("нельзя атаковать себя")
	}
	hp := m[target].HP
	res, err := e.Attack(&seq{v: []int{19, 3}}, get, target, w, dice.Normal) // 20 → попадание
	if err != nil || !res.Hit || m[target].HP != max(0, hp-res.Damage) {
		t.Fatalf("%+v %v hp=%d", res, err, m[target].HP)
	}
	if _, err := e.Attack(&seq{v: []int{19}}, get, target, w, dice.Normal); err == nil {
		t.Error("вторая атака за ход должна отклоняться")
	}
	e.Next(r, get)
	m[target].Dead = true
	if _, err := e.Attack(&seq{v: []int{19}}, get, target, w, dice.Normal); err == nil {
		t.Error("нельзя атаковать погибшего")
	}
}

func TestAttackOnDownedHeroCostsDeathSave(t *testing.T) {
	m, get := arena(t)
	r := &seq{v: []int{5, 10, 15}}
	e, _ := Start(r, []string{"a", "b", "c"}, get) // ходит c
	m["b"].HP = 0
	w := model.Weapon{Name: "Кинжал", Dice: "1d4", Ability: "dex"}
	if res, err := e.Attack(&seq{v: []int{9, 1}}, get, "b", w, dice.Normal); err != nil || !res.Hit || !res.Crit {
		t.Fatalf("удар по лежащему – критический: %+v %v", res, err)
	}
	if m["b"].DeathFail != 2 {
		t.Errorf("крит по лежащему даёт 2 провала, у героя %d", m["b"].DeathFail)
	}
}

func TestMonsterDiesAtZero(t *testing.T) {
	m, get := arena(t)
	g, _ := bestiary.Make("goblin", 1, func() string { return "g" })
	m["g"] = g
	r := &seq{v: []int{19, 1}}
	e, err := Start(r, []string{"a", "g"}, get)
	if err != nil {
		t.Fatal(err)
	}
	cur := e.Current()
	target := map[string]string{"a": "g", "g": "a"}[cur]
	if target == "a" {
		g.Stat.AC = 100
		m["a"].HP = 1
		m["g"].Weapons[0].Dice = "1d2+10"
	} else {
		g.HP = 1
	}
	if _, err := e.Attack(&seq{v: []int{19, 1}}, get, target, model.Weapon{Name: "Меч", Dice: "1d8", Ability: "str"}, dice.Normal); err != nil {
		t.Fatal(err)
	}
	if target == "g" && !g.Dead {
		t.Error("существо должно погибать сразу на 0 HP")
	}
}

func TestDragonbornBreathOncePerRest(t *testing.T) {
	m, get := arena(t)
	m["d"] = hero(t, "d", "dnd5e", "dragonborn", "fighter")
	m["d"].Level = 1
	r := &seq{v: []int{19, 1, 1, 1}}
	e, err := Start(r, []string{"d", "a", "b"}, get)
	if err != nil {
		t.Fatal(err)
	}
	for e.Current() != "d" {
		e.Next(r, get)
	}
	if err := e.Special(&seq{v: []int{3}}, get, "breath"); err != nil {
		t.Fatal(err)
	}
	if err := e.Special(&seq{v: []int{3}}, get, "breath"); err == nil {
		t.Error("второй выдох в тот же ход – ошибка")
	}
	e.Acted = false
	if err := e.Special(&seq{v: []int{3}}, get, "breath"); err == nil {
		t.Error("выдох расходуется до отдыха")
	}
}

func TestDeletedCombatantDoesNotCrash(t *testing.T) {
	m, get := arena(t)
	r := &seq{v: []int{5, 10, 15}}
	e, _ := Start(r, []string{"a", "b", "c"}, get)
	delete(m, e.Current())
	e.Next(r, get)
	e.seek(r, get)
}

func TestExtraAttacksAndMultiattack(t *testing.T) {
	m, get := arena(t)
	m["a"].Level = 5 // воин: две атаки
	m["t"] = func() *model.Character {
		c, err := bestiary.Make("troll", 1, func() string { return "t" })
		if err != nil {
			t.Fatal(err)
		}
		return c
	}()
	// инициатива: a = 20, t = 15, остальные ниже
	r := &seq{v: []int{19, 14, 0, 0}}
	e, err := Start(r, []string{"a", "t", "b", "c"}, get)
	if err != nil {
		t.Fatal(err)
	}
	if e.Current() != "a" || e.Left != 2 {
		t.Fatalf("ход %s, осталось атак %d", e.Current(), e.Left)
	}
	w := model.Weapon{Name: "Меч", Dice: "1d8", Ability: "str", Type: "slashing"}
	if _, err := e.Attack(&seq{v: []int{0}}, get, "b", w, dice.Normal); err != nil || e.Acted || e.Left != 1 {
		t.Fatalf("после первой атаки ход продолжается: acted=%v left=%d err=%v", e.Acted, e.Left, err)
	}
	if _, err := e.Attack(&seq{v: []int{0}}, get, "b", w, dice.Normal); err != nil || !e.Acted {
		t.Fatalf("вторая атака завершает действие: %v", err)
	}
	if _, err := e.Attack(&seq{v: []int{0}}, get, "b", w, dice.Normal); err == nil {
		t.Error("третья атака невозможна")
	}
	e.Next(&seq{v: []int{1}}, get)
	if e.Current() != "t" || e.Left != 3 {
		t.Errorf("тролль: ход %s, атак %d (укус и два когтя)", e.Current(), e.Left)
	}
}

func TestSpellInEncounter(t *testing.T) {
	m, get := arena(t)
	w := m["b"]
	w.Level, w.Abilities["int"] = 5, 17
	w.Spells = []model.Spell{{Name: "Огненный шар", Level: 3, Ref: "fireball"}}
	e, err := Start(&seq{v: []int{0, 19, 0}}, []string{"a", "b", "c"}, get) // первым ходит b
	if err != nil || e.Current() != "b" {
		t.Fatalf("%v %s", err, e.Current())
	}
	if err := e.Cast(&seq{v: []int{3}}, get, 0, 3, nil, dice.Normal); err == nil {
		t.Error("нужна цель")
	}
	if err := e.Cast(&seq{v: []int{3}}, get, 5, 3, []string{"a"}, dice.Normal); err == nil {
		t.Error("нет такого заклинания")
	}
	if err := e.Cast(&seq{v: []int{3}}, get, 0, 3, []string{"zzz"}, dice.Normal); err == nil {
		t.Error("цели нет в бою")
	}
	if w.SlotsUsed != [9]int{} || e.Acted {
		t.Fatal("ошибки не тратят ячейку и действие")
	}
	if err := e.Cast(&seq{v: []int{3}}, get, 0, 3, []string{"a", "c"}, dice.Normal); err != nil {
		t.Fatal(err)
	}
	if !e.Acted || w.SlotsUsed[2] != 1 || m["a"].HP != 0 || m["c"].HP != 0 {
		t.Errorf("шар: acted=%v слоты=%v HP a=%d c=%d", e.Acted, w.SlotsUsed, m["a"].HP, m["c"].HP)
	}
	if err := e.Cast(&seq{v: []int{3}}, get, 0, 3, []string{"a"}, dice.Normal); err == nil {
		t.Error("второе действие за ход")
	}
}

func TestAddRemoveAndWinner(t *testing.T) {
	chars := map[string]*model.Character{}
	mk := func(id, kind string) {
		rs, _ := rules.Get("dnd5e")
		c := &model.Character{ID: id, Name: id, Kind: kind, Ruleset: "dnd5e", Race: "human", Class: "fighter", Level: 1}
		if kind == "monster" {
			c.Race, c.Class = "", ""
			c.Stat = &model.MonsterStat{AC: 10, MaxHP: 5, Attack: 2, Speed: 30}
			c.Abilities = map[string]int{"str": 10, "dex": 10, "con": 10, "int": 10, "wis": 10, "cha": 10}
			c.Weapons = []model.Weapon{{Name: "Удар", Dice: "1d4", Ability: "str", Type: "bludgeoning"}}
			c.HP = 5
		} else {
			_ = rs.Build(c)
			c.HP = rs.Derive(c).MaxHP
		}
		chars[id] = c
	}
	for _, id := range []string{"h1", "m1", "m2"} {
		kind := "monster"
		if id == "h1" {
			kind = ""
		}
		mk(id, kind)
	}
	get := func(id string) *model.Character { return chars[id] }
	e, err := Start(&seq{v: []int{10}}, []string{"h1", "m1"}, get)
	if err != nil {
		t.Fatal(err)
	}
	cur := e.Current()
	if err := e.Add(&seq{v: []int{10}}, []string{"m2", "m2", "m1"}, get); err != nil || len(e.Order) != 3 {
		t.Fatalf("дубли не добавляются: %v %v", err, e.Order)
	}
	if e.Current() != cur {
		t.Error("ход не должен переходить к новичку")
	}
	if f := e.Foes(get, "h1"); len(f) != 2 {
		t.Errorf("у героя два врага: %d", len(f))
	}
	if f := e.Foes(get, "m1"); len(f) != 1 || f[0].ID != "h1" {
		t.Errorf("монстры не атакуют друг друга: %+v", f)
	}
	if e.Winner(get) != "" {
		t.Error("бой идёт")
	}
	chars["m1"].HP, chars["m2"].HP = 0, 0
	chars["m1"].Dead, chars["m2"].Dead = true, true
	if e.Winner(get) != "heroes" {
		t.Error("герои победили")
	}
	e.Remove(&seq{v: []int{10}}, "m2", get)
	if len(e.Order) != 2 {
		t.Error("удаление участника")
	}
	for _, id := range e.Order {
		if id == "m2" {
			t.Error("m2 остался")
		}
	}
}

func TestConcentrationBreaksWhenHitAndEndsEffects(t *testing.T) {
	m, get := arena(t)
	cl, tgt, foe := m["b"], m["a"], m["c"]
	cl.Level, cl.Abilities["wis"] = 5, 16
	cl.Class = "cleric"
	cl.Spells = []model.Spell{{Name: "Благословение", Level: 1, Ref: "bless"}}
	e, err := Start(&seq{v: []int{0, 19, 0}}, []string{"a", "b", "c"}, get)
	if err != nil || e.Current() != "b" {
		t.Fatalf("%v %s", err, e.Current())
	}
	if err := e.Cast(&seq{v: []int{1}}, get, 0, 1, []string{"a"}, dice.Normal); err != nil {
		t.Fatal(err)
	}
	if cl.Conc != "Благословение" || len(tgt.Effects) != 1 {
		t.Fatalf("благословение: conc=%q эффекты=%d", cl.Conc, len(tgt.Effects))
	}
	// жрец получает удар
	cl.HP = 100
	rs, _ := rules.Get("dnd5e")
	rules.Strike(rs, cl, 30, false)
	e.settle(&seq{v: []int{0}}, get) // бросок 1 против СЛ 15 – концентрация сорвана
	if cl.Conc != "" || len(tgt.Effects) != 0 {
		t.Errorf("эффект должен исчезнуть вместе с концентрацией: conc=%q эффекты=%d", cl.Conc, len(tgt.Effects))
	}
	_ = foe
}

func TestNewConcentrationReplacesOld(t *testing.T) {
	m, get := arena(t)
	cl := m["b"]
	cl.Level, cl.Abilities["wis"], cl.Class = 5, 16, "cleric"
	cl.Spells = []model.Spell{{Name: "Благословение", Level: 1, Ref: "bless"}, {Name: "Щит веры", Level: 1, Ref: "shieldoffaith"}}
	e, _ := Start(&seq{v: []int{0, 19, 0}}, []string{"a", "b", "c"}, get)
	e.Cast(&seq{v: []int{1}}, get, 0, 1, []string{"a"}, dice.Normal)
	e.Acted = false
	if err := e.Cast(&seq{v: []int{1}}, get, 1, 1, []string{"c"}, dice.Normal); err != nil {
		t.Fatal(err)
	}
	if len(m["a"].Effects) != 0 || len(m["c"].Effects) != 1 || cl.Conc != "Щит веры" {
		t.Errorf("старая концентрация снимается: a=%d c=%d conc=%q", len(m["a"].Effects), len(m["c"].Effects), cl.Conc)
	}
}

func TestParalyzedFoeSkipsTurnAndRollsRepeatSave(t *testing.T) {
	m, get := arena(t)
	e, _ := Start(&seq{v: []int{19, 10, 0}}, []string{"a", "b", "c"}, get)
	victim := m[e.Order[1]]
	rules.AddEffect(victim, model.Effect{Name: "Паралич", Rounds: 10, Cond: "Парализован", Save: "con", DC: 30, Src: "x"})
	e.Next(&seq{v: []int{5}}, get) // ход переходит на парализованного – пропускается, спасбросок провален
	if e.Current() == victim.ID {
		t.Error("парализованный не может ходить")
	}
	if !victim.Has("Парализован") {
		t.Error("СЛ 30 слишком велика, паралич держится")
	}
}

func TestReactionResetsOnOwnTurn(t *testing.T) {
	m, get := arena(t)
	_ = m
	e, _ := Start(&seq{v: []int{19, 10, 0}}, []string{"a", "b", "c"}, get)
	other := e.Order[1]
	if err := e.Reaction(other); err != nil || !e.React[other] {
		t.Fatal("реакция отмечена")
	}
	e.Next(&seq{v: []int{5}}, get)
	if e.React[other] {
		t.Error("в свой ход реакция возвращается")
	}
	if err := e.Reaction("zzz"); err == nil {
		t.Error("чужой участник")
	}
}
