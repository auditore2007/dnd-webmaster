package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"heroesbook/internal/combat"
	"heroesbook/internal/model"
	"heroesbook/internal/rules"
	"heroesbook/internal/store"
	"heroesbook/internal/treasure"
)

// ---------- бой: способности по цели, реакции, эффекты ----------

// UseSpecialAt — способность существа «по одной цели» (укус с ядом, парализующие когти).
func (a *App) UseSpecialAt(key, targetID string) (*combat.Encounter, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e := a.state.Encounter
	if e == nil {
		return nil, errors.New("бой не начат")
	}
	a.checkpoint()
	before := e.Seq
	if err := e.SpecialAt(a.rng, a.find, key, targetID); err != nil {
		a.rollback()
		return nil, err
	}
	a.logNew(e, before)
	return e, a.persist()
}

// ToggleReaction отмечает реакцию участника потраченной или возвращает её.
func (a *App) ToggleReaction(id string) (*combat.Encounter, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e := a.state.Encounter
	if e == nil {
		return nil, errors.New("бой не начат")
	}
	if err := e.Reaction(id); err != nil {
		return nil, err
	}
	return e, a.persist()
}

// DropConcentration — мастер обрывает концентрацию заклинателя вручную.
func (a *App) DropConcentration(id string) (*combat.Encounter, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	c := a.find(id)
	if c == nil {
		return nil, errNotFound
	}
	a.checkpoint()
	if e := a.state.Encounter; e != nil {
		before := e.Seq
		e.DropConc(c, a.find)
		a.logNew(e, before)
	} else {
		for i := range a.state.Characters {
			o := &a.state.Characters[i]
			for k := len(o.Effects) - 1; k >= 0; k-- {
				if o.Effects[k].Conc && o.Effects[k].Src == id {
					rules.DropEffect(o, k)
				}
			}
		}
		c.Conc = ""
	}
	return a.state.Encounter, a.persist()
}

// AddEffect — мастер вешает на существо свой эффект (горение, проклятие, чужое заклинание) с длительностью в ходах.
func (a *App) AddEffect(id, name string, rounds int, cond string) (CharView, error) {
	return a.with(id, func(c *model.Character, rs rules.Ruleset) (string, error) {
		name = strings.TrimSpace(name)
		switch {
		case name == "" || len([]rune(name)) > 60:
			return "", errors.New("эффект: нужно название до 60 символов")
		case rounds < 0 || rounds > 1000:
			return "", errors.New("эффект: длительность от 0 до 1000 ходов")
		case cond != "" && !slices.Contains(rs.Catalog().Conditions, cond):
			return "", errors.New("эффект: неизвестное состояние")
		case len(c.Effects) >= 20:
			return "", errors.New("эффектов не может быть больше 20")
		}
		rules.AddEffect(c, model.Effect{Name: name, Rounds: rounds, Cond: cond})
		return fmt.Sprintf("%s: эффект «%s»", c.Name, name), nil
	})
}

// RemoveEffect снимает эффект по номеру.
func (a *App) RemoveEffect(id string, idx int) (CharView, error) {
	return a.with(id, func(c *model.Character, rs rules.Ruleset) (string, error) {
		if idx < 0 || idx >= len(c.Effects) {
			return "", errors.New("эффект не найден")
		}
		name := c.Effects[idx].Name
		rules.DropEffect(c, idx)
		return fmt.Sprintf("%s: эффект «%s» снят", c.Name, name), nil
	})
}

// ---------- инвентарь ----------

func validCoin(kind string) bool { return slices.Contains(model.CoinOrder, kind) }

// AdjustCoins меняет кошелёк героя на delta монет (минус — потратить).
func (a *App) AdjustCoins(id, kind string, delta int) (CharView, error) {
	return a.with(id, func(c *model.Character, rs rules.Ruleset) (string, error) {
		if !validCoin(kind) {
			return "", errors.New("неизвестная монета")
		}
		if delta < -1_000_000_000 || delta > 1_000_000_000 {
			return "", errors.New("слишком большая сумма")
		}
		if c.Purse[kind]+delta < 0 {
			return "", fmt.Errorf("не хватает монет: есть %d", c.Purse[kind])
		}
		c.AddCoins(kind, delta)
		return "", nil
	})
}

// SetCoins задаёт число монет вручную.
func (a *App) SetCoins(id, kind string, n int) (CharView, error) {
	return a.with(id, func(c *model.Character, rs rules.Ruleset) (string, error) {
		if !validCoin(kind) || n < 0 || n > 1_000_000_000 {
			return "", errors.New("монеты: число от 0 до миллиарда")
		}
		if c.Purse == nil {
			c.Purse = map[string]int{}
		}
		c.Purse[kind] = n
		return "", nil
	})
}

// TransferItem передаёт qty штук предмета с номером idx от одного героя другому.
func (a *App) TransferItem(fromID, toID string, idx, qty int) ([]CharView, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	from, to := a.find(fromID), a.find(toID)
	switch {
	case from == nil || to == nil:
		return nil, errNotFound
	case fromID == toID:
		return nil, errors.New("выберите другого героя")
	case idx < 0 || idx >= len(from.Inventory):
		return nil, errors.New("предмет не найден")
	}
	it := from.Inventory[idx]
	if qty < 1 || qty > max(1, it.Qty) {
		return nil, fmt.Errorf("можно передать от 1 до %d", max(1, it.Qty))
	}
	a.checkpoint()
	give := it
	give.Qty, give.Equipped = qty, false
	if it.Qty <= qty {
		from.Inventory = slices.Delete(from.Inventory, idx, idx+1)
	} else {
		from.Inventory[idx].Qty -= qty
	}
	addToInventory(to, give)
	a.note("%s передаёт %s ×%d → %s", from.Name, it.Name, qty, to.Name)
	return []CharView{a.view(from), a.view(to)}, a.persist()
}

func addToInventory(c *model.Character, it model.Item) {
	for i := range c.Inventory {
		x := &c.Inventory[i]
		if x.Name == it.Name && x.Weight == it.Weight && x.Price == it.Price && x.Desc == it.Desc {
			x.Qty += max(1, it.Qty)
			return
		}
	}
	it.Qty = max(1, it.Qty)
	c.Inventory = append(c.Inventory, it)
}

// ---------- сокровища ----------

// RollTreasure бросает добычу за существ с указанным уровнем опасности.
func (a *App) RollTreasure(crs []string, hoard bool) (treasure.Treasure, error) {
	if len(crs) == 0 || len(crs) > 100 {
		return treasure.Treasure{}, errors.New("укажите от 1 до 100 существ")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	t := treasure.ForCRs(a.rng, crs, hoard)
	a.note("Добыча: %s", t.Text)
	return t, a.persist()
}

// EncounterTreasure — добыча за всех существ текущего боя.
func (a *App) EncounterTreasure(hoard bool) (treasure.Treasure, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e := a.state.Encounter
	if e == nil {
		return treasure.Treasure{}, errors.New("бой не начат")
	}
	var crs []string
	for _, id := range e.Order {
		if c := a.find(id); c != nil && c.Stat != nil {
			crs = append(crs, c.Stat.CR)
		}
	}
	if len(crs) == 0 {
		return treasure.Treasure{}, errors.New("в бою нет существ")
	}
	t := treasure.ForCRs(a.rng, crs, hoard)
	a.note("Добыча за бой: %s", t.Text)
	return t, a.persist()
}

// GiveLoot кладёт предмет из добычи в инвентарь героя.
func (a *App) GiveLoot(charID string, it model.Item, qty int) (CharView, error) {
	return a.with(charID, func(c *model.Character, rs rules.Ruleset) (string, error) {
		if strings.TrimSpace(it.Name) == "" || len(it.Name) > 200 {
			return "", errors.New("предмет без названия")
		}
		if qty < 1 || qty > 9999 {
			return "", errors.New("количество от 1 до 9999")
		}
		it.ID, it.Qty, it.Equipped = "", qty, false
		addToInventory(c, it)
		return fmt.Sprintf("%s получает %s ×%d", c.Name, it.Name, qty), nil
	})
}

// SplitCoins делит монеты поровну между героями (остаток достаётся первым).
func (a *App) SplitCoins(heroIDs []string, coins map[string]int) ([]CharView, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	var hs []*model.Character
	for _, id := range heroIDs {
		c := a.find(id)
		if c == nil {
			return nil, errNotFound
		}
		if !slices.Contains(hs, c) {
			hs = append(hs, c)
		}
	}
	if len(hs) == 0 {
		return nil, errors.New("выберите хотя бы одного героя")
	}
	for k, v := range coins {
		if !validCoin(k) || v < 0 || v > 1_000_000_000 {
			return nil, errors.New("монеты: неверные значения")
		}
	}
	a.checkpoint()
	var out []CharView
	for k, v := range coins {
		each, rest := v/len(hs), v%len(hs)
		for i, h := range hs {
			h.AddCoins(k, each+b2i(i < rest))
		}
	}
	for _, h := range hs {
		out = append(out, a.view(h))
	}
	a.note("Монеты поделены между героями: %d", len(hs))
	return out, a.persist()
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ---------- сохранение ----------

const quickPrefix = "⚡ "

// QuickSave — быстрое сохранение одной кнопкой (Ctrl+S). Хранится пять последних.
func (a *App) QuickSave() (SnapshotInfo, error) {
	name := quickPrefix + time.Now().Format("02.01 15:04:05")
	if err := a.SaveSnapshot(name); err != nil {
		return SnapshotInfo{}, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	var quick []string
	for _, s := range a.state.Snapshots {
		if strings.HasPrefix(s.Name, quickPrefix) {
			quick = append(quick, s.ID)
		}
	}
	if len(quick) > 5 {
		old := quick[:len(quick)-5]
		a.state.Snapshots = slices.DeleteFunc(a.state.Snapshots, func(s store.Snapshot) bool { return slices.Contains(old, s.ID) })
	}
	s := a.state.Snapshots[len(a.state.Snapshots)-1]
	return SnapshotInfo{s.ID, s.Name, s.Time}, a.persist()
}

// ---------- карта: сетка и туман войны ----------

// SetMapGrid включает сетку (size — пикселей в клетке, 0 выключает) и задаёт футы в клетке.
func (a *App) SetMapGrid(id string, size, feet int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.mapIdx(id)
	switch {
	case i < 0:
		return errors.New("карта не найдена")
	case size < 0 || size > 2000:
		return errors.New("сетка: размер клетки от 0 до 2000 пикселей")
	case feet < 1 || feet > 1000:
		return errors.New("сетка: от 1 до 1000 футов в клетке")
	}
	a.checkpoint()
	a.state.Maps[i].Grid, a.state.Maps[i].Feet = size, feet
	return a.persist()
}

const maxFogChars = 2 << 20

// SetFog сохраняет маску тумана (data:image/png, непрозрачное — скрыто); пустая строка убирает туман.
func (a *App) SetFog(id, data string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.mapIdx(id) < 0 {
		return errors.New("карта не найдена")
	}
	if data != "" && (!strings.HasPrefix(data, "data:image/png") || len(data) > maxFogChars) {
		return errors.New("туман: нужна PNG-маска до 2 МБ")
	}
	bs, ok := a.store.(blobStore)
	if !ok {
		return errors.New("хранилище не поддерживает карты")
	}
	return bs.SaveBlob("fog-"+id, data)
}

func (a *App) GetFog(id string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.mapIdx(id) < 0 {
		return "", errors.New("карта не найдена")
	}
	if bs, ok := a.store.(blobStore); ok {
		return bs.LoadBlob("fog-" + id)
	}
	return "", nil
}
