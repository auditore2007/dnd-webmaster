// Package rules — системы правил. Ruleset — стратегия: бой и интерфейс не знают,
// по каким правилам считается КД, здоровье, попадание и смерть. Интерфейс разбит на роли.
package rules

import (
	"fmt"
	"strings"

	"heroesbook/internal/dice"
	"heroesbook/internal/model"
)

type Ability struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Race struct {
	ID      string         `json:"id"`
	Name    string         `json:"name"`
	Speed   int            `json:"speed"`
	Bonus   map[string]int `json:"bonus"` // D&D: бонусы к характеристикам; Алдоран: базовые значения
	Traits  []string       `json:"traits"`
	Resist  []string       `json:"resist"`
	HPLevel int            `json:"-"` // доп. HP за уровень (холмовой дварф)
	Subs    []Race         `json:"subs"`
	Desc    string         `json:"desc"` // коротко: кто это
	Lore    string         `json:"lore"` // история народа
}

type Feat struct {
	Level int    `json:"level"`
	Name  string `json:"name"`
	Desc  string `json:"desc"`
	Key   string `json:"-"` // непустой ключ — способность с механикой (rage, secondwind)
}

type Subclass struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Feats []Feat `json:"features"`
}

type Class struct {
	ID     string     `json:"id"`
	Name   string     `json:"name"`
	HitDie int        `json:"hitDie"`
	Saves  []string   `json:"saves"`
	Feats  []Feat     `json:"features"`
	Subs   []Subclass `json:"subclasses"`
	Desc   string     `json:"desc"`
	Lore   string     `json:"lore"`
}

type Skill struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Ability string `json:"ability"`
}

type Catalog struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	ACLabel     string     `json:"acLabel"`
	DeathSaves  bool       `json:"deathSaves"`
	Abilities   []Ability  `json:"abilities"`
	Races       []Race     `json:"races"`
	Classes     []Class    `json:"classes"`
	Conditions  []string   `json:"conditions"`
	DamageTypes []Ability  `json:"damageTypes"`
	Rests       []Ability  `json:"rests"`
	Skills      []Skill    `json:"skills"`
	Spells      []SpellDef `json:"spells"`
}

// Core — создание героя и производные значения.
type Core interface {
	ID() string
	Catalog() Catalog
	Build(c *model.Character) error
	Derive(c *model.Character) model.Derived
}

// Combatant — всё, что нужно бою.
type Combatant interface {
	Initiative(r dice.Roller, c *model.Character) int
	Attack(r dice.Roller, atk, def *model.Character, w model.Weapon, m dice.Mode) model.AttackResult
	Resist(def *model.Character, damageType string) float64
	AfterDamage(c *model.Character, wasDown, crit bool)
	TurnStart(r dice.Roller, c *model.Character) (lines []string, skip bool)
	// EndTurn — конец хода: повторные спасброски от эффектов.
	EndTurn(r dice.Roller, c *model.Character) []string
	// ConcCheck — проверка концентрации после урона; lost — заклинание сорвано.
	ConcCheck(r dice.Roller, c *model.Character) (lines []string, lost bool)
	Survives(c *model.Character) (msg string, counter, ok bool)
	Special(r dice.Roller, key string, src *model.Character, foes []*model.Character) ([]string, error)
	DeathSave(r dice.Roller, c *model.Character) (string, error)
	// Ability применяет способность класса или подкласса (кости превосходства, ки, канал…). extra — дополнительные атаки в этот ход.
	Ability(r dice.Roller, key string, src, tgt *model.Character) (msg string, extra int, err error)
	// Spell творит заклинание из каталога: тратит ячейку и применяет эффект к целям.
	Spell(r dice.Roller, src *model.Character, s model.Spell, slot int, targets []*model.Character, m dice.Mode) ([]string, error)
}

// Progression — опыт, отдых, заклинания.
type Progression interface {
	XPFor(level int) int
	LevelPoints(level int) int
	Rest(c *model.Character, kind string)
	Cast(c *model.Character, s model.Spell, slot int) (string, error)
	// Toggle включает и выключает способность, учитывая её запас (ярость).
	Toggle(c *model.Character, key string) (string, error)
	// Act применяет способность-действие (второе дыхание).
	Act(r dice.Roller, c *model.Character, key string) (string, error)
}

type Ruleset interface {
	Core
	Combatant
	Progression
}

var registry []Ruleset

func Register(r Ruleset) { registry = append(registry, r) }
func All() []Ruleset     { return registry }

func Get(id string) (Ruleset, error) {
	for _, r := range registry {
		if r.ID() == id {
			return r, nil
		}
	}
	return nil, fmt.Errorf("неизвестная система правил: %q", id)
}

func init() {
	Register(DnD5e{})
}

// Strike применяет урон по правилам системы: временные HP, «устоять», смерть.
// counter == true, если цель имеет право на немедленный ответный удар.
func Strike(rs Ruleset, def *model.Character, dmg int, crit bool) (lines []string, counter bool) {
	if dmg <= 0 || def.Dead {
		return nil, false
	}
	wasDown := def.Down()
	def.Damage(dmg)
	if def.Down() && !wasDown {
		if msg, c, ok := rs.Survives(def); ok {
			def.HP, counter = 1, c
			lines = append(lines, msg)
		}
	}
	rs.AfterDamage(def, wasDown, crit)
	if def.Conc != "" {
		if def.Down() {
			def.ConcDmg = -1
		} else {
			def.ConcDmg = max(def.ConcDmg, dmg)
		}
	}
	switch {
	case def.Dead:
		lines = append(lines, def.Name+" погибает")
	case def.Down() && !wasDown:
		lines = append(lines, def.Name+" без сознания")
	}
	return lines, counter
}

type effect struct {
	Src, Ab string
	V       int
}

func (e effect) text(names []Ability) string {
	n := e.Ab
	for _, a := range names {
		if a.ID == e.Ab {
			n = a.Name
		}
	}
	return fmt.Sprintf("%s: %+d %s", e.Src, e.V, n)
}

func applyEffects(base map[string]int, fx []effect) map[string]int {
	out := make(map[string]int, len(base))
	for k, v := range base {
		out[k] = v
	}
	for _, e := range fx {
		out[e.Ab] += e.V
	}
	return out
}

func itemEffects(c *model.Character) (fx []effect, ac int) {
	for _, it := range c.Inventory {
		if !it.Equipped {
			continue
		}
		ac += it.ACBonus
		if it.Ability != "" && it.AbilityBonus != 0 {
			fx = append(fx, effect{it.Name, strings.ToLower(it.Ability), it.AbilityBonus})
		}
	}
	return
}

// allWeapons — собственное оружие героя плюс оружие из надетых предметов.
func allWeapons(c *model.Character) []model.Weapon {
	out := append([]model.Weapon(nil), c.Weapons...)
	for _, it := range c.Inventory {
		if it.Equipped && it.Weapon != nil {
			w := *it.Weapon
			if w.Name == "" {
				w.Name = it.Name
			}
			out = append(out, w)
		}
	}
	return out
}

func totalWeight(c *model.Character) (w float64) {
	for _, it := range c.Inventory {
		w += it.Weight * float64(max(1, it.Qty))
	}
	w += c.PurseWeight()
	return
}

// merge возвращает расу вместе с подрасой. Карты копируются: данные каталога менять нельзя.
func merge(races []Race, id, sub string) (Race, error) {
	for _, r := range races {
		if r.ID != id {
			continue
		}
		out := r
		out.Bonus = map[string]int{}
		for k, v := range r.Bonus {
			out.Bonus[k] = v
		}
		out.Traits = append([]string(nil), r.Traits...)
		out.Resist = append([]string(nil), r.Resist...)
		if len(r.Subs) == 0 {
			return out, nil
		}
		for _, s := range r.Subs {
			if s.ID != sub {
				continue
			}
			if s.Speed > 0 {
				out.Speed = s.Speed
			}
			for k, v := range s.Bonus {
				out.Bonus[k] += v
			}
			out.Traits = append(out.Traits, s.Traits...)
			out.Resist = append(out.Resist, s.Resist...)
			out.HPLevel += s.HPLevel
			return out, nil
		}
		return Race{}, fmt.Errorf("для расы %q нужно выбрать подрасу", r.Name)
	}
	return Race{}, fmt.Errorf("неизвестная раса: %q", id)
}

func abilityOr(a, def string) string {
	if a == "" {
		return def
	}
	return strings.ToLower(a)
}
