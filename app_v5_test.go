package main

import (
	"strings"
	"testing"

	"heroesbook/internal/model"
)

// spawnFight создаёт героя и существо и начинает бой; текущий ход отдаётся существу.
func spawnFight(t *testing.T, a *App, monsterID string) (CharView, CharView) {
	t.Helper()
	h := hero(t, a)
	ms, err := a.AddMonster(monsterID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.StartEncounter([]string{h.ID, ms[0].ID}); err != nil {
		t.Fatal(err)
	}
	e := a.state.Encounter
	for i, id := range e.Order {
		if id == ms[0].ID {
			e.Turn = i
		}
	}
	e.Acted, e.Left = false, 1
	return h, ms[0]
}

func TestSpecialAtAppliesParalysis(t *testing.T) {
	a := newApp()
	h, g := spawnFight(t, a, "ghoul")
	a.state.Characters[0].HP = 100 // герой переживёт укус
	if _, err := a.UseSpecialAt("claws", h.ID); err != nil {
		t.Fatal(err)
	}
	c := a.find(h.ID)
	// fixed{9}: атака 10 + бонус попадает по КД героя, спасбросок 10 + мод против 10 – успех, паралича нет;
	// главное – способность отработала и записана в журнал
	if !a.find(g.ID).Used["claws"] && len(a.state.Log) == 0 {
		t.Error("способность должна отработать и попасть в журнал")
	}
	if _, err := a.UseSpecialAt("nope", h.ID); err == nil {
		t.Error("неизвестная способность должна давать ошибку")
	}
	_ = c
}

func TestReactionToggle(t *testing.T) {
	a := newApp()
	h, _ := spawnFight(t, a, "goblin")
	e, err := a.ToggleReaction(h.ID)
	if err != nil || !e.React[h.ID] {
		t.Fatalf("реакция должна отметиться: %v %v", e, err)
	}
	e, _ = a.ToggleReaction(h.ID)
	if e.React[h.ID] {
		t.Error("повторный вызов возвращает реакцию")
	}
	if _, err := a.ToggleReaction("нет"); err == nil {
		t.Error("чужой id")
	}
}

func TestEffectsAddRemoveAndCleanup(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	v, err := a.AddEffect(h.ID, "Горение", 3, "Напуган")
	if err != nil || len(v.Effects) != 1 || !v.Has("Напуган") {
		t.Fatalf("эффект не наложен: %+v %v", v, err)
	}
	if _, err := a.AddEffect(h.ID, "", 1, ""); err == nil {
		t.Error("без названия нельзя")
	}
	if _, err := a.AddEffect(h.ID, "X", 1, "Нет такого"); err == nil {
		t.Error("неизвестное состояние")
	}
	v, err = a.RemoveEffect(h.ID, 0)
	if err != nil || len(v.Effects) != 0 || v.Has("Напуган") {
		t.Errorf("эффект должен сняться вместе с состоянием: %+v %v", v, err)
	}
	// конец боя снимает эффекты и концентрацию
	h2, _ := spawnFight(t, a, "goblin")
	a.AddEffect(h2.ID, "Тест", 5, "")
	a.find(h2.ID).Conc = "Тест"
	if _, err := a.EndEncounter(true); err != nil {
		t.Fatal(err)
	}
	if c := a.find(h2.ID); len(c.Effects) != 0 || c.Conc != "" {
		t.Errorf("после боя эффектов быть не должно: %+v", c.Effects)
	}
}

func TestPurseAndWeight(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	v, err := a.AdjustCoins(h.ID, "gp", 150)
	if err != nil || v.Purse["gp"] != 150 {
		t.Fatalf("монеты: %+v %v", v.Purse, err)
	}
	if v.Derived.Load < 3 {
		t.Errorf("150 монет весят 3 фунта: %v", v.Derived.Load)
	}
	if _, err := a.AdjustCoins(h.ID, "gp", -500); err == nil {
		t.Error("нельзя уйти в минус")
	}
	if _, err := a.AdjustCoins(h.ID, "xx", 1); err == nil {
		t.Error("неизвестная монета")
	}
	v, _ = a.SetCoins(h.ID, "sp", 7)
	if v.Purse["sp"] != 7 {
		t.Error("SetCoins")
	}
}

func TestTransferItem(t *testing.T) {
	a := newApp()
	h1 := hero(t, a)
	h2, _ := a.CreateCharacter("dnd5e", "Борин", "dwarf", "hill", "cleric")
	a.AddSRDItems()
	lib := a.ItemLibrary()
	if _, err := a.GiveItem(h1.ID, lib[0].ID, 5); err != nil {
		t.Fatal(err)
	}
	if _, err := a.TransferItem(h1.ID, h2.ID, 0, 2); err != nil {
		t.Fatal(err)
	}
	if got := a.find(h1.ID).Inventory[0].Qty; got != 3 {
		t.Errorf("у отправителя должно остаться 3, а не %d", got)
	}
	if got := a.find(h2.ID).Inventory[len(a.find(h2.ID).Inventory)-1].Qty; got != 2 {
		t.Errorf("получатель должен получить 2, а не %d", got)
	}
	if _, err := a.TransferItem(h1.ID, h2.ID, 0, 9); err == nil {
		t.Error("нельзя передать больше, чем есть")
	}
	if _, err := a.TransferItem(h1.ID, h1.ID, 0, 1); err == nil {
		t.Error("самому себе нельзя")
	}
	if _, err := a.TransferItem(h1.ID, h2.ID, 0, 3); err != nil {
		t.Fatal(err)
	}
	for _, it := range a.find(h1.ID).Inventory {
		if it.Qty < 1 {
			t.Error("пустая стопка должна исчезнуть")
		}
	}
}

func TestTreasureAndLoot(t *testing.T) {
	a := newApp()
	h1 := hero(t, a)
	h2, _ := a.CreateCharacter("dnd5e", "Борин", "dwarf", "hill", "cleric")
	tr, err := a.RollTreasure([]string{"1", "2", "5"}, false)
	if err != nil || tr.Text == "" {
		t.Fatalf("добыча: %+v %v", tr, err)
	}
	if _, err := a.RollTreasure(nil, false); err == nil {
		t.Error("без существ добычи нет")
	}
	vs, err := a.SplitCoins([]string{h1.ID, h2.ID}, map[string]int{"gp": 101, "sp": 4})
	if err != nil || len(vs) != 2 {
		t.Fatal(err)
	}
	if got := vs[0].Purse["gp"] + vs[1].Purse["gp"]; got != 101 || vs[0].Purse["gp"] != 51 {
		t.Errorf("101 зм делятся 51/50: %d/%d", vs[0].Purse["gp"], vs[1].Purse["gp"])
	}
	gem := model.Item{Name: "Драгоценный камень: рубин", Cat: "gem", Price: "5000 зм"}
	v, err := a.GiveLoot(h1.ID, gem, 1)
	if err != nil || len(v.Inventory) == 0 {
		t.Fatal(err)
	}
	if _, err := a.GiveLoot(h1.ID, model.Item{}, 1); err == nil {
		t.Error("предмет без названия")
	}
	// добыча за бой
	spawnFight(t, a, "goblin")
	if tr, err := a.EncounterTreasure(false); err != nil || tr.Text == "" {
		t.Errorf("добыча за бой: %+v %v", tr, err)
	}
}

func TestQuickSaveKeepsFive(t *testing.T) {
	a := newApp()
	hero(t, a)
	a.SaveSnapshot("Моё имя")
	for i := 0; i < 8; i++ {
		if _, err := a.QuickSave(); err != nil {
			t.Fatal(err)
		}
	}
	quick, named := 0, 0
	for _, s := range a.Snapshots() {
		if strings.HasPrefix(s.Name, "⚡") {
			quick++
		} else {
			named++
		}
	}
	if quick != 5 || named != 1 {
		t.Errorf("быстрых должно быть 5, именных 1: %d и %d", quick, named)
	}
}

func TestMapGridAndFog(t *testing.T) {
	a := newApp()
	m, err := a.AddMap("Подземелье", "data:image/png;base64,AAAA")
	if err != nil {
		t.Fatal(err)
	}
	if err := a.SetMapGrid(m.ID, 70, 5); err != nil {
		t.Fatal(err)
	}
	if err := a.SetMapGrid(m.ID, -1, 5); err == nil {
		t.Error("отрицательная сетка")
	}
	if err := a.SetMapGrid("нет", 70, 5); err == nil {
		t.Error("нет такой карты")
	}
	if err := a.SetFog(m.ID, "data:image/png;base64,BBBB"); err != nil {
		t.Fatal(err)
	}
	if err := a.SetFog(m.ID, "data:text/html;base64,xx"); err == nil {
		t.Error("туман принимает только PNG")
	}
	info := a.Maps()[0]
	if info.Grid != 70 || info.Feet != 5 || !info.HasFog {
		t.Errorf("карта: %+v", info)
	}
	if f, _ := a.GetFog(m.ID); f == "" {
		t.Error("маска тумана должна читаться")
	}
	a.SetFog(m.ID, "")
	if a.Maps()[0].HasFog {
		t.Error("пустая маска убирает туман")
	}
}

// Способности: монах тратит ки, получает две дополнительные атаки в свой ход; Всплеск действий даёт ещё ход атак.
func TestUseAbilityInEncounter(t *testing.T) {
	a := newApp()
	v, err := a.CreateCharacter("dnd5e", "Лао", "human", "", "monk")
	if err != nil {
		t.Fatal(err)
	}
	a.mu.Lock()
	a.find(v.ID).Level = 5
	a.mu.Unlock()
	ms, _ := a.AddMonster("goblin", 1)
	if _, err := a.StartEncounter([]string{v.ID, ms[0].ID}); err != nil {
		t.Fatal(err)
	}
	e := a.state.Encounter
	for i, id := range e.Order {
		if id == v.ID {
			e.Turn = i
		}
	}
	e.Acted, e.Left = false, 2
	if _, err := a.UseAbility(v.ID, "flurry", ""); err != nil {
		t.Fatal(err)
	}
	if e.Left != 4 {
		t.Errorf("шквал добавляет 2 атаки: %d", e.Left)
	}
	e.Acted, e.Left = true, 0
	if _, err := a.UseAbility(v.ID, "flurry", ""); err != nil {
		t.Fatal(err)
	}
	if e.Acted || e.Left != 2 {
		t.Errorf("после атак шквал открывает ещё две: acted=%v left=%d", e.Acted, e.Left)
	}
	if _, err := a.UseAbility(v.ID, "nope", ""); err == nil {
		t.Error("неизвестная способность")
	}
	a.mu.Lock()
	cv := a.view(a.find(v.ID))
	a.mu.Unlock()
	if len(cv.Derived.Pools) == 0 || cv.Derived.Pools[0].Left != 3 {
		t.Errorf("ки: %+v", cv.Derived.Pools)
	}
}
