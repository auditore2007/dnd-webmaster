package bestiary

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"heroesbook/internal/dice"
	"heroesbook/internal/model"
)

var abilityIDs = []string{"str", "dex", "con", "int", "wis", "cha"}
var condIDs = []string{"", "Ослеплён", "Очарован", "Оглохший", "Напуган", "Схвачен", "Недееспособен", "Парализован", "Окаменел", "Отравлен", "Сбит с ног", "Опутан", "Ошеломлён"}

const (
	maxLegendary = 5
	maxLegActs   = 6
	maxLair      = 4
)

// validBoss проверяет легендарные действия, логово и фазу собственного существа.
func validBoss(m *Monster) error {
	switch {
	case m.Legendary < 0 || m.Legendary > maxLegendary || m.LegRes < 0 || m.LegRes > maxLegendary:
		return fmt.Errorf("существо: легендарных действий и сопротивлений от 0 до %d", maxLegendary)
	case len(m.LegActs) > maxLegActs || len(m.Lair) > maxLair:
		return fmt.Errorf("существо: не больше %d легендарных действий и %d действий логова", maxLegActs, maxLair)
	}
	for i := range m.LegActs {
		a := &m.LegActs[i]
		if a.Name = strings.TrimSpace(a.Name); a.Name == "" {
			return errors.New("легендарное действие: нужно название")
		}
		a.Key, a.Cost = fmt.Sprintf("la%d", i), max(1, min(3, a.Cost))
		if a.Special == nil {
			if a.Weapon < 0 || a.Weapon >= len(m.Weapons) {
				return fmt.Errorf("легендарное действие «%s»: нет такой атаки", a.Name)
			}
			continue
		}
		a.Weapon = -1
		a.Special.Name = a.Name
		if err := validSpecial(a.Special, i); err != nil {
			return err
		}
		a.Special.Key = a.Key
	}
	for i := range m.Lair {
		if err := validSpecial(&m.Lair[i], i); err != nil {
			return err
		}
		m.Lair[i].Key = fmt.Sprintf("lair%d", i)
	}
	if p := m.Phase; p != nil {
		if p.Name = strings.TrimSpace(p.Name); p.Name == "" {
			p.Name = "Вторая фаза"
		}
		if p.Pct < 10 || p.Pct > 90 || p.Atk < 0 || p.Atk > 10 || p.AC < 0 || p.AC > 10 || p.Extra < 0 || p.Extra > 3 || p.Temp < 0 || p.Temp > 1000 {
			return errors.New("фаза: порог 10–90%, атака и КД до +10, атак до +3, временные HP до 1000")
		}
	}
	return nil
}

// validSpecial проверяет и нормализует особую способность собственного существа.
func validSpecial(s *model.Special, i int) error {
	s.Name = strings.TrimSpace(s.Name)
	if s.Name == "" {
		return errors.New("способность: нужно название")
	}
	s.Key = fmt.Sprintf("sp%d", i)
	if s.Mode != "area" {
		s.Mode = "single"
	}
	for _, d := range []*string{&s.Dmg, &s.XDmg} {
		if *d = strings.TrimSpace(*d); *d != "" {
			if _, err := dice.Parse(*d); err != nil {
				return fmt.Errorf("способность «%s»: %w", s.Name, err)
			}
		}
	}
	for _, t := range []string{s.Type, s.XType} {
		if t != "" && !slices.Contains(damageTypes, t) {
			return fmt.Errorf("способность «%s»: неизвестный тип урона %q", s.Name, t)
		}
	}
	if s.Dmg != "" && s.Type == "" {
		s.Type = "force"
	}
	if s.XDmg != "" && s.XType == "" {
		s.XType = "poison"
	}
	if s.Save != "" && !slices.Contains(abilityIDs, s.Save) {
		return fmt.Errorf("способность «%s»: неизвестная характеристика %q", s.Name, s.Save)
	}
	if s.RSave != "" && !slices.Contains(abilityIDs, s.RSave) {
		s.RSave = ""
	}
	if !slices.Contains(condIDs, s.Cond) {
		return fmt.Errorf("способность «%s»: неизвестное состояние %q", s.Name, s.Cond)
	}
	if s.DC < 0 || s.DC > 40 || s.Rounds < 0 || s.Rounds > 1000 {
		return fmt.Errorf("способность «%s»: СЛ 0–40, длительность 0–1000 ходов", s.Name)
	}
	if s.Save != "" && s.DC == 0 {
		s.DC = 10
	}
	if s.Dmg == "" && s.XDmg == "" && s.Cond == "" {
		return fmt.Errorf("способность «%s»: нужен урон или состояние", s.Name)
	}
	return nil
}
