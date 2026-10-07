package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"

	"heroesbook/internal/bestiary"
	"heroesbook/internal/combat"
	"heroesbook/internal/dice"
	"heroesbook/internal/model"
	"heroesbook/internal/rules"
	"heroesbook/internal/store"
)

// ---------- бой: состав и завершение ----------

// leaveEncounter выводит участника из боя; если остаётся меньше двух, бой заканчивается.
func (a *App) leaveEncounter(id string) {
	e := a.state.Encounter
	if e == nil {
		return
	}
	if _, in := e.Init[id]; !in {
		return
	}
	before := e.Seq
	e.Remove(a.rng, id, a.find)
	if len(e.Order) < 2 {
		a.state.Encounter = nil
		return
	}
	a.logNew(e, before)
}

// AddToEncounter вводит в идущий бой уже созданных героев или существ.
func (a *App) AddToEncounter(ids []string) (*combat.Encounter, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e := a.state.Encounter
	if e == nil {
		return nil, errors.New("бой не начат")
	}
	a.checkpoint()
	before := e.Seq
	if err := e.Add(a.rng, ids, a.find); err != nil {
		a.rollback()
		return nil, err
	}
	a.logNew(e, before)
	return e.Clone(), a.persist()
}

// SpawnInBattle создаёт count существ и сразу вводит их в бой.
func (a *App) SpawnInBattle(monsterID string, count int) (*combat.Encounter, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e := a.state.Encounter
	if e == nil {
		return nil, errors.New("бой не начат")
	}
	a.checkpoint()
	ids, err := a.spawn(monsterID, count)
	if err == nil {
		before := e.Seq
		if err = e.Add(a.rng, ids, a.find); err == nil {
			a.logNew(e, before)
		}
	}
	if err != nil {
		a.restore() // spawn мог уже добавить существ — возвращаем состояние из истории
		return nil, err
	}
	return e.Clone(), a.persist()
}

// RemoveFromEncounter выводит участника из боя (убежал, убран мастером), не удаляя его из игры.
func (a *App) RemoveFromEncounter(id string) (*combat.Encounter, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e := a.state.Encounter
	if e == nil {
		return nil, errors.New("бой не начат")
	}
	if _, in := e.Init[id]; !in {
		return nil, errors.New("участник не в бою")
	}
	a.checkpoint()
	a.leaveEncounter(id)
	return a.state.Encounter.Clone(), a.persist()
}

// EndEncounter заканчивает бой. Погибшие существа убираются всегда; clearMonsters убирает и уцелевших.
// Возвращает число убранных существ.
func (a *App) EndEncounter(clearMonsters bool) (int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.checkpoint()
	removed := map[string]bool{}
	if e := a.state.Encounter; e != nil {
		e.Cleanup(a.find)
		for _, id := range e.Order {
			c := a.find(id)
			if c == nil {
				continue
			}
			if c.Kind == "monster" {
				if c.Dead || c.Down() || clearMonsters {
					removed[id] = true
				}
				continue
			}
			if rs, err := rules.Get(c.Ruleset); err == nil {
				rs.Rest(c, "battle")
			}
		}
		a.note("Бой окончен")
	}
	a.dropCharacters(removed)
	a.state.Encounter = nil
	if len(removed) > 0 {
		a.note("Убрано существ: %d", len(removed))
	}
	return len(removed), a.persist()
}

// ClearMonsters убирает со стола всех существ (когда боя нет).
func (a *App) ClearMonsters() (int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.state.Encounter != nil {
		return 0, errors.New("идёт бой — сначала завершите его")
	}
	a.checkpoint()
	removed := map[string]bool{}
	for _, c := range a.state.Characters {
		if c.Kind == "monster" {
			removed[c.ID] = true
		}
	}
	a.dropCharacters(removed)
	if len(removed) > 0 {
		a.note("Убраны все существа: %d", len(removed))
	}
	return len(removed), a.persist()
}

func (a *App) dropCharacters(ids map[string]bool) {
	if len(ids) == 0 {
		return
	}
	a.state.Characters = slices.DeleteFunc(a.state.Characters, func(c model.Character) bool { return ids[c.ID] })
}

// ---------- собственные существа ----------

// SaveMonster создаёт или обновляет собственное существо.
func (a *App) SaveMonster(m bestiary.Monster) (bestiary.Monster, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := bestiary.Validate(&m); err != nil {
		return bestiary.Monster{}, err
	}
	a.checkpoint()
	if i := slices.IndexFunc(a.state.Monsters, func(x bestiary.Monster) bool { return x.ID == m.ID && m.ID != "" }); i >= 0 {
		a.state.Monsters[i] = m
	} else {
		m.ID = "custom-" + newID()
		a.state.Monsters = append(a.state.Monsters, m)
		a.note("Создано своё существо: %s", m.Name)
	}
	return m, a.persist()
}

func (a *App) DeleteMonster(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	i := slices.IndexFunc(a.state.Monsters, func(x bestiary.Monster) bool { return x.ID == id })
	if i < 0 {
		return errors.New("своё существо не найдено (встроенных удалять нельзя)")
	}
	a.checkpoint()
	a.state.Monsters = slices.Delete(a.state.Monsters, i, i+1)
	return a.persist()
}

// ---------- собственные заклинания ----------

var spellModes = []string{"", "attack", "save", "auto", "heal"}

func checkAuto(s *model.SpellAuto) error {
	if s == nil {
		return nil
	}
	if !slices.Contains(spellModes, s.Mode) {
		return fmt.Errorf("неизвестный режим %q", s.Mode)
	}
	if s.Mode == "" {
		return nil
	}
	if _, err := dice.Parse(s.Dmg); err != nil {
		return err
	}
	if s.Up != "" {
		if _, err := dice.Parse(s.Up); err != nil {
			return err
		}
	}
	if s.Mode == "save" && !slices.Contains([]string{"str", "dex", "con", "int", "wis", "cha"}, s.Save) {
		return errors.New("для спасброска выберите характеристику")
	}
	if s.Mode != "heal" && s.Type != "" {
		ok := false
		for _, t := range rules.DamageTypeIDs() {
			ok = ok || t == s.Type
		}
		if !ok {
			return fmt.Errorf("неизвестный тип урона %q", s.Type)
		}
	}
	return nil
}

func (a *App) SpellLibrary() []model.SpellTemplate {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]model.SpellTemplate{}, a.state.Spells...)
}

func (a *App) SaveSpellTemplate(t model.SpellTemplate) (model.SpellTemplate, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	t.Name = strings.TrimSpace(t.Name)
	if t.Name == "" || len([]rune(t.Name)) > 60 || t.Level < 0 || t.Level > 9 || len([]rune(t.Desc)) > 4000 {
		return model.SpellTemplate{}, errors.New("заклинание: нужно имя до 60 символов, уровень 0–9, описание до 4000 символов")
	}
	if err := checkAuto(t.Auto); err != nil {
		return model.SpellTemplate{}, err
	}
	if t.Auto != nil && t.Auto.Mode == "" {
		t.Auto = nil
	}
	a.checkpoint()
	if i := slices.IndexFunc(a.state.Spells, func(x model.SpellTemplate) bool { return x.ID == t.ID && t.ID != "" }); i >= 0 {
		a.state.Spells[i] = t
	} else {
		t.ID = newID()
		a.state.Spells = append(a.state.Spells, t)
	}
	return t, a.persist()
}

func (a *App) DeleteSpellTemplate(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	i := slices.IndexFunc(a.state.Spells, func(x model.SpellTemplate) bool { return x.ID == id })
	if i < 0 {
		return errors.New("заклинание не найдено")
	}
	a.checkpoint()
	a.state.Spells = slices.Delete(a.state.Spells, i, i+1)
	return a.persist()
}

// GiveSpellTemplate записывает собственное заклинание в книгу героя (с автоматикой).
func (a *App) GiveSpellTemplate(charID, tplID string) (CharView, error) {
	return a.with(charID, func(c *model.Character, rs rules.Ruleset) (string, error) {
		i := slices.IndexFunc(a.state.Spells, func(x model.SpellTemplate) bool { return x.ID == tplID })
		if i < 0 {
			return "", errors.New("заклинание не найдено в библиотеке")
		}
		t := a.state.Spells[i]
		if slices.ContainsFunc(c.Spells, func(s model.Spell) bool { return s.Name == t.Name }) {
			return "", errors.New("это заклинание уже есть у героя")
		}
		sp := model.Spell{Name: t.Name, Level: t.Level, Note: t.Desc}
		if t.Auto != nil {
			au := *t.Auto
			sp.Auto = &au
		}
		c.Spells = append(c.Spells, sp)
		return fmt.Sprintf("%s изучает «%s»", c.Name, t.Name), nil
	})
}

// ---------- карты ----------

type blobStore interface {
	SaveBlob(name, data string) error
	LoadBlob(name string) (string, error)
}

const maxMapChars = 14 << 20

// MapInfo — карта без картинки (картинка запрашивается отдельно).
type MapInfo struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Pins   []store.Pin `json:"pins"`
	Grid   int         `json:"grid"`
	Feet   int         `json:"feet"`
	HasFog bool        `json:"hasFog"`
}

func (a *App) mapInfo(m store.MapMeta) MapInfo {
	has := false
	if bs, ok := a.store.(blobStore); ok {
		if f, err := bs.LoadBlob("fog-" + m.ID); err == nil && f != "" {
			has = true
		}
	}
	return MapInfo{m.ID, m.Name, append([]store.Pin{}, m.Pins...), m.Grid, max(5, m.Feet), has}
}

func (a *App) Maps() []MapInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := []MapInfo{}
	for _, m := range a.state.Maps {
		out = append(out, a.mapInfo(m))
	}
	return out
}

func (a *App) mapIdx(id string) int {
	return slices.IndexFunc(a.state.Maps, func(m store.MapMeta) bool { return m.ID == id })
}

// AddMap добавляет свою карту (data:image/…, до 10 МБ).
func (a *App) AddMap(name, data string) (MapInfo, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 60 {
		return MapInfo{}, errors.New("карта: нужно название до 60 символов")
	}
	if !strings.HasPrefix(data, "data:image/") || len(data) > maxMapChars {
		return MapInfo{}, errors.New("карта: нужна картинка размером не больше 10 МБ")
	}
	if len(a.state.Maps) >= 30 {
		return MapInfo{}, errors.New("карт не может быть больше 30")
	}
	bs, ok := a.store.(blobStore)
	if !ok {
		return MapInfo{}, errors.New("хранилище не поддерживает карты")
	}
	m := store.MapMeta{ID: newID(), Name: name}
	if err := bs.SaveBlob("map-"+m.ID, data); err != nil {
		return MapInfo{}, err
	}
	a.checkpoint()
	a.state.Maps = append(a.state.Maps, m)
	return a.mapInfo(m), a.persist()
}

func (a *App) RenameMap(id, name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	name = strings.TrimSpace(name)
	i := a.mapIdx(id)
	switch {
	case i < 0:
		return errors.New("карта не найдена")
	case name == "" || len([]rune(name)) > 60:
		return errors.New("карта: нужно название до 60 символов")
	}
	a.checkpoint()
	a.state.Maps[i].Name = name
	return a.persist()
}

func (a *App) DeleteMap(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.mapIdx(id)
	if i < 0 {
		return errors.New("карта не найдена")
	}
	a.checkpoint()
	a.state.Maps = slices.Delete(a.state.Maps, i, i+1)
	if bs, ok := a.store.(blobStore); ok {
		_ = bs.SaveBlob("map-"+id, "") // картинка удаляется; отмена вернёт карту без картинки
		_ = bs.SaveBlob("fog-"+id, "")
	}
	return a.persist()
}

func (a *App) GetMapImage(id string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.mapIdx(id) < 0 {
		return "", errors.New("карта не найдена")
	}
	if bs, ok := a.store.(blobStore); ok {
		return bs.LoadBlob("map-" + id)
	}
	return "", nil
}

func (a *App) AddPin(mapID string, x, y float64, text string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	text = strings.TrimSpace(text)
	i := a.mapIdx(mapID)
	switch {
	case i < 0:
		return errors.New("карта не найдена")
	case text == "" || len([]rune(text)) > 60 || math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) || x < 0 || y < 0:
		return errors.New("метка: нужна подпись до 60 символов и координаты на карте")
	}
	a.checkpoint()
	a.state.Maps[i].Pins = append(a.state.Maps[i].Pins, store.Pin{X: x, Y: y, Text: text})
	return a.persist()
}

func (a *App) RemovePin(mapID string, n int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.mapIdx(mapID)
	if i < 0 {
		return errors.New("карта не найдена")
	}
	if n < 0 || n >= len(a.state.Maps[i].Pins) {
		return errors.New("метка не найдена")
	}
	a.checkpoint()
	a.state.Maps[i].Pins = slices.Delete(a.state.Maps[i].Pins, n, n+1)
	return a.persist()
}

// restore возвращает состояние на последнюю контрольную точку и убирает её из истории.
func (a *App) restore() {
	n := len(a.hist)
	if n == 0 {
		return
	}
	var st store.State
	if json.Unmarshal(a.hist[n-1], &st) == nil {
		st.Snapshots = a.state.Snapshots
		a.state = st
	}
	a.hist = a.hist[:n-1]
}
