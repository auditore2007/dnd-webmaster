package bestiary

import "heroesbook/internal/model"

type sp = model.Special

// poison — укус или жало с ядом: попадание наносит обычный урон, затем спасбросок Телосложения против ядовитого урона.
func poison(key, name, dmg, ty, xdmg string, dc int) sp {
	return sp{Key: key, Name: name, Mode: "single", Attack: true, Dmg: dmg, Type: ty, Save: "con", DC: dc, XDmg: xdmg, XType: "poison", Half: true}
}

// breath — дыхание: дальняя область по всем врагам, перезарядка 5–6.
func breath(name, dmg, ty, save string, dc int) sp {
	return sp{Key: "breath", Name: name, Mode: "area", Dmg: dmg, Type: ty, Save: save, DC: dc, Half: true, Recharge: true}
}

// fright — устрашающее присутствие.
func fright(dc int) sp {
	return sp{Key: "fright", Name: "Устрашающее присутствие", Mode: "area", Save: "wis", DC: dc, Cond: "Напуган", Rounds: 10, Repeat: true, Once: true}
}

// Особые способности существ, которые применяются из боя одной кнопкой. Значения СЛ и урона — по памяти Monster Manual.
var builtin = map[string][]sp{
	"poison-snake": {poison("bite", "Ядовитый укус", "1d4", "piercing", "2d4", 10)},
	"giant-wasp":   {poison("sting", "Ядовитое жало", "1d6", "piercing", "3d6", 11)},
	"imp":          {poison("sting", "Жало", "1d4", "piercing", "3d6", 11)},
	"giant-spider": {poison("bite", "Ядовитый укус", "1d8", "piercing", "2d8", 11),
		{Key: "web", Name: "Паутина", Mode: "single", Attack: true, Cond: "Опутан", Repeat: true, Save: "", RSave: "str", DC: 12, Rounds: 0, Recharge: true, Desc: "Опутанный освобождается проверкой Силы СЛ 12."}},
	"giant-scorpion":    {poison("sting", "Жало", "1d10", "piercing", "4d10", 12)},
	"wyvern":            {poison("sting", "Ядовитое жало", "2d6", "piercing", "7d6", 15)},
	"phase-spider":      {poison("bite", "Ядовитый укус", "1d10", "piercing", "4d8", 11)},
	"pseudodragon":      {{Key: "sting", Name: "Усыпляющее жало", Mode: "single", Attack: true, Dmg: "1d4", Type: "piercing", Save: "con", DC: 11, Cond: "Отравлен", Rounds: 10, Repeat: true}},
	"ghoul":             {{Key: "claws", Name: "Парализующие когти", Mode: "single", Attack: true, Dmg: "2d4", Type: "slashing", Save: "con", DC: 10, Cond: "Парализован", Rounds: 10, Repeat: true}},
	"ghast":             {{Key: "claws", Name: "Парализующие когти", Mode: "single", Attack: true, Dmg: "3d6", Type: "slashing", Save: "con", DC: 10, Cond: "Парализован", Rounds: 10, Repeat: true}},
	"cockatrice":        {{Key: "bite", Name: "Окаменяющий укус", Mode: "single", Attack: true, Dmg: "1d4", Type: "piercing", Save: "con", DC: 11, Cond: "Опутан", Rounds: 10, Repeat: true, Desc: "При втором провале — окаменение: решает мастер."}},
	"basilisk":          {{Key: "gaze", Name: "Окаменяющий взгляд", Mode: "single", Save: "con", DC: 12, Cond: "Опутан", Rounds: 10, Repeat: true, Desc: "Опутанный окаменевает при втором провале."}},
	"medusa":            {{Key: "gaze", Name: "Окаменяющий взгляд", Mode: "area", Save: "con", DC: 14, Cond: "Опутан", Rounds: 10, Repeat: true, Desc: "Опутанный окаменевает при втором провале."}},
	"wolf":              {{Key: "trip", Name: "Укус с подсечкой", Mode: "single", Attack: true, Dmg: "2d4", Type: "piercing", Save: "str", DC: 11, Cond: "Сбит с ног", Rounds: 1}},
	"dire-wolf":         {{Key: "trip", Name: "Укус с подсечкой", Mode: "single", Attack: true, Dmg: "2d6", Type: "piercing", Save: "str", DC: 13, Cond: "Сбит с ног", Rounds: 1}},
	"constrictor-snake": {{Key: "squeeze", Name: "Сдавливание", Mode: "single", Attack: true, Dmg: "1d8", Type: "bludgeoning", Cond: "Схвачен", Repeat: true, RSave: "str", DC: 14, Desc: "Схваченный освобождается проверкой Силы или Ловкости СЛ 14."}},
	"gelatinous-cube":   {{Key: "engulf", Name: "Поглощение", Mode: "single", Save: "dex", DC: 12, XDmg: "6d6", XType: "acid", Cond: "Схвачен", Repeat: true, RSave: "str", Desc: "Поглощённый получает кислоту в начале каждого хода куба."}},
	"harpy":             {{Key: "song", Name: "Завлекающая песня", Mode: "area", Save: "wis", DC: 11, Cond: "Очарован", Rounds: 10, Repeat: true}},
	"mummy": {{Key: "glare", Name: "Ужасающий взгляд", Mode: "single", Save: "wis", DC: 11, Cond: "Напуган", Rounds: 2},
		{Key: "fist", Name: "Гнилой кулак", Mode: "single", Attack: true, Dmg: "2d6", Type: "bludgeoning", Save: "con", DC: 12, XDmg: "3d6", XType: "necrotic"}},
	"ghost":         {{Key: "visage", Name: "Ужасный лик", Mode: "area", Save: "wis", DC: 13, Cond: "Напуган", Rounds: 10, Repeat: true}},
	"banshee":       {{Key: "visage", Name: "Ужасный лик", Mode: "area", Save: "wis", DC: 13, Cond: "Напуган", Rounds: 10, Repeat: true}},
	"cloaker":       {{Key: "moan", Name: "Стон", Mode: "area", Save: "wis", DC: 13, Cond: "Напуган", Rounds: 2, Recharge: true}},
	"succubus":      {{Key: "charm", Name: "Чарующий взгляд", Mode: "single", Save: "wis", DC: 15, Cond: "Очарован", Rounds: 100, Repeat: true}},
	"dryad":         {{Key: "charm", Name: "Чарующая речь", Mode: "single", Save: "wis", DC: 14, Cond: "Очарован", Rounds: 100, Repeat: true}},
	"ankheg":        {{Key: "spray", Name: "Кислотный плевок", Mode: "area", Dmg: "3d6", Type: "acid", Save: "dex", DC: 13, Half: true, Recharge: true}},
	"hell-hound":    {breath("Огненное дыхание", "6d6", "fire", "dex", 12)},
	"chimera":       {breath("Огненное дыхание", "7d8", "fire", "dex", 15)},
	"behir":         {breath("Дыхание молнии", "12d10", "lightning", "dex", 16)},
	"dragon-turtle": {breath("Паровое дыхание", "15d6", "fire", "con", 18)},
	"flameskull":    {{Key: "fireball", Name: "Огненный шар", Mode: "area", Dmg: "8d6", Type: "fire", Save: "dex", DC: 13, Half: true, Once: true}},
	"death-knight":  {{Key: "hellfire", Name: "Адский шар", Mode: "area", Dmg: "10d6", Type: "fire", XDmg: "10d6", XType: "necrotic", Save: "dex", DC: 18, Half: true, Once: true}},
	"pit-fiend":     {fright(21)},
	"tarrasque":     {fright(17)},

	"young-white-dragon":   {breath("Ледяное дыхание", "10d8", "cold", "con", 15)},
	"young-black-dragon":   {breath("Кислотное дыхание", "11d8", "acid", "dex", 14)},
	"young-green-dragon":   {breath("Ядовитое дыхание", "12d6", "poison", "con", 14)},
	"young-blue-dragon":    {breath("Дыхание молнии", "12d10", "lightning", "dex", 16)},
	"young-red-dragon":     {breath("Огненное дыхание", "16d6", "fire", "dex", 17)},
	"young-gold-dragon":    {breath("Огненное дыхание", "16d6", "fire", "dex", 17)},
	"adult-white-dragon":   {breath("Ледяное дыхание", "12d8", "cold", "con", 19), fright(14)},
	"adult-black-dragon":   {breath("Кислотное дыхание", "12d8", "acid", "dex", 18), fright(16)},
	"adult-green-dragon":   {breath("Ядовитое дыхание", "16d6", "poison", "con", 18), fright(16)},
	"adult-red-dragon":     {breath("Огненное дыхание", "18d6", "fire", "dex", 21), fright(19)},
	"ancient-white-dragon": {breath("Ледяное дыхание", "16d8", "cold", "con", 22), fright(16)},
	"ancient-black-dragon": {breath("Кислотное дыхание", "15d8", "acid", "dex", 22), fright(18)},
	"ancient-red-dragon":   {breath("Огненное дыхание", "26d6", "fire", "dex", 24), fright(21)},
}
