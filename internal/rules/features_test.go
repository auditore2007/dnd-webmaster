package rules

import (
	"slices"
	"strings"
	"testing"

	"heroesbook/internal/dice"
	"heroesbook/internal/model"
)

// dummy – существо-мишень без особенностей.
func dummy(ac, hp int) *model.Character {
	return &model.Character{ID: "m", Name: "Манекен", Ruleset: "dnd5e", Level: 1, HP: hp, Abilities: map[string]int{"str": 10, "dex": 10, "con": 10, "int": 10, "wis": 10, "cha": 10},
		Stat: &model.MonsterStat{AC: ac, MaxHP: hp}}
}

func TestArmorClassRules(t *testing.T) {
	d := DnD5e{}
	h := build(t, d, "human", "", "fighter")
	h.Abilities["dex"] = 17 // +3 после расового бонуса
	if got := d.Derive(h).AC; got != 13 {
		t.Errorf("без доспехов: %d", got)
	}
	leather := model.Item{Name: "Кожа", Equipped: true, ArmorBase: 11, ArmorType: "light"}
	half := model.Item{Name: "Полулаты", Equipped: true, ArmorBase: 15, ArmorType: "medium"}
	chain := model.Item{Name: "Кольчуга", Equipped: true, ArmorBase: 16, ArmorType: "heavy"}
	shield := model.Item{Name: "Щит", Equipped: true, ACBonus: 2}
	for name, tc := range map[string]struct {
		inv  []model.Item
		want int
	}{"лёгкий": {[]model.Item{leather}, 14}, "средний, Лов ограничена +2": {[]model.Item{half}, 17},
		"тяжёлый, Лов не учитывается": {[]model.Item{chain}, 16}, "тяжёлый и щит": {[]model.Item{chain, shield}, 18},
		"щит без доспеха": {[]model.Item{shield}, 15}} {
		h.Inventory = tc.inv
		if got := d.Derive(h).AC; got != tc.want {
			t.Errorf("%s: КД %d, ожидалось %d", name, got, tc.want)
		}
	}
	h.Inventory = []model.Item{{Name: "Снятый доспех", ArmorBase: 18, ArmorType: "heavy"}}
	if got := d.Derive(h).AC; got != 13 {
		t.Errorf("снятый доспех не даёт КД: %d", got)
	}
	b := build(t, d, "human", "", "barbarian")
	b.Abilities["dex"], b.Abilities["con"] = 15, 15 // +2, +2
	if got := d.Derive(b).AC; got != 14 {
		t.Errorf("защита без доспехов варвара: %d", got)
	}
	mk := build(t, d, "human", "", "monk")
	mk.Abilities["dex"], mk.Abilities["wis"] = 15, 17
	if got := d.Derive(mk).AC; got != 15 {
		t.Errorf("защита монаха: %d", got)
	}
	s := build(t, d, "human", "", "sorcerer")
	s.Subclass, s.Abilities["dex"] = "draconic", 15
	if got := d.Derive(s).AC; got != 15 {
		t.Errorf("драконья стойкость: %d", got)
	}
}

func TestExtraAttacksAndRage(t *testing.T) {
	d := DnD5e{}
	b := build(t, d, "halforc", "", "barbarian")
	b.Level = 5
	if d.Derive(b).Attacks != 2 {
		t.Error("варвар 5 бьёт дважды")
	}
	f := build(t, d, "human", "", "fighter")
	for lvl, want := range map[int]int{1: 1, 5: 2, 11: 3, 20: 4} {
		f.Level = lvl
		if got := d.Derive(f).Attacks; got != want {
			t.Errorf("воин %d: %d атак, ожидалось %d", lvl, got, want)
		}
	}
	// ярость: +2 урона Силой и половина физического урона
	b.Level = 1
	foe := dummy(5, 40)
	w := model.Weapon{Name: "Топор", Dice: "1d12", Ability: "str", Type: "slashing"}
	r := &seq{v: []int{14, 5}} // попадание, 1d12 = 6
	plain := d.Attack(r, b, foe, w, dice.Normal)
	if _, err := d.Toggle(b, "rage"); err != nil {
		t.Fatal(err)
	}
	raged := d.Attack(&seq{v: []int{14, 5}}, b, foe, w, dice.Normal)
	if raged.Damage != plain.Damage+2 {
		t.Errorf("ярость даёт +2: было %d, стало %d", plain.Damage, raged.Damage)
	}
	if d.Resist(b, "slashing") != .5 || d.Resist(b, "fire") != 1 {
		t.Error("ярость: сопротивление только физическому урону")
	}
	d.Rest(b, "battle")
	if b.Active["rage"] {
		t.Error("ярость кончается с боем")
	}
	// запас ярости: 2 на первом уровне
	if _, err := d.Toggle(b, "rage"); err != nil {
		t.Fatal(err) // вторая ярость
	}
	d.Toggle(b, "rage") // выключить
	if _, err := d.Toggle(b, "rage"); err == nil {
		t.Error("третья ярость на 1 уровне невозможна")
	}
	d.Rest(b, "long")
	if _, err := d.Toggle(b, "rage"); err != nil {
		t.Errorf("долгий отдых возвращает ярость: %v", err)
	}
	b.Inventory = []model.Item{{Name: "Латы", Equipped: true, ArmorBase: 18, ArmorType: "heavy"}}
	d.Toggle(b, "rage") // выключить
	if _, err := d.Toggle(b, "rage"); err == nil {
		t.Error("в тяжёлых доспехах ярость невозможна")
	}
	if _, err := d.Toggle(f, "rage"); err == nil {
		t.Error("воин не умеет впадать в ярость")
	}
}

func TestSneakAttackAndFinesse(t *testing.T) {
	d := DnD5e{}
	p := build(t, d, "human", "", "rogue")
	p.Level, p.Abilities["dex"] = 5, 17 // +3, 3d6 скрытой атаки
	foe := dummy(10, 100)
	dagger := model.Weapon{Name: "Кинжал", Dice: "1d4", Ability: "str", Type: "piercing", Finesse: true}
	// преимущество: d20 = 15 и 10; кинжал 1d4 = 3; скрытая атака 3d6 = 4+4+4
	res := d.Attack(&seq{v: []int{14, 9, 2, 3, 3, 3}}, p, foe, dagger, dice.Advantage)
	if !res.Hit || res.Damage != 3+3+12 {
		t.Errorf("скрытая атака: %+v (ожидалось 18)", res)
	}
	if !p.Active["sneaked"] {
		t.Error("скрытая атака помечается использованной")
	}
	again := d.Attack(&seq{v: []int{14, 9, 2}}, p, foe, dagger, dice.Advantage)
	if again.Damage != 3+3 {
		t.Errorf("второй раз за ход без скрытой атаки: %+v", again)
	}
	d.TurnStart(&seq{v: []int{1}}, p)
	if p.Active["sneaked"] {
		t.Error("в начале хода скрытая атака снова доступна")
	}
	plain := d.Attack(&seq{v: []int{14, 2}}, p, foe, dagger, dice.Normal)
	if plain.Damage != 3+3 {
		t.Errorf("без преимущества нет скрытой атаки: %+v", plain)
	}
	club := model.Weapon{Name: "Дубина", Dice: "1d4", Ability: "str", Type: "bludgeoning"}
	if res := d.Attack(&seq{v: []int{14, 9, 2, 3, 3, 3}}, p, foe, club, dice.Advantage); res.Damage != 3 {
		// у дубины нет Ловкости и «фехтования»: модификатор Силы 0
		t.Errorf("дубина не получает скрытой атаки: %+v", res)
	}
}

func TestChampionCrit(t *testing.T) {
	d := DnD5e{}
	f := build(t, d, "human", "", "fighter")
	f.Level, f.Subclass = 3, "champion"
	foe := dummy(30, 100) // 19 + мастерство не пробивают КД 30, но крит попадает всегда
	w := model.Weapon{Name: "Меч", Dice: "1d8", Ability: "str", Type: "slashing"}
	res := d.Attack(&seq{v: []int{18, 0, 0}}, f, foe, w, dice.Normal) // 19
	if !res.Hit || !res.Crit {
		t.Errorf("чемпион критует на 19: %+v", res)
	}
	f.Subclass = ""
	if res := d.Attack(&seq{v: []int{18, 0, 0}}, f, foe, w, dice.Normal); res.Hit {
		t.Errorf("без подкласса 19 не критует: %+v", res)
	}
	f.Subclass, f.Level = "champion", 15
	if res := d.Attack(&seq{v: []int{17, 0, 0}}, f, foe, w, dice.Normal); !res.Crit {
		t.Errorf("превосходный крит на 18: %+v", res)
	}
}

func TestSkillsAndPactSlots(t *testing.T) {
	d := DnD5e{}
	h := build(t, d, "human", "", "rogue")
	h.Abilities["dex"] = 17 // +3
	h.Skills = []string{"stealth", "acrobatics"}
	h.Expert = []string{"stealth"}
	sk := d.Derive(h).Skills
	if sk["stealth"] != 3+4 || sk["acrobatics"] != 3+2 || sk["athletics"] != 0 {
		t.Errorf("навыки: %v", sk)
	}
	if len(sk) != len(dndSkills) {
		t.Errorf("навыков %d, ожидалось %d", len(sk), len(dndSkills))
	}
	w := build(t, d, "human", "", "warlock")
	for lvl, want := range map[int][2]int{1: {1, 1}, 2: {1, 2}, 5: {3, 2}, 9: {5, 2}, 11: {5, 3}, 17: {5, 4}} {
		w.Level = lvl
		s := d.Derive(w).Slots
		if len(s) != 1 || s[0].Level != want[0] || s[0].Max != want[1] {
			t.Errorf("колдун %d: %+v, ожидалось %v", lvl, s, want)
		}
	}
	w.Level, w.SlotsUsed = 5, [9]int{0, 0, 2}
	d.Rest(w, "short")
	if w.SlotsUsed != [9]int{} {
		t.Error("колдун восстанавливает ячейки на коротком отдыхе")
	}
	m := build(t, d, "human", "", "wizard")
	m.SlotsUsed = [9]int{1}
	d.Rest(m, "short")
	if m.SlotsUsed[0] != 1 {
		t.Error("волшебник на коротком отдыхе ячейки не получает")
	}
	// драконья кровь: +1 HP за уровень
	s := build(t, d, "human", "", "sorcerer")
	base := d.Derive(s).MaxHP
	s.Subclass = "draconic"
	if got := d.Derive(s).MaxHP; got != base+1 {
		t.Errorf("драконья стойкость: %d → %d", base, got)
	}
}

func caster(t *testing.T, class string, lvl int, ab string, score int) *model.Character {
	c := build(t, DnD5e{}, "human", "", class)
	c.Level, c.Abilities[ab] = lvl, score
	return c
}

func TestSpellDCAndAttack(t *testing.T) {
	w := caster(t, "wizard", 5, "int", 17) // +3, мастерство +3
	d := (DnD5e{}).Derive(w)
	if d.SpellDC != 14 || d.SpellAtk != 6 {
		t.Errorf("СЛ %d, атака %d", d.SpellDC, d.SpellAtk)
	}
	if f := (DnD5e{}).Derive(build(t, DnD5e{}, "human", "", "fighter")); f.SpellDC != 0 {
		t.Error("у воина нет СЛ заклинаний")
	}
}

func TestFireballAreaAndHalf(t *testing.T) {
	d := DnD5e{}
	w := caster(t, "wizard", 5, "int", 17)
	w.Spells = []model.Spell{{Name: "Огненный шар", Level: 3, Ref: "fireball"}}
	a, b := dummy(10, 40), dummy(10, 40)
	a.ID, b.ID = "a", "b"
	// 8d6 по 4 = 32; первая цель проваливает спасбросок (1), вторая проходит (20 → половина)
	r := &seq{v: []int{3, 3, 3, 3, 3, 3, 3, 3, 0, 19}}
	lines, err := d.Spell(r, w, w.Spells[0], 3, []*model.Character{a, b}, dice.Normal)
	if err != nil {
		t.Fatal(err)
	}
	if a.HP != 8 || b.HP != 24 {
		t.Errorf("HP целей %d и %d, ожидалось 8 и 24\n%s", a.HP, b.HP, strings.Join(lines, "\n"))
	}
	if w.SlotsUsed[2] != 1 {
		t.Error("ячейка 3 уровня потрачена")
	}
	// улучшение: ячейка 4 уровня – 9d6
	w.Spells[0].Level = 3
	a.HP = 100
	d.Spell(&seq{v: []int{3, 3, 3, 3, 3, 3, 3, 3, 3, 0}}, w, w.Spells[0], 3, []*model.Character{a}, dice.Normal)
}

func TestSpellErrorsDoNotSpendSlot(t *testing.T) {
	d := DnD5e{}
	w := caster(t, "wizard", 5, "int", 17)
	sp := model.Spell{Name: "Огненный шар", Level: 3, Ref: "fireball"}
	if _, err := d.Spell(&seq{v: []int{1}}, w, sp, 3, nil, dice.Normal); err == nil {
		t.Error("без цели заклинание невозможно")
	}
	if w.SlotsUsed != [9]int{} {
		t.Error("ячейка не должна тратиться при ошибке")
	}
	if _, err := d.Spell(&seq{v: []int{1}}, w, sp, 2, []*model.Character{dummy(10, 10)}, dice.Normal); err == nil {
		t.Error("шар во 2-й ячейке невозможен")
	}
	w.SlotsUsed = [9]int{0, 0, 2}
	if _, err := d.Spell(&seq{v: []int{1}}, w, sp, 3, []*model.Character{dummy(10, 10)}, dice.Normal); err == nil {
		t.Error("ячейки закончились")
	}
	// описательное заклинание только тратит ячейку
	lines, err := d.Spell(&seq{v: []int{1}}, w, model.Spell{Name: "Щит", Level: 1, Ref: "shield"}, 1, nil, dice.Normal)
	if err != nil || len(lines) != 2 || w.SlotsUsed[0] != 1 || len(w.Effects) != 1 {
		t.Errorf("щит: %v %v", lines, err)
	}
}

func TestMagicMissileCantripScalingAndHeal(t *testing.T) {
	d := DnD5e{}
	w := caster(t, "wizard", 5, "int", 17)
	foe := dummy(10, 100)
	mm := model.Spell{Name: "Волшебная стрела", Level: 1, Ref: "magicmissile"}
	d.Spell(&seq{v: []int{1}}, w, mm, 1, []*model.Character{foe}, dice.Normal) // 3 снаряда по 1d4+1 = 3
	if foe.HP != 91 {
		t.Errorf("3 снаряда: HP %d", foe.HP)
	}
	d.Spell(&seq{v: []int{1}}, w, mm, 2, []*model.Character{foe}, dice.Normal) // 4 снаряда
	if foe.HP != 79 {
		t.Errorf("4 снаряда: HP %d", foe.HP)
	}
	// огненный снаряд на 5 уровне: 2d10; d20 = 15 + 6 против КД 10
	fb := model.Spell{Name: "Огненный снаряд", Ref: "firebolt"}
	d.Spell(&seq{v: []int{14, 4, 4}}, w, fb, 0, []*model.Character{foe}, dice.Normal)
	if foe.HP != 79-10 {
		t.Errorf("заговор 2d10: HP %d", foe.HP)
	}
	if w.SlotsUsed != [9]int{1, 1} {
		t.Errorf("заговор не тратит ячейки: %v", w.SlotsUsed)
	}
	// лечение: 1d8 + мод. Мудрости
	cl := caster(t, "cleric", 1, "wis", 15) // +2
	cl.HP = 1
	cw := model.Spell{Name: "Лечение ран", Level: 1, Ref: "curewounds"}
	if _, err := d.Spell(&seq{v: []int{3}}, cl, cw, 1, nil, dice.Normal); err != nil {
		t.Fatal(err)
	}
	if cl.HP != 1+4+2 {
		t.Errorf("лечение себя: HP %d", cl.HP)
	}
}

func TestSecondWind(t *testing.T) {
	d := DnD5e{}
	f := build(t, d, "human", "", "fighter")
	f.Level = 3
	f.HP = 1
	msg, err := d.Act(&seq{v: []int{4}}, f, "secondwind") // 1d10 = 5, +3
	if err != nil || f.HP != 9 {
		t.Fatalf("%q %v HP=%d", msg, err, f.HP)
	}
	if _, err := d.Act(&seq{v: []int{4}}, f, "secondwind"); err == nil {
		t.Error("второе дыхание раз за короткий отдых")
	}
	d.Rest(f, "short")
	if _, err := d.Act(&seq{v: []int{4}}, f, "secondwind"); err != nil {
		t.Errorf("после короткого отдыха доступно: %v", err)
	}
	if _, err := d.Act(&seq{v: []int{4}}, build(t, d, "human", "", "wizard"), "secondwind"); err == nil {
		t.Error("волшебник не владеет вторым дыханием")
	}
}

func TestMonsterResistances(t *testing.T) {
	d := DnD5e{}
	sk := dummy(10, 20)
	sk.Stat.Vuln, sk.Stat.Immune, sk.Stat.Resist = []string{"bludgeoning"}, []string{"poison"}, []string{"slashing"}
	for ty, want := range map[string]float64{"bludgeoning": 2, "poison": 0, "slashing": .5, "fire": 1} {
		if got := d.Resist(sk, ty); got != want {
			t.Errorf("%s: ×%v, ожидалось ×%v", ty, got, want)
		}
	}
	h := build(t, d, "human", "", "fighter")
	w := model.Weapon{Name: "Яд", Dice: "1d6", Ability: "str", Type: "poison"}
	if res := d.Attack(&seq{v: []int{19, 5}}, h, sk, w, dice.Normal); res.Damage != 0 {
		t.Errorf("иммунитет к яду: %+v", res)
	}
}

func TestCatalogDataIntegrity(t *testing.T) {
	ids := map[string]bool{}
	for _, c := range dndClasses {
		if c.HitDie == 0 || len(c.Saves) != 2 || len(c.Feats) == 0 || len(c.Subs) == 0 {
			t.Errorf("класс %s заполнен не полностью", c.ID)
		}
		for _, f := range c.Feats {
			if f.Level < 1 || f.Level > 20 || f.Name == "" || f.Desc == "" {
				t.Errorf("%s: умение %+v", c.ID, f)
			}
		}
		for _, s := range c.Subs {
			if len(s.Feats) == 0 {
				t.Errorf("%s/%s без умений", c.ID, s.ID)
			}
		}
		ids[c.ID] = true
	}
	ab := map[string]bool{}
	for _, a := range dndAbilities {
		ab[a.ID] = true
	}
	for _, s := range dndSkills {
		if !ab[s.Ability] {
			t.Errorf("навык %s: характеристика %q", s.ID, s.Ability)
		}
	}
	seen := map[string]bool{}
	for _, s := range spellDefs {
		if seen[s.ID] || s.Name == "" || s.Level < 0 || s.Level > 9 {
			t.Errorf("заклинание %+v", s)
		}
		seen[s.ID] = true
		for _, c := range s.Classes {
			if !ids[c] {
				t.Errorf("%s: неизвестный класс %q", s.ID, c)
			}
		}
		if s.Mode != "" {
			if _, err := dice.Parse(s.Dmg); err != nil {
				t.Errorf("%s: %v", s.ID, err)
			}
		}
		if s.Mode == "save" && !ab[s.Save] {
			t.Errorf("%s: спасбросок %q", s.ID, s.Save)
		}
		if s.Mode == "attack" || s.Mode == "save" || s.Mode == "auto" {
			if !slices.ContainsFunc(dndDamage, func(a Ability) bool { return a.ID == s.Type }) {
				t.Errorf("%s: тип урона %q", s.ID, s.Type)
			}
		}
	}
}

func TestSubclassCatalogIntegrity(t *testing.T) {
	total := 0
	for _, c := range dndClasses {
		seen := map[string]bool{}
		for _, s := range c.Subs {
			if seen[s.ID] || s.ID == "" || s.Name == "" || len(s.Feats) == 0 {
				t.Errorf("%s/%s: повтор или пустой подкласс", c.ID, s.ID)
			}
			seen[s.ID] = true
			total++
			for _, f := range s.Feats {
				if f.Level < 1 || f.Level > 20 || f.Name == "" || f.Desc == "" {
					t.Errorf("%s/%s: кривое умение %+v", c.ID, s.ID, f)
				}
			}
		}
		if len(c.Subs) < 2 {
			t.Errorf("%s: подклассов мало: %d", c.ID, len(c.Subs))
		}
		if tx := classText[c.ID]; tx.Desc == "" || tx.Lore == "" {
			t.Errorf("%s: нет описания или истории", c.ID)
		}
	}
	if len(dndClasses) < 16 || total < 80 {
		t.Errorf("классов %d, подклассов %d – ожидалось больше", len(dndClasses), total)
	}
	// новые классы создаются и считаются
	for _, id := range []string{"mystic", "gunslinger"} {
		h := build(t, DnD5e{}, "human", "", id)
		if d := (DnD5e{}).Derive(h); d.MaxHP < 1 || d.AC < 10 {
			t.Errorf("%s: %+v", id, d)
		}
	}
}
