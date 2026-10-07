package main

import (
	"encoding/json"
	"slices"
	"sync"
	"testing"
)

// Wails вызывает методы App в отдельных горутинах и сериализует ответ уже после выхода из метода,
// то есть без мьютекса. Ответы не должны ссылаться на живое состояние – иначе гонка
// (а для map – аварийное завершение «concurrent map read and map write»). Запускать с -race.
func TestResponsesDoNotShareLiveState(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	mons, err := a.AddMonster("goblin", 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.StartEncounter([]string{h.ID, mons[0].ID}); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				_, _ = a.AdjustCoins(h.ID, "gp", 1)
				_, _ = a.ToggleReaction(h.ID)
				_, _ = a.ApplyDamage(mons[0].ID, 0)
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				v, _ := a.AdjustCoins(h.ID, "sp", 1)
				_, _ = json.Marshal(v)
				_, _ = json.Marshal(a.Characters())
				_, _ = json.Marshal(a.Encounter())
				e, _ := a.ToggleReaction(mons[0].ID)
				_, _ = json.Marshal(e)
			}
		}()
	}
	wg.Wait()
}

func TestCloneIsDeep(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	if _, err := a.AdjustCoins(h.ID, "gp", 5); err != nil {
		t.Fatal(err)
	}
	v := a.Characters()[0]
	v.Purse["gp"] = 999
	v.Abilities["str"] = 1
	if got := a.Characters()[0]; got.Purse["gp"] != 5 || got.Abilities["str"] == 1 {
		t.Errorf("изменение ответа задело состояние: кошелёк %d, сила %d", got.Purse["gp"], got.Abilities["str"])
	}
}

// Неудачное добавление в бой не должно оставлять участников «полузаписанными»: иначе их потом нельзя ввести в бой.
func TestFailedAddToEncounterLeavesNoTrace(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	mons, err := a.AddMonster("goblin", 2)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.StartEncounter([]string{h.ID, mons[0].ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddToEncounter([]string{mons[1].ID, "нет-такого"}); err == nil {
		t.Fatal("ожидалась ошибка для неизвестного участника")
	}
	e, err := a.AddToEncounter([]string{mons[1].ID})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(e.Order, mons[1].ID) {
		t.Errorf("гоблин не вошёл в бой после неудачной попытки: %v", e.Order)
	}
}

func TestMonsterTurnThroughApp(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	mons, err := a.AddMonster("goblin", 1)
	if err != nil {
		t.Fatal(err)
	}
	e, err := a.StartEncounter([]string{h.ID, mons[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	if e.Current() != mons[0].ID {
		if e, err = a.NextTurn(); err != nil {
			t.Fatal(err)
		}
	}
	v, err := a.MonsterTurn()
	if err != nil {
		t.Fatal(err)
	}
	if v.Encounter.Current() == mons[0].ID || v.Actor == "" {
		t.Errorf("после хода существа очередь должна перейти дальше: %+v", v)
	}
	if _, err := a.SetEncounterOptions(false, false, true); err != nil {
		t.Fatal(err)
	}
	if got := a.Encounter(); got.Auto || got.Morale || !got.Lair {
		t.Errorf("настройки боя не сохранились: %+v", got)
	}
}
