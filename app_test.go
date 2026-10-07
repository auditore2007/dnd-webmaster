package main

import (
	"strings"
	"testing"

	"heroesbook/internal/bestiary"
	"heroesbook/internal/model"
	"heroesbook/internal/store"
)

type memStore struct {
	s     store.State
	blobs map[string]string
}

func (m *memStore) SaveBlob(name, data string) error {
	if m.blobs == nil {
		m.blobs = map[string]string{}
	}
	if data == "" {
		delete(m.blobs, name)
	} else {
		m.blobs[name] = data
	}
	return nil
}
func (m *memStore) LoadBlob(name string) (string, error) { return m.blobs[name], nil }

func (m *memStore) Load() (store.State, error) { return m.s, nil }
func (m *memStore) Save(s store.State) error   { m.s = s; return nil }

type fixed struct{ v int }

func (f fixed) Intn(n int) int { return f.v % n }

func newApp() *App { return &App{store: &memStore{}, rng: fixed{9}} }

func hero(t *testing.T, a *App) CharView {
	t.Helper()
	v, err := a.CreateCharacter("dnd5e", "Лира", "human", "", "fighter")
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestUndoRestoresState(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	hp := h.HP
	if _, err := a.ApplyDamage(h.ID, 4); err != nil {
		t.Fatal(err)
	}
	if err := a.Undo(); err != nil {
		t.Fatal(err)
	}
	if got := a.Characters()[0].HP; got != hp {
		t.Errorf("после отмены HP=%d, ожидалось %d", got, hp)
	}
	a.Undo() // отмена создания героя
	if len(a.Characters()) != 0 {
		t.Error("отмена должна убрать героя")
	}
	if err := a.Undo(); err == nil {
		t.Error("пустая история — ошибка")
	}
}

func TestFailedOperationDoesNotPolluteHistory(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	n := len(a.hist)
	if _, err := a.ApplyDamage(h.ID, -5); err == nil {
		t.Fatal("отрицательный урон должен отклоняться")
	}
	if len(a.hist) != n {
		t.Error("неудачная операция оставила запись в истории")
	}
}

func TestUpdateValidation(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	bad := h.Character
	bad.Level = 21
	if _, err := a.UpdateCharacter(bad); err == nil {
		t.Error("уровень 21")
	}
	bad = h.Character
	bad.Abilities = map[string]int{"str": 99}
	if _, err := a.UpdateCharacter(bad); err == nil {
		t.Error("характеристика 99")
	}
	bad = h.Character
	bad.Abilities = map[string]int{"bogus": 10}
	if _, err := a.UpdateCharacter(bad); err == nil {
		t.Error("неизвестная характеристика")
	}
	bad = h.Character
	bad.Weapons = []model.Weapon{{Name: "X", Dice: "abc"}}
	if _, err := a.UpdateCharacter(bad); err == nil {
		t.Error("неверные кубики оружия")
	}
	bad = h.Character
	bad.Race, bad.Dead, bad.Name = "elf", true, "Новое"
	v, err := a.UpdateCharacter(bad)
	if err != nil || v.Race != "human" || v.Dead || v.Name != "Новое" {
		t.Errorf("раса и смерть не редактируются: %+v %v", v, err)
	}
}

func TestLevelUpByXP(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	v, err := a.AddXP(h.ID, 900)
	if err != nil || v.Level != 3 || v.Derived.MaxHP <= h.Derived.MaxHP || v.HP != v.Derived.MaxHP {
		t.Errorf("900 опыта = 3 уровень с ростом HP: %+v %v", v.Level, err)
	}
	if v, _ = a.AddXP(h.ID, 2000); v.Level != 4 || v.Points != 2 {
		t.Errorf("4 уровень даёт 2 очка: ур. %d, очки %d", v.Level, v.Points)
	}
	if _, err := a.AddXP(h.ID, 0); err == nil {
		t.Error("нулевой опыт")
	}
}

func TestMonstersAndItems(t *testing.T) {
	a := newApp()
	ms, err := a.AddMonster("goblin", 2)
	if err != nil || len(ms) != 2 || ms[1].Name != "Гоблин 2" {
		t.Fatalf("%v %v", ms, err)
	}
	more, _ := a.AddMonster("goblin", 1)
	if more[0].Name != "Гоблин 3" {
		t.Errorf("нумерация продолжается: %s", more[0].Name)
	}
	if _, err := a.AddMonster("goblin", 0); err == nil {
		t.Error("ноль существ")
	}
	if _, err := a.AddXP(ms[0].ID, 100); err == nil {
		t.Error("существа не получают опыт")
	}
	h := hero(t, a)
	it, err := a.SaveLibraryItem(model.Item{Name: "Щит", Weight: 6, ACBonus: 2})
	if err != nil {
		t.Fatal(err)
	}
	a.GiveItem(h.ID, it.ID, 1)
	v, _ := a.GiveItem(h.ID, it.ID, 2)
	if len(v.Inventory) != 1 || v.Inventory[0].Qty != 3 || v.Derived.Load != 18 {
		t.Errorf("одинаковые предметы складываются: %+v", v.Inventory)
	}
	if _, err := a.SaveLibraryItem(model.Item{Name: "Плохой", Weight: -1}); err == nil {
		t.Error("отрицательный вес")
	}
	if _, err := a.GiveItem(h.ID, "нет", 1); err == nil {
		t.Error("предмета нет в библиотеке")
	}
}

func TestSnapshots(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	if err := a.SaveSnapshot("до боя"); err != nil {
		t.Fatal(err)
	}
	a.DeleteCharacter(h.ID)
	if err := a.LoadSnapshot(a.Snapshots()[0].ID); err != nil || len(a.Characters()) != 1 {
		t.Fatalf("снимок возвращает героя: %v", err)
	}
	if len(a.Snapshots()) != 1 {
		t.Error("список снимков должен сохраняться при загрузке")
	}
}

func TestEncounterFlowAndLog(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	ms, _ := a.AddMonster("goblin", 1)
	e, err := a.StartEncounter([]string{h.ID, ms[0].ID})
	if err != nil || len(e.Order) != 2 {
		t.Fatal(err)
	}
	if len(a.Log()) == 0 {
		t.Error("начало боя должно попасть в журнал")
	}
	foe := ms[0].ID
	if e.Current() == ms[0].ID {
		foe = h.ID
	}
	v, err := a.Attack(foe, -1, "")
	if err != nil || v.Result == nil {
		t.Fatalf("%v", err)
	}
	if _, err := a.Attack(foe, 7, ""); err == nil {
		t.Error("неверный номер оружия")
	}
	if _, err := a.EndEncounter(false); err != nil || a.Encounter() != nil {
		t.Error("конец боя")
	}
}

func TestClassDataValidation(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	in := h.Character
	in.Subclass = "champion"
	in.Skills, in.Expert = []string{"athletics", "athletics", "stealth"}, []string{"stealth"}
	v, err := a.UpdateCharacter(in)
	if err != nil {
		t.Fatal(err)
	}
	if v.Subclass != "champion" || len(v.Skills) != 2 {
		t.Errorf("дубликаты навыков убираются: %+v", v.Skills)
	}
	for name, mut := range map[string]func(c *model.Character){
		"подкласс другого класса": func(c *model.Character) { c.Subclass = "thief" },
		"неизвестный навык":       func(c *model.Character) { c.Skills = []string{"cooking"} },
		"неизвестное заклинание":  func(c *model.Character) { c.Spells = []model.Spell{{Name: "X", Ref: "nope"}} },
		"портрет не картинка":     func(c *model.Character) { c.Portrait = "http://evil.example/x.png" },
		"портрет слишком большой": func(c *model.Character) { c.Portrait = "data:image/png;base64," + strings.Repeat("A", 150000) },
	} {
		bad := v.Character
		mut(&bad)
		if _, err := a.UpdateCharacter(bad); err == nil {
			t.Errorf("%s: ожидалась ошибка", name)
		}
	}
	// уровень заклинания берётся из каталога
	sp := v.Character
	sp.Spells = []model.Spell{{Name: "Шар", Level: 9, Ref: "fireball"}}
	if v2, err := a.UpdateCharacter(sp); err != nil || v2.Spells[0].Level != 3 {
		t.Errorf("уровень из каталога: %+v %v", v2.Spells, err)
	}
}

func TestSkillCheckAndFeatures(t *testing.T) {
	a := newApp()
	b, err := a.CreateCharacter("dnd5e", "Торг", "halforc", "", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	in := b.Character
	in.Skills = []string{"athletics"}
	b, _ = a.UpdateCharacter(in)
	r, err := a.Check(b.ID, "athletics", "skill", "", 15)
	if err != nil || r.Bonus != b.Derived.Skills["athletics"] || r.Bonus != 1+2 {
		t.Errorf("проверка навыка: %+v %v", r, err)
	}
	if _, err := a.Check(b.ID, "cooking", "skill", "", 0); err == nil {
		t.Error("неизвестный навык")
	}
	v, err := a.UseFeature(b.ID, "rage")
	if err != nil || !v.Active["rage"] || !strings.Contains(v.Msg, "ярость") {
		t.Fatalf("ярость: %+v %v", v.Msg, err)
	}
	if v, err = a.UseFeature(b.ID, "rage"); err != nil || v.Active["rage"] {
		t.Errorf("повторное нажатие выключает ярость: %v", err)
	}
	f := hero(t, a)
	if _, err := a.ApplyDamage(f.ID, 5); err != nil {
		t.Fatal(err)
	}
	if v, err := a.UseFeature(f.ID, "secondwind"); err != nil || v.HP <= f.HP-5 {
		t.Errorf("второе дыхание: %v HP=%d", err, v.HP)
	}
	if _, err := a.UseFeature(f.ID, "secondwind"); err == nil {
		t.Error("второе дыхание только раз")
	}
}

func TestMultiattackAndSpellInApp(t *testing.T) {
	a := newApp()
	a.rng = fixed{14}
	f := hero(t, a)
	mons, err := a.AddMonster("owlbear", 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.StartEncounter([]string{f.ID, mons[0].ID}); err != nil {
		t.Fatal(err)
	}
	e := a.Encounter()
	for e.Current() != mons[0].ID { // ждём ход совомедведя
		if _, err := a.NextTurn(); err != nil {
			t.Fatal(err)
		}
		e = a.Encounter()
	}
	if e.Left != 2 {
		t.Fatalf("у совомедведя две атаки, осталось %d", e.Left)
	}
	mv, err := a.AttackAll(f.ID, 0, "")
	if err != nil || len(mv.Results) != 2 || !mv.Encounter.Acted {
		t.Fatalf("мультиатака: %d результатов, %v", len(mv.Results), err)
	}
	if _, err := a.AttackAll(f.ID, 0, ""); err == nil {
		t.Error("ход уже потрачен")
	}
	if _, err := a.CastSpell(0, 1, []string{f.ID}, ""); err == nil {
		t.Error("у совомедведя нет заклинаний, а ход потрачен")
	}
}

func TestSRDLibrary(t *testing.T) {
	a := newApp()
	a.startup(nil)
	if n := len(a.ItemLibrary()); n < 400 {
		t.Errorf("SRD-предметы добавлены при первом запуске, в библиотеке %d", n)
	}
	lib := a.ItemLibrary()
	if err := a.DeleteLibraryItem(lib[0].ID); err != nil {
		t.Fatal(err)
	}
	a.startup(nil)
	if len(a.ItemLibrary()) != len(lib)-1 {
		t.Error("удалённый SRD-предмет не должен возвращаться при перезапуске")
	}
	if n, err := a.AddSRDItems(); err != nil || n != 1 {
		t.Errorf("кнопка возвращает недостающее: %d %v", n, err)
	}
	if n, _ := a.AddSRDItems(); n != 0 {
		t.Error("повторное добавление ничего не дублирует")
	}
}

func TestMaps(t *testing.T) {
	a := newApp()
	if _, err := a.AddMap("Плохая", "javascript:alert(1)"); err == nil {
		t.Error("карта только data:image")
	}
	if _, err := a.AddMap(" ", "data:image/png;base64,AAAA"); err == nil {
		t.Error("нужно название")
	}
	m, err := a.AddMap("Королевство", "data:image/png;base64,AAAA")
	if err != nil {
		t.Fatal(err)
	}
	m2, _ := a.AddMap("Подземелье", "data:image/png;base64,BBBB")
	if got, _ := a.GetMapImage(m.ID); got != "data:image/png;base64,AAAA" {
		t.Errorf("картинка: %q", got)
	}
	if a.AddPin(m.ID, 10, 20, " ") == nil || a.AddPin(m.ID, -1, 5, "x") == nil || a.AddPin("нет", 1, 1, "x") == nil {
		t.Error("плохие метки принимаются")
	}
	if a.AddPin(m.ID, 10, 20, "Город") != nil || a.RemovePin(m.ID, 5) == nil {
		t.Error("метки")
	}
	if ms := a.Maps(); len(ms) != 2 || len(ms[0].Pins) != 1 || len(ms[1].Pins) != 0 {
		t.Errorf("метки принадлежат своей карте: %+v", ms)
	}
	if a.RenameMap(m2.ID, "Бездна") != nil || a.Maps()[1].Name != "Бездна" {
		t.Error("переименование")
	}
	if a.DeleteMap(m.ID) != nil || len(a.Maps()) != 1 {
		t.Error("удаление")
	}
	if _, err := a.GetMapImage(m.ID); err == nil {
		t.Error("картинка удалённой карты")
	}
}

func TestMigrationFromV3(t *testing.T) {
	ms := &memStore{s: store.State{
		Pins:       []store.Pin{{X: 1, Y: 2, Text: "Старая метка"}},
		Characters: []model.Character{{ID: "x", Name: "Грош", Ruleset: "aldoran"}, {ID: "y", Name: "Лира", Ruleset: "dnd5e", Race: "human", Class: "fighter", Level: 1, Abilities: map[string]int{"str": 10}}},
	}}
	ms.SaveBlob("map", "data:image/png;base64,OLD")
	a := &App{store: ms, rng: fixed{9}}
	a.startup(nil)
	if len(a.Characters()) != 1 || a.Warning() == "" {
		t.Errorf("герои Алдорана убираются с предупреждением: %d %q", len(a.Characters()), a.Warning())
	}
	maps := a.Maps()
	if len(maps) != 1 || len(maps[0].Pins) != 1 {
		t.Fatalf("старая карта переехала в галерею: %+v", maps)
	}
	if img, _ := a.GetMapImage(maps[0].ID); img != "data:image/png;base64,OLD" {
		t.Error("картинка старой карты потеряна")
	}
	if old, _ := ms.LoadBlob("map"); old != "" {
		t.Error("старый blob должен быть убран")
	}
	a.startup(nil)
	if len(a.Maps()) != 1 {
		t.Error("повторный запуск не дублирует карту")
	}
}

func TestBattleTeamsAndCleanup(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	ms, _ := a.AddMonster("goblin", 2)
	e, err := a.StartEncounter([]string{h.ID, ms[0].ID, ms[1].ID})
	if err != nil {
		t.Fatal(err)
	}
	foes := e.Foes(a.find, h.ID)
	if len(foes) != 2 {
		t.Fatalf("у героя два врага-гоблина: %d", len(foes))
	}
	if f := e.Foes(a.find, ms[0].ID); len(f) != 1 || f[0].ID != h.ID {
		t.Errorf("гоблин не считает своих врагами: %+v", f)
	}
	// третий гоблин приходит на помощь прямо в бою
	e2, err := a.SpawnInBattle("goblin", 1)
	if err != nil || len(e2.Order) != 4 {
		t.Fatalf("подкрепление: %v", err)
	}
	cur := e2.Current()
	for _, id := range e2.Order {
		if _, ok := e2.Init[id]; !ok {
			t.Error("инициатива нового участника не брошена")
		}
	}
	if e2.Current() != cur {
		t.Error("новый участник не должен красть ход")
	}
	// убиваем одного гоблина и уводим другого из боя
	dead := ms[0].ID
	a.ApplyDamage(dead, 1000)
	if _, err := a.RemoveFromEncounter(ms[1].ID); err != nil {
		t.Fatal(err)
	}
	if len(a.Encounter().Order) != 3 {
		t.Errorf("после ухода участника: %v", a.Encounter().Order)
	}
	n, err := a.EndEncounter(false)
	if err != nil || n != 1 {
		t.Fatalf("после боя убираются только погибшие: %d %v", n, err)
	}
	left := 0
	for _, c := range a.Characters() {
		if c.Kind == "monster" {
			left++
			if c.ID == dead {
				t.Error("погибший враг остался на столе")
			}
		}
	}
	if left != 2 {
		t.Errorf("живые враги остаются: %d", left)
	}
	if k, err := a.ClearMonsters(); err != nil || k != 2 {
		t.Errorf("очистка врагов: %d %v", k, err)
	}
	if len(a.Characters()) != 1 {
		t.Error("герой должен остаться")
	}
}

func TestEndEncounterClearsAllEnemies(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	ms, _ := a.AddMonster("wolf", 2)
	a.StartEncounter([]string{h.ID, ms[0].ID, ms[1].ID})
	if n, _ := a.EndEncounter(true); n != 2 {
		t.Errorf("убраны оба волка: %d", n)
	}
	if _, err := a.ClearMonsters(); err != nil {
		t.Error(err)
	}
}

func TestVictoryDetected(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	ms, _ := a.AddMonster("rat", 1)
	a.StartEncounter([]string{h.ID, ms[0].ID})
	a.ApplyDamage(ms[0].ID, 100)
	if _, err := a.NextTurn(); err != nil {
		t.Fatal(err)
	}
	if a.Encounter().Won != "heroes" {
		t.Errorf("победа героев: %q", a.Encounter().Won)
	}
}

func TestCustomMonsterAndSpell(t *testing.T) {
	a := newApp()
	h := hero(t, a)
	m, err := a.SaveMonster(bestiary.Monster{Name: "Мой дракончик", CR: "2", Kind: "dragon", AC: 14, HP: 40, Attack: 5, Speed: 30,
		Abilities: [6]int{12, 12, 12, 8, 8, 8}, Weapons: []model.Weapon{{Name: "Укус", Dice: "1d8+2", Type: "piercing"}}})
	if err != nil || !m.Custom || m.ID == "" {
		t.Fatalf("%v %+v", err, m)
	}
	bs := a.Bestiary()
	idx := -1
	for i, b := range bs {
		if b.ID == m.ID {
			idx = i
		}
	}
	if idx < 0 || bestiary.CRValue(bs[idx].CR) < bestiary.CRValue(bs[max(0, idx-1)].CR) {
		t.Error("своё существо попало в общий порядок по CR")
	}
	if cs, err := a.AddMonster(m.ID, 2); err != nil || len(cs) != 2 || cs[0].Stat.MaxHP != 40 {
		t.Errorf("создание из шаблона: %v", err)
	}
	if _, err := a.SaveMonster(bestiary.Monster{Name: "Плохой"}); err == nil {
		t.Error("пустое существо")
	}
	if a.DeleteMonster("goblin") == nil {
		t.Error("встроенных удалять нельзя")
	}
	if a.DeleteMonster(m.ID) != nil {
		t.Error("своё удаляется")
	}

	if _, err := a.SaveSpellTemplate(model.SpellTemplate{Name: "Гнев", Level: 2, Auto: &model.SpellAuto{Mode: "save", Dmg: "xx", Save: "dex"}}); err == nil {
		t.Error("плохие кубики заклинания")
	}
	if _, err := a.SaveSpellTemplate(model.SpellTemplate{Name: "Гнев", Level: 2, Auto: &model.SpellAuto{Mode: "save", Dmg: "4d6", Type: "fire", Save: "foo"}}); err == nil {
		t.Error("плохая характеристика")
	}
	t1, err := a.SaveSpellTemplate(model.SpellTemplate{Name: "Гнев", Level: 2, Auto: &model.SpellAuto{Mode: "save", Dmg: "4d6", Type: "fire", Save: "dex", Half: true}})
	if err != nil {
		t.Fatal(err)
	}
	v, err := a.GiveSpellTemplate(h.ID, t1.ID)
	if err != nil || len(v.Spells) != 1 || v.Spells[0].Auto == nil {
		t.Fatalf("%v %+v", err, v.Spells)
	}
	if _, err := a.GiveSpellTemplate(h.ID, t1.ID); err == nil {
		t.Error("дубликат заклинания")
	}
	if a.DeleteSpellTemplate(t1.ID) != nil || len(a.SpellLibrary()) != 0 {
		t.Error("удаление шаблона")
	}
}
