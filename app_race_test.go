package main

import (
	"encoding/json"
	"sync"
	"testing"
)

// Wails вызывает методы App в отдельных горутинах и сериализует ответ уже после выхода из метода,
// то есть без мьютекса. Ответы не должны ссылаться на живое состояние — иначе гонка
// (а для map — аварийное завершение «concurrent map read and map write»). Запускать с -race.
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
