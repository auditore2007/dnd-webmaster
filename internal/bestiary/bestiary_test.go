package bestiary

import (
	"testing"

	"heroesbook/internal/dice"
	"heroesbook/internal/model"
)

func TestBestiaryIntegrity(t *testing.T) {
	seen := map[string]bool{}
	for _, m := range List() {
		if seen[m.ID] || m.Name == "" || m.HP < 1 || m.AC < 5 || len(m.Weapons) == 0 {
			t.Errorf("%s: неполные данные %+v", m.ID, m)
		}
		seen[m.ID] = true
		for _, w := range m.Weapons {
			if _, err := dice.Parse(w.Dice); err != nil {
				t.Errorf("%s/%s: %v", m.ID, w.Name, err)
			}
		}
		for _, i := range m.Multi {
			if i < 0 || i >= len(m.Weapons) {
				t.Errorf("%s: мультиатака ссылается на оружие %d", m.ID, i)
			}
		}
		c, err := Make(m.ID, 1, func() string { return "x" })
		if err != nil || c.Stat.ID != m.ID || c.Stat.MaxHP != m.HP || len(c.Weapons) != len(m.Weapons) {
			t.Errorf("%s: Make: %v %+v", m.ID, err, c)
		}
	}
	if len(seen) < 150 {
		t.Errorf("в бестиарии только %d существ", len(seen))
	}
	if _, err := Make("nope", 1, func() string { return "x" }); err == nil {
		t.Error("неизвестное существо — ошибка")
	}
}

func TestMakeDoesNotShareState(t *testing.T) {
	a, _ := Make("goblin", 1, func() string { return "a" })
	a.Weapons[0].Dice = "9d9"
	b, _ := Make("goblin", 2, func() string { return "b" })
	if b.Weapons[0].Dice == "9d9" {
		t.Error("существа делят оружие через общий срез")
	}
}

func TestOrderedByDanger(t *testing.T) {
	prev := -1.0
	for _, m := range List() {
		if v := CRValue(m.CR); v < prev {
			t.Errorf("%s (CR %s) стоит не на своём месте", m.ID, m.CR)
		} else {
			prev = v
		}
		if CRValue(m.CR) == 0 && m.CR != "0" {
			t.Errorf("%s: CR %q не разобран", m.ID, m.CR)
		}
	}
	if CRValue("1/8") >= CRValue("1/4") || CRValue("1/2") >= CRValue("1") {
		t.Error("порядок дробных CR")
	}
}

func TestValidateCustom(t *testing.T) {
	ok := Monster{Name: "Тест", CR: "3", Kind: "beast", AC: 12, HP: 30, Attack: 4, Speed: 30, Abilities: [6]int{10, 10, 10, 10, 10, 10},
		Weapons: []model.Weapon{{Name: "Укус", Dice: "1d6+2", Type: "piercing"}}, Multi: []int{0, 0}}
	if err := Validate(&ok); err != nil || !ok.Custom {
		t.Fatal(err)
	}
	bad := ok
	bad.Weapons = []model.Weapon{{Dice: "xx", Type: "piercing"}}
	if Validate(&bad) == nil {
		t.Error("кривые кубики должны отклоняться")
	}
	bad = ok
	bad.Multi = []int{3}
	if Validate(&bad) == nil {
		t.Error("мультиатака с несуществующей атакой")
	}
	bad = ok
	bad.HP = 0
	if Validate(&bad) == nil {
		t.Error("HP 0")
	}
}

func TestBuiltinSpecialsAreValid(t *testing.T) {
	ids := map[string]bool{}
	for _, m := range List() {
		ids[m.ID] = true
	}
	for id, ss := range builtin {
		if !ids[id] {
			t.Errorf("особая способность для несуществующего существа %q", id)
		}
		seen := map[string]bool{}
		for _, s := range ss {
			if seen[s.Key] {
				t.Errorf("%s: ключ %q повторяется", id, s.Key)
			}
			seen[s.Key] = true
			s := s
			cp := s
			if err := validSpecial(&cp, 0); err != nil {
				t.Errorf("%s/%s: %v", id, s.Key, err)
			}
			if s.Mode != "area" && s.Mode != "single" {
				t.Errorf("%s/%s: режим %q", id, s.Key, s.Mode)
			}
			if s.Mode == "area" && s.Save == "" {
				t.Errorf("%s/%s: область без спасброска", id, s.Key)
			}
		}
	}
	m, err := Make("young-red-dragon", 1, func() string { return "x" })
	if err != nil || len(m.Stat.Specials) != 1 || !m.Stat.Specials[0].Recharge {
		t.Errorf("дракон должен дышать с перезарядкой: %+v %v", m, err)
	}
}

func TestCustomMonsterSpecialsValidated(t *testing.T) {
	m := Monster{Name: "Слизень", CR: "1", Kind: "ooze", AC: 8, HP: 30, Abilities: [6]int{10, 10, 10, 10, 10, 10}, Weapons: []model.Weapon{{Name: "Касание", Dice: "1d6", Type: "acid"}},
		Specials: []model.Special{{Name: "Плевок", Mode: "area", Dmg: "2d6", Type: "acid", Save: "dex", DC: 12, Half: true}}}
	if err := Validate(&m); err != nil || m.Specials[0].Key != "sp0" {
		t.Fatalf("%v %+v", err, m.Specials)
	}
	bad := m
	bad.Specials = []model.Special{{Name: "Ошибка", Dmg: "ой", Type: "acid"}}
	if Validate(&bad) == nil {
		t.Error("плохие кубики")
	}
	bad.Specials = []model.Special{{Name: "Пусто"}}
	if Validate(&bad) == nil {
		t.Error("способность без эффекта")
	}
}
