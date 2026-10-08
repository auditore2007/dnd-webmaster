package main

import (
	"strings"
	"testing"

)

func TestGenerateWorldAndDungeonMaps(t *testing.T) {
	a := newApp()
	m, err := a.GenerateWorld("Эльдория", 7, []string{"human", "dwarf"}, 3, "small", "")
	if err != nil {
		t.Fatal(err)
	}
	if m.Name != "Эльдория" || len(m.Places) < 4 {
		t.Fatalf("мир: %q, мест %d", m.Name, len(m.Places))
	}
	if img, _ := a.GetMapImage(m.ID); !strings.HasPrefix(img, "data:image/jpeg") {
		t.Error("у мира нет картинки")
	}
	if _, err := a.GenerateWorld("", 1, []string{"дракон-эльф"}, 3, "small", ""); err == nil {
		t.Error("неизвестная раса должна отклоняться")
	}
	d, err := a.GenerateDungeon("", 3, 6, 2)
	if err != nil || d.Name != "Подземелье" || len(d.Places) < 3 {
		t.Fatalf("подземелье: %+v %v", d.Name, err)
	}
}

func TestPlacesEditAndPopulate(t *testing.T) {
	a := newApp()
	m, err := a.GenerateDungeon("x", 3, 5, 2)
	if err != nil {
		t.Fatal(err)
	}
	m, err = a.AddPlace(m.ID, 10, 10, "tavern", "")
	if err != nil {
		t.Fatal(err)
	}
	p := m.Places[len(m.Places)-1]
	if p.Name != "Таверна" || len(p.NPCs) != 0 {
		t.Fatalf("новое место: %+v", p)
	}
	if _, err := a.AddPlace(m.ID, 1, 1, "космодром", ""); err == nil {
		t.Error("неизвестный вид места – ошибка")
	}
	m, err = a.PopulateMap(m.ID, 3, true)
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Places[len(m.Places)-1]; len(got.NPCs) == 0 {
		t.Error("в таверне после населения должны появиться люди")
	}
	p = m.Places[len(m.Places)-1]
	p.Name, p.Race = "Пьяный гусь", "dwarf"
	if m, err = a.UpdatePlace(m.ID, p); err != nil || m.Places[len(m.Places)-1].Name != "Пьяный гусь" {
		t.Fatalf("переименование: %v", err)
	}
	n := len(m.Places)
	if m, err = a.RemovePlace(m.ID, p.ID); err != nil || len(m.Places) != n-1 {
		t.Fatalf("удаление: %v", err)
	}
	if err := a.Undo(); err != nil || len(a.Maps()[0].Places) != n {
		t.Errorf("отмена должна вернуть место: %v", err)
	}
}
