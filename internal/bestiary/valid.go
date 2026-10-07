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
