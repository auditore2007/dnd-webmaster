package rules

import (
	"slices"
	"testing"

	"heroesbook/internal/dice"
	"heroesbook/internal/model"
)

func TestModRoundsDown(t *testing.T) {
	for score, want := range map[int]int{1: -5, 7: -2, 8: -1, 9: -1, 10: 0, 11: 0, 12: 1, 20: 5, 30: 10} {
		if got := Mod(score); got != want {
			t.Errorf("Mod(%d) = %d, want %d", score, got, want)
		}
	}
}

func build(t *testing.T, rs Ruleset, race, sub, class string) *model.Character {
	t.Helper()
	c := &model.Character{Ruleset: rs.ID(), Race: race, Subrace: sub, Class: class}
	if err := rs.Build(c); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestDnDBuildAndHP(t *testing.T) {
	d := DnD5e{}
	c := build(t, d, "halforc", "", "barbarian") // СЛ 11 (мод 0), d12
	if c.Abilities["str"] != 12 || c.Abilities["con"] != 11 || c.HP != 12 {
		t.Errorf("полуорк-варвар: %+v HP=%d", c.Abilities, c.HP)
	}
	h := build(t, d, "dwarf", "hill", "cleric") // СЛ 12 (+1), d8, +1 за холмового
	if got := d.Derive(h).MaxHP; got != 8+1+1 {
		t.Errorf("холмовой дварф-жрец: HP %d, ожидалось 10", got)
	}
	h.Level = 3
	if got := d.Derive(h).MaxHP; got != 10+2*(5+1+1) {
		t.Errorf("3 уровень: HP %d", got)
	}
}

func TestProficiencyByLevel(t *testing.T) {
	c := build(t, DnD5e{}, "human", "", "fighter")
	for lvl, want := range map[int]int{1: 2, 4: 2, 5: 3, 9: 4, 13: 5, 17: 6, 20: 6} {
		c.Level = lvl
		if got := (DnD5e{}).Derive(c).Prof; got != want {
			t.Errorf("уровень %d: мастерство %d, want %d", lvl, got, want)
		}
	}
}

func TestCatalogNotMutated(t *testing.T) {
	build(t, DnD5e{}, "dwarf", "mountain", "fighter")
	if dndRaces[1].Bonus["str"] != 0 || len(dndRaces[1].Bonus) != 1 {
		t.Errorf("каталог изменился: %v", dndRaces[1].Bonus)
	}
}

func TestSubraceRequired(t *testing.T) {
	c := &model.Character{Race: "elf", Class: "wizard"}
	if err := (DnD5e{}).Build(c); err == nil {
		t.Error("у эльфа без подрасы должна быть ошибка")
	}
	if err := (DnD5e{}).Build(&model.Character{Race: "elf", Subrace: "wood", Class: "nope"}); err == nil {
		t.Error("неизвестный класс должен давать ошибку")
	}
}

func TestSavesAndInitiative(t *testing.T) {
	c := build(t, DnD5e{}, "elf", "wood", "rogue") // ЛВК 12, ИНТ 10
	d := (DnD5e{}).Derive(c)
	if d.Saves["dex"] != 1+2 || d.Saves["int"] != 0+2 || d.Saves["str"] != 0 || d.Speed != 35 {
		t.Errorf("%+v", d)
	}
}

type seq struct {
	v []int
	i int
}

func (s *seq) Intn(n int) int { x := s.v[s.i%len(s.v)]; s.i++; return x % n }

func TestAttackRules(t *testing.T) {
	d := DnD5e{}
	atk := build(t, d, "human", "", "fighter") // СИЛ 11
	def := build(t, d, "tiefling", "", "wizard")
	w := model.Weapon{Name: "Меч", Dice: "1d8", Ability: "str", Type: "fire"}
	// 20 – автоматическое попадание и крит; кубики удваиваются; тифлинг получает половину огненного урона
	res := d.Attack(&seq{v: []int{19, 5, 5}}, atk, def, w, dice.Normal)
	if !res.Hit || !res.Crit || res.Damage != (6+6+0)/2 {
		t.Errorf("крит: %+v", res)
	}
	// 1 – всегда промах, даже против КД 0
	def.ACBonus = -10
	if res := d.Attack(&seq{v: []int{0}}, atk, def, w, dice.Normal); res.Hit {
		t.Errorf("натуральная 1 должна промахиваться: %+v", res)
	}
}

func TestDeathSavesFlow(t *testing.T) {
	d := DnD5e{}
	c := build(t, d, "human", "", "fighter")
	if lines, _ := Strike(d, c, 999, false); c.HP != 0 || c.Dead || len(lines) == 0 {
		t.Fatalf("падение на 0 не убивает: %+v", c)
	}
	Strike(d, c, 1, false)
	if c.DeathFail != 1 {
		t.Errorf("урон по лежащему = провал, а не %d", c.DeathFail)
	}
	Strike(d, c, 1, true)
	if !c.Dead {
		t.Error("три провала (крит = 2) должны убивать")
	}
	if c.Heal(5, 10); c.HP != 0 {
		t.Error("погибшего лечить нельзя")
	}
}

func TestDeathSaveRolls(t *testing.T) {
	d := DnD5e{}
	c := build(t, d, "human", "", "fighter")
	c.HP = 0
	d.DeathSave(&seq{v: []int{19}}, c) // 20
	if c.HP != 1 {
		t.Errorf("20 возвращает 1 HP, HP=%d", c.HP)
	}
	c.HP = 0
	d.DeathSave(&seq{v: []int{0}}, c) // 1
	if c.DeathFail != 2 {
		t.Errorf("1 = два провала, а не %d", c.DeathFail)
	}
	d.DeathSave(&seq{v: []int{9}}, c) // 10
	d.DeathSave(&seq{v: []int{9}}, c)
	d.DeathSave(&seq{v: []int{9}}, c)
	if !c.Stable || c.Dead {
		t.Errorf("три успеха стабилизируют: %+v", c)
	}
	if _, err := d.DeathSave(&seq{v: []int{9}}, c); err == nil {
		t.Error("стабильному спасбросок не нужен")
	}
	c.Stable, c.HP = false, 0
	c.Heal(3, 10)
	if c.DeathFail != 0 || c.DeathOk != 0 {
		t.Error("лечение сбрасывает счётчики")
	}
}

func TestRelentlessEnduranceOnce(t *testing.T) {
	d := DnD5e{}
	c := build(t, d, "halforc", "", "barbarian")
	Strike(d, c, 999, false)
	if c.HP != 1 {
		t.Fatalf("первый смертельный удар: HP=%d", c.HP)
	}
	Strike(d, c, 999, false)
	if c.HP != 0 {
		t.Error("второй раз стойкость не срабатывает")
	}
}

func TestSlotsXPAndCast(t *testing.T) {
	d := DnD5e{}
	w := build(t, d, "human", "", "wizard")
	if s := d.Derive(w).Slots; len(s) != 1 || s[0].Max != 2 {
		t.Errorf("волшебник 1: %+v", s)
	}
	w.Level = 5
	if s := d.Derive(w).Slots; len(s) != 3 || s[2].Max != 2 {
		t.Errorf("волшебник 5: %+v", s)
	}
	if s := slotsFor("paladin", 1); s != nil {
		t.Error("паладин 1 без ячеек")
	}
	if s := slotsFor("paladin", 2); len(s) != 1 || s[0].Max != 2 {
		t.Errorf("паладин 2: %+v", s)
	}
	if s := slotsFor("fighter", 10); s != nil {
		t.Error("воин без ячеек")
	}
	spell := model.Spell{Name: "Огненная стрела", Level: 1}
	for i := 0; i < 4; i++ {
		d.Cast(w, spell, 1)
	}
	if _, err := d.Cast(w, spell, 1); err == nil {
		t.Error("пятая ячейка 1 уровня невозможна (их 4)")
	}
	if _, err := d.Cast(w, model.Spell{Name: "Луч", Level: 3}, 2); err == nil {
		t.Error("нельзя кастовать 3-й уровень во 2-й ячейке")
	}
	if d.XPFor(2) != 300 || d.XPFor(20) != 355000 || d.XPFor(21) < 1<<20 {
		t.Error("таблица опыта")
	}
	d.Rest(w, "long")
	if w.SlotsUsed != [9]int{} {
		t.Error("долгий отдых восстанавливает ячейки")
	}
}

func TestInventoryEffects(t *testing.T) {
	d := DnD5e{}
	c := build(t, d, "human", "", "fighter") // СИЛ 11, ЛВК 11
	base := d.Derive(c)
	c.Inventory = []model.Item{
		{Name: "Пояс", Qty: 1, Weight: 1, Ability: "STR", AbilityBonus: 4, Equipped: true},
		{Name: "Щит", Qty: 1, Weight: 6, ACBonus: 2, Equipped: true},
		{Name: "Меч", Qty: 2, Weight: 3, Weapon: &model.Weapon{Dice: "1d8", Ability: "str"}, Equipped: true},
		{Name: "Кольцо", Qty: 1, Ability: "dex", AbilityBonus: 5},
	}
	got := d.Derive(c)
	if got.Mods["str"] != base.Mods["str"]+2 || got.AC != base.AC+2 || got.Mods["dex"] != base.Mods["dex"] {
		t.Errorf("надетое действует, ненадетое нет: %+v", got)
	}
	if got.Load != 1+6+6 || len(got.Weapons) != 1 || got.Weapons[0].Name != "Меч" {
		t.Errorf("вес и оружие: %+v", got)
	}
}

func TestConditionsAttackMode(t *testing.T) {
	a, tg := &model.Character{}, &model.Character{HP: 5}
	if condMode(a, tg, dice.Normal) != dice.Normal {
		t.Error("без состояний – обычный бросок")
	}
	a.Conditions = []string{"Отравлен"}
	if condMode(a, tg, dice.Normal) != dice.Disadvantage {
		t.Error("отравленный атакует с помехой")
	}
	tg.Conditions = []string{"Парализован"}
	if condMode(a, tg, dice.Normal) != dice.Normal {
		t.Error("преимущество и помеха гасят друг друга")
	}
	a.Conditions = nil
	if condMode(a, tg, dice.Normal) != dice.Advantage {
		t.Error("по парализованному – преимущество")
	}
}

func TestMonsterDerive(t *testing.T) {
	g := &model.Character{Ruleset: "dnd5e", Abilities: map[string]int{"str": 8, "dex": 14}, Stat: &model.MonsterStat{AC: 15, MaxHP: 7, Attack: 4, Speed: 30}}
	d := (DnD5e{}).Derive(g)
	if d.AC != 15 || d.MaxHP != 7 || d.AtkBonus != 4 || d.Mods["dex"] != 2 || d.Mods["str"] != -1 {
		t.Errorf("%+v", d)
	}
}

func TestSpellCatalogIntegrity(t *testing.T) {
	cat := DnD5e{}.Catalog()
	ids, names := map[string]bool{}, map[string]bool{}
	classes := map[string]bool{}
	for _, c := range cat.Classes {
		classes[c.ID] = true
	}
	auto := 0
	for i, s := range cat.Spells {
		if s.ID == "" || s.Name == "" || s.School == "" || s.Level < 0 || s.Level > 9 || s.Desc == "" {
			t.Errorf("заклинание %d неполное: %+v", i, s)
		}
		if ids[s.ID] || names[s.Name] {
			t.Errorf("дубликат: %s / %s", s.ID, s.Name)
		}
		ids[s.ID], names[s.Name] = true, true
		if i > 0 && cat.Spells[i-1].Level > s.Level {
			t.Errorf("%s: порядок по уровню нарушен", s.ID)
		}
		for _, c := range s.Classes {
			if !classes[c] {
				t.Errorf("%s: неизвестный класс %q", s.ID, c)
			}
		}
		if len(s.Classes) == 0 {
			t.Errorf("%s: нет классов", s.ID)
		}
		if s.Mode != "" {
			auto++
			if _, err := dice.Parse(s.Dmg); err != nil {
				t.Errorf("%s: %v", s.ID, err)
			}
			if s.Mode == "save" && !slices.Contains([]string{"str", "dex", "con", "int", "wis", "cha"}, s.Save) {
				t.Errorf("%s: спасбросок %q", s.ID, s.Save)
			}
			if s.Mode != "heal" && !slices.Contains(DamageTypeIDs(), s.Type) {
				t.Errorf("%s: тип урона %q", s.ID, s.Type)
			}
		}
	}
	if len(cat.Spells) < 330 || auto < 80 {
		t.Errorf("заклинаний %d, автоматизировано %d", len(cat.Spells), auto)
	}
}

func TestLoreForEveryRaceAndClass(t *testing.T) {
	cat := DnD5e{}.Catalog()
	if len(cat.Races) < 25 || len(cat.Classes) < 12 {
		t.Errorf("рас %d, классов %d", len(cat.Races), len(cat.Classes))
	}
	for _, r := range cat.Races {
		if len([]rune(r.Desc)) < 40 || len([]rune(r.Lore)) < 100 {
			t.Errorf("раса %s: нет описания или истории", r.ID)
		}
	}
	for _, c := range cat.Classes {
		if len([]rune(c.Desc)) < 40 || len([]rune(c.Lore)) < 100 {
			t.Errorf("класс %s: нет описания или истории", c.ID)
		}
	}
}

func TestEveryRaceAndClassBuilds(t *testing.T) {
	rs := DnD5e{}
	for _, r := range rs.Catalog().Races {
		for _, sub := range append([]Race{{}}, r.Subs...) {
			if len(r.Subs) > 0 && sub.ID == "" {
				continue
			}
			for _, cl := range rs.Catalog().Classes {
				c := &model.Character{Name: "т", Ruleset: "dnd5e", Race: r.ID, Subrace: sub.ID, Class: cl.ID, Level: 1}
				if err := rs.Build(c); err != nil {
					t.Fatalf("%s/%s/%s: %v", r.ID, sub.ID, cl.ID, err)
				}
				d := rs.Derive(c)
				if d.MaxHP < 1 || d.AC < 5 {
					t.Errorf("%s/%s: HP %d КД %d", r.ID, cl.ID, d.MaxHP, d.AC)
				}
			}
		}
	}
}
