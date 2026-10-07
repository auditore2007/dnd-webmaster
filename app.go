package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"heroesbook/internal/bestiary"
	"heroesbook/internal/combat"
	"heroesbook/internal/dice"
	"heroesbook/internal/model"
	"heroesbook/internal/rules"
	"heroesbook/internal/srd"
	"heroesbook/internal/store"
)

var errNotFound = errors.New("герой не найден")

// App — фасад для интерфейса: единственная точка входа, все методы под одним мьютексом.
// Каждое изменение сначала откладывает копию состояния (Undo), затем пишет на диск.
type App struct {
	mu    sync.Mutex
	store store.Store
	state store.State
	rng   dice.Roller
	warn  string
	hist  [][]byte
}

func NewApp() *App {
	fs, err := store.Default()
	if err != nil {
		fs = &store.FileStore{Path: "heroes-book-state.json"}
	}
	return &App{store: fs, rng: dice.RNG{}}
}

func (a *App) startup(_ context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	st, err := a.store.Load()
	a.state = st
	if err != nil {
		a.warn = err.Error()
	}
	a.migrate()
	if a.state.SeedVer < seedVersion { // каталог предметов добавляется при первом запуске и при расширении каталога
		a.addSRD()
		a.state.SeedVer = seedVersion
	}
	_ = a.persist()
}

// seedVersion растёт, когда в каталог SRD добавляются новые предметы: недостающие подтянутся один раз.
const seedVersion = 2

// migrate приводит старые сохранения к текущему виду: убирает героев несуществующих систем правил
// и переносит единственную карту со старыми метками в галерею карт.
func (a *App) migrate() {
	kept := a.state.Characters[:0:0]
	dropped := 0
	for _, c := range a.state.Characters {
		if _, err := rules.Get(c.Ruleset); err != nil {
			dropped++
			continue
		}
		kept = append(kept, c)
	}
	if dropped > 0 {
		a.state.Characters = kept
		a.state.Encounter = nil
		a.warn = fmt.Sprintf("Из сохранения убрано героев несуществующей системы правил: %d (мир «Алдоран» удалён)", dropped)
	}
	if bs, ok := a.store.(blobStore); ok && len(a.state.Maps) == 0 {
		old, _ := bs.LoadBlob("map")
		if old != "" || len(a.state.Pins) > 0 {
			m := store.MapMeta{ID: newID(), Name: "Моя карта", Pins: a.state.Pins}
			if old != "" && bs.SaveBlob("map-"+m.ID, old) == nil {
				_ = bs.SaveBlob("map", "")
			}
			a.state.Maps = append(a.state.Maps, m)
			a.state.Pins = nil
		}
	}
}

// addSRD добавляет в библиотеку недостающие предметы SRD (по названию) и возвращает их число.
func (a *App) addSRD() int {
	n := 0
	for _, it := range srd.Items() {
		if slices.ContainsFunc(a.state.Library, func(x model.Item) bool { return x.Name == it.Name }) {
			continue
		}
		it.ID = newID()
		a.state.Library = append(a.state.Library, it)
		n++
	}
	return n
}

// AddSRDItems: кнопка «Добавить SRD-оружие и доспехи».
func (a *App) AddSRDItems() (int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.checkpoint()
	n := a.addSRD()
	return n, a.persist()
}

type CharView struct {
	model.Character
	Derived model.Derived `json:"derived"`
	Msg     string        `json:"msg,omitempty"`
}

type RollView struct {
	Label   string `json:"label"`
	Rolls   []int  `json:"rolls"`
	Kept    int    `json:"kept"`
	Bonus   int    `json:"bonus"`
	Total   int    `json:"total"`
	DC      int    `json:"dc"`
	Success bool   `json:"success"`
	Sides   int    `json:"sides"` // 20 для d20-проверок; для свободного броска — число граней
	Mode    string `json:"mode"`  // adv | dis | ""
	Kind    string `json:"kind"`  // roll | check | save | skill
}

type EncounterView struct {
	Encounter *combat.Encounter   `json:"encounter"`
	Result    *model.AttackResult `json:"result,omitempty"`
}

type SnapshotInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Time string `json:"time"`
}

func (a *App) find(id string) *model.Character {
	for i := range a.state.Characters {
		if a.state.Characters[i].ID == id {
			return &a.state.Characters[i]
		}
	}
	return nil
}

func (a *App) view(c *model.Character) CharView {
	rs, err := rules.Get(c.Ruleset)
	if err != nil {
		return CharView{Character: *c}
	}
	return CharView{Character: *c, Derived: rs.Derive(c)}
}

func (a *App) persist() error { return a.store.Save(a.state) }

func (a *App) checkpoint() {
	s := a.state
	s.Snapshots = nil
	if b, err := json.Marshal(s); err == nil {
		a.hist = append(a.hist, b)
		if len(a.hist) > 50 {
			a.hist = a.hist[1:]
		}
	}
}

func (a *App) rollback() {
	if n := len(a.hist); n > 0 {
		a.hist = a.hist[:n-1]
	}
}

func (a *App) note(f string, v ...any) {
	a.state.Log = append(a.state.Log, time.Now().Format("15:04")+"  "+fmt.Sprintf(f, v...))
	if len(a.state.Log) > 1000 {
		a.state.Log = a.state.Log[len(a.state.Log)-1000:]
	}
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// with выполняет изменение одного героя: поиск, система правил, откат при ошибке, запись в журнал.
// Функция f обязана проверить данные до изменений: при ошибке состояние не восстанавливается.
func (a *App) with(id string, f func(c *model.Character, rs rules.Ruleset) (string, error)) (CharView, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	c := a.find(id)
	if c == nil {
		return CharView{}, errNotFound
	}
	rs, err := rules.Get(c.Ruleset)
	if err != nil {
		return CharView{}, err
	}
	a.checkpoint()
	msg, err := f(c, rs)
	if err != nil {
		a.rollback()
		return CharView{}, err
	}
	if msg != "" {
		a.note("%s", msg)
	}
	v := a.view(c)
	v.Msg = msg
	return v, a.persist()
}

// ---------- справочники ----------

func (a *App) Catalog() []rules.Catalog {
	var out []rules.Catalog
	for _, r := range rules.All() {
		out = append(out, r.Catalog())
	}
	return out
}

// Bestiary — встроенные и собственные существа по возрастанию опасности (CR), затем HP.
func (a *App) Bestiary() []bestiary.Monster {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.bestiary()
}

func (a *App) bestiary() []bestiary.Monster {
	out := append(bestiary.List(), a.state.Monsters...)
	bestiary.Sort(out)
	return out
}

func (a *App) Warning() string { a.mu.Lock(); defer a.mu.Unlock(); return a.warn }

// ---------- герои ----------

func (a *App) Characters() []CharView {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := []CharView{}
	for i := range a.state.Characters {
		out = append(out, a.view(&a.state.Characters[i]))
	}
	return out
}

func (a *App) CreateCharacter(rulesetID, name, raceID, subraceID, classID string) (CharView, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	rs, err := rules.Get(rulesetID)
	if err != nil {
		return CharView{}, err
	}
	c := model.Character{ID: newID(), Name: strings.TrimSpace(name), Ruleset: rulesetID, Race: raceID, Subrace: subraceID, Class: classID}
	if c.Name == "" {
		c.Name = "Новый герой"
	}
	if err := rs.Build(&c); err != nil {
		return CharView{}, err
	}
	a.checkpoint()
	a.state.Characters = append(a.state.Characters, c)
	a.note("Создан герой %s", c.Name)
	return a.view(&a.state.Characters[len(a.state.Characters)-1]), a.persist()
}

// AddMonster добавляет count существ (встроенных или собственных) с нумерацией имён.
func (a *App) AddMonster(id string, count int) ([]CharView, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.checkpoint()
	ids, err := a.spawn(id, count)
	if err != nil {
		a.rollback()
		return nil, err
	}
	var out []CharView
	for _, cid := range ids {
		out = append(out, a.view(a.find(cid)))
	}
	return out, a.persist()
}

// spawn создаёт существ и возвращает их ID. Вызывающий держит мьютекс и делает checkpoint.
func (a *App) spawn(id string, count int) ([]string, error) {
	if count < 1 || count > 20 {
		return nil, errors.New("количество должно быть от 1 до 20")
	}
	var tpl bestiary.Monster
	for _, m := range a.bestiary() {
		if m.ID == id {
			tpl = m
		}
	}
	base := tpl.Name
	if base == "" {
		return nil, fmt.Errorf("неизвестное существо: %q", id)
	}
	next := 1
	for _, c := range a.state.Characters {
		if n, err := strconv.Atoi(strings.TrimPrefix(c.Name, base+" ")); err == nil && strings.HasPrefix(c.Name, base+" ") && n >= next {
			next = n + 1
		}
	}
	var ids []string
	for i := 0; i < count; i++ {
		m := bestiary.Build(tpl, next+i, newID)
		a.state.Characters = append(a.state.Characters, *m)
		ids = append(ids, m.ID)
	}
	a.note("Добавлено: %s ×%d", base, count)
	return ids, nil
}

// UpdateCharacter принимает только редактируемые поля: раса, класс, система, смерть и ID менять нельзя.
func (a *App) UpdateCharacter(in model.Character) (CharView, error) {
	return a.with(in.ID, func(c *model.Character, rs rules.Ruleset) (string, error) {
		if in.Level < 1 || in.Level > 20 {
			return "", errors.New("уровень должен быть от 1 до 20")
		}
		if err := checkClassData(rs, c, &in); err != nil {
			return "", err
		}
		inc := 0
		for k, v := range in.Abilities {
			old, ok := c.Abilities[k]
			if !ok || v < 0 || v > 30 {
				return "", fmt.Errorf("недопустимое значение характеристики %q: %d", k, v)
			}
			inc += max(0, v-old)
		}
		for i := range in.Weapons {
			if _, err := dice.Parse(in.Weapons[i].Dice); err != nil {
				return "", err
			}
			in.Weapons[i].Ability = strings.ToLower(in.Weapons[i].Ability)
		}
		inv := in.Inventory[:0:0]
		for _, it := range in.Inventory {
			if strings.TrimSpace(it.Name) == "" {
				continue
			}
			if err := checkItem(&it); err != nil {
				return "", err
			}
			it.Qty = max(1, it.Qty)
			inv = append(inv, it)
		}
		spells := in.Spells[:0:0]
		for _, s := range in.Spells {
			if strings.TrimSpace(s.Name) == "" || s.Level < 0 || s.Level > 9 || s.Cost < 0 {
				return "", errors.New("заклинание: нужно имя, уровень 0–9 и неотрицательная стоимость")
			}
			if err := checkAuto(s.Auto); err != nil {
				return "", fmt.Errorf("заклинание «%s»: %w", s.Name, err)
			}
			spells = append(spells, s)
		}
		before := rs.Derive(c)
		if n := strings.TrimSpace(in.Name); n != "" {
			c.Name = n
		}
		c.Level, c.XP, c.HP, c.TempHP, c.MP = in.Level, max(0, in.XP), in.HP, in.TempHP, in.MP
		c.ACBonus = max(-10, min(30, in.ACBonus))
		c.Conditions, c.Weapons, c.Inventory, c.Spells, c.Notes = in.Conditions, in.Weapons, inv, spells, in.Notes
		c.SlotsUsed = in.SlotsUsed
		c.Subclass, c.Skills, c.Expert, c.Portrait = in.Subclass, in.Skills, in.Expert, in.Portrait
		c.Points = max(0, c.Points-inc)
		if len(in.Abilities) > 0 {
			c.Abilities = in.Abilities
		}
		d := rs.Derive(c)
		if !c.Dead { // рост максимума (уровень, телосложение, предметы) поднимает и текущие значения
			c.HP += max(0, d.MaxHP-before.MaxHP)
			c.MP += max(0, d.MaxMP-before.MaxMP)
		}
		c.Normalize(d.MaxHP, d.MaxMP)
		for i := range c.SlotsUsed {
			limit := 0
			for _, s := range d.Slots {
				if s.Level == i+1 {
					limit = s.Max
				}
			}
			c.SlotsUsed[i] = max(0, min(c.SlotsUsed[i], limit))
		}
		return "", nil
	})
}

// checkClassData проверяет подкласс, навыки, ссылки на каталог заклинаний и портрет. Уровень заклинания берётся из каталога.
func checkClassData(rs rules.Ruleset, c *model.Character, in *model.Character) error {
	cat := rs.Catalog()
	if in.Subclass != "" {
		ok := false
		for _, cl := range cat.Classes {
			if cl.ID != c.Class {
				continue
			}
			for _, sc := range cl.Subs {
				ok = ok || sc.ID == in.Subclass
			}
		}
		if !ok {
			return fmt.Errorf("подкласс %q не подходит классу героя", in.Subclass)
		}
	}
	known := map[string]bool{}
	for _, sk := range cat.Skills {
		known[sk.ID] = true
	}
	uniq := func(list []string) ([]string, error) {
		out := list[:0:0]
		for _, id := range list {
			if !known[id] {
				return nil, fmt.Errorf("неизвестный навык: %q", id)
			}
			if !slices.Contains(out, id) {
				out = append(out, id)
			}
		}
		return out, nil
	}
	var err error
	if in.Skills, err = uniq(in.Skills); err != nil {
		return err
	}
	if in.Expert, err = uniq(in.Expert); err != nil {
		return err
	}
	for i := range in.Spells {
		ref := in.Spells[i].Ref
		if ref == "" {
			continue
		}
		j := slices.IndexFunc(cat.Spells, func(sd rules.SpellDef) bool { return sd.ID == ref })
		if j < 0 {
			return fmt.Errorf("заклинание %q нет в каталоге", ref)
		}
		in.Spells[i].Level = cat.Spells[j].Level
	}
	if p := in.Portrait; p != "" && (!strings.HasPrefix(p, "data:image/") || len(p) > 150000) {
		return errors.New("портрет: нужна картинка data:image/… не больше 150 КБ")
	}
	return nil
}

func checkItem(it *model.Item) error {
	it.Name, it.Ability = strings.TrimSpace(it.Name), strings.ToLower(it.Ability)
	if it.Weight < 0 || math.IsNaN(it.Weight) || math.IsInf(it.Weight, 0) || it.Weight > 100000 {
		return fmt.Errorf("«%s»: недопустимый вес", it.Name)
	}
	if it.AbilityBonus < -10 || it.AbilityBonus > 10 || it.ACBonus < -5 || it.ACBonus > 10 {
		return fmt.Errorf("«%s»: бонус вне допустимых границ", it.Name)
	}
	if it.ArmorBase < 0 || it.ArmorBase > 30 || !slices.Contains([]string{"", "light", "medium", "heavy"}, it.ArmorType) {
		return fmt.Errorf("«%s»: недопустимые параметры доспеха", it.Name)
	}
	if it.Weapon != nil {
		if _, err := dice.Parse(it.Weapon.Dice); err != nil {
			return err
		}
		it.Weapon.Ability = strings.ToLower(it.Weapon.Ability)
	}
	return nil
}

func (a *App) DeleteCharacter(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	chars := a.state.Characters
	for i := range chars {
		if chars[i].ID == id {
			a.checkpoint()
			a.note("Удалён: %s", chars[i].Name)
			a.state.Characters = append(chars[:i:i], chars[i+1:]...)
			a.leaveEncounter(id)
			return a.persist()
		}
	}
	return errNotFound
}

func (a *App) ApplyDamage(id string, n int) (CharView, error) {
	return a.with(id, func(c *model.Character, rs rules.Ruleset) (string, error) {
		if n < 0 {
			return "", errors.New("значение не может быть отрицательным")
		}
		lines, _ := rules.Strike(rs, c, n, false)
		return strings.Join(append([]string{fmt.Sprintf("%s: −%d HP", c.Name, n)}, lines...), "; "), nil
	})
}

func (a *App) ApplyHeal(id string, n int) (CharView, error) {
	return a.with(id, func(c *model.Character, rs rules.Ruleset) (string, error) {
		if n < 0 {
			return "", errors.New("значение не может быть отрицательным")
		}
		if c.Dead {
			return "", errors.New(c.Name + " погиб: используйте «Воскресить»")
		}
		c.Heal(n, rs.Derive(c).MaxHP)
		return fmt.Sprintf("%s: +%d HP", c.Name, n), nil
	})
}

func (a *App) Revive(id string) (CharView, error) {
	return a.with(id, func(c *model.Character, rs rules.Ruleset) (string, error) {
		c.Dead, c.Stable, c.DeathOk, c.DeathFail = false, false, 0, 0
		c.HP = max(1, c.HP)
		return c.Name + " возвращается к жизни", nil
	})
}

func (a *App) DeathSave(id string) (CharView, error) {
	return a.with(id, func(c *model.Character, rs rules.Ruleset) (string, error) {
		m, err := rs.DeathSave(a.rng, c)
		if err != nil {
			return "", err
		}
		return c.Name + ": " + m, nil
	})
}

func (a *App) AddXP(id string, n int) (CharView, error) {
	return a.with(id, func(c *model.Character, rs rules.Ruleset) (string, error) {
		if n < 1 || n > 1000000 {
			return "", errors.New("опыт должен быть от 1 до 1 000 000")
		}
		if c.Stat != nil {
			return "", errors.New("существа не получают опыт")
		}
		c.XP += n
		msg := fmt.Sprintf("%s: +%d опыта", c.Name, n)
		for c.Level < 20 && c.XP >= rs.XPFor(c.Level+1) {
			before := rs.Derive(c)
			c.Level++
			after := rs.Derive(c)
			if !c.Dead {
				c.HP += max(0, after.MaxHP-before.MaxHP)
				c.MP += max(0, after.MaxMP-before.MaxMP)
			}
			c.Points += rs.LevelPoints(c.Level)
			msg += fmt.Sprintf("; уровень %d!", c.Level)
		}
		return msg, nil
	})
}

func (a *App) Rest(id, kind string) (CharView, error) {
	return a.with(id, func(c *model.Character, rs rules.Ruleset) (string, error) {
		name, ok := restName(rs, kind)
		if !ok {
			return "", fmt.Errorf("неизвестный вид отдыха: %q", kind)
		}
		rs.Rest(c, kind)
		return c.Name + ": " + strings.ToLower(name), nil
	})
}

func restName(rs rules.Ruleset, kind string) (string, bool) {
	for _, r := range rs.Catalog().Rests {
		if r.ID == kind {
			return r.Name, true
		}
	}
	return "", false
}

// RestAll отдыхает всех героев (существа не отдыхают).
func (a *App) RestAll(kind string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.checkpoint()
	for i := range a.state.Characters {
		c := &a.state.Characters[i]
		if rs, err := rules.Get(c.Ruleset); err == nil && c.Stat == nil {
			if _, ok := restName(rs, kind); ok {
				rs.Rest(c, kind)
			}
		}
	}
	a.note("Отдых для всех: %s", kind)
	return a.persist()
}

func (a *App) UseFeature(id, key string) (CharView, error) {
	return a.with(id, func(c *model.Character, rs rules.Ruleset) (string, error) {
		for _, f := range rs.Derive(c).Features {
			if f.Key != key {
				continue
			}
			switch f.Kind {
			case "toggle":
				msg, err := rs.Toggle(c, key)
				if err != nil {
					return "", err
				}
				if msg == "" {
					msg = fmt.Sprintf("%s: %s %s", c.Name, f.Name, map[bool]string{true: "включено", false: "выключено"}[c.Active[key]])
				}
				return msg, nil
			case "once":
				return rs.Act(a.rng, c, key)
			case "ability":
				return a.ability(c, rs, key, "")
			case "special":
				return "", errors.New("эта способность применяется в бою")
			}
			return "", errors.New("эта способность срабатывает сама")
		}
		return "", errors.New("нет такой способности")
	})
}

// UseAbility — способность класса или подкласса (кости превосходства, ки, наложение рук…).
// targetID — союзник или враг, если способности нужна цель. Дополнительные атаки сразу добавляются в текущий ход боя.
func (a *App) UseAbility(id, key, targetID string) (CharView, error) {
	return a.with(id, func(c *model.Character, rs rules.Ruleset) (string, error) {
		return a.ability(c, rs, key, targetID)
	})
}

func (a *App) ability(c *model.Character, rs rules.Ruleset, key, targetID string) (string, error) {
	var t *model.Character
	if targetID != "" {
		if t = a.find(targetID); t == nil {
			return "", errors.New("цель не найдена")
		}
	}
	msg, extra, err := rs.Ability(a.rng, key, c, t)
	if err != nil {
		return "", err
	}
	if e := a.state.Encounter; e != nil {
		if _, in := e.Init[c.ID]; in {
			e.Say(msg)
			if extra > 0 && e.Current() == c.ID {
				e.AddAttacks(extra)
			}
		}
	}
	return msg, nil
}

func (a *App) Cast(id string, idx, slot int) (CharView, error) {
	return a.with(id, func(c *model.Character, rs rules.Ruleset) (string, error) {
		if idx < 0 || idx >= len(c.Spells) {
			return "", errors.New("заклинание не найдено")
		}
		return rs.Cast(c, c.Spells[idx], slot)
	})
}

// ---------- предметы ----------

func (a *App) ItemLibrary() []model.Item {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]model.Item{}, a.state.Library...)
}

func (a *App) SaveLibraryItem(it model.Item) (model.Item, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if strings.TrimSpace(it.Name) == "" {
		return model.Item{}, errors.New("у предмета должно быть название")
	}
	if err := checkItem(&it); err != nil {
		return model.Item{}, err
	}
	it.Qty, it.Equipped = 1, false
	a.checkpoint()
	for i := range a.state.Library {
		if it.ID != "" && a.state.Library[i].ID == it.ID {
			a.state.Library[i] = it
			return it, a.persist()
		}
	}
	it.ID = newID()
	a.state.Library = append(a.state.Library, it)
	return it, a.persist()
}

func (a *App) DeleteLibraryItem(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.checkpoint()
	lib := a.state.Library
	for i := range lib {
		if lib[i].ID == id {
			a.state.Library = append(lib[:i:i], lib[i+1:]...)
			return a.persist()
		}
	}
	a.rollback()
	return errors.New("предмет не найден")
}

func (a *App) GiveItem(charID, libID string, qty int) (CharView, error) {
	return a.with(charID, func(c *model.Character, rs rules.Ruleset) (string, error) {
		if qty < 1 || qty > 9999 {
			return "", errors.New("количество должно быть от 1 до 9999")
		}
		for _, it := range a.state.Library {
			if it.ID != libID {
				continue
			}
			for i := range c.Inventory {
				if c.Inventory[i].Name == it.Name && c.Inventory[i].Weight == it.Weight {
					c.Inventory[i].Qty += qty
					return fmt.Sprintf("%s получает %s ×%d", c.Name, it.Name, qty), nil
				}
			}
			it.Qty = qty
			c.Inventory = append(c.Inventory, it)
			return fmt.Sprintf("%s получает %s ×%d", c.Name, it.Name, qty), nil
		}
		return "", errors.New("предмет не найден в библиотеке")
	})
}

// ---------- броски ----------

// Roll — свободный бросок. Запись d20 понимает преимущество и помеху.
func (a *App) Roll(expr, mode string) (RollView, error) {
	e, err := dice.Parse(expr)
	if err != nil {
		return RollView{}, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	v := RollView{Label: expr, Sides: e.Sides, Kind: "roll"}
	if e.Count == 1 && e.Sides == 20 {
		if m := dice.ParseMode(mode); m != dice.Normal {
			v.Mode = strings.ToLower(mode)
		}
		k, rolls := dice.D20(a.rng, dice.ParseMode(mode))
		v.Rolls, v.Kept, v.Bonus, v.Total = rolls, k, e.Mod, k+e.Mod
	} else {
		r := e.Roll(a.rng, false)
		v.Rolls, v.Kept, v.Bonus, v.Total = r.Rolls, r.Total-r.Mod, r.Mod, r.Total
	}
	a.note("Бросок %s: %v → %d", expr, v.Rolls, v.Total)
	return v, nil
}

// Check — проверка характеристики или спасбросок (kind = "save").
func (a *App) Check(id, ability, kind, mode string, dc int) (RollView, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	c := a.find(id)
	if c == nil {
		return RollView{}, errNotFound
	}
	rs, err := rules.Get(c.Ruleset)
	if err != nil {
		return RollView{}, err
	}
	d := rs.Derive(c)
	bonus, ok := d.Mods[ability]
	label := c.Name + ": " + ability
	if kind == "skill" {
		sk, has := d.Skills[ability]
		if !has {
			return RollView{}, fmt.Errorf("неизвестный навык: %q", ability)
		}
		bonus, ok = sk, true
		for _, def := range rs.Catalog().Skills {
			if def.ID == ability {
				label = c.Name + ": " + def.Name
			}
		}
	}
	if !ok {
		return RollView{}, fmt.Errorf("неизвестная характеристика: %q", ability)
	}
	if s, has := d.Saves[ability]; kind == "save" && has {
		bonus, label = s, c.Name+": спасбросок "+ability
	}
	k, rolls := dice.D20(a.rng, dice.ParseMode(mode))
	v := RollView{Label: label, Rolls: rolls, Kept: k, Bonus: bonus, Total: k + bonus, DC: dc, Sides: 20, Kind: kind}
	if m := dice.ParseMode(mode); m != dice.Normal {
		v.Mode = strings.ToLower(mode)
	}
	v.Success = dc > 0 && v.Total >= dc
	a.note("%s: %v %+d = %d", label, rolls, bonus, v.Total)
	return v, nil
}

// ---------- бой ----------

func (a *App) Encounter() *combat.Encounter {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.state.Encounter
}

// logNew переносит новые строки боя в журнал сессии.
func (a *App) logNew(e *combat.Encounter, before int) {
	if w := e.Winner(a.find); w != e.Won {
		e.Won = w
		switch w {
		case "heroes":
			e.Log = append([]string{"🏆 Победа! Все враги повержены — завершите бой, чтобы убрать их со стола"}, e.Log...)
			e.Seq++
		case "monsters":
			e.Log = append([]string{"☠️ Герои пали"}, e.Log...)
			e.Seq++
		}
	}
	n := min(e.Seq-before, len(e.Log))
	for i := n - 1; i >= 0; i-- {
		a.note("%s", e.Log[i])
	}
}

func (a *App) StartEncounter(ids []string) (*combat.Encounter, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e, err := combat.Start(a.rng, ids, a.find)
	if err != nil {
		return nil, err
	}
	a.checkpoint()
	a.state.Encounter = e
	a.logNew(e, 0)
	return e, a.persist()
}

// Attack: weapon — индекс в списке оружия атакующего (derived.weapons), -1 означает безоружный удар.
func (a *App) Attack(targetID string, weapon int, mode string) (EncounterView, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e := a.state.Encounter
	if e == nil {
		return EncounterView{}, errors.New("бой не начат")
	}
	att := a.find(e.Current())
	if att == nil {
		return EncounterView{}, errors.New("атакующий не найден")
	}
	rs, err := rules.Get(att.Ruleset)
	if err != nil {
		return EncounterView{}, err
	}
	ws := rs.Derive(att).Weapons
	w := model.Unarmed
	if weapon >= len(ws) || weapon < -1 {
		return EncounterView{}, errors.New("неверный номер оружия")
	} else if weapon >= 0 {
		w = ws[weapon]
	}
	a.checkpoint()
	before := e.Seq
	res, err := e.Attack(a.rng, a.find, targetID, w, dice.ParseMode(mode))
	if err != nil {
		a.rollback()
		return EncounterView{}, err
	}
	a.logNew(e, before)
	return EncounterView{e, &res}, a.persist()
}

type MultiView struct {
	Encounter *combat.Encounter    `json:"encounter"`
	Results   []model.AttackResult `json:"results"`
}

// AttackAll выполняет все оставшиеся атаки хода по одной цели: мультиатака существа или дополнительные атаки героя.
// Серия прерывается, если цель погибла; оставшиеся атаки можно направить в другую цель.
func (a *App) AttackAll(targetID string, weapon int, mode string) (MultiView, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e := a.state.Encounter
	if e == nil {
		return MultiView{}, errors.New("бой не начат")
	}
	if e.Acted {
		return MultiView{}, errors.New("в этот ход герой уже действовал")
	}
	att := a.find(e.Current())
	if att == nil {
		return MultiView{}, errors.New("атакующий не найден")
	}
	rs, err := rules.Get(att.Ruleset)
	if err != nil {
		return MultiView{}, err
	}
	ws := rs.Derive(att).Weapons
	var plan []model.Weapon
	if att.Stat != nil {
		for _, i := range att.Stat.Multi {
			if i >= 0 && i < len(ws) {
				plan = append(plan, ws[i])
			}
		}
	}
	if len(plan) == 0 {
		w := model.Unarmed
		if weapon >= len(ws) || weapon < -1 {
			return MultiView{}, errors.New("неверный номер оружия")
		} else if weapon >= 0 {
			w = ws[weapon]
		}
		for i := 0; i < max(1, e.Left); i++ {
			plan = append(plan, w)
		}
	}
	if e.Left > 0 && len(plan) > e.Left {
		plan = plan[len(plan)-e.Left:]
	}
	a.checkpoint()
	before := e.Seq
	out := MultiView{Encounter: e}
	md := dice.ParseMode(mode)
	for _, w := range plan {
		if e.Acted {
			break
		}
		res, err := e.Attack(a.rng, a.find, targetID, w, md)
		if err != nil {
			if len(out.Results) == 0 {
				a.rollback()
				return MultiView{}, err
			}
			break
		}
		out.Results = append(out.Results, res)
		if def := a.find(targetID); def == nil || def.Dead {
			break
		}
	}
	a.logNew(e, before)
	return out, a.persist()
}

// CastSpell: заклинание текущего участника в бою. Ячейка тратится только если все проверки пройдены.
func (a *App) CastSpell(spell, slot int, targetIDs []string, mode string) (*combat.Encounter, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e := a.state.Encounter
	if e == nil {
		return nil, errors.New("бой не начат")
	}
	a.checkpoint()
	before := e.Seq
	if err := e.Cast(a.rng, a.find, spell, slot, targetIDs, dice.ParseMode(mode)); err != nil {
		a.rollback()
		return nil, err
	}
	a.logNew(e, before)
	return e, a.persist()
}

// UseSpecial — особая способность текущего участника (например, дыхание дракона).
func (a *App) UseSpecial(key string) (*combat.Encounter, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e := a.state.Encounter
	if e == nil {
		return nil, errors.New("бой не начат")
	}
	a.checkpoint()
	before := e.Seq
	if err := e.Special(a.rng, a.find, key); err != nil {
		a.rollback()
		return nil, err
	}
	a.logNew(e, before)
	return e, a.persist()
}

func (a *App) NextTurn() (*combat.Encounter, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e := a.state.Encounter
	if e == nil {
		return nil, errors.New("бой не начат")
	}
	a.checkpoint()
	before := e.Seq
	e.Next(a.rng, a.find)
	a.logNew(e, before)
	return e, a.persist()
}

// ---------- журнал, отмена, карта, снимки ----------

func (a *App) Undo() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	n := len(a.hist)
	if n == 0 {
		return errors.New("нечего отменять")
	}
	var s store.State
	if err := json.Unmarshal(a.hist[n-1], &s); err != nil {
		return err
	}
	s.Snapshots = a.state.Snapshots
	a.hist, a.state = a.hist[:n-1], s
	return a.persist()
}

func (a *App) Log() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]string{}, a.state.Log...)
}

// ExportSession сохраняет журнал в текстовый файл (Документы или папка настроек) и возвращает путь.
func (a *App) ExportSession() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	dir, err := os.UserHomeDir()
	if err == nil {
		dir = filepath.Join(dir, "Documents")
	}
	if err != nil || os.MkdirAll(dir, 0o755) != nil {
		if dir, err = os.UserConfigDir(); err != nil {
			return "", err
		}
	}
	path := filepath.Join(dir, "heroes-book-session-"+time.Now().Format("2006-01-02-1504")+".md")
	text := "# Журнал сессии\n\n" + strings.Join(a.state.Log, "\n") + "\n"
	return path, os.WriteFile(path, []byte(text), 0o644)
}

func (a *App) ClearLog() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.checkpoint()
	a.state.Log = nil
	return a.persist()
}

func (a *App) Snapshots() []SnapshotInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := []SnapshotInfo{}
	for _, s := range a.state.Snapshots {
		out = append(out, SnapshotInfo{s.ID, s.Name, s.Time})
	}
	return out
}

func (a *App) SaveSnapshot(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.state
	s.Snapshots = nil
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	now := time.Now().Format("02.01.2006 15:04")
	if name = strings.TrimSpace(name); name == "" {
		name = "Сохранение " + now
	}
	a.state.Snapshots = append(a.state.Snapshots, store.Snapshot{ID: newID(), Name: name, Time: now, Data: b})
	if n := len(a.state.Snapshots); n > 20 {
		a.state.Snapshots = a.state.Snapshots[n-20:]
	}
	return a.persist()
}

// LoadSnapshot заменяет игру снимком; текущее состояние попадает в историю, так что загрузку можно отменить.
func (a *App) LoadSnapshot(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, sn := range a.state.Snapshots {
		if sn.ID != id {
			continue
		}
		var s store.State
		if err := json.Unmarshal(sn.Data, &s); err != nil {
			return err
		}
		a.checkpoint()
		s.Snapshots = a.state.Snapshots
		a.state = s
		return a.persist()
	}
	return errors.New("сохранение не найдено")
}

func (a *App) DeleteSnapshot(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	sn := a.state.Snapshots
	for i := range sn {
		if sn[i].ID == id {
			a.state.Snapshots = append(sn[:i:i], sn[i+1:]...)
			return a.persist()
		}
	}
	return errors.New("сохранение не найдено")
}
