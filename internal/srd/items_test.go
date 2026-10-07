package srd

import (
	"testing"

	"heroesbook/internal/dice"
)

func TestItems(t *testing.T) {
	names := map[string]bool{}
	for _, it := range Items() {
		if names[it.Name] || it.Name == "" || it.Weight < 0 {
			t.Errorf("предмет %+v", it)
		}
		names[it.Name] = true
		if it.Weapon != nil {
			if _, err := dice.Parse(it.Weapon.Dice); err != nil {
				t.Errorf("%s: %v", it.Name, err)
			}
		}
		if it.ArmorBase > 0 && it.ArmorType == "" {
			t.Errorf("%s: у доспеха нет типа", it.Name)
		}
		if it.Weapon != nil && it.ArmorBase > 0 {
			t.Errorf("%s: и оружие, и доспех", it.Name)
		}
	}
}
