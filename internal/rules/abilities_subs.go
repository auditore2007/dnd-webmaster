package rules

import (
	"fmt"
	"strconv"

	"heroesbook/internal/model"
)

// Механика подклассов, у которых раньше было только описание. Правила упрощены так же, как в abilities.go:
// реакции и бонусные действия – кнопки, которые мастер нажимает в нужный момент; описание говорит, что именно считается.

// byLvl – значение, растущее с уровнем: base, затем steps {уровень, значение}.
func byLvl(lvl, base int, steps ...[2]int) int {
	v := base
	for _, s := range steps {
		if lvl >= s[0] {
			v = s[1]
		}
	}
	return v
}

// nd – «NdS» с числом кубиков, растущим с уровнем.
func nd(sides int, f func(x ctx) int) val {
	return func(x ctx) string { return fmt.Sprintf("%dd%d", max(1, f(x)), sides) }
}

// plus – кубики плюс модификатор характеристики (не ниже нуля).
func plus(d val, ab string) val {
	return func(x ctx) string { return d(x) + "+" + strconv.Itoa(max(0, x.M[ab])) }
}

func prof(x ctx) int     { return x.Prof }
func profPool(x ctx) int { return x.Prof }
func one1(ctx) int       { return 1 }
func two(ctx) int        { return 2 }
func modOr1(ab string) func(x ctx) int {
	return func(x ctx) int { return atLeast(1, x.M[ab]) }
}

func psiDie(x ctx) string {
	return stepDie(x.Lvl, [2]int{1, 6}, [2]int{5, 8}, [2]int{11, 10}, [2]int{17, 12})
}

// divine – «божественный удар» жреца: +1d8 (2d8 с 14-го уровня) нужного типа к ближайшему попаданию.
func divine(sub, ty, name string) ab {
	return ab{Class: "cleric", Sub: sub, Lvl: 8, Key: "divine-" + sub, Name: "Божественный удар: " + name,
		Rider: &riderSpec{Dice: nd(8, func(x ctx) int { return byLvl(x.Lvl, 1, [2]int{14, 2}) }), Type: ty},
		Desc:  "Раз за ход: ближайшее попадание оружием +1d8 (2d8 с 14-го уровня), тип – " + name + "."}
}

func area(save, cond string, rounds int, repeat bool) *model.Special {
	return &model.Special{Mode: "area", Save: save, Cond: cond, Rounds: rounds, Repeat: repeat}
}

func single(save, cond string, rounds int, repeat bool) *model.Special {
	return &model.Special{Mode: "single", Save: save, Cond: cond, Rounds: rounds, Repeat: repeat}
}

func blast(mode, ty, save string, half bool) *model.Special {
	return &model.Special{Mode: mode, Type: ty, Save: save, Half: half}
}

var subPools = map[string]poolDef{
	"wildaug":   {Name: "Магическая подпитка", Rest: "long", Max: profPool},
	"barbsurge": {Name: "Волна магии (за бой)", Rest: "battle", Max: one1},
	"beastrage": {Name: "Заразная ярость (за бой)", Rest: "battle", Max: one1},
	"terror":    {Name: "Слова ужаса", Rest: "short", Max: one1},
	"dance":     {Name: "Танцующий предмет", Rest: "long", Max: one1},
	"forge":     {Name: "Благословение кузни", Rest: "long", Max: one1},
	"bond":      {Name: "Узы мира", Rest: "long", Max: profPool},
	"balm":      {Name: "Бальзам летнего двора (кубики)", Rest: "long", Max: func(x ctx) int { return x.Lvl }},
	"totem":     {Name: "Тотем духа", Rest: "short", Max: one1},
	"starmap":   {Name: "Путеводный знак", Rest: "long", Max: profPool},
	"arrow":     {Name: "Магические стрелы", Rest: "short", Max: two},
	"unwaver":   {Name: "Ответный удар кавалериста", Rest: "long", Max: modOr1("con")},
	"echo":      {Name: "Высвобождение воплощения", Rest: "long", Max: modOr1("con")},
	"psidie":    {Name: "Кости псионной энергии", Rest: "long", ShortLvl: 1, Max: func(x ctx) int { return 2 * x.Prof }},
	"giant":     {Name: "Гигантская мощь", Rest: "long", Max: profPool},
	"rune":      {Name: "Руны", Rest: "short", Max: two},
	"favored":   {Name: "Предпочтение богов", Rest: "short", Max: one1},
	"balance":   {Name: "Восстановление равновесия", Rest: "long", Max: profPool},
	"entropy":   {Name: "Энтропийная защита", Rest: "short", Max: one1},
	"hlight":    {Name: "Исцеляющий свет (кубики)", Rest: "long", Max: func(x ctx) int { return 1 + x.Lvl }},
	"tentacle":  {Name: "Щупальце бездны", Rest: "long", Max: profPool},
	"undying":   {Name: "Отказ от смерти", Rest: "long", Max: one1},
	"teleport":  {Name: "Благотворное перемещение", Rest: "long", Max: one1},
	"portent":   {Name: "Знамение", Rest: "long", Max: func(x ctx) int { return byLvl(x.Lvl, 2, [2]int{14, 3}) }},
	"gaze":      {Name: "Гипнотический взгляд", Rest: "short", Max: one1},
	"illusory":  {Name: "Иллюзорный двойник", Rest: "short", Max: one1},
	"stone":     {Name: "Камень превращения", Rest: "long", Max: one1},
	"chrono":    {Name: "Хроносдвиг", Rest: "long", Max: two},
	"field":     {Name: "Защитное поле", Rest: "long", Max: profPool},
	"hybrid":    {Name: "Гибридная форма", Rest: "long", Max: one1},
	"mutagen":   {Name: "Мутагены", Rest: "long", Max: func(x ctx) int { return byLvl(x.Lvl, 1, [2]int{7, 2}, [2]int{11, 3}, [2]int{18, 4}) }},
	"wails":     {Name: "Стенания из могилы", Rest: "long", Max: profPool},
	"psi":       {Name: "Психическая энергия", Rest: "long", Max: func(x ctx) int { return 2 + 2*x.Lvl }},
	"grit":      {Name: "Выдержка", Rest: "short", Max: func(x ctx) int { return max(2, 1+x.M["wis"]) }},
}

func init() {
	for k, v := range subPools {
		pools[k] = v
	}
}

var subAbTable = []ab{
	// ---------- Варвар ----------
	{Class: "barbarian", Sub: "ancestral", Lvl: 3, Key: "ancestors", Name: "Предки-защитники", Need: "rage", Target: "foe",
		Fx: efx("Предки-защитники", 1, "Помеха"), Desc: "В ярости: первая цель, которую вы ударили, атакует с помехой до начала вашего хода."},
	{Class: "barbarian", Sub: "ancestral", Lvl: 6, Key: "spiritshield", Name: "Духовный щит", Need: "rage", Target: "ally",
		Temp: nd(6, func(x ctx) int { return byLvl(x.Lvl, 2, [2]int{10, 3}, [2]int{14, 4}) }),
		Desc: "Реакция в ярости: союзник получает временные HP 2d6 (3d6 с 10-го, 4d6 с 14-го) – упрощённо вместо снижения урона."},
	{Class: "barbarian", Sub: "storm", Lvl: 3, Key: "desert", Name: "Аура бури: пустыня", Need: "rage", Special: blast("area", "fire", "", false),
		SpecialDmg: num(func(x ctx) int { return byLvl(x.Lvl, 2, [2]int{5, 3}, [2]int{10, 4}, [2]int{15, 5}, [2]int{20, 6}) }),
		Desc:       "В ярости: огонь по всем врагам рядом – 2 урона (растёт с уровнем), без спасброска."},
	{Class: "barbarian", Sub: "storm", Lvl: 3, Key: "tundra", Name: "Аура бури: тундра", Need: "rage", Target: "ally",
		Temp: num(func(x ctx) int { return byLvl(x.Lvl, 2, [2]int{5, 3}, [2]int{10, 4}, [2]int{15, 5}, [2]int{20, 6}) }),
		Desc: "В ярости: союзник получает временные HP (2, растёт с уровнем)."},
	{Class: "barbarian", Sub: "wild", Lvl: 3, Key: "wildsurge", Name: "Волна дикой магии", Need: "rage", Pool: "barbsurge", Cost: 1,
		Pick: []ab{
			{Name: "призрачная защита", Fx: &effectSpec{Name: "Призрачная защита", Rounds: 10, AC: func(ctx) int { return 1 }}},
			{Name: "магическое оружие", Fx: &effectSpec{Name: "Магическое оружие", Rounds: 10, Atk: k("1")}},
			{Name: "живительный свет", Temp: nd(12, one1)},
			{Name: "буйство", Fx: efx("Буйство дикой магии", 1, "Преимущество")},
		},
		Desc: "Раз за ярость (за бой): случайный эффект – КД +1, +1 к атакам, временные HP 1d12 или преимущество."},
	{Class: "barbarian", Sub: "wild", Lvl: 6, Key: "bolster", Name: "Магическая подпитка", Pool: "wildaug", Cost: 1, Target: "ally",
		Fx:   &effectSpec{Name: "Магическая подпитка", Rounds: 10, Atk: k("1d3")},
		Desc: "Союзник добавляет 1d3 к броскам атаки на 10 ходов."},
	{Class: "barbarian", Sub: "beast", Lvl: 3, Key: "claws", Name: "Форма зверя: когти", Need: "rage", Extra: 1,
		Desc: "В ярости: ещё одна атака когтями в этот ход."},
	{Class: "barbarian", Sub: "beast", Lvl: 3, Key: "bite", Name: "Форма зверя: укус", Need: "rage",
		Rider: &riderSpec{Dice: k("1d8"), Type: "piercing"}, Temp: num(prof),
		Desc: "В ярости: следующее попадание – укус +1d8, вы получаете временные HP, равные бонусу мастерства."},
	{Class: "barbarian", Sub: "beast", Lvl: 10, Key: "infectious", Name: "Заразная ярость", Pool: "beastrage", Cost: 1, Need: "rage", DC: "con",
		Special: single("wis", "Очарован", 1, false), Desc: "В ярости раз за бой: враг (спасбросок Мудрости) в ярости бьёт своих – очарован на ход."},

	// ---------- Бард ----------
	{Class: "bard", Sub: "whispers", Lvl: 3, Key: "psyblades", Name: "Психические клинки", Pool: "insp", Cost: 1,
		Rider: &riderSpec{Dice: nd(6, func(x ctx) int { return byLvl(x.Lvl, 2, [2]int{5, 3}, [2]int{10, 5}, [2]int{15, 8}) }), Type: "psychic"},
		Desc:  "Следующее попадание: +2d6 психической энергией (3d6 с 5-го, 5d6 с 10-го, 8d6 с 15-го)."},
	{Class: "bard", Sub: "whispers", Lvl: 3, Key: "terror", Name: "Слова ужаса", Pool: "terror", Cost: 1, DC: "cha",
		Special: single("wis", "Напуган", 10, true), Desc: "Шёпот пугает врага (спасбросок Мудрости, повтор в конце хода)."},
	{Class: "bard", Sub: "creation", Lvl: 3, Key: "notedestroy", Name: "Нота разрушения", Pool: "insp", Cost: 1, DC: "cha",
		Special: blast("single", "thunder", "con", true), SpecialDmg: one(inspDie),
		Desc: "Нота созидания: кубик вдохновения громом по врагу, спасбросок Телосложения – половина."},
	{Class: "bard", Sub: "creation", Lvl: 3, Key: "dancing", Name: "Танцующий предмет", Pool: "dance", Cost: 1,
		Rider: &riderSpec{Dice: plus(k("1d10"), "cha"), Type: "force", Persist: true},
		Desc:  "Оживлённый предмет бьёт вместе с вами: каждое попадание +1d10 + Харизма силовым полем, пока не отпустите."},
	{Class: "bard", Sub: "eloquence", Lvl: 3, Key: "unsettle", Name: "Тревожащие слова", Pool: "insp", Cost: 1, Target: "foe",
		Fx:   &effectSpec{Name: "Тревожащие слова", Rounds: 1, Once: true, Atk: func(x ctx) string { return "-1" + inspDie(x) }},
		Desc: "Враг вычитает кубик вдохновения из следующего броска (упрощённо – броска атаки)."},
	{Class: "bard", Sub: "eloquence", Lvl: 3, Key: "silver", Name: "Серебряный язык", Passive: true, Desc: "Проверки Убеждения и Обмана не ниже 10 – учитывайте при бросках."},

	// ---------- Жрец ----------
	{Class: "cleric", Sub: "knowledge", Lvl: 1, Key: "blessknow", Name: "Благословение знаний", Passive: true, Desc: "Два навыка знаний с экспертизой – отметьте их на листе."},
	{Class: "cleric", Sub: "knowledge", Lvl: 6, Key: "readthoughts", Name: "Канал: чтение мыслей", Pool: "cd", Cost: 1, DC: "wis",
		Special: single("wis", "Очарован", 10, true), Desc: "Канал: читаете мысли врага и подчиняете его (спасбросок Мудрости)."},
	{Class: "cleric", Sub: "nature", Lvl: 2, Key: "charmnature", Name: "Канал: очарование природы", Pool: "cd", Cost: 1, DC: "wis",
		Special: &model.Special{Mode: "area", Save: "wis", Cond: "Очарован", Rounds: 10, Repeat: true, OnlyKind: "beast"}, Desc: "Канал: звери среди врагов очарованы (спасбросок Мудрости)."},
	divine("nature", "cold", "холод"),
	{Class: "cleric", Sub: "trickery", Lvl: 2, Key: "duplicity", Name: "Канал: двойник", Pool: "cd", Cost: 1, Fx: efx("Иллюзорный двойник", 10, "Преимущество"),
		Desc: "Канал: иллюзорная копия отвлекает – преимущество на ваши атаки 10 ходов."},
	{Class: "cleric", Sub: "trickery", Lvl: 6, Key: "cloakshadow", Name: "Канал: плащ теней", Pool: "cd", Cost: 1, Fx: efx("Плащ теней", 1, "Невидим"),
		Desc: "Канал: невидимость до начала следующего хода."},
	divine("trickery", "poison", "яд"),
	{Class: "cleric", Sub: "forge", Lvl: 1, Key: "forgebless", Name: "Благословение кузни", Pool: "forge", Cost: 1, Target: "ally",
		Fx: &effectSpec{Name: "Благословение кузни", AC: func(ctx) int { return 1 }, Atk: k("1")}, Desc: "Оружие или доспех союзника +1 (упрощённо: +1 к КД и атакам до отдыха)."},
	divine("forge", "fire", "огонь"),
	{Class: "cleric", Sub: "grave", Lvl: 2, Key: "pathgrave", Name: "Канал: путь к могиле", Pool: "cd", Cost: 1, Target: "foe",
		Fx: efx("Путь к могиле", 1, "Подсвечен"), Desc: "Канал: проклятие – атаки по врагу с преимуществом до конца вашего следующего хода."},
	{Class: "cleric", Sub: "grave", Lvl: 1, Key: "spareddying", Name: "Круг смертности", Target: "ally", Heal: nd(4, func(ctx) int { return 1 }),
		Desc: "Лечение союзника на 0 HP – с максимумом кубиков (упрощённо: 1d4 сверху)."},
	{Class: "cleric", Sub: "order", Lvl: 2, Key: "orderdemand", Name: "Канал: требование порядка", Pool: "cd", Cost: 1, DC: "wis",
		Special: area("wis", "Очарован", 1, false), Desc: "Канал: враги вокруг очарованы до конца вашего следующего хода (спасбросок Мудрости)."},
	divine("order", "psychic", "психическая энергия"),
	{Class: "cleric", Sub: "peace", Lvl: 1, Key: "bond", Name: "Узы мира", Pool: "bond", Cost: 1, Target: "ally",
		Fx: &effectSpec{Name: "Узы мира", Rounds: 100, Atk: k("1d4")}, Desc: "Союзник связан узами: +1d4 к броскам атаки до конца боя."},
	{Class: "cleric", Sub: "peace", Lvl: 2, Key: "balmpeace", Name: "Канал: бальзам мира", Pool: "cd", Cost: 1, Target: "ally",
		Heal: plus(k("2d6"), "wis"), Desc: "Канал: союзник исцеляется на 2d6 + Мудрость."},
	{Class: "cleric", Sub: "twilight", Lvl: 2, Key: "sanctuary", Name: "Канал: сумеречное святилище", Pool: "cd", Cost: 1, Target: "ally",
		Temp: func(x ctx) string { return "1d6+" + strconv.Itoa(x.Lvl) }, Desc: "Канал: союзник в сумеречной сфере получает 1d6 + уровень временных HP."},
	divine("twilight", "radiant", "излучение"),

	// ---------- Друид ----------
	{Class: "druid", Sub: "dreams", Lvl: 2, Key: "balmcourt", Name: "Бальзам летнего двора", Pool: "balm", Cost: 1, Target: "ally",
		Heal: k("1d6"), Temp: k("1"), Desc: "Тратит кубик бальзама: союзник исцеляется на 1d6 (запас кубиков = уровень друида)."},
	{Class: "druid", Sub: "shepherd", Lvl: 2, Key: "bearspirit", Name: "Тотем духа: медведь", Pool: "totem", Cost: 1, Target: "ally",
		Temp: num(func(x ctx) int { return 5 + x.Lvl }), Desc: "Союзник в ауре медведя получает 5 + уровень временных HP."},
	{Class: "druid", Sub: "shepherd", Lvl: 2, Key: "hawkspirit", Name: "Тотем духа: ястреб", Pool: "totem", Cost: 1, Target: "ally",
		Fx: efx("Дух ястреба", 10, "Преимущество"), Desc: "Ястреб направляет удары союзника: преимущество на атаки 10 ходов."},
	{Class: "druid", Sub: "shepherd", Lvl: 2, Key: "unicornspirit", Name: "Тотем духа: единорог", Pool: "totem", Cost: 1, Target: "ally",
		Heal: num(func(x ctx) int { return x.Lvl }), Desc: "Дух единорога лечит союзника на уровень друида."},
	{Class: "druid", Sub: "spores", Lvl: 2, Key: "halo", Name: "Ореол спор", DC: "wis",
		Special: blast("single", "necrotic", "con", false), SpecialDmg: func(x ctx) string {
			return "1" + stepDie(x.Lvl, [2]int{1, 4}, [2]int{6, 6}, [2]int{10, 8}, [2]int{14, 10})
		},
		Desc: "Реакция: споры по врагу рядом – 1d4 некротикой (1d6 с 6-го, 1d8 с 10-го, 1d10 с 14-го), спасбросок Телосложения – без урона."},
	{Class: "druid", Sub: "spores", Lvl: 2, Key: "symbiotic", Name: "Симбиотическая сущность", Pool: "wild", Cost: 1,
		Temp: num(func(x ctx) int { return 4 * x.Lvl }), Rider: &riderSpec{Dice: k("1d6"), Type: "necrotic", Persist: true},
		Desc: "Тратит дикий облик: 4 × уровень временных HP и +1d6 некротикой к каждому попаданию, пока не отпустите."},
	{Class: "druid", Sub: "stars", Lvl: 2, Key: "archer", Name: "Звёздная форма: лучник", Pool: "wild", Cost: 1, Extra: 1,
		Rider: &riderSpec{Dice: plus(k("1d8"), "wis"), Type: "radiant"}, Desc: "Тратит дикий облик: ещё одна атака – светящаяся стрела +1d8 + Мудрость излучением."},
	{Class: "druid", Sub: "stars", Lvl: 2, Key: "chalice", Name: "Звёздная форма: чаша", Pool: "wild", Cost: 1, Target: "ally",
		Heal: plus(k("1d8"), "wis"), Desc: "Тратит дикий облик: союзник исцеляется на 1d8 + Мудрость."},
	{Class: "druid", Sub: "stars", Lvl: 2, Key: "guiding", Name: "Звёздная карта: путеводный знак", Pool: "starmap", Cost: 1,
		Fx: &effectSpec{Name: "Путеводный знак", Rounds: 1, Once: true, Atk: k("1d6")}, Desc: "+1d6 к следующему броску атаки."},
	{Class: "druid", Sub: "wildfire", Lvl: 2, Key: "firespirit", Name: "Призыв огненного духа", Pool: "wild", Cost: 1, DC: "wis",
		Special: blast("area", "fire", "dex", true), SpecialDmg: k("2d6"), Desc: "Тратит дикий облик: дух появляется во вспышке – 2d6 огнём по врагам, спасбросок Ловкости – половина."},
	{Class: "druid", Sub: "wildfire", Lvl: 2, Key: "flameseed", Name: "Огненный дух: удар пламенем",
		Rider: &riderSpec{Dice: plus(k("1d6"), "wis"), Type: "fire"}, Desc: "Дух бьёт вместе с вами: следующее попадание +1d6 + Мудрость огнём."},

	// ---------- Воин ----------
	{Class: "fighter", Sub: "champion", Lvl: 3, Key: "impcrit", Name: "Улучшенный критический удар", Passive: true, Desc: "Критический удар на 19–20 (с 15-го уровня – на 18–20). Считается автоматически."},
	{Class: "fighter", Sub: "champion", Lvl: 18, Key: "survivor", Name: "Выживший", Target: "", Heal: num(func(x ctx) int { return 5 + max(0, x.M["con"]) }),
		Desc: "В начале хода, если HP не больше половины: +5 + Телосложение HP (нажмите сами)."},
	{Class: "fighter", Sub: "arcanearcher", Lvl: 3, Key: "bursting", Name: "Стрела: взрывная", Pool: "arrow", Cost: 1,
		Rider: &riderSpec{Dice: nd(6, func(x ctx) int { return byLvl(x.Lvl, 2, [2]int{18, 4}) }), Type: "force"}, Desc: "Следующее попадание: +2d6 силовым полем (4d6 с 18-го)."},
	{Class: "fighter", Sub: "arcanearcher", Lvl: 3, Key: "grasping", Name: "Стрела: опутывающая", Pool: "arrow", Cost: 1,
		Rider: &riderSpec{Dice: nd(6, func(x ctx) int { return byLvl(x.Lvl, 2, [2]int{18, 4}) }), Type: "poison", Save: "str", Cond: "Опутан", Rounds: 10, DC: "int"},
		Desc:  "Следующее попадание: +2d6 ядом, цель опутана при провале спасброска Силы."},
	{Class: "fighter", Sub: "arcanearcher", Lvl: 3, Key: "shadowarrow", Name: "Стрела: теневая", Pool: "arrow", Cost: 1,
		Rider: &riderSpec{Dice: nd(6, func(x ctx) int { return byLvl(x.Lvl, 2, [2]int{18, 4}) }), Type: "psychic", Save: "wis", Cond: "Ослеплён", Rounds: 1, DC: "int"},
		Desc:  "Следующее попадание: +2d6 психической энергией, цель ослеплена на ход при провале спасброска Мудрости."},
	{Class: "fighter", Sub: "cavalier", Lvl: 3, Key: "unwavering", Name: "Метка кавалериста", Target: "foe", Fx: efx("Метка кавалериста", 1, "Помеха"),
		Desc: "Попадание отмечает врага: он атакует с помехой до конца вашего следующего хода."},
	{Class: "fighter", Sub: "cavalier", Lvl: 3, Key: "markstrike", Name: "Ответный удар по метке", Pool: "unwaver", Cost: 1, Extra: 1,
		Rider: &riderSpec{Dice: num(func(x ctx) int { return x.Lvl / 2 })}, Desc: "Помеченный враг ударил другого: ваша дополнительная атака с преимуществом, + половина уровня к урону."},
	{Class: "fighter", Sub: "echoknight", Lvl: 3, Key: "unleash", Name: "Высвобождение воплощения", Pool: "echo", Cost: 1, Extra: 1,
		Desc: "Эхо бьёт вместе с вами: ещё одна атака в этот ход."},
	{Class: "fighter", Sub: "echoknight", Lvl: 3, Key: "echoshield", Name: "Эхо: подмена", Fx: efx("Эхо-двойник", 1, "Уклонение"),
		Desc: "Меняетесь местами с эхом – атаки по вам с помехой до начала следующего хода (упрощённо)."},
	{Class: "fighter", Sub: "psiwarrior", Lvl: 3, Key: "psistrike", Name: "Псионный удар", Pool: "psidie", Cost: 1,
		Rider: &riderSpec{Dice: plus(one(psiDie), "int"), Type: "force"}, Desc: "Следующее попадание: + кость псионики + Интеллект силовым полем."},
	{Class: "fighter", Sub: "psiwarrior", Lvl: 3, Key: "psifield", Name: "Защитное поле", Pool: "psidie", Cost: 1, Target: "ally",
		Temp: plus(one(psiDie), "int"), Desc: "Реакция: союзник получает временные HP = кость псионики + Интеллект (вместо снижения урона)."},
	{Class: "fighter", Sub: "runeknight", Lvl: 3, Key: "giantmight", Name: "Гигантская мощь", Pool: "giant", Cost: 1,
		Rider: &riderSpec{Dice: k("1d6"), Persist: true}, Desc: "Вы вырастаете: каждое попадание +1d6, пока не отпустите (снимается повторным нажатием)."},
	{Class: "fighter", Sub: "runeknight", Lvl: 3, Key: "firerune", Name: "Руна огня", Pool: "rune", Cost: 1,
		Rider: &riderSpec{Dice: k("2d6"), Type: "fire", Save: "str", Cond: "Опутан", Rounds: 10, DC: "con"}, Desc: "Следующее попадание: +2d6 огнём, цель скована огненными путами (спасбросок Силы)."},
	{Class: "fighter", Sub: "runeknight", Lvl: 3, Key: "stonerune", Name: "Руна камня", Pool: "rune", Cost: 1, Target: "foe", DC: "con",
		Special: single("wis", "Очарован", 10, true), Desc: "Реакция: враг в трансе (спасбросок Мудрости, повтор в конце хода)."},

	// ---------- Монах ----------
	{Class: "monk", Sub: "shadow", Lvl: 3, Key: "shadowarts", Name: "Теневые искусства: тьма", Pool: "ki", Cost: 2, Fx: efx("Тьма", 10, "Невидим"),
		Desc: "2 ки: вы скрыты магической тьмой – невидимы 10 ходов (упрощённо)."},
	{Class: "monk", Sub: "shadow", Lvl: 6, Key: "shadowstep", Name: "Шаг тени", Fx: efx("Шаг тени", 1, "Преимущество"),
		Desc: "Из тени в тень: преимущество на ближайшую атаку в этот ход."},
	{Class: "monk", Sub: "drunken", Lvl: 3, Key: "drunkflurry", Name: "Пьяный шквал", Pool: "ki", Cost: 1, Extra: 2, Fx: efx("Пьяная походка", 1, "Уклонение"),
		Desc: "Две дополнительные атаки, а враги промахиваются по шатающемуся монаху (уклонение до следующего хода)."},
	{Class: "monk", Sub: "drunken", Lvl: 6, Key: "tipsy", Name: "Шатание", Passive: true, Desc: "Промах врага можно перенаправить в другого врага (решает мастер)."},
	{Class: "monk", Sub: "kensei", Lvl: 3, Key: "agileparry", Name: "Ловкое парирование", Fx: &effectSpec{Name: "Ловкое парирование", Rounds: 1, AC: func(ctx) int { return 2 }},
		Desc: "После безоружного удара: КД +2 до начала следующего хода."},
	{Class: "monk", Sub: "kensei", Lvl: 3, Key: "kenseishot", Name: "Выстрел кэнсэя", Rider: &riderSpec{Dice: k("1d4")},
		Desc: "Бонусное действие: следующее попадание дальнобойным оружием кэнсэя +1d4."},
	{Class: "monk", Sub: "kensei", Lvl: 6, Key: "sharpen", Name: "Заострить клинок", Pool: "ki", Cost: 1, Fx: &effectSpec{Name: "Заострённый клинок", Rounds: 10, Atk: k("1")},
		Rider: &riderSpec{Dice: k("1"), Persist: true}, Desc: "Ки: оружие кэнсэя +1 к атаке и урону на 10 ходов (урон – пока не снимете)."},
	{Class: "monk", Sub: "sunsoul", Lvl: 3, Key: "sunbolt", Name: "Сияющий болт", Special: blast("single", "radiant", "", false), SpecialDmg: plus(one(martialDie), "dex"),
		Desc: "Луч солнечного света по врагу: кубик боевых искусств + Ловкость излучением (упрощённо без броска атаки)."},
	{Class: "monk", Sub: "sunsoul", Lvl: 6, Key: "searing", Name: "Обжигающая вспышка", Pool: "ki", Cost: 2, DC: "wis",
		Special: blast("area", "fire", "con", true), SpecialDmg: k("2d6"), Desc: "2 ки: сфера света – 2d6 огнём по врагам, спасбросок Телосложения – половина."},
	{Class: "monk", Sub: "longdeath", Lvl: 3, Key: "touchdeath", Name: "Прикосновение смерти", Temp: num(func(x ctx) int { return max(1, x.M["wis"]+x.Lvl) }),
		Desc: "Когда враг рядом падает на 0 HP: временные HP = Мудрость + уровень (нажмите сами)."},
	{Class: "monk", Sub: "longdeath", Lvl: 6, Key: "hourreaping", Name: "Час жатвы", Pool: "ki", Cost: 1, DC: "wis",
		Special: area("wis", "Напуган", 1, false), Desc: "Ки: враги вокруг напуганы до конца вашего следующего хода (спасбросок Мудрости)."},
	{Class: "monk", Sub: "astral", Lvl: 3, Key: "astralarms", Name: "Руки астрала", Pool: "ki", Cost: 1, DC: "wis",
		Special: blast("area", "force", "dex", false), SpecialDmg: func(x ctx) string { return "2" + martialDie(x) },
		Desc: "Ки: призрачные руки – два кубика боевых искусств силовым полем по врагам рядом, спасбросок Ловкости – без урона."},
	{Class: "monk", Sub: "astral", Lvl: 3, Key: "astralblow", Name: "Удар астральной руки", Rider: &riderSpec{Dice: plus(one(martialDie), "wis"), Type: "force"},
		Desc: "Следующее попадание астральной рукой: кубик боевых искусств + Мудрость силовым полем."},

	// ---------- Паладин ----------
	{Class: "paladin", Sub: "oathbreaker", Lvl: 3, Key: "controlundead", Name: "Канал: управление нежитью", Pool: "cd", Cost: 1, DC: "cha",
		Special: &model.Special{Mode: "single", Save: "wis", Cond: "Очарован", Rounds: 100, OnlyKind: "undead"}, Desc: "Канал: нежить подчиняется вам (спасбросок Мудрости)."},
	{Class: "paladin", Sub: "oathbreaker", Lvl: 3, Key: "dreadful", Name: "Канал: ужасающий облик", Pool: "cd", Cost: 1, DC: "cha",
		Special: area("wis", "Напуган", 10, true), Desc: "Канал: враги вокруг напуганы (спасбросок Мудрости, повтор в конце хода)."},
	{Class: "paladin", Sub: "redemption", Lvl: 3, Key: "rebuke", Name: "Канал: возмездие насилию", Pool: "cd", Cost: 1, DC: "cha",
		Special: blast("single", "radiant", "wis", true), SpecialDmg: k("3d10"), Desc: "Канал: напавший получает 3d10 излучением, спасбросок Мудрости – половина (упрощённо)."},
	{Class: "paladin", Sub: "redemption", Lvl: 3, Key: "emissary", Name: "Канал: эмиссар мира", Pool: "cd", Cost: 1, Fx: efx("Эмиссар мира", 10, "Уклонение"),
		Desc: "Канал: враги не решаются бить вас – атаки по вам с помехой 10 ходов (упрощённо)."},
	{Class: "paladin", Sub: "glory", Lvl: 3, Key: "inspsmite", Name: "Канал: вдохновляющая кара", Pool: "cd", Cost: 1, Target: "ally",
		Temp: func(x ctx) string { return "2d8+" + strconv.Itoa(x.Lvl) }, Desc: "Канал после кары: союзник получает 2d8 + уровень временных HP."},
	{Class: "paladin", Sub: "glory", Lvl: 3, Key: "athlete", Name: "Канал: несравненный атлет", Pool: "cd", Cost: 1, Fx: efx("Несравненный атлет", 10, "Преимущество"),
		Desc: "Канал: 10 ходов преимущество (на атлетику и, упрощённо, на атаки)."},
	{Class: "paladin", Sub: "watchers", Lvl: 3, Key: "abjure", Name: "Канал: отторжение иномирца", Pool: "cd", Cost: 1, DC: "cha",
		Special: &model.Special{Mode: "area", Save: "wis", Cond: "Напуган", Rounds: 10, Repeat: true, OnlyKind: "fiend"}, Desc: "Канал: исчадия среди врагов в страхе (спасбросок Мудрости)."},
	{Class: "paladin", Sub: "watchers", Lvl: 3, Key: "watchwill", Name: "Канал: воля наблюдателей", Pool: "cd", Cost: 1, Target: "ally",
		Fx: efx("Воля наблюдателей", 10, "Преимущество"), Desc: "Канал: союзник собран и решителен – преимущество 10 ходов (упрощённо на атаки)."},

	// ---------- Следопыт ----------
	{Class: "ranger", Sub: "beastmaster", Lvl: 3, Key: "beastattack", Name: "Атака зверя-спутника", Special: blast("single", "piercing", "", false),
		SpecialDmg: func(x ctx) string { return "1d8+" + strconv.Itoa(x.Prof+2) }, Desc: "Спутник кусает врага: 1d8 + 2 + мастерство (упрощённо без броска атаки)."},
	{Class: "ranger", Sub: "horizonwalker", Lvl: 3, Key: "planar", Name: "Планарный воин",
		Rider: &riderSpec{Dice: nd(8, func(x ctx) int { return byLvl(x.Lvl, 1, [2]int{11, 2}) }), Type: "force"}, Desc: "Раз за ход: следующее попадание +1d8 силовым полем (2d8 с 11-го)."},
	{Class: "ranger", Sub: "monsterslayer", Lvl: 3, Key: "slayerprey", Name: "Добыча убийцы", Rider: &riderSpec{Dice: k("1d6"), Persist: true},
		Desc: "Помеченный враг: каждое попадание +1d6, пока не снимете."},
	{Class: "ranger", Sub: "feywanderer", Lvl: 3, Key: "dreadful", Name: "Жуткие удары",
		Rider: &riderSpec{Dice: func(x ctx) string { return "1" + stepDie(x.Lvl, [2]int{1, 4}, [2]int{11, 6}) }, Type: "psychic", Persist: true}, Desc: "Каждое попадание +1d4 психической энергией (1d6 с 11-го), пока не снимете."},
	{Class: "ranger", Sub: "swarmkeeper", Lvl: 3, Key: "swarm", Name: "Собранный рой",
		Rider: &riderSpec{Dice: nd(6, func(x ctx) int { return 1 }), Type: "piercing", Persist: true}, Desc: "Рой кусает вместе с вами: каждое попадание +1d6 колющим, пока не снимете."},

	// ---------- Плут ----------
	{Class: "rogue", Sub: "thief", Lvl: 3, Key: "fasthands", Name: "Быстрые руки: зелье", Heal: k("2d4+2"),
		Desc: "Бонусным действием выпить зелье лечения: 2d4 + 2 HP (зелье вычеркните из инвентаря)."},
	{Class: "rogue", Sub: "inquisitive", Lvl: 3, Key: "insightful", Name: "Проницательный боец", Target: "foe", Fx: efx("Разгадан", 10, "Подсвечен"),
		Desc: "Вы разгадали врага: атаки по нему с преимуществом 10 ходов (для скрытой атаки)."},
	{Class: "rogue", Sub: "mastermind", Lvl: 3, Key: "tactics", Name: "Мастер тактики", Target: "ally", Fx: efx("Помощь", 1, "Преимущество"),
		Desc: "Помощь бонусным действием: союзник атакует с преимуществом до конца хода."},
	{Class: "rogue", Sub: "scout", Lvl: 3, Key: "skirmish", Name: "Застрельщик", Fx: efx("Застрельщик", 1, "Уклонение"),
		Desc: "Реакция: отскок от врага – атаки по вам с помехой до начала следующего хода."},
	{Class: "rogue", Sub: "swashbuckler", Lvl: 3, Key: "footwork", Name: "Изящная работа ног", Fx: efx("Изящная работа ног", 1, "Уклонение"),
		Desc: "Ударили врага – он не может ответить: атаки по вам с помехой до начала следующего хода (упрощённо)."},
	{Class: "rogue", Sub: "swashbuckler", Lvl: 3, Key: "rakish", Name: "Дерзость", Fx: efx("Дерзость", 1, "Преимущество"),
		Desc: "Один на один: скрытая атака без преимущества – упрощённо преимущество на атаку в этот ход."},
	{Class: "rogue", Sub: "swashbuckler", Lvl: 9, Key: "panache", Name: "Панаш", DC: "cha", Special: single("wis", "Очарован", 10, true),
		Desc: "Обаяние дуэлянта: враг очарован (спасбросок Мудрости)."},
	{Class: "rogue", Sub: "phantom", Lvl: 3, Key: "wails", Name: "Стенания из могилы", Pool: "wails", Cost: 1,
		Special: blast("single", "necrotic", "", false), SpecialDmg: nd(6, func(x ctx) int { return max(1, (x.Lvl+1)/4) }),
		Desc: "После скрытой атаки: второй враг получает половину кубиков скрытой атаки некротикой."},

	// ---------- Чародей ----------
	{Class: "sorcerer", Sub: "draconic", Lvl: 1, Key: "draconicres", Name: "Драконья стойкость", Passive: true, Desc: "+1 HP за уровень и КД 13 + Ловкость без доспехов – считается автоматически."},
	{Class: "sorcerer", Sub: "draconic", Lvl: 6, Key: "affinity", Name: "Стихийное родство", Pool: "sorc", Cost: 1,
		Rider: &riderSpec{Dice: num(func(x ctx) int { return max(1, x.M["cha"]) }), Type: "fire"}, Desc: "Очко чародейства: ближайший урон +Харизма стихией предка."},
	{Class: "sorcerer", Sub: "divine", Lvl: 1, Key: "favored", Name: "Предпочтение богов", Pool: "favored", Cost: 1,
		Fx: &effectSpec{Name: "Предпочтение богов", Rounds: 1, Once: true, Atk: k("2d4")}, Desc: "+2d4 к следующему броску атаки (после промаха)."},
	{Class: "sorcerer", Sub: "divine", Lvl: 6, Key: "empowered", Name: "Исцеление сверхъестественное", Pool: "sorc", Cost: 1, Target: "ally",
		Heal: k("2d8"), Desc: "Очко чародейства: союзник исцеляется на 2d8 (упрощённо)."},
	{Class: "sorcerer", Sub: "shadow", Lvl: 3, Key: "darkness", Name: "Тьма чародея", Pool: "sorc", Cost: 2, Fx: efx("Тьма чародея", 10, "Невидим"),
		Desc: "2 очка: тьма, сквозь которую видите только вы – невидимость 10 ходов (упрощённо)."},
	{Class: "sorcerer", Sub: "shadow", Lvl: 6, Key: "hound", Name: "Пёс дурного предзнаменования", Pool: "sorc", Cost: 3,
		Special: blast("single", "psychic", "", false), SpecialDmg: k("2d6+3"), Desc: "3 очка: теневой пёс терзает врага – 2d6 + 3 психической энергией."},
	{Class: "sorcerer", Sub: "storm", Lvl: 6, Key: "heartstorm", Name: "Сердце бури", Special: blast("area", "lightning", "", false),
		SpecialDmg: num(func(x ctx) int { return max(1, x.Lvl/2) }), Desc: "После заклинания молнии или грома: половина уровня урона молнией врагам рядом."},
	{Class: "sorcerer", Sub: "storm", Lvl: 1, Key: "tempestuous", Name: "Буйная магия", Fx: efx("Буйная магия", 1, "Уклонение"),
		Desc: "После заклинания – взлёт на 10 футов: атаки по вам с помехой до начала хода (упрощённо)."},
	{Class: "sorcerer", Sub: "aberrant", Lvl: 1, Key: "mindsliver", Name: "Осколок разума", DC: "cha",
		Special: blast("single", "psychic", "int", false), SpecialDmg: nd(6, func(x ctx) int { return tier(x.Lvl) }),
		Desc: "Заговор: 1d6 психической энергией (растёт с уровнем), спасбросок Интеллекта – без урона."},
	{Class: "sorcerer", Sub: "aberrant", Lvl: 6, Key: "revelation", Name: "Откровение плоти", Pool: "sorc", Cost: 1, Fx: efx("Откровение плоти", 10, "Уклонение"),
		Desc: "Очко: чешуя и щупальца – атаки по вам с помехой 10 ходов (упрощённо)."},
	{Class: "sorcerer", Sub: "clockwork", Lvl: 1, Key: "restorebal", Name: "Восстановление равновесия", Pool: "balance", Cost: 1, Target: "ally",
		Fx: efx("Равновесие", 1, "Уклонение"), Desc: "Реакция: гасите преимущество врага – атаки по союзнику с помехой до конца хода (упрощённо)."},
	{Class: "sorcerer", Sub: "clockwork", Lvl: 6, Key: "bastion", Name: "Бастион закона", Pool: "sorc", Cost: 2, Target: "ally",
		Temp: k("2d8"), Desc: "2 очка: союзник получает щит – 2d8 временных HP."},

	// ---------- Колдун ----------
	{Class: "warlock", Sub: "greatold", Lvl: 1, Key: "awakened", Name: "Пробуждённый разум", Passive: true, Desc: "Телепатия на 30 футов."},
	{Class: "warlock", Sub: "greatold", Lvl: 1, Key: "whispers", Name: "Шёпот безумия", DC: "cha", Special: single("wis", "Напуган", 1, false),
		Desc: "Чужие мысли в голове врага: напуган до конца вашего хода (спасбросок Мудрости)."},
	{Class: "warlock", Sub: "greatold", Lvl: 6, Key: "entropic", Name: "Энтропийная защита", Pool: "entropy", Cost: 1, Fx: efx("Энтропийная защита", 1, "Уклонение"),
		Desc: "Реакция: атака по вам с помехой; при промахе – ваша следующая атака с преимуществом."},
	{Class: "warlock", Sub: "celestial", Lvl: 1, Key: "healinglight", Name: "Исцеляющий свет", Pool: "hlight", Cost: 1, Target: "ally",
		Heal: k("1d6"), Desc: "Тратит кубик света: союзник исцеляется на 1d6 (запас кубиков = 1 + уровень)."},
	{Class: "warlock", Sub: "celestial", Lvl: 6, Key: "radiantsoul", Name: "Сияющая душа", Rider: &riderSpec{Dice: num(func(x ctx) int { return max(1, x.M["cha"]) }), Type: "radiant"},
		Desc: "Раз за ход: ближайший урон огнём или излучением +Харизма."},
	{Class: "warlock", Sub: "fathomless", Lvl: 1, Key: "tentacle", Name: "Щупальце бездны", Pool: "tentacle", Cost: 1,
		Special: blast("single", "cold", "", false), SpecialDmg: nd(8, func(x ctx) int { return byLvl(x.Lvl, 1, [2]int{10, 2}) }),
		Desc: "Призрачное щупальце бьёт врага: 1d8 холодом (2d8 с 10-го)."},
	{Class: "warlock", Sub: "genie", Lvl: 1, Key: "geniewrath", Name: "Гнев джинна", Rider: &riderSpec{Dice: num(prof), Type: "force"},
		Desc: "Раз за ход: ближайшее попадание + бонус мастерства (тип – по вашему джинну)."},
	{Class: "warlock", Sub: "undying", Lvl: 6, Key: "defydeath", Name: "Отказ от смерти", Pool: "undying", Cost: 1,
		Heal: plus(k("1d8"), "con"), Desc: "Когда спасаетесь от смерти: исцеление 1d8 + Телосложение."},
	{Class: "warlock", Sub: "undying", Lvl: 1, Key: "amongdead", Name: "Среди мёртвых", Passive: true, Desc: "Нежить с трудом решается напасть на вас (спасбросок Мудрости – решает мастер)."},

	// ---------- Волшебник ----------
	{Class: "wizard", Sub: "evocation", Lvl: 2, Key: "sculpt", Name: "Ваяние заклинаний", Passive: true, Desc: "Союзники в области ваших заклинаний воплощения не страдают – не выбирайте их целями."},
	{Class: "wizard", Sub: "evocation", Lvl: 10, Key: "empevoc", Name: "Усиленное воплощение", Rider: &riderSpec{Dice: num(func(x ctx) int { return max(1, x.M["int"]) }), Type: "force"},
		Desc: "Урон заклинания воплощения +Интеллект (нажмите перед заклинанием-атакой)."},
	{Class: "wizard", Sub: "conjuration", Lvl: 6, Key: "transpose", Name: "Благотворное перемещение", Pool: "teleport", Cost: 1, Fx: efx("Перемещение", 1, "Уклонение"),
		Desc: "Телепорт на 30 футов: атаки по вам с помехой до начала хода (упрощённо)."},
	{Class: "wizard", Sub: "conjuration", Lvl: 2, Key: "minorconj", Name: "Малое вызывание", Passive: true, Desc: "Создаёт простой предмет на час."},
	{Class: "wizard", Sub: "divination", Lvl: 2, Key: "portentfoe", Name: "Знамение: врагу", Pool: "portent", Cost: 1, Target: "foe",
		Fx: &effectSpec{Name: "Дурное знамение", Rounds: 10, Once: true, Atk: k("-1d10")}, Desc: "Видение будущего: следующий бросок атаки врага −1d10 (упрощённо вместо замены броска)."},
	{Class: "wizard", Sub: "divination", Lvl: 2, Key: "portentally", Name: "Знамение: союзнику", Pool: "portent", Cost: 1, Target: "ally",
		Fx: &effectSpec{Name: "Доброе знамение", Rounds: 10, Once: true, Atk: k("1d10")}, Desc: "Следующий бросок атаки союзника +1d10 (упрощённо)."},
	{Class: "wizard", Sub: "enchantment", Lvl: 2, Key: "hypnotic", Name: "Гипнотический взгляд", Pool: "gaze", Cost: 1, DC: "int",
		Special: single("wis", "Недееспособен", 1, false), Desc: "Враг рядом заворожён и недееспособен до конца вашего хода (спасбросок Мудрости)."},
	{Class: "wizard", Sub: "illusion", Lvl: 2, Key: "illusoryself", Name: "Иллюзорный двойник", Pool: "illusory", Cost: 1, Fx: efx("Иллюзорный двойник", 1, "Уклонение"),
		Desc: "Иллюзия принимает удар: атаки по вам с помехой до начала хода (упрощённо)."},
	{Class: "wizard", Sub: "necromancy", Lvl: 2, Key: "harvest", Name: "Мрачная жатва", Heal: num(func(x ctx) int { return max(2, x.Lvl/2) }),
		Desc: "Заклинание убило врага: вы исцеляетесь (упрощённо – на половину уровня, не меньше 2)."},
	{Class: "wizard", Sub: "transmutation", Lvl: 6, Key: "stone", Name: "Камень превращения", Pool: "stone", Cost: 1, Target: "ally",
		Temp: num(func(x ctx) int { return x.Lvl }), Desc: "Камень даёт стойкость: союзник получает временные HP, равные вашему уровню (упрощённо)."},
	{Class: "wizard", Sub: "warmagic", Lvl: 2, Key: "deflection", Name: "Тайный отпор", Fx: &effectSpec{Name: "Тайный отпор", Rounds: 1, AC: func(ctx) int { return 2 }},
		Desc: "Реакция: КД +2 до начала вашего хода."},
	{Class: "wizard", Sub: "chronurgy", Lvl: 2, Key: "chronal", Name: "Хроносдвиг", Pool: "chrono", Cost: 1, Target: "foe",
		Fx: &effectSpec{Name: "Хроносдвиг", Rounds: 1, Once: true, Atk: k("-1d6")}, Desc: "Реакция: время сдвигается – следующий бросок атаки врага −1d6 (упрощённо вместо переброса)."},

	// ---------- Изобретатель ----------
	{Class: "artificer", Sub: "armorer", Lvl: 3, Key: "thunder", Name: "Доспех стража: громовая перчатка",
		Rider: &riderSpec{Dice: k("1d8"), Type: "thunder", Persist: true}, Desc: "Каждое попадание перчаткой +1d8 громом, пока не снимете."},
	{Class: "artificer", Sub: "armorer", Lvl: 3, Key: "defield", Name: "Защитное поле", Pool: "field", Cost: 1,
		Temp: num(func(x ctx) int { return x.Lvl }), Desc: "Бонусное действие: временные HP, равные уровню изобретателя."},
	{Class: "artificer", Sub: "artillerist", Lvl: 3, Key: "flamethrower", Name: "Пушка: огнемёт", DC: "int",
		Special: blast("area", "fire", "dex", true), SpecialDmg: nd(8, func(x ctx) int { return byLvl(x.Lvl, 2, [2]int{9, 3}) }),
		Desc: "Волшебная пушка: 2d8 огнём по врагам (3d8 с 9-го), спасбросок Ловкости – половина."},
	{Class: "artificer", Sub: "artillerist", Lvl: 3, Key: "protector", Name: "Пушка: защитник", Target: "ally",
		Temp: plus(k("1d8"), "int"), Desc: "Пушка окутывает союзника: 1d8 + Интеллект временных HP."},
	{Class: "artificer", Sub: "battlesmith", Lvl: 3, Key: "defender", Name: "Стальной защитник", Special: blast("single", "force", "", false),
		SpecialDmg: func(x ctx) string { return "1d8+" + strconv.Itoa(x.Prof) }, Desc: "Стальной защитник бьёт врага: 1d8 + мастерство силовым полем (упрощённо без броска атаки)."},
	{Class: "artificer", Sub: "battlesmith", Lvl: 3, Key: "battleready", Name: "Боевая готовность", Passive: true, Desc: "Магическое оружие – атаки Интеллектом (выберите характеристику оружия)."},

	// ---------- Охотник на кровь ----------
	{Class: "bloodhunter", Sub: "ghostslayer", Lvl: 3, Key: "ritedawn", Name: "Обряд рассвета", HPCost: one(bloodDie),
		Rider: &riderSpec{Dice: one(bloodDie), Type: "radiant", Persist: true}, Desc: "Цена: кубик обряда в HP. Каждое попадание + кубик излучением (против нежити – вдвойне решает мастер)."},
	{Class: "bloodhunter", Sub: "lycan", Lvl: 3, Key: "hybrid", Name: "Гибридная форма", Pool: "hybrid", Cost: 1,
		Fx: &effectSpec{Name: "Гибридная форма", Rounds: 100, AC: func(ctx) int { return 1 }}, Rider: &riderSpec{Dice: k("1d6"), Persist: true},
		Desc: "Обращение: КД +1 и +1d6 к каждому попаданию когтями до конца боя (снимается повторным нажатием)."},
	{Class: "bloodhunter", Sub: "mutant", Lvl: 3, Key: "potency", Name: "Мутаген: мощь", Pool: "mutagen", Cost: 1,
		Fx: &effectSpec{Name: "Мутаген мощи", Rounds: 100, Atk: k("1")}, Desc: "Мутаген: +1 к атакам до конца боя (побочный эффект – на усмотрение мастера)."},
	{Class: "bloodhunter", Sub: "mutant", Lvl: 3, Key: "aether", Name: "Мутаген: стойкость", Pool: "mutagen", Cost: 1,
		Temp: num(func(x ctx) int { return x.Lvl }), Desc: "Мутаген: временные HP, равные уровню."},
	{Class: "bloodhunter", Sub: "profane", Lvl: 3, Key: "possession", Name: "Кровавое проклятие одержимости", HPCost: one(bloodDie), DC: "int",
		Special: single("wis", "Напуган", 1, false), Desc: "Цена: кубик обряда в HP. Дух-спутник пугает врага (спасбросок Мудрости)."},
	{Class: "bloodhunter", Sub: "profane", Lvl: 3, Key: "spiritstrike", Name: "Удар духа-спутника", Special: blast("single", "necrotic", "", false),
		SpecialDmg: one(bloodDie), Desc: "Дух-спутник бьёт врага: кубик обряда некротикой."},

	// ---------- Мистик (неофициальный класс) ----------
	{Class: "mystic", Lvl: 1, Key: "psyfocus", Name: "Психический фокус", Pool: "psi", Cost: 1, Temp: num(func(x ctx) int { return x.Lvl + max(0, x.M["int"]) }),
		Desc: "1 очко: сосредоточенность – временные HP = уровень + Интеллект."},
	{Class: "mystic", Lvl: 1, Key: "mindspike", Name: "Удар разума", Pool: "psi", Cost: 1, DC: "int",
		Special: blast("single", "psychic", "wis", true), SpecialDmg: nd(8, func(x ctx) int { return tier(x.Lvl) + 1 }),
		Desc: "1 очко: психический удар по врагу – 2d8 (растёт с уровнем), спасбросок Мудрости – половина."},
	{Class: "mystic", Sub: "immortal", Lvl: 3, Key: "immortalwill", Name: "Стойкость бессмертных", Pool: "psi", Cost: 2,
		Temp: num(func(x ctx) int { return 2 * x.Lvl }), Desc: "2 очка: временные HP = 2 × уровень."},
	{Class: "mystic", Sub: "immortal", Lvl: 6, Key: "unbreak", Name: "Несокрушимость", Passive: true, Desc: "Раз за отдых остаётесь на ногах с 1 HP (решает мастер)."},
	{Class: "mystic", Sub: "awakened", Lvl: 3, Key: "psyattack", Name: "Психическая атака", Pool: "psi", Cost: 2, DC: "int",
		Special: blast("single", "psychic", "int", true), SpecialDmg: nd(10, func(x ctx) int { return tier(x.Lvl) + 1 }),
		Desc: "2 очка: дистанционный психический удар – 2d10 (растёт), спасбросок Интеллекта – половина."},
	{Class: "mystic", Sub: "awakened", Lvl: 6, Key: "mindwave", Name: "Волна разума", Pool: "psi", Cost: 4, DC: "int",
		Special: area("int", "Ошеломлён", 1, false), Desc: "4 очка: волна ошеломляет врагов до конца вашего хода (спасбросок Интеллекта)."},
	{Class: "mystic", Sub: "elements", Lvl: 3, Key: "firestream", Name: "Огненный поток", Pool: "psi", Cost: 2, DC: "int",
		Special: blast("area", "fire", "dex", true), SpecialDmg: nd(6, func(x ctx) int { return 2 + tier(x.Lvl) }), Desc: "2 очка: поток огня – 3d6 (растёт), спасбросок Ловкости – половина."},
	{Class: "mystic", Sub: "elements", Lvl: 6, Key: "elemshield", Name: "Стихийный щит", Pool: "psi", Cost: 1,
		Temp: num(func(x ctx) int { return x.Lvl }), Fx: &effectSpec{Name: "Стихийный щит", Rounds: 10, AC: func(ctx) int { return 1 }},
		Desc: "1 очко: КД +1 на 10 ходов и временные HP = уровень."},

	// ---------- Стрелок (неофициальный класс) ----------
	{Class: "gunslinger", Lvl: 2, Key: "deadeye", Name: "Трюк: меткий выстрел", Pool: "grit", Cost: 1,
		Fx: &effectSpec{Name: "Меткий выстрел", Rounds: 1, Once: true, Atk: k("1d10")}, Desc: "1 выдержка: +1d10 к следующему броску атаки."},
	{Class: "gunslinger", Lvl: 2, Key: "disarm", Name: "Трюк: обезоруживающий выстрел", Pool: "grit", Cost: 1,
		Rider: &riderSpec{Save: "str", Cond: "Помеха", Rounds: 1, DC: "dex"}, Desc: "1 выдержка: попадание выбивает оружие – цель атакует с помехой (спасбросок Силы)."},
	{Class: "gunslinger", Sub: "sharpshooter", Lvl: 3, Key: "headshot", Name: "Выстрел в голову", Pool: "grit", Cost: 1,
		Rider: &riderSpec{Dice: k("2d10")}, Desc: "1 выдержка: следующее попадание +2d10."},
	{Class: "gunslinger", Sub: "sharpshooter", Lvl: 7, Key: "weakspot", Name: "Выстрел в слабое место", Pool: "grit", Cost: 2,
		Rider: &riderSpec{Crit: true}, Desc: "2 выдержки: следующее попадание – критическое."},
	{Class: "gunslinger", Sub: "desperado", Lvl: 3, Key: "lightning", Name: "Быстрый как молния", Pool: "grit", Cost: 1, Extra: 1,
		Desc: "1 выдержка: ещё один выстрел из пистолета в этот ход."},
	{Class: "gunslinger", Sub: "desperado", Lvl: 7, Key: "smoke", Name: "Дым и пули", Pool: "grit", Cost: 2, DC: "dex",
		Special: &model.Special{Mode: "area", Dmg: "2d6", Type: "piercing", Save: "dex", Half: true, Cond: "Ослеплён", Rounds: 1},
		Desc:    "2 выдержки: шквал огня в дыму – 2d6 по врагам и ослепление на ход (спасбросок Ловкости – половина без ослепления)."},
	{Class: "gunslinger", Sub: "trickshot", Lvl: 3, Key: "ricochet", Name: "Рикошет", Pool: "grit", Cost: 1,
		Rider: &riderSpec{Dice: k("1d8")}, Desc: "1 выдержка: выстрел от стены в укрывшегося врага – попадание +1d8."},
	{Class: "gunslinger", Sub: "trickshot", Lvl: 7, Key: "firevolley", Name: "Огненный залп", Pool: "grit", Cost: 2,
		Rider: &riderSpec{Dice: k("2d6"), Type: "fire"}, Desc: "2 выдержки: зажигательный заряд – следующее попадание +2d6 огнём."},
}
