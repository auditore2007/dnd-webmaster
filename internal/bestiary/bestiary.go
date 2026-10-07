// Package bestiary — существа D&D 5e для боя: встроенный список (SRD / Monster Manual, значения по памяти — сверяйте с книгой)
// и собственные существа мастера.
package bestiary

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"heroesbook/internal/dice"
	"heroesbook/internal/model"
)

type Monster struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	CR        string          `json:"cr"`
	Kind      string          `json:"kind"` // beast, humanoid, undead, dragon, giant, fiend, fey, construct, ooze, elemental, monstrosity, aberration, plant, celestial
	AC        int             `json:"ac"`
	HP        int             `json:"hp"`
	Attack    int             `json:"attack"`
	Speed     int             `json:"speed"`
	Abilities [6]int          `json:"abilities"` // СИЛ ЛВК ТЕЛ ИНТ МДР ХАР
	Weapon    model.Weapon    `json:"weapon"`    // первое оружие, для краткого вида
	Weapons   []model.Weapon  `json:"weapons"`
	Multi     []int           `json:"multi"` // индексы оружия в мультиатаке
	Traits    []string        `json:"traits"`
	Resist    []string        `json:"resist"`
	Vuln      []string        `json:"vuln"`
	Immune    []string        `json:"immune"`
	Specials  []model.Special `json:"specials"`
	Custom    bool            `json:"custom"` // создано мастером
	Desc      string          `json:"desc"`
}

var kinds = []string{"beast", "humanoid", "undead", "dragon", "giant", "fiend", "fey", "construct", "ooze", "elemental", "monstrosity", "aberration", "plant", "celestial"}

// Kinds — допустимые типы существ.
func Kinds() []string { return slices.Clone(kinds) }

var damageTypes = []string{"slashing", "piercing", "bludgeoning", "fire", "cold", "lightning", "thunder", "acid", "poison", "necrotic", "radiant", "force", "psychic"}

var list []*Monster

func init() {
	for _, ln := range strings.Split(table, "\n") {
		p := strings.Split(strings.TrimSpace(ln), "|")
		if len(p) != 15 {
			continue
		}
		x := &Monster{ID: p[0], Name: p[1], CR: p[2], Kind: p[3]}
		x.AC, _ = strconv.Atoi(p[4])
		x.HP, _ = strconv.Atoi(p[5])
		x.HP = (x.HP*3 + 1) / 2 // здоровье встроенных существ ×1,5 (с округлением вверх): в таблице значения из книг
		x.Attack, _ = strconv.Atoi(p[6])
		x.Speed, _ = strconv.Atoi(p[7])
		for i, v := range strings.Split(p[8], ",") {
			x.Abilities[i], _ = strconv.Atoi(v)
		}
		for _, w := range strings.Split(p[9], ";") {
			f := strings.Split(w, ":")
			if len(f) != 3 {
				continue
			}
			ranged := strings.HasSuffix(f[2], "~")
			wp := model.Weapon{Name: f[0], Dice: f[1], Type: strings.TrimSuffix(f[2], "~"), Ability: "str", Ranged: ranged}
			if ranged {
				wp.Ability = "dex"
			}
			x.Weapons = append(x.Weapons, wp)
		}
		if len(x.Weapons) > 0 {
			x.Weapon = x.Weapons[0]
		}
		for _, s := range strings.Split(p[10], ",") {
			if n, err := strconv.Atoi(s); err == nil {
				x.Multi = append(x.Multi, n)
			}
		}
		split := func(s, sep string) []string {
			if s == "" {
				return nil
			}
			return strings.Split(s, sep)
		}
		x.Traits, x.Resist, x.Vuln, x.Immune = split(p[11], ";"), split(p[12], ","), split(p[13], ","), split(p[14], ",")
		x.Specials = slices.Clone(builtin[x.ID])
		list = append(list, x)
	}
	sort.SliceStable(list, func(i, j int) bool {
		a, b := CRValue(list[i].CR), CRValue(list[j].CR)
		if a != b {
			return a < b
		}
		return list[i].HP < list[j].HP
	})
}

// CRValue переводит «1/4» в число для сортировки.
func CRValue(cr string) float64 {
	if n, d, ok := strings.Cut(cr, "/"); ok {
		a, _ := strconv.ParseFloat(n, 64)
		b, _ := strconv.ParseFloat(d, 64)
		if b > 0 {
			return a / b
		}
	}
	v, _ := strconv.ParseFloat(cr, 64)
	return v
}

// List — встроенные существа по возрастанию уровня опасности (CR), затем по HP.
func List() []Monster {
	out := make([]Monster, len(list))
	for i, x := range list {
		out[i] = *x
	}
	return out
}

// Sort упорядочивает любой список существ так же, как встроенный.
func Sort(ms []Monster) {
	sort.SliceStable(ms, func(i, j int) bool {
		a, b := CRValue(ms[i].CR), CRValue(ms[j].CR)
		if a != b {
			return a < b
		}
		return ms[i].HP < ms[j].HP
	})
}

// Build создаёт существо из описания; номер n добавляется к имени, чтобы в бою их можно было различать.
func Build(x Monster, n int, newID func() string) *model.Character {
	ab := map[string]int{}
	for i, k := range []string{"str", "dex", "con", "int", "wis", "cha"} {
		ab[k] = x.Abilities[i]
	}
	return &model.Character{ID: newID(), Name: fmt.Sprintf("%s %d", x.Name, n), Kind: "monster", Ruleset: "dnd5e", Level: 1,
		Abilities: ab, HP: x.HP, Weapons: append([]model.Weapon(nil), x.Weapons...),
		Stat: &model.MonsterStat{CR: x.CR, Kind: x.Kind, AC: x.AC, MaxHP: x.HP, Attack: x.Attack, Speed: x.Speed, Traits: slices.Clone(x.Traits),
			ID: x.ID, Specials: slices.Clone(x.Specials), Multi: slices.Clone(x.Multi), Resist: slices.Clone(x.Resist), Vuln: slices.Clone(x.Vuln), Immune: slices.Clone(x.Immune)}}
}

// Make создаёт встроенное существо.
func Make(id string, n int, newID func() string) (*model.Character, error) {
	for _, x := range list {
		if x.ID == id {
			return Build(*x, n, newID), nil
		}
	}
	return nil, fmt.Errorf("неизвестное существо: %q", id)
}

// Validate проверяет и нормализует собственное существо.
func Validate(m *Monster) error {
	m.Name = strings.TrimSpace(m.Name)
	switch {
	case m.Name == "" || len([]rune(m.Name)) > 60:
		return errors.New("существо: нужно имя до 60 символов")
	case m.AC < 1 || m.AC > 40:
		return errors.New("существо: КД от 1 до 40")
	case m.HP < 1 || m.HP > 3000:
		return errors.New("существо: HP от 1 до 3000")
	case m.Attack < -5 || m.Attack > 30:
		return errors.New("существо: бонус атаки от -5 до 30")
	case m.Speed < 0 || m.Speed > 300:
		return errors.New("существо: скорость от 0 до 300")
	case len(m.Weapons) == 0 || len(m.Weapons) > 8:
		return errors.New("существо: нужна хотя бы одна атака (до 8)")
	}
	if m.CR = strings.TrimSpace(m.CR); m.CR == "" {
		m.CR = "0"
	}
	if CRValue(m.CR) < 0 || CRValue(m.CR) > 30 || (m.CR != "0" && CRValue(m.CR) == 0) {
		return fmt.Errorf("существо: непонятный уровень опасности %q", m.CR)
	}
	if !slices.Contains(kinds, m.Kind) {
		m.Kind = "monstrosity"
	}
	for i, v := range m.Abilities {
		if v < 1 || v > 30 {
			return errors.New("существо: характеристики от 1 до 30")
		}
		_ = i
	}
	for i := range m.Weapons {
		w := &m.Weapons[i]
		if w.Name = strings.TrimSpace(w.Name); w.Name == "" {
			w.Name = "Атака"
		}
		if _, err := dice.Parse(w.Dice); err != nil {
			return fmt.Errorf("атака «%s»: %w", w.Name, err)
		}
		if !slices.Contains(damageTypes, w.Type) {
			return fmt.Errorf("атака «%s»: неизвестный тип урона %q", w.Name, w.Type)
		}
		w.Ability = "str"
		if w.Ranged {
			w.Ability = "dex"
		}
		w.Bonus = 0
	}
	m.Weapon = m.Weapons[0]
	for _, i := range m.Multi {
		if i < 0 || i >= len(m.Weapons) {
			return errors.New("существо: мультиатака ссылается на несуществующую атаку")
		}
	}
	if len(m.Multi) > 10 {
		return errors.New("существо: в мультиатаке не больше 10 ударов")
	}
	for _, list := range [][]string{m.Resist, m.Vuln, m.Immune} {
		for _, t := range list {
			if !slices.Contains(damageTypes, t) {
				return fmt.Errorf("существо: неизвестный тип урона %q", t)
			}
		}
	}
	if len(m.Specials) > 6 {
		return errors.New("существо: не больше 6 особых способностей")
	}
	for i := range m.Specials {
		if err := validSpecial(&m.Specials[i], i); err != nil {
			return err
		}
	}
	m.Custom = true
	return nil
}
