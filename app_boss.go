package main

import (
	"errors"

	"heroesbook/internal/combat"
	"heroesbook/internal/model"
)

// ---------- бой: автоход существ, легендарные действия, настройки ----------

// MonsterTurnView – ход существа: броски его атак (для кубиков) и бой после хода.
type MonsterTurnView struct {
	Encounter *combat.Encounter    `json:"encounter"`
	Actor     string               `json:"actor"`
	Results   []model.AttackResult `json:"results"`
	Targets   []string             `json:"targets"`
}

// MonsterTurn – текущее существо делает свой ход само (способность или атаки), затем ход переходит дальше.
func (a *App) MonsterTurn() (MonsterTurnView, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e := a.state.Encounter
	if e == nil {
		return MonsterTurnView{}, errors.New("бой не начат")
	}
	if e.Won != "" {
		return MonsterTurnView{}, errors.New("бой уже решён")
	}
	actor := ""
	if c := a.find(e.Current()); c != nil {
		actor = c.Name
	}
	a.checkpoint()
	before := e.Seq
	res, targets, err := e.MonsterTurn(a.rng, a.find)
	if err != nil {
		a.rollback()
		return MonsterTurnView{}, err
	}
	a.logNew(e, before)
	return MonsterTurnView{Encounter: e.Clone(), Actor: actor, Results: res, Targets: targets}, a.persist()
}

// LegendaryView – бой после легендарного действия и бросок атаки, если это была атака оружием.
type LegendaryView struct {
	Encounter *combat.Encounter   `json:"encounter"`
	Result    *model.AttackResult `json:"result,omitempty"`
}

// LegendaryAction – мастер тратит легендарное действие босса между ходами других участников.
func (a *App) LegendaryAction(actorID, key, targetID string) (LegendaryView, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e := a.state.Encounter
	switch {
	case e == nil:
		return LegendaryView{}, errors.New("бой не начат")
	case e.Won != "":
		return LegendaryView{}, errors.New("бой уже решён")
	}
	a.checkpoint()
	before := e.Seq
	res, err := e.Legendary(a.rng, a.find, actorID, key, targetID)
	if err != nil {
		a.rollback()
		return LegendaryView{}, err
	}
	a.logNew(e, before)
	return LegendaryView{Encounter: e.Clone(), Result: res}, a.persist()
}

// SetEncounterOptions: auto – существа ходят сами, morale – враги бегут и сдаются, lair – бой идёт в логове босса.
func (a *App) SetEncounterOptions(auto, morale, lair bool) (*combat.Encounter, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e := a.state.Encounter
	if e == nil {
		return nil, errors.New("бой не начат")
	}
	a.checkpoint()
	e.Auto, e.Morale = auto, morale
	if lair && !e.Lair {
		e.LairRound = e.Round // логово начнёт действовать со следующего раунда
	}
	e.Lair = lair
	return e.Clone(), a.persist()
}
