package rules

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"heroesbook/internal/dice"
	"heroesbook/internal/model"
)

// Способности классов и подклассов. Каждая — строка таблицы abTable; движок общий:
//   ресурс (ки, кости превосходства…) → тратится при применении, восстанавливается на отдыхе;
//   надбавка к удару (rider) → срабатывает на следующем попадании;
//   лечение, временные HP, эффект на себя/союзника/врага, дополнительные атаки, возврат ячеек;
//   особая способность по области или цели (изгнание нежити, сияющий рассвет).
// Правила упрощены: описание у каждой способности говорит, что именно считается.

// ctx — всё, что нужно, чтобы вычислить число или кубики способности.
type ctx struct {
	C    *model.Character
	Lvl  int
	Prof int
	M    map[string]int
}

// val вычисляет запись кубиков или число («2d8», «5», «-1d6»).
type val func(x ctx) string

func k(s string) val { return func(ctx) string { return s } }
func num(f func(x ctx) int) val {
	return func(x ctx) string { return strconv.Itoa(f(x)) }
}

// ступенчатые кубики, растущие с уровнем
func stepDie(lvl int, steps ...[2]int) string {
	sides := steps[0][1]
	for _, s := range steps {
		if lvl >= s[0] {
			sides = s[1]
		}
	}
	return "d" + strconv.Itoa(sides)
}

func supDie(x ctx) string { return stepDie(x.Lvl, [2]int{1, 8}, [2]int{10, 10}, [2]int{18, 12}) }
func inspDie(x ctx) string {
	return stepDie(x.Lvl, [2]int{1, 6}, [2]int{5, 8}, [2]int{10, 10}, [2]int{15, 12})
}
func martialDie(x ctx) string {
	return stepDie(x.Lvl, [2]int{1, 4}, [2]int{5, 6}, [2]int{11, 8}, [2]int{17, 10})
}
func bloodDie(x ctx) string {
	return stepDie(x.Lvl, [2]int{1, 4}, [2]int{5, 6}, [2]int{11, 8}, [2]int{17, 10})
}
func huntDie(x ctx) string {
	return stepDie(x.Lvl, [2]int{1, 6}, [2]int{5, 8}, [2]int{11, 10}, [2]int{17, 12})
}

func one(f func(x ctx) string) val { return func(x ctx) string { return "1" + f(x) } }

// poolDef — ресурс класса.
type poolDef struct {
	Name, Rest string // Rest: battle | short | long
	ShortLvl   int    // с какого уровня восстанавливается и на коротком отдыхе
	Max        func(x ctx) int
}

func atLeast(n int, v int) int { return max(n, v) }

var pools = map[string]poolDef{
	"sup": {Name: "Кости превосходства", Rest: "short", Max: func(x ctx) int { return 4 + b2i(x.Lvl >= 7) + b2i(x.Lvl >= 15) }},
	"ki":  {Name: "Очки ки", Rest: "short", Max: func(x ctx) int { return x.Lvl }},
	"cd": {Name: "Божественный канал", Rest: "short", Max: func(x ctx) int {
		if x.C.Class == "paladin" {
			return 1
		}
		return 1 + b2i(x.Lvl >= 6) + b2i(x.Lvl >= 18)
	}},
	"lay":    {Name: "Наложение рук (запас HP)", Rest: "long", Max: func(x ctx) int { return 5 * x.Lvl }},
	"insp":   {Name: "Вдохновение барда", Rest: "long", ShortLvl: 5, Max: func(x ctx) int { return atLeast(1, x.M["cha"]) }},
	"wild":   {Name: "Дикий облик", Rest: "short", Max: func(ctx) int { return 2 }},
	"sorc":   {Name: "Очки чародейства", Rest: "long", Max: func(x ctx) int { return x.Lvl }},
	"surge":  {Name: "Всплеск действий", Rest: "short", Max: func(x ctx) int { return 1 + b2i(x.Lvl >= 17) }},
	"flash":  {Name: "Вспышка гениальности", Rest: "long", Max: func(x ctx) int { return atLeast(1, x.M["int"]) }},
	"recov":  {Name: "Восстановление ячеек", Rest: "long", Max: func(ctx) int { return 1 }},
	"ward":   {Name: "Магический барьер", Rest: "long", Max: func(ctx) int { return 1 }},
	"elix":   {Name: "Экспериментальный эликсир", Rest: "long", Max: func(x ctx) int { return 1 + b2i(x.Lvl >= 5) }},
	"tides":  {Name: "Приливы хаоса", Rest: "long", Max: func(ctx) int { return 1 }},
	"blade":  {Name: "Песнь клинка", Rest: "long", Max: func(x ctx) int { return x.Prof }},
	"assn":   {Name: "Убийство (раз за бой)", Rest: "battle", Max: func(ctx) int { return 1 }},
	"ambush": {Name: "Ужас из засады (раз за бой)", Rest: "battle", Max: func(ctx) int { return 1 }},
	"curse":  {Name: "Проклятие колдуна клинка", Rest: "short", Max: func(ctx) int { return 1 }},
	"fey":    {Name: "Присутствие фей", Rest: "short", Max: func(ctx) int { return 1 }},
	"dob":    {Name: "Благословение тёмного (за бой)", Rest: "battle", Max: func(ctx) int { return 3 }},
	"spirit": {Name: "Боевой дух", Rest: "long", Max: func(ctx) int { return 3 }},
	"wrath":  {Name: "Гнев бури", Rest: "long", Max: func(x ctx) int { return atLeast(1, x.M["wis"]) }},
}

func (p poolDef) restFor(lvl int) string {
	if p.ShortLvl > 0 && lvl >= p.ShortLvl {
		return "short"
	}
	return p.Rest
}

// effectSpec — эффект, который способность накладывает на цель.
type effectSpec struct {
	Name   string
	Rounds int
	Cond   string
	AC     func(x ctx) int
	Atk    val
	Once   bool
}

// riderSpec — надбавка к ближайшему попаданию (или ко всем, пока не снимут, если Persist).
type riderSpec struct {
	Dice    val
	Type    string // "" — тип оружия
	Persist bool
	Crit    bool // попадание станет критическим
	Save    string
	Cond    string
	Rounds  int
	DC      string // характеристика СЛ: str, dex, … или "sd" — большая из Силы и Ловкости
}

type ab struct {
	Class, Sub string
	Lvl        int
	Key, Name  string
	Desc       string
	Passive    bool
	Pool       string
	Cost       int
	Slot       bool   // тратит самую низкую свободную ячейку заклинаний
	Target     string // "" — на себя; ally; foe
	Heal       val
	HealPool   bool // лечит столько, сколько не хватает, из запаса пула
	Temp       val
	Fx         *effectSpec
	Extra      int // дополнительные атаки в этот ход; -1 — полный набор атак
	Rider      *riderSpec
	HPCost     val
	Regain     string // arcane | slot1…slot5
	ToPool     string // ячейка → очки этого пула
	Special    *model.Special
	SpecialDmg val
	DC         string
}

func (a ab) String() string { return a.Class + "/" + a.Sub + "/" + a.Key }

func efx(name string, rounds int, cond string) *effectSpec {
	return &effectSpec{Name: name, Rounds: rounds, Cond: cond}
}

// ---------- таблица способностей ----------

var abTable = []ab{
	// Варвар
	{Class: "barbarian", Lvl: 2, Key: "reckless", Name: "Безрассудная атака", Fx: efx("Безрассудная атака", 1, "Безрассудство"),
		Desc: "Преимущество на ваши атаки, но и по вам с преимуществом — до начала вашего следующего хода."},
	{Class: "barbarian", Sub: "berserker", Lvl: 3, Key: "frenzy", Name: "Бешенство", Passive: true, Desc: "В ярости +1 атака за ход (атака бонусным действием)."},
	{Class: "barbarian", Sub: "totem", Lvl: 3, Key: "bear", Name: "Тотемный дух: медведь", Passive: true, Desc: "В ярости сопротивление всему урону, кроме психического."},
	{Class: "barbarian", Sub: "zealot", Lvl: 3, Key: "fury", Name: "Божественная ярость", Passive: true, Desc: "В ярости раз за ход +1d6 + половина уровня излучением к первому попаданию."},

	// Бард
	{Class: "bard", Lvl: 1, Key: "inspire", Name: "Вдохновение барда", Pool: "insp", Cost: 1, Target: "ally",
		Fx:   &effectSpec{Name: "Вдохновение", Rounds: 10, Once: true, Atk: func(x ctx) string { return "1" + inspDie(x) }},
		Desc: "Союзник добавляет кубик вдохновения к броскам атаки на 10 ходов."},
	{Class: "bard", Sub: "lore", Lvl: 3, Key: "cutting", Name: "Рассекающие слова", Pool: "insp", Cost: 1, Target: "foe",
		Fx:   &effectSpec{Name: "Рассекающие слова", Rounds: 1, Once: true, Atk: func(x ctx) string { return "-1" + inspDie(x) }},
		Desc: "Враг вычитает кубик вдохновения из следующего броска атаки."},
	{Class: "bard", Sub: "valor", Lvl: 6, Key: "bardextra", Name: "Дополнительная атака", Passive: true, Desc: "Две атаки за действие."},
	{Class: "bard", Sub: "swords", Lvl: 3, Key: "flourish", Name: "Клинковый расцвет", Pool: "insp", Cost: 1,
		Rider: &riderSpec{Dice: func(x ctx) string { return "1" + inspDie(x) }}, Desc: "Следующее попадание: + кубик вдохновения к урону."},
	{Class: "bard", Sub: "glamour", Lvl: 3, Key: "mantle", Name: "Мантия вдохновения", Pool: "insp", Cost: 1, Target: "ally",
		Temp: num(func(x ctx) int { return 5 + x.M["cha"] }), Desc: "Союзник получает 5 + модификатор Харизмы временных HP (упрощённо)."},

	// Жрец
	{Class: "cleric", Lvl: 2, Key: "turn", Name: "Изгнание нежити", Pool: "cd", Cost: 1, DC: "wis",
		Special: &model.Special{Mode: "area", Save: "wis", Cond: "Напуган", Rounds: 10, Repeat: true, OnlyKind: "undead"},
		Desc:    "Нежить рядом делает спасбросок Мудрости или бежит в страхе (вся нежить во вражеской команде)."},
	{Class: "cleric", Sub: "life", Lvl: 1, Key: "discipleoflife", Name: "Ученик жизни", Passive: true, Desc: "Лечение заклинаниями получает +2 + уровень заклинания."},
	{Class: "cleric", Sub: "life", Lvl: 2, Key: "preserve", Name: "Сохранение жизни", Pool: "cd", Cost: 1, Target: "ally",
		Heal: num(func(x ctx) int { return 5 * x.Lvl }), Desc: "Канал: союзник получает 5 × уровень HP."},
	{Class: "cleric", Sub: "war", Lvl: 2, Key: "guided", Name: "Направляющий удар", Pool: "cd", Cost: 1,
		Fx: &effectSpec{Name: "Направляющий удар", Rounds: 1, Once: true, Atk: k("10")}, Desc: "Канал: +10 к следующему броску атаки."},
	{Class: "cleric", Sub: "tempest", Lvl: 1, Key: "wrathstorm", Name: "Гнев бури", Pool: "wrath", Cost: 1, DC: "wis",
		Special: &model.Special{Mode: "single", Type: "lightning", Save: "dex", Half: true}, SpecialDmg: k("2d8"),
		Desc: "Реакция: 2d8 молнией по врагу, спасбросок Ловкости — половина."},
	{Class: "cleric", Sub: "light", Lvl: 2, Key: "dawn", Name: "Сияющий рассвет", Pool: "cd", Cost: 1, DC: "wis",
		Special: &model.Special{Mode: "area", Type: "radiant", Save: "con", Half: true}, SpecialDmg: num2("2d10", func(x ctx) int { return x.Lvl }),
		Desc: "Канал: 2d10 + уровень излучением по всем врагам, спасбросок Телосложения — половина."},

	// Друид
	{Class: "druid", Lvl: 2, Key: "wildshape", Name: "Дикий облик", Pool: "wild", Cost: 1, Temp: num(func(x ctx) int { return 3 * x.Lvl }),
		Desc: "Упрощённо: временные HP зверя = 3 × уровень (форму и её атаки ведите вручную)."},
	{Class: "druid", Sub: "moon", Lvl: 2, Key: "moonshape", Name: "Боевой облик луны", Pool: "wild", Cost: 1, Temp: num(func(x ctx) int { return 6 * x.Lvl }),
		Desc: "Упрощённо: сильные формы Круга луны дают 6 × уровень временных HP."},
	{Class: "druid", Sub: "land", Lvl: 2, Key: "recovery", Name: "Естественное восстановление", Pool: "recov", Cost: 1, Regain: "arcane",
		Desc: "Раз за долгий отдых: вернуть потраченные ячейки суммой уровней до половины уровня друида (не выше 5-го)."},

	// Воин
	{Class: "fighter", Lvl: 2, Key: "surge", Name: "Всплеск действий", Pool: "surge", Cost: 1, Extra: -1,
		Desc: "Ещё один полный набор атак в этот ход."},
	{Class: "fighter", Sub: "battlemaster", Lvl: 3, Key: "trip", Name: "Манёвр: подсечка", Pool: "sup", Cost: 1,
		Rider: &riderSpec{Dice: one(supDie), Save: "str", Cond: "Сбит с ног", Rounds: 1, DC: "sd"},
		Desc:  "Следующее попадание: + кость превосходства к урону, цель сбивается с ног при провале спасброска Силы."},
	{Class: "fighter", Sub: "battlemaster", Lvl: 3, Key: "menacing", Name: "Манёвр: грозная атака", Pool: "sup", Cost: 1,
		Rider: &riderSpec{Dice: one(supDie), Save: "wis", Cond: "Напуган", Rounds: 1, DC: "sd"},
		Desc:  "Следующее попадание: + кость к урону, цель напугана при провале спасброска Мудрости."},
	{Class: "fighter", Sub: "battlemaster", Lvl: 3, Key: "riposte", Name: "Манёвр: контрудар", Pool: "sup", Cost: 1,
		Rider: &riderSpec{Dice: one(supDie)}, Desc: "Следующее попадание: + кость превосходства к урону."},
	{Class: "fighter", Sub: "battlemaster", Lvl: 3, Key: "precision", Name: "Манёвр: точная атака", Pool: "sup", Cost: 1,
		Fx: &effectSpec{Name: "Точная атака", Rounds: 1, Once: true, Atk: one(supDie)}, Desc: "+ кость превосходства к следующему броску атаки."},
	{Class: "fighter", Sub: "battlemaster", Lvl: 3, Key: "parry", Name: "Манёвр: парирование", Pool: "sup", Cost: 1,
		Temp: func(x ctx) string { return "1" + supDie(x) + "+" + strconv.Itoa(max(0, x.M["dex"])) }, Desc: "Упрощённо: временные HP = кость превосходства + Ловкость (гасит урон)."},
	{Class: "fighter", Sub: "battlemaster", Lvl: 3, Key: "rally", Name: "Манёвр: сплочение", Pool: "sup", Cost: 1, Target: "ally",
		Temp: func(x ctx) string { return "1" + supDie(x) + "+" + strconv.Itoa(max(0, x.M["cha"])) }, Desc: "Союзник получает временные HP: кость превосходства + Харизма."},
	{Class: "fighter", Sub: "eldritchknight", Lvl: 3, Key: "ekcast", Name: "Колдовство воина", Passive: true, Desc: "Ячейки заклинаний как у третьего заклинателя, характеристика Интеллект. Заклинания — из списка волшебника."},
	{Class: "fighter", Sub: "samurai", Lvl: 3, Key: "spirit", Name: "Боевой дух", Pool: "spirit", Cost: 1,
		Temp: num(func(x ctx) int { return 5 * (1 + b2i(x.Lvl >= 10) + b2i(x.Lvl >= 15)) }), Fx: efx("Боевой дух", 1, "Преимущество"),
		Desc: "Преимущество на атаки до следующего хода и временные HP: 5 (10 с 10-го, 15 с 15-го уровня)."},

	// Монах
	{Class: "monk", Lvl: 1, Key: "martial", Name: "Боевые искусства", Passive: true, Desc: "Безоружный удар монаха: кубик растёт с уровнем, бьёт Ловкостью."},
	{Class: "monk", Lvl: 2, Key: "flurry", Name: "Шквал ударов", Pool: "ki", Cost: 1, Extra: 2, Desc: "Две дополнительные атаки в этот ход."},
	{Class: "monk", Lvl: 2, Key: "patient", Name: "Терпеливая защита", Pool: "ki", Cost: 1, Fx: efx("Терпеливая защита", 1, "Уклонение"),
		Desc: "Атаки по вам идут с помехой до начала следующего хода."},
	{Class: "monk", Lvl: 5, Key: "stun", Name: "Ошеломляющий удар", Pool: "ki", Cost: 1,
		Rider: &riderSpec{Save: "con", Cond: "Ошеломлён", Rounds: 1, DC: "wis"}, Desc: "Следующее попадание: цель ошеломлена при провале спасброска Телосложения."},
	{Class: "monk", Sub: "openhand", Lvl: 3, Key: "palm", Name: "Шквал с ладонью", Pool: "ki", Cost: 1, Extra: 2,
		Rider: &riderSpec{Save: "dex", Cond: "Сбит с ног", Rounds: 1, DC: "wis"}, Desc: "Две дополнительные атаки; попадание сбивает с ног при провале спасброска Ловкости."},
	{Class: "monk", Sub: "fourelements", Lvl: 3, Key: "snake", Name: "Клыки огненной змеи", Pool: "ki", Cost: 1,
		Rider: &riderSpec{Dice: k("1d10"), Type: "fire"}, Desc: "Следующее попадание: +1d10 огнём."},
	{Class: "monk", Sub: "mercy", Lvl: 3, Key: "healhand", Name: "Руки здоровья", Pool: "ki", Cost: 1, Target: "ally",
		Heal: func(x ctx) string { return "1" + martialDie(x) + "+" + strconv.Itoa(max(0, x.M["wis"])) }, Desc: "Союзник исцеляется: кубик боевых искусств + Мудрость."},
	{Class: "monk", Sub: "mercy", Lvl: 3, Key: "harmhand", Name: "Руки вреда", Pool: "ki", Cost: 1,
		Rider: &riderSpec{Dice: func(x ctx) string { return "1" + martialDie(x) + "+" + strconv.Itoa(max(0, x.M["wis"])) }, Type: "necrotic"},
		Desc:  "Следующее попадание: + кубик боевых искусств + Мудрость некротикой."},

	// Паладин
	{Class: "paladin", Lvl: 1, Key: "layhands", Name: "Наложение рук", Pool: "lay", Target: "ally", HealPool: true,
		Desc: "Лечит союзника на недостающее здоровье, тратя запас (5 × уровень HP за долгий отдых)."},
	{Class: "paladin", Lvl: 2, Key: "smite", Name: "Божественная кара", Slot: true,
		Rider: &riderSpec{Dice: func(x ctx) string { return strconv.Itoa(min(5, 1+max(1, x.C.Counters["rdslot:smite"]))) + "d8" }, Type: "radiant"},
		Desc:  "Тратит самую низкую ячейку: следующее попадание +2d8 излучением (+1d8 за уровень ячейки выше 1-го, до 5d8)."},
	{Class: "paladin", Lvl: 6, Key: "aura", Name: "Аура защиты", Passive: true, Desc: "Модификатор Харизмы добавляется к вашим спасброскам."},
	{Class: "paladin", Sub: "devotion", Lvl: 3, Key: "sacredweapon", Name: "Священное оружие", Pool: "cd", Cost: 1,
		Fx: &effectSpec{Name: "Священное оружие", Rounds: 10, Atk: num(func(x ctx) int { return max(1, x.M["cha"]) })}, Desc: "Канал: бонус Харизмы к броскам атаки на 10 ходов."},
	{Class: "paladin", Sub: "vengeance", Lvl: 3, Key: "vow", Name: "Клятва врага", Pool: "cd", Cost: 1, Fx: efx("Клятва врага", 10, "Преимущество"),
		Desc: "Канал: преимущество на ваши атаки на 10 ходов."},
	{Class: "paladin", Sub: "ancients", Lvl: 3, Key: "naturewrath", Name: "Ярость природы", Pool: "cd", Cost: 1, DC: "cha",
		Special: &model.Special{Mode: "area", Save: "str", Cond: "Опутан", Rounds: 10, Repeat: true}, Desc: "Канал: лианы опутывают врагов (спасбросок Силы)."},
	{Class: "paladin", Sub: "conquest", Lvl: 3, Key: "conquer", Name: "Покоряющее присутствие", Pool: "cd", Cost: 1, DC: "cha",
		Special: &model.Special{Mode: "area", Save: "wis", Cond: "Напуган", Rounds: 10, Repeat: true}, Desc: "Канал: враги напуганы (спасбросок Мудрости)."},

	// Следопыт
	{Class: "ranger", Lvl: 2, Key: "mark", Name: "Метка охотника", Slot: true,
		Rider: &riderSpec{Dice: k("1d6"), Persist: true}, Desc: "Тратит ячейку 1-го уровня и более; каждое попадание +1d6, пока не снимете (или до конца боя). Повторное применение снимает метку."},
	{Class: "ranger", Sub: "hunter", Lvl: 3, Key: "colossus", Name: "Убийца колоссов", Passive: true, Desc: "Раз за ход +1d8 к урону по раненой цели."},
	{Class: "ranger", Sub: "gloomstalker", Lvl: 3, Key: "dread", Name: "Ужас из засады", Pool: "ambush", Cost: 1, Extra: 1,
		Rider: &riderSpec{Dice: k("1d8")}, Desc: "Раз за бой: ещё одна атака в этот ход, попадание +1d8."},

	// Плут
	{Class: "rogue", Lvl: 2, Key: "cunning", Name: "Хитрое действие", Fx: efx("Хитрое действие", 1, "Уклонение"), Desc: "Уклонение до начала следующего хода: атаки по вам с помехой."},
	{Class: "rogue", Sub: "assassin", Lvl: 3, Key: "assassinate", Name: "Убийство", Pool: "assn", Cost: 1, Fx: efx("Убийство", 1, "Преимущество"),
		Rider: &riderSpec{Crit: true}, Desc: "Раз за бой: преимущество и гарантированный критический удар по ближайшему попаданию."},
	{Class: "rogue", Sub: "arcanetrickster", Lvl: 3, Key: "atcast", Name: "Колдовство плута", Passive: true, Desc: "Ячейки заклинаний как у третьего заклинателя, характеристика Интеллект."},
	{Class: "rogue", Sub: "soulknife", Lvl: 3, Key: "blades", Name: "Психические клинки", Passive: true, Desc: "В оружии появляется «Психический клинок»: 1d6 психической энергии, фехтовальный."},

	// Чародей
	{Class: "sorcerer", Lvl: 2, Key: "slot1", Name: "Ячейка 1 ур. (2 очка)", Pool: "sorc", Cost: 2, Regain: "slot1", Desc: "Возвращает потраченную ячейку 1-го уровня."},
	{Class: "sorcerer", Lvl: 3, Key: "slot2", Name: "Ячейка 2 ур. (3 очка)", Pool: "sorc", Cost: 3, Regain: "slot2", Desc: "Возвращает потраченную ячейку 2-го уровня."},
	{Class: "sorcerer", Lvl: 5, Key: "slot3", Name: "Ячейка 3 ур. (5 очков)", Pool: "sorc", Cost: 5, Regain: "slot3", Desc: "Возвращает потраченную ячейку 3-го уровня."},
	{Class: "sorcerer", Lvl: 7, Key: "slot4", Name: "Ячейка 4 ур. (6 очков)", Pool: "sorc", Cost: 6, Regain: "slot4", Desc: "Возвращает потраченную ячейку 4-го уровня."},
	{Class: "sorcerer", Lvl: 9, Key: "slot5", Name: "Ячейка 5 ур. (7 очков)", Pool: "sorc", Cost: 7, Regain: "slot5", Desc: "Возвращает потраченную ячейку 5-го уровня."},
	{Class: "sorcerer", Lvl: 2, Key: "topoints", Name: "Ячейка в очки", ToPool: "sorc", Desc: "Сжигает самую низкую свободную ячейку: очки чародейства = её уровень."},
	{Class: "sorcerer", Sub: "wild", Lvl: 1, Key: "tides", Name: "Приливы хаоса", Pool: "tides", Cost: 1, Fx: efx("Приливы хаоса", 1, "Преимущество"),
		Desc: "Раз за долгий отдых: преимущество на ваши атаки до следующего хода."},

	// Колдун
	{Class: "warlock", Lvl: 1, Key: "hex", Name: "Сглаз", Slot: true, Rider: &riderSpec{Dice: k("1d6"), Type: "necrotic", Persist: true},
		Desc: "Тратит ячейку договора; каждое попадание +1d6 некротикой, пока не снимете. Повторное применение снимает сглаз."},
	{Class: "warlock", Sub: "fiend", Lvl: 1, Key: "darkone", Name: "Благословение тёмного", Pool: "dob", Cost: 1,
		Temp: num(func(x ctx) int { return max(1, x.M["cha"]) + x.Lvl }), Desc: "После убийства врага: временные HP = Харизма + уровень (до 3 раз за бой)."},
	{Class: "warlock", Sub: "archfey", Lvl: 1, Key: "feypresence", Name: "Присутствие фей", Pool: "fey", Cost: 1, DC: "cha",
		Special: &model.Special{Mode: "area", Save: "wis", Cond: "Напуган", Rounds: 1}, Desc: "Враги вокруг напуганы до конца вашего следующего хода (спасбросок Мудрости)."},
	{Class: "warlock", Sub: "hexblade", Lvl: 1, Key: "hexcurse", Name: "Проклятие колдуна клинка", Pool: "curse", Cost: 1,
		Rider: &riderSpec{Dice: num(func(x ctx) int { return x.Prof }), Persist: true}, Desc: "Каждое попадание +бонус мастерства к урону, пока не снимете."},

	// Волшебник
	{Class: "wizard", Lvl: 1, Key: "arcane", Name: "Магическое восстановление", Pool: "recov", Cost: 1, Regain: "arcane",
		Desc: "Раз за долгий отдых: вернуть потраченные ячейки суммой уровней до половины уровня волшебника (не выше 5-го)."},
	{Class: "wizard", Sub: "abjuration", Lvl: 2, Key: "warding", Name: "Магический барьер", Pool: "ward", Cost: 1,
		Temp: num(func(x ctx) int { return 2*x.Lvl + max(0, x.M["int"]) }), Desc: "Временные HP = 2 × уровень + Интеллект."},
	{Class: "wizard", Sub: "bladesinging", Lvl: 2, Key: "bladesong", Name: "Песнь клинка", Pool: "blade", Cost: 1,
		Fx:   &effectSpec{Name: "Песнь клинка", Rounds: 10, AC: func(x ctx) int { return max(1, x.M["int"]) }},
		Desc: "10 ходов: КД +Интеллект."},

	// Изобретатель
	{Class: "artificer", Lvl: 7, Key: "flashgenius", Name: "Вспышка гениальности", Pool: "flash", Cost: 1, Target: "ally",
		Fx: &effectSpec{Name: "Вспышка гениальности", Rounds: 1, Once: true, Atk: num(func(x ctx) int { return max(1, x.M["int"]) })}, Desc: "Союзник добавляет Интеллект к следующему броску атаки."},
	{Class: "artificer", Sub: "alchemist", Lvl: 3, Key: "elixir", Name: "Экспериментальный эликсир", Pool: "elix", Cost: 1, Target: "ally",
		Heal: func(x ctx) string { return "2d4+" + strconv.Itoa(max(0, x.M["int"])) }, Desc: "Эликсир исцеления: 2d4 + Интеллект."},

	// Охотник на кровь
	{Class: "bloodhunter", Lvl: 1, Key: "prey", Name: "Охотничья добыча", Rider: &riderSpec{Dice: one(huntDie), Persist: true},
		Desc: "Каждое попадание по помеченной цели + кубик добычи, пока не снимете."},
	{Class: "bloodhunter", Lvl: 1, Key: "rite", Name: "Кровавый обряд", HPCost: one(bloodDie), Rider: &riderSpec{Dice: one(bloodDie), Type: "fire", Persist: true},
		Desc: "Цена: бросок кубика обряда в HP. Каждое попадание + кубик огнём, пока не снимете."},
}

func num2(base string, add func(x ctx) int) val {
	return func(x ctx) string { return fmt.Sprintf("%s+%d", base, add(x)) }
}

var abIndex = func() map[string]ab {
	m := map[string]ab{}
	for _, a := range abTable {
		if _, dup := m[a.Class+"/"+a.Key]; dup {
			panic("повтор способности: " + a.String())
		}
		m[a.Class+"/"+a.Key] = a
	}
	return m
}()

func (d DnD5e) ctxOf(c *model.Character) ctx {
	ds := d.mods(c)
	lvl := clampLevel(c.Level)
	return ctx{C: c, Lvl: lvl, Prof: 2 + (lvl-1)/4, M: ds}
}

// mods — модификаторы без вызова Derive (чтобы не зациклиться): базовые значения плюс эффекты предметов.
func (d DnD5e) mods(c *model.Character) map[string]int {
	fx, _ := itemEffects(c)
	eff := applyEffects(c.Abilities, fx)
	out := map[string]int{}
	for _, a := range dndAbilities {
		out[a.ID] = Mod(eff[a.ID])
	}
	return out
}

// abilitiesFor — способности героя на его уровне (в порядке таблицы).
func abilitiesFor(c *model.Character) (out []ab) {
	if c.Stat != nil {
		return nil
	}
	for _, a := range abTable {
		if a.Class == c.Class && (a.Sub == "" || a.Sub == c.Subclass) && c.Level >= a.Lvl {
			out = append(out, a)
		}
	}
	return out
}

func findAb(c *model.Character, key string) (ab, bool) {
	for _, a := range abilitiesFor(c) {
		if a.Key == key {
			return a, true
		}
	}
	return ab{}, false
}

func poolKey(p string) string { return "p:" + p }

func poolLeft(c *model.Character, x ctx, p string) (left, mx int) {
	def, ok := pools[p]
	if !ok {
		return 0, 0
	}
	mx = def.Max(x)
	return max(0, mx-c.Counters[poolKey(p)]), mx
}

// poolViews — ресурсы, которыми пользуются способности героя.
func (d DnD5e) poolViews(c *model.Character) (out []model.PoolView) {
	x := d.ctxOf(c)
	seen := map[string]bool{}
	for _, a := range abilitiesFor(c) {
		for _, p := range []string{a.Pool, a.ToPool} {
			if p == "" || seen[p] {
				continue
			}
			seen[p] = true
			left, mx := poolLeft(c, x, p)
			if mx > 0 {
				out = append(out, model.PoolView{Key: p, Name: pools[p].Name, Left: left, Max: mx, Rest: pools[p].restFor(x.Lvl)})
			}
		}
	}
	return out
}

// abilityFeatures — кнопки способностей на листе героя и в бою.
func (d DnD5e) abilityFeatures(c *model.Character) (fs []model.Feature) {
	x := d.ctxOf(c)
	for _, a := range abilitiesFor(c) {
		f := model.Feature{Key: a.Key, Name: a.Name, Desc: a.Desc, Kind: "ability", Target: a.Target, Cost: a.Cost}
		switch {
		case a.Passive:
			f.Kind = "passive"
		case a.Special != nil:
			f.Kind, f.Mode = "special", a.Special.Mode
		}
		if a.Rider != nil && a.Rider.Persist {
			f.On = c.Active["rd:"+a.Key]
		} else if a.Rider != nil {
			f.On = c.Active["rd:"+a.Key]
		}
		if a.Pool != "" {
			left, mx := poolLeft(c, x, a.Pool)
			f.Pool, f.Left, f.Max, f.Rest = pools[a.Pool].Name, left, mx, pools[a.Pool].restFor(x.Lvl)
			if a.Special != nil && left < max(1, a.Cost) {
				f.On = true // для особых способностей On означает «недоступно»
			}
		}
		if a.Slot {
			f.Pool, f.Rest = "Ячейка заклинаний", "long"
		}
		fs = append(fs, f)
	}
	return fs
}

// ---------- применение ----------

func rollStr(r dice.Roller, s string, crit bool) (int, []int) {
	s = strings.TrimSpace(s)
	sign := 1
	if strings.HasPrefix(s, "-") {
		sign, s = -1, s[1:]
	}
	s = strings.TrimPrefix(s, "+")
	if n, err := strconv.Atoi(s); err == nil {
		return sign * n, nil
	}
	e, err := dice.Parse(s)
	if err != nil {
		return 0, nil
	}
	res := e.Roll(r, crit)
	return sign * res.Total, res.Rolls
}

func dcOf(x ctx, ab string) int {
	m := x.M[ab]
	if ab == "sd" {
		m = max(x.M["str"], x.M["dex"])
	}
	return 8 + x.Prof + m
}

// spendSlot тратит самую низкую свободную ячейку и возвращает её уровень.
func (d DnD5e) spendSlot(c *model.Character) (int, error) {
	for _, sv := range d.Derive(c).Slots {
		if c.SlotsUsed[sv.Level-1] < sv.Max {
			c.SlotsUsed[sv.Level-1]++
			return sv.Level, nil
		}
	}
	return 0, errors.New("нет свободных ячеек заклинаний")
}

// Ability применяет способность. tgt — выбранная цель (союзник или враг); для способностей на себя игнорируется.
// extra — сколько атак добавить в текущий ход (-1 — полный набор).
func (d DnD5e) Ability(r dice.Roller, key string, src, tgt *model.Character) (msg string, extra int, err error) {
	a, ok := findAb(src, key)
	switch {
	case !ok || a.Passive:
		return "", 0, errors.New("у героя нет такой способности")
	case a.Special != nil:
		return "", 0, errors.New("эта способность применяется в бою")
	case src.Dead || src.Down():
		return "", 0, errors.New("герой без сознания")
	}
	x := d.ctxOf(src)
	// повторное применение постоянной надбавки снимает её без затрат
	if a.Rider != nil && a.Rider.Persist && src.Active["rd:"+key] {
		delete(src.Active, "rd:"+key)
		return fmt.Sprintf("%s: «%s» снято", src.Name, a.Name), 0, nil
	}
	t := tgt
	switch a.Target {
	case "":
		t = src
	case "ally":
		if t == nil {
			t = src
		}
	case "foe":
		if t == nil {
			return "", 0, errors.New("выберите цель")
		}
	}
	if t.Dead {
		return "", 0, errors.New("цель погибла")
	}
	// проверки до траты
	cost := a.Cost
	var healAmt int
	if a.HealPool {
		left, _ := poolLeft(src, x, a.Pool)
		missing := d.Derive(t).MaxHP - t.HP
		healAmt = min(left, missing)
		if healAmt <= 0 {
			if left <= 0 {
				return "", 0, errors.New("запас исчерпан — нужен долгий отдых")
			}
			return "", 0, errors.New(t.Name + " не нуждается в лечении")
		}
		cost = healAmt
	}
	if a.Pool != "" {
		if left, _ := poolLeft(src, x, a.Pool); left < cost {
			return "", 0, fmt.Errorf("не хватает ресурса «%s» (нужно %d, есть %d)", pools[a.Pool].Name, cost, left)
		}
	}
	var hpCost int
	if a.HPCost != nil {
		hpCost, _ = rollStr(r, a.HPCost(x), false)
		if src.HP <= hpCost {
			return "", 0, errors.New("не хватает HP, чтобы заплатить цену")
		}
	}
	if a.Regain != "" {
		if err := d.canRegain(src, a.Regain); err != nil {
			return "", 0, err
		}
	}
	slot := 0
	if a.Slot {
		if slot, err = d.spendSlot(src); err != nil {
			return "", 0, err
		}
	}
	if a.ToPool != "" {
		lv, err := d.spendSlot(src)
		if err != nil {
			return "", 0, err
		}
		got := min(lv, src.Counters[poolKey(a.ToPool)])
		src.Count(poolKey(a.ToPool), -got)
		return fmt.Sprintf("%s: ячейка %d ур. → +%d очков (%s)", src.Name, lv, got, pools[a.ToPool].Name), 0, nil
	}
	// затраты
	if a.Pool != "" {
		src.Count(poolKey(a.Pool), cost)
	}
	if hpCost > 0 {
		src.HP -= hpCost
	}
	parts := []string{}
	if a.Heal != nil || a.HealPool {
		if a.HealPool {
			before := t.HP
			t.Heal(healAmt, d.Derive(t).MaxHP)
			parts = append(parts, fmt.Sprintf("%s +%d HP (%d → %d)", t.Name, t.HP-before, before, t.HP))
		} else {
			n, _ := rollStr(r, a.Heal(x), false)
			before := t.HP
			t.Heal(max(1, n), d.Derive(t).MaxHP)
			parts = append(parts, fmt.Sprintf("%s +%d HP (%d → %d)", t.Name, t.HP-before, before, t.HP))
		}
	}
	if a.Temp != nil {
		n, _ := rollStr(r, a.Temp(x), false)
		if n > t.TempHP {
			t.TempHP = n
		}
		parts = append(parts, fmt.Sprintf("%s: временные HP %d", t.Name, t.TempHP))
	}
	if a.Fx != nil {
		e := model.Effect{Name: a.Fx.Name, Rounds: a.Fx.Rounds, Cond: a.Fx.Cond, Src: src.ID, Once: a.Fx.Once}
		if a.Fx.AC != nil {
			e.AC = a.Fx.AC(x)
		}
		if a.Fx.Atk != nil {
			e.Atk = a.Fx.Atk(x)
		}
		AddEffect(t, e)
		parts = append(parts, fmt.Sprintf("%s: «%s»", t.Name, e.Name))
	}
	if a.Regain != "" {
		parts = append(parts, d.regain(src, a.Regain))
	}
	if a.Rider != nil {
		setActive(src, "rd:"+key, true)
		if slot > 0 {
			src.Count("rdslot:"+key, -src.Counters["rdslot:"+key])
			src.Count("rdslot:"+key, slot)
		}
		parts = append(parts, "готово к следующему попаданию")
	}
	extra = a.Extra
	if extra < 0 {
		extra = d.Derive(src).Attacks
	}
	if extra > 0 {
		parts = append(parts, fmt.Sprintf("+%d атак в этот ход", extra))
	}
	if a.Slot {
		parts = append(parts, fmt.Sprintf("ячейка %d ур.", slot))
	}
	if hpCost > 0 {
		parts = append(parts, fmt.Sprintf("цена %d HP", hpCost))
	}
	return fmt.Sprintf("%s: %s — %s", src.Name, a.Name, strings.Join(parts, "; ")), extra, nil
}

func (d DnD5e) canRegain(c *model.Character, what string) error {
	if strings.HasPrefix(what, "slot") {
		n, _ := strconv.Atoi(strings.TrimPrefix(what, "slot"))
		if n < 1 || n > 9 || c.SlotsUsed[n-1] == 0 {
			return errors.New("эта ячейка не потрачена")
		}
		return nil
	}
	budget := (clampLevel(c.Level) + 1) / 2
	for lv := min(5, budget); lv >= 1; lv-- {
		if c.SlotsUsed[lv-1] > 0 {
			return nil
		}
	}
	return errors.New("нечего восстанавливать")
}

func (d DnD5e) regain(c *model.Character, what string) string {
	if strings.HasPrefix(what, "slot") {
		n, _ := strconv.Atoi(strings.TrimPrefix(what, "slot"))
		c.SlotsUsed[n-1]--
		return fmt.Sprintf("ячейка %d ур. возвращена", n)
	}
	budget := (clampLevel(c.Level) + 1) / 2
	var got []string
	for lv := 5; lv >= 1; lv-- {
		for c.SlotsUsed[lv-1] > 0 && budget >= lv {
			c.SlotsUsed[lv-1]--
			budget -= lv
			got = append(got, strconv.Itoa(lv))
		}
	}
	return "возвращены ячейки " + strings.Join(got, ", ") + " ур."
}

// ---------- надбавки к удару ----------

// activeRiders — активные надбавки героя в стабильном порядке.
func activeRiders(c *model.Character) (out []ab) {
	var keys []string
	for k, on := range c.Active {
		if on && strings.HasPrefix(k, "rd:") {
			keys = append(keys, strings.TrimPrefix(k, "rd:"))
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		if a, ok := findAb(c, key); ok && a.Rider != nil {
			out = append(out, a)
		}
	}
	return out
}

func riderCrit(c *model.Character) bool {
	return slices.ContainsFunc(activeRiders(c), func(a ab) bool { return a.Rider.Crit })
}

// applyRiders вызывается при попадании: добавляет кубики, накладывает состояния. Возвращает урон и пояснение.
func (d DnD5e) applyRiders(r dice.Roller, a, t *model.Character, w model.Weapon, crit bool) (int, string) {
	total, note := 0, ""
	x := d.ctxOf(a)
	for _, ab := range activeRiders(a) {
		rd := ab.Rider
		if rd.Dice != nil {
			n, rolls := rollStr(r, rd.Dice(x), crit)
			ty := rd.Type
			if ty == "" {
				ty = w.Type
			}
			n = int(float64(n) * d.Resist(t, ty))
			total += n
			if len(rolls) > 0 {
				note += fmt.Sprintf(" + %s %v", ab.Name, rolls)
			} else {
				note += fmt.Sprintf(" + %s %d", ab.Name, n)
			}
		}
		if rd.Save != "" && rd.Cond != "" {
			dc := dcOf(x, rd.DC)
			tot := d.saveRoll(r, t, rd.Save)
			if tot >= dc {
				note += fmt.Sprintf(" [%s: спасбросок %s против %d — успех]", ab.Name, saveText(tot), dc)
			} else {
				AddEffect(t, model.Effect{Name: ab.Name, Rounds: rd.Rounds, Cond: rd.Cond, Src: a.ID})
				note += fmt.Sprintf(" [%s: спасбросок %s против %d — %s]", ab.Name, saveText(tot), dc, rd.Cond)
			}
		}
		if !rd.Persist {
			delete(a.Active, "rd:"+ab.Key)
		}
	}
	return total, note
}

// autoBonus — пассивные надбавки подклассов: убийца колоссов, божественная ярость.
func (d DnD5e) autoBonus(r dice.Roller, a, t *model.Character, w model.Weapon, crit bool) (int, string) {
	lvl := clampLevel(a.Level)
	if a.Active["tb"] {
		return 0, ""
	}
	switch {
	case a.Class == "ranger" && a.Subclass == "hunter" && lvl >= 3 && t.HP < d.Derive(t).MaxHP:
		setActive(a, "tb", true)
		n, rolls := rollStr(r, "1d8", crit)
		return int(float64(n) * d.Resist(t, w.Type)), fmt.Sprintf(" + убийца колоссов %v", rolls)
	case a.Class == "barbarian" && a.Subclass == "zealot" && lvl >= 3 && a.Active["rage"]:
		setActive(a, "tb", true)
		n, rolls := rollStr(r, "1d6", crit)
		n += lvl / 2
		return int(float64(n) * d.Resist(t, "radiant")), fmt.Sprintf(" + божественная ярость %v+%d", rolls, lvl/2)
	}
	return 0, ""
}

// ---------- особые способности героев ----------

func (d DnD5e) heroSpecial(r dice.Roller, key string, src *model.Character, foes []*model.Character) ([]string, error) {
	a, ok := findAb(src, key)
	if !ok || a.Special == nil {
		return nil, errors.New("у героя нет такой способности")
	}
	if src.Dead || src.Down() {
		return nil, errors.New("герой без сознания")
	}
	x := d.ctxOf(src)
	if left, _ := poolLeft(src, x, a.Pool); left < max(1, a.Cost) {
		return nil, fmt.Errorf("не хватает ресурса «%s»", pools[a.Pool].Name)
	}
	s := *a.Special
	s.Key, s.Name, s.DC = key, a.Name, dcOf(x, a.DC)
	if a.SpecialDmg != nil {
		s.Dmg = a.SpecialDmg(x)
	}
	if s.Save != "" && s.Repeat {
		s.RSave = s.Save
	}
	lines, err := d.runSpecial(r, s, src, foes)
	if err != nil {
		return nil, err
	}
	src.Count(poolKey(a.Pool), max(1, a.Cost))
	return lines, nil
}

// ---------- отдых ----------

// restPools сбрасывает ресурсы: battle — конец боя, short — короткий отдых, long — долгий.
func restPools(c *model.Character, kind string) {
	lvl := clampLevel(c.Level)
	for p, def := range pools {
		switch def.restFor(lvl) {
		case "battle":
		case "short":
			if kind == "battle" {
				continue
			}
		case "long":
			if kind != "long" {
				continue
			}
		}
		delete(c.Counters, poolKey(p))
	}
	for k := range c.Active {
		if strings.HasPrefix(k, "rd:") && kind == "battle" {
			delete(c.Active, k)
		}
	}
	for k := range c.Counters {
		if strings.HasPrefix(k, "rdslot:") && kind == "battle" {
			delete(c.Counters, k)
		}
	}
}
