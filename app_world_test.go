package main

import (
	"slices"
	"strings"
	"testing"

	"heroesbook/internal/model"
	"heroesbook/internal/worldgen"
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
	m, err = a.PopulatePlace(m.ID, p.ID, 3)
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

func TestOpenPlaceMakesLocationMapOnce(t *testing.T) {
	a := newApp()
	w, err := a.GenerateWorld("Мир", 5, []string{"human", "dwarf"}, 3, "small", "")
	if err != nil {
		t.Fatal(err)
	}
	var capital, cave model.Place
	for _, p := range w.Places {
		if p.Kind == "capital" && capital.ID == "" {
			capital = p
		}
		if worldgen.Wild(p.Kind) && cave.ID == "" {
			cave = p
		}
	}
	town, err := a.OpenPlace(w.ID, capital.ID, 3)
	if err != nil {
		t.Fatal(err)
	}
	if town.Parent != w.ID || town.ParentPlace != capital.ID || town.Name != capital.Name || len(town.Places) < 3 {
		t.Fatalf("карта столицы: %+v", town)
	}
	if !slices.ContainsFunc(town.Places, func(p model.Place) bool { return p.Kind == "tavern" && len(p.NPCs) > 0 }) {
		t.Error("в городе должна быть населённая таверна")
	}
	again, err := a.OpenPlace(w.ID, capital.ID, 3)
	if err != nil || again.ID != town.ID {
		t.Fatalf("повторное открытие должно вернуть ту же карту: %v", err)
	}
	tavern := town.Places[slices.IndexFunc(town.Places, func(p model.Place) bool { return p.Kind == "tavern" })]
	inside, err := a.OpenPlace(town.ID, tavern.ID, 3)
	if err != nil || inside.Grid == 0 || inside.Parent != town.ID {
		t.Fatalf("таверна изнутри – боевая карта с сеткой: %+v %v", inside.Grid, err)
	}
	if _, err := a.OpenPlace(w.ID, cave.ID, 3); err != nil {
		t.Fatalf("дикое место %s: %v", cave.Kind, err)
	}
	if err := a.DeleteMap(w.ID); err != nil {
		t.Fatal(err)
	}
	if n := len(a.Maps()); n != 0 {
		t.Errorf("с картой мира удаляются и карты её мест, осталось %d", n)
	}
}

func TestPartyTravelsAcrossWorld(t *testing.T) {
	a := newApp()
	w, err := a.GenerateWorld("Мир", 9, []string{"human", "elf"}, 3, "small", "")
	if err != nil {
		t.Fatal(err)
	}
	from, to := w.Places[0], w.Places[1]
	if _, err := a.PlanJourney(w.ID, to.X, to.Y, "normal"); err == nil {
		t.Error("без отряда маршрута нет")
	}
	if w, err = a.SetParty(w.ID, from.X, from.Y); err != nil || w.Party == nil {
		t.Fatalf("отряд: %v", err)
	}
	plan, err := a.PlanJourney(w.ID, to.X, to.Y, "normal")
	if err != nil {
		t.Fatal(err)
	}
	if plan.Miles <= 0 || plan.Days <= 0 || len(plan.Path) < 2 || len(plan.Legs) == 0 {
		t.Fatalf("маршрут: %+v", plan)
	}
	j, err := a.Travel(w.ID, to.X, to.Y, "fast", 3)
	if err != nil {
		t.Fatal(err)
	}
	if j.Days > plan.Days {
		t.Error("быстрым темпом не дольше")
	}
	for _, e := range j.Encounters {
		if e.Day < 1 || len(e.Foes) == 0 || e.Text == "" {
			t.Errorf("встреча: %+v", e)
		}
	}
	m := a.Maps()[0]
	if m.Party == nil || m.Party.X != to.X || m.Party.Y != to.Y {
		t.Errorf("отряд должен прийти: %+v", m.Party)
	}
	if err := a.DeleteMap(w.ID); err != nil {
		t.Fatal(err)
	}
}
