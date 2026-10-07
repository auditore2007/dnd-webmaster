package rules

import (
	"strings"
	"testing"

	"heroesbook/internal/dice"
	"heroesbook/internal/model"
)

// У каждого подкласса должна быть хотя бы одна строка механики (кнопка или пассивное умение).
func TestEverySubclassHasMechanics(t *testing.T) {
	has := map[string]bool{}
	for _, a := range abTable {
		has[a.Class+"/"+a.Sub] = true
	}
	for _, c := range dndCatalog().Classes {
		for _, s := range c.Subs {
			if !has[c.ID+"/"+s.ID] {
				t.Errorf("у подкласса %s/%s (%s) нет механики", c.ID, s.ID, s.Name)
			}
		}
	}
}

// Каждую способность каждого подкласса можно применить: без паники и без неожиданных ошибок.
func TestEverySubclassAbilityWorks(t *testing.T) {
	d := DnD5e{}
	r := dice.RNG{}
	allowed := []string{"нечего восстанавливать", "не потрачена", "не нуждается в лечении", "нет подходящих целей", "нет свободных ячеек"}
	for _, cl := range dndCatalog().Classes {
		for _, sub := range cl.Subs {
			hero := &model.Character{ID: "h", Name: "Герой", Ruleset: "dnd5e", Race: "human", Class: cl.ID}
			if err := d.Build(hero); err != nil {
				t.Fatal(err)
			}
			hero.Level, hero.Subclass = 20, sub.ID
			hero.HP = d.Derive(hero).MaxHP
			ally := &model.Character{ID: "a", Name: "Союзник", Ruleset: "dnd5e", Race: "human", Class: "fighter"}
			_ = d.Build(ally)
			ally.HP = 1
			for _, a := range abilitiesFor(hero) {
				if a.Passive {
					continue
				}
				setActive(hero, "rage", true)
				foe := &model.Character{ID: "f", Name: "Враг", Ruleset: "dnd5e", Kind: "monster", HP: 500, Abilities: map[string]int{"str": 10, "dex": 10, "con": 10, "int": 10, "wis": 10, "cha": 10},
					Stat: &model.MonsterStat{CR: "1", Kind: "undead", AC: 12, MaxHP: 500}}
				var err error
				if a.Special != nil {
					_, err = d.Special(r, a.Key, hero, []*model.Character{foe})
				} else {
					tgt := ally
					if a.Target == "foe" {
						tgt = foe
					}
					_, _, err = d.Ability(r, a.Key, hero, tgt)
				}
				if err != nil && !containsAny(err.Error(), allowed) {
					t.Errorf("%s/%s «%s»: %v", cl.ID, sub.ID, a.Name, err)
				}
				hero.Counters = nil // ресурсы не мешают проверить следующую способность
			}
		}
	}
}

func containsAny(s string, parts []string) bool {
	for _, p := range parts {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}
