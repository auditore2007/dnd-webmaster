package rules

import (
	"slices"
	"strings"
	"testing"

	"heroesbook/internal/model"
)

// flat — бросок с постоянным результатом: кубик n граней даёт min(5, n) (d20 → 5, d4 → 4).
type flat struct{}

func (flat) Intn(n int) int { return min(4, n-1) }

func hero(t *testing.T, class, sub string, lvl int) *model.Character {
	t.Helper()
	c := build(t, DnD5e{}, "human", "", class)
	c.Subclass, c.Level = sub, lvl
	c.HP = DnD5e{}.Derive(c).MaxHP
	return c
}

func poolOf(d DnD5e, c *model.Character, key string) model.PoolView {
	for _, p := range d.Derive(c).Pools {
		if p.Key == key {
			return p
		}
	}
	return model.PoolView{}
}

func foe(ac, hp int, kind string) *model.Character {
	m := dummy(ac, hp)
	m.Stat.Kind = kind
	return m
}

func TestAbilityTableIntegrity(t *testing.T) {
	cat := DnD5e{}.Catalog()
	for _, a := range abTable {
		ci := slices.IndexFunc(cat.Classes, func(c Class) bool { return c.ID == a.Class })
		if ci < 0 {
			t.Errorf("%s: нет такого класса", a)
			continue
		}
		if a.Sub != "" && !slices.ContainsFunc(cat.Classes[ci].Subs, func(s Subclass) bool { return s.ID == a.Sub }) {
			t.Errorf("%s: нет такого подкласса", a)
		}
		if a.Pool != "" {
			if _, ok := pools[a.Pool]; !ok {
				t.Errorf("%s: нет пула %s", a, a.Pool)
			}
		}
		if a.Name == "" || a.Desc == "" {
			t.Errorf("%s: нет названия или описания", a)
		}
		if !a.Passive && a.Pool == "" && !a.Slot && a.Rider == nil && a.Fx == nil && a.Heal == nil && a.Temp == nil && a.Regain == "" && a.ToPool == "" {
			t.Errorf("%s: способность ничего не делает", a)
		}
	}
	for _, c := range cat.Classes {
		if c.ID == "mystic" || c.ID == "gunslinger" {
			continue
		}
		if !slices.ContainsFunc(abTable, func(a ab) bool { return a.Class == c.ID && a.Sub == "" }) {
			t.Errorf("у класса %s нет ни одной активной способности", c.ID)
		}
	}
}

func TestBattleMasterDice(t *testing.T) {
	d := DnD5e{}
	h := hero(t, "fighter", "battlemaster", 3)
	if _, _, err := d.Ability(flat{}, "riposte", h, nil); err != nil {
		t.Fatal(err)
	}
	if pv := poolOf(d, h, "sup"); pv.Left != 3 || pv.Max != 4 {
		t.Fatalf("кости превосходства: %+v", pv)
	}
	m := foe(5, 100, "beast")
	w := model.Weapon{Name: "Меч", Dice: "1d8", Ability: "str", Type: "slashing"}
	r1 := d.Attack(flat{}, h, m, w, 0)
	if !r1.Hit || !strings.Contains(r1.DamageText, "Манёвр: контрудар") {
		t.Fatalf("надбавка не сработала: %+v", r1)
	}
	r2 := d.Attack(flat{}, h, m, w, 0)
	if strings.Contains(r2.DamageText, "Манёвр") {
		t.Errorf("надбавка должна сниматься после удара: %+v", r2)
	}
	if r1.Damage <= r2.Damage {
		t.Errorf("урон с костью должен быть больше: %d и %d", r1.Damage, r2.Damage)
	}
	for i := 0; i < 3; i++ {
		if _, _, err := d.Ability(flat{}, "riposte", h, nil); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := d.Ability(flat{}, "riposte", h, nil); err == nil {
		t.Error("кости должны закончиться")
	}
	d.Rest(h, "short")
	if poolOf(d, h, "sup").Left != 4 {
		t.Error("короткий отдых возвращает кости")
	}
}

func TestTripManeuver(t *testing.T) {
	d := DnD5e{}
	h := hero(t, "fighter", "battlemaster", 3)
	m := foe(5, 100, "beast")
	if _, _, err := d.Ability(flat{}, "trip", h, nil); err != nil {
		t.Fatal(err)
	}
	r := d.Attack(flat{}, h, m, model.Weapon{Dice: "1d8", Ability: "str", Type: "slashing"}, 0)
	if !r.Hit || !m.Has("Сбит с ног") {
		t.Errorf("цель должна быть сбита с ног: %+v", r)
	}
}

func TestEldritchKnightSlots(t *testing.T) {
	d := DnD5e{}
	if got := d.Derive(hero(t, "fighter", "champion", 7)).Slots; len(got) != 0 {
		t.Errorf("чемпион не колдует: %+v", got)
	}
	ek := hero(t, "fighter", "eldritchknight", 3)
	got := d.Derive(ek).Slots
	if len(got) != 1 || got[0].Max != 2 {
		t.Errorf("мистический рыцарь 3 ур.: %+v", got)
	}
	ek.Level = 7
	got = d.Derive(ek).Slots
	if len(got) != 2 || got[0].Max != 4 || got[1].Max != 2 {
		t.Errorf("мистический рыцарь 7 ур.: %+v", got)
	}
	if ds := d.Derive(ek); ds.SpellDC != 8+ds.Prof+ds.Mods["int"] {
		t.Errorf("СЛ на Интеллекте: %d", ds.SpellDC)
	}
	at := hero(t, "rogue", "arcanetrickster", 3)
	if len(d.Derive(at).Slots) != 1 {
		t.Error("мистический ловкач колдует")
	}
	ek.Level = 2
	if len(d.Derive(ek).Slots) != 0 {
		t.Error("до 3 уровня ячеек нет")
	}
}

func TestPaladinSmite(t *testing.T) {
	d := DnD5e{}
	h := hero(t, "paladin", "", 2)
	m := foe(5, 100, "beast")
	w := model.Weapon{Name: "Меч", Dice: "1d8", Ability: "str", Type: "slashing"}
	if _, _, err := d.Ability(flat{}, "smite", h, nil); err != nil {
		t.Fatal(err)
	}
	if h.SlotsUsed[0] != 1 {
		t.Error("кара тратит ячейку")
	}
	r := d.Attack(flat{}, h, m, w, 0)
	if !strings.Contains(r.DamageText, "Божественная кара") {
		t.Fatalf("нет кары: %+v", r)
	}
	h.SlotsUsed = [9]int{2}
	if _, _, err := d.Ability(flat{}, "smite", h, nil); err == nil {
		t.Error("без ячеек кара невозможна")
	}
}

func TestMonkKiAndFlurry(t *testing.T) {
	d := DnD5e{}
	h := hero(t, "monk", "", 5)
	ds := d.Derive(h)
	if !slices.ContainsFunc(ds.Weapons, func(w model.Weapon) bool { return w.Name == "Удар монаха" }) {
		t.Error("у монаха есть удар боевыми искусствами")
	}
	if ds.Pools[0].Max != 5 {
		t.Errorf("ки = уровень: %+v", ds.Pools)
	}
	_, extra, err := d.Ability(flat{}, "flurry", h, nil)
	if err != nil || extra != 2 {
		t.Fatalf("шквал: %d %v", extra, err)
	}
	if d.Derive(h).Pools[0].Left != 4 {
		t.Error("шквал тратит ки")
	}
	h.Counters["p:ki"] = 5
	if _, _, err := d.Ability(flat{}, "flurry", h, nil); err == nil {
		t.Error("без ки нельзя")
	}
}

func TestLayOnHands(t *testing.T) {
	d := DnD5e{}
	p := hero(t, "paladin", "", 2)
	ally := hero(t, "wizard", "", 1)
	ally.HP = ally.HP - 4
	maxHP := d.Derive(ally).MaxHP
	if _, _, err := d.Ability(flat{}, "layhands", p, ally); err != nil {
		t.Fatal(err)
	}
	if ally.HP != maxHP {
		t.Errorf("исцелён до максимума: %d из %d", ally.HP, maxHP)
	}
	if left := d.Derive(p).Pools[0].Left; left != 10-4 {
		t.Errorf("запас уменьшился на 4: %d", left)
	}
	if _, _, err := d.Ability(flat{}, "layhands", p, ally); err == nil {
		t.Error("лечить некого")
	}
}

func TestAssassinateAndReckless(t *testing.T) {
	d := DnD5e{}
	r := hero(t, "rogue", "assassin", 3)
	m := foe(30, 100, "beast")
	if _, _, err := d.Ability(flat{}, "assassinate", r, nil); err != nil {
		t.Fatal(err)
	}
	if !r.Has("Преимущество") {
		t.Fatal("убийство даёт преимущество")
	}
	w := model.Weapon{Dice: "1d6", Ability: "dex", Finesse: true, Type: "piercing"}
	m.Stat.AC = 5
	res := d.Attack(flat{}, r, m, w, 0)
	if !res.Hit || !res.Crit {
		t.Errorf("убийство гарантирует крит: %+v", res)
	}
	b := hero(t, "barbarian", "", 2)
	if _, _, err := d.Ability(flat{}, "reckless", b, nil); err != nil {
		t.Fatal(err)
	}
	if condModeR(b, m, 0, false) <= 0 || condModeR(m, b, 0, false) <= 0 {
		t.Error("безрассудство: преимущество у обеих сторон")
	}
}

func TestTotemRage(t *testing.T) {
	d := DnD5e{}
	b := hero(t, "barbarian", "totem", 3)
	b.Active = map[string]bool{"rage": true}
	if d.Resist(b, "fire") != .5 || d.Resist(b, "psychic") != 1 {
		t.Error("тотем медведя: сопротивление всему, кроме психического")
	}
	z := hero(t, "barbarian", "berserker", 3)
	z.Active = map[string]bool{"rage": true}
	if d.Resist(z, "fire") != 1 {
		t.Error("берсерк не сопротивляется огню")
	}
	if d.Derive(z).Attacks != 2 {
		t.Error("бешенство: лишняя атака в ярости")
	}
}

func TestTurnUndead(t *testing.T) {
	d := DnD5e{}
	c := hero(t, "cleric", "", 2)
	sk, gob := foe(5, 20, "undead"), foe(5, 20, "humanoid")
	sk.ID, gob.ID = "s", "g"
	sk.Abilities["wis"] = 1
	if _, err := d.Special(flat{}, "turn", c, []*model.Character{sk, gob}); err != nil {
		t.Fatal(err)
	}
	if !sk.Has("Напуган") || gob.Has("Напуган") {
		t.Errorf("изгнание действует только на нежить: %v / %v", sk.Conditions, gob.Conditions)
	}
	if d.Derive(c).Pools[0].Left != 0 {
		t.Error("канал потрачен")
	}
	if _, err := d.Special(flat{}, "turn", c, []*model.Character{sk}); err == nil {
		t.Error("канал исчерпан")
	}
	if _, err := d.Special(flat{}, "turn", hero(t, "cleric", "", 2), []*model.Character{gob}); err == nil {
		t.Error("нет нежити — нет цели")
	}
}

func TestInspirationConsumed(t *testing.T) {
	d := DnD5e{}
	b := hero(t, "bard", "", 1)
	ally := hero(t, "fighter", "", 1)
	if _, _, err := d.Ability(flat{}, "inspire", b, ally); err != nil {
		t.Fatal(err)
	}
	if len(ally.Effects) != 1 {
		t.Fatal("эффект вдохновения")
	}
	m := foe(5, 100, "beast")
	r := d.Attack(flat{}, ally, m, model.Weapon{Dice: "1d6", Ability: "str", Type: "slashing"}, 0)
	if !r.Hit || len(ally.Effects) != 0 {
		t.Errorf("кубик тратится на первом броске: %+v, эффектов %d", r, len(ally.Effects))
	}
}

func TestSorcererSlotsAndPoints(t *testing.T) {
	d := DnD5e{}
	s := hero(t, "sorcerer", "", 3)
	s.SlotsUsed = [9]int{2, 0}
	if _, _, err := d.Ability(flat{}, "slot1", s, nil); err != nil {
		t.Fatal(err)
	}
	if s.SlotsUsed[0] != 1 {
		t.Error("ячейка возвращена")
	}
	if left := d.Derive(s).Pools[0].Left; left != 1 {
		t.Errorf("3 очка − 2: %d", left)
	}
	if _, _, err := d.Ability(flat{}, "slot2", s, nil); err == nil {
		t.Error("нет потраченной ячейки 2 ур. / не хватает очков")
	}
}

func TestWizardRecoveryAndTemp(t *testing.T) {
	d := DnD5e{}
	w := hero(t, "wizard", "abjuration", 4)
	w.SlotsUsed = [9]int{4, 1}
	if _, _, err := d.Ability(flat{}, "arcane", w, nil); err != nil {
		t.Fatal(err)
	}
	if w.SlotsUsed[1] != 0 || w.SlotsUsed[0] != 4 {
		t.Errorf("бюджет 2: вернулась ячейка 2 ур.: %v", w.SlotsUsed)
	}
	if _, _, err := d.Ability(flat{}, "warding", w, nil); err != nil || w.TempHP < 8 {
		t.Errorf("барьер: %d %v", w.TempHP, err)
	}
}

func TestRiderPersistAndToggle(t *testing.T) {
	d := DnD5e{}
	r := hero(t, "ranger", "", 2)
	m := foe(5, 100, "beast")
	w := model.Weapon{Dice: "1d8", Ability: "dex", Ranged: true, Type: "piercing"}
	if _, _, err := d.Ability(flat{}, "mark", r, nil); err != nil {
		t.Fatal(err)
	}
	a1, a2 := d.Attack(flat{}, r, m, w, 0), d.Attack(flat{}, r, m, w, 0)
	if !strings.Contains(a1.DamageText, "Метка") || !strings.Contains(a2.DamageText, "Метка") {
		t.Error("метка держится на всех ударах")
	}
	if msg, _, _ := d.Ability(flat{}, "mark", r, nil); !strings.Contains(msg, "снято") {
		t.Error("повторное применение снимает метку")
	}
	d.Rest(r, "battle")
}

func TestPassiveAndDeadErrors(t *testing.T) {
	d := DnD5e{}
	b := hero(t, "barbarian", "berserker", 3)
	if _, _, err := d.Ability(flat{}, "frenzy", b, nil); err == nil {
		t.Error("пассивную способность нельзя применить")
	}
	b.HP = 0
	if _, _, err := d.Ability(flat{}, "reckless", b, nil); err == nil {
		t.Error("герой без сознания")
	}
	if _, _, err := d.Ability(flat{}, "nothing", hero(t, "wizard", "", 1), nil); err == nil {
		t.Error("неизвестная способность")
	}
}
