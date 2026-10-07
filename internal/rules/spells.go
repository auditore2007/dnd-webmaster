package rules

import (
	"slices"
	"strconv"
	"strings"
)

// SpellDef – заклинание каталога. Mode задаёт автоматизацию: attack (бросок атаки заклинанием), save (спасбросок цели),
// auto (без броска, магический снаряд), heal (лечение). Пустой Mode – описательное заклинание: тратится ячейка, эффект решает мастер.
type SpellDef struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Level   int      `json:"level"`
	School  string   `json:"school"`
	Classes []string `json:"classes"`
	Desc    string   `json:"desc"`
	Mode    string   `json:"mode"`
	Area    bool     `json:"area"`
	Dmg     string   `json:"-"`
	Type    string   `json:"-"`
	Save    string   `json:"-"`
	Half    bool     `json:"-"`
	Up      string   `json:"-"` // доп. кубики за уровень ячейки выше базового; "ray" – доп. луч
	Scale   bool     `json:"-"` // заговор растёт с 5/11/17 уровня
	Rays    int      `json:"-"`
	Beams   bool     `json:"-"`       // число лучей растёт как у заговора
	Conc    bool     `json:"conc"`    // требует концентрации
	Tgt     string   `json:"targets"` // ally | foe | self – кого выбирать для эффекта, если он есть
}

var (
	cSW = []string{"sorcerer", "wizard"}
)

var spellDefs = []SpellDef{
	{ID: "firebolt", Name: "Огненный снаряд", School: "Воплощение", Classes: cSW, Mode: "attack", Dmg: "1d10", Type: "fire", Scale: true, Desc: "Дальняя атака заклинанием: 1d10 огнём, растёт с 5/11/17 уровня."},
	{ID: "sacredflame", Name: "Священное пламя", School: "Воплощение", Classes: []string{"cleric"}, Mode: "save", Save: "dex", Dmg: "1d8", Type: "radiant", Scale: true, Desc: "Спасбросок Ловкости, иначе 1d8 излучением."},
	{ID: "eldritchblast", Name: "Мистический заряд", School: "Воплощение", Classes: []string{"warlock"}, Mode: "attack", Dmg: "1d10", Type: "force", Beams: true, Desc: "Луч 1d10 силой; лучей 2/3/4 с 5/11/17 уровня."},
	{ID: "rayoffrost", Name: "Луч холода", School: "Воплощение", Classes: cSW, Mode: "attack", Dmg: "1d8", Type: "cold", Scale: true, Desc: "Атака заклинанием: 1d8 холодом, скорость цели падает."},
	{ID: "shockinggrasp", Name: "Шоковое прикосновение", School: "Воплощение", Classes: cSW, Mode: "attack", Dmg: "1d8", Type: "lightning", Scale: true, Desc: "Атака заклинанием: 1d8 молнией, цель теряет реакцию."},
	{ID: "viciousmockery", Name: "Едкая насмешка", School: "Очарование", Classes: []string{"bard"}, Mode: "save", Save: "wis", Dmg: "1d4", Type: "psychic", Scale: true, Desc: "Спасбросок Мудрости, иначе 1d4 психическим и помеха на следующую атаку."},
	{ID: "poisonspray", Name: "Ядовитые брызги", School: "Вызов", Classes: []string{"druid", "sorcerer", "warlock", "wizard"}, Mode: "save", Save: "con", Dmg: "1d12", Type: "poison", Scale: true, Desc: "Спасбросок Телосложения, иначе 1d12 ядом."},
	{ID: "acidsplash", Name: "Брызги кислоты", School: "Вызов", Classes: cSW, Mode: "save", Save: "dex", Dmg: "1d6", Type: "acid", Scale: true, Desc: "Спасбросок Ловкости, иначе 1d6 кислотой."},
	{ID: "magehand", Name: "Волшебная рука", School: "Вызов", Classes: []string{"bard", "sorcerer", "warlock", "wizard"}, Desc: "Призрачная рука двигает предметы до 10 фунтов."},
	{ID: "light", Name: "Свет", School: "Воплощение", Classes: []string{"bard", "cleric", "sorcerer", "wizard"}, Desc: "Предмет светится как факел час."},
	{ID: "prestidigitation", Name: "Фокусы", School: "Преобразование", Classes: []string{"bard", "sorcerer", "warlock", "wizard"}, Desc: "Мелкие магические трюки."},
	{ID: "guidance", Name: "Указание", School: "Прорицание", Classes: []string{"cleric", "druid"}, Desc: "Союзник добавляет 1d4 к проверке."},
	{ID: "magicmissile", Name: "Волшебная стрела", Level: 1, School: "Воплощение", Classes: cSW, Mode: "auto", Dmg: "1d4+1", Type: "force", Desc: "Три снаряда по 1d4+1 силой без броска; +1 снаряд за ячейку выше."},
	{ID: "burninghands", Name: "Огненные ладони", Level: 1, School: "Воплощение", Classes: cSW, Mode: "save", Save: "dex", Dmg: "3d6", Type: "fire", Half: true, Area: true, Up: "1d6", Desc: "Конус огня 15 футов, 3d6; половина при спасброске."},
	{ID: "thunderwave", Name: "Громовая волна", Level: 1, School: "Воплощение", Classes: []string{"bard", "druid", "sorcerer", "wizard"}, Mode: "save", Save: "con", Dmg: "2d8", Type: "thunder", Half: true, Area: true, Up: "1d8", Desc: "Волна 15 футов, 2d8 звуком и отбрасывание."},
	{ID: "guidingbolt", Name: "Направляющий снаряд", Level: 1, School: "Воплощение", Classes: []string{"cleric"}, Mode: "attack", Dmg: "4d6", Type: "radiant", Up: "1d6", Desc: "Атака заклинанием 4d6 излучением, следующая атака по цели с преимуществом."},
	{ID: "inflictwounds", Name: "Нанесение ран", Level: 1, School: "Некромантия", Classes: []string{"cleric"}, Mode: "attack", Dmg: "3d10", Type: "necrotic", Up: "1d10", Desc: "Атака ближнего боя заклинанием: 3d10 некротическим."},
	{ID: "curewounds", Name: "Лечение ран", Level: 1, School: "Воплощение", Classes: []string{"bard", "cleric", "druid", "paladin", "ranger"}, Mode: "heal", Dmg: "1d8", Up: "1d8", Desc: "Касание: 1d8 + модификатор колдовства HP."},
	{ID: "healingword", Name: "Лечащее слово", Level: 1, School: "Воплощение", Classes: []string{"bard", "cleric", "druid"}, Mode: "heal", Dmg: "1d4", Up: "1d4", Desc: "Бонусным действием на 60 футов: 1d4 + модификатор HP."},
	{ID: "shield", Name: "Щит", Level: 1, School: "Ограждение", Classes: cSW, Desc: "Реакция: +5 к КД до начала вашего хода."},
	{ID: "magearmor", Name: "Доспех мага", Level: 1, School: "Ограждение", Classes: cSW, Desc: "КД цели 13 + Ловкость на 8 часов."},
	{ID: "sleep", Name: "Сон", Level: 1, School: "Очарование", Classes: []string{"bard", "sorcerer", "wizard"}, Desc: "5d8 HP существ засыпают, начиная с самых слабых."},
	{ID: "bless", Name: "Благословение", Level: 1, School: "Очарование", Classes: []string{"cleric", "paladin"}, Desc: "Три союзника добавляют 1d4 к атакам и спасброскам."},
	{ID: "hex", Name: "Сглаз", Level: 1, School: "Очарование", Classes: []string{"warlock"}, Desc: "+1d6 некротическим по цели, помеха на проверки одной характеристики."},
	{ID: "huntersmark", Name: "Метка охотника", Level: 1, School: "Прорицание", Classes: []string{"ranger"}, Desc: "+1d6 урона по помеченной цели."},
	{ID: "detectmagic", Name: "Обнаружение магии", Level: 1, School: "Прорицание", Classes: []string{"bard", "cleric", "druid", "paladin", "ranger", "sorcerer", "wizard"}, Desc: "Чувствуете магию в 30 футах."},
	{ID: "charmperson", Name: "Очарование личности", Level: 1, School: "Очарование", Classes: []string{"bard", "druid", "sorcerer", "warlock", "wizard"}, Desc: "Гуманоид считает вас другом на час."},
	{ID: "scorchingray", Name: "Палящий луч", Level: 2, School: "Воплощение", Classes: cSW, Mode: "attack", Dmg: "2d6", Type: "fire", Rays: 3, Up: "ray", Desc: "Три луча по 2d6 огнём, атака на каждый; +1 луч за ячейку выше."},
	{ID: "mistystep", Name: "Туманный шаг", Level: 2, School: "Колдовство", Classes: []string{"sorcerer", "warlock", "wizard"}, Desc: "Бонусным действием телепорт на 30 футов."},
	{ID: "invisibility", Name: "Невидимость", Level: 2, School: "Иллюзия", Classes: []string{"bard", "sorcerer", "warlock", "wizard"}, Desc: "Существо невидимо до атаки или заклинания."},
	{ID: "holdperson", Name: "Удержание личности", Level: 2, School: "Очарование", Classes: []string{"bard", "cleric", "druid", "sorcerer", "warlock", "wizard"}, Desc: "Гуманоид парализован, спасбросок Мудрости."},
	{ID: "spiritualweapon", Name: "Духовное оружие", Level: 2, School: "Воплощение", Classes: []string{"cleric"}, Desc: "Парящее оружие бьёт бонусным действием 1d8 + мод."},
	{ID: "fireball", Name: "Огненный шар", Level: 3, School: "Воплощение", Classes: cSW, Mode: "save", Save: "dex", Dmg: "8d6", Type: "fire", Half: true, Area: true, Up: "1d6", Desc: "Сфера 20 футов, 8d6 огнём; половина при спасброске."},
	{ID: "lightningbolt", Name: "Молния", Level: 3, School: "Воплощение", Classes: cSW, Mode: "save", Save: "dex", Dmg: "8d6", Type: "lightning", Half: true, Area: true, Up: "1d6", Desc: "Линия 100 футов, 8d6 молнией."},
	{ID: "counterspell", Name: "Контрзаклинание", Level: 3, School: "Ограждение", Classes: []string{"sorcerer", "warlock", "wizard"}, Desc: "Реакция: срывает заклинание 3 уровня и ниже."},
	{ID: "dispelmagic", Name: "Рассеивание магии", Level: 3, School: "Ограждение", Classes: []string{"bard", "cleric", "druid", "paladin", "sorcerer", "warlock", "wizard"}, Desc: "Снимает заклинание с цели."},
	{ID: "haste", Name: "Ускорение", Level: 3, School: "Преобразование", Classes: []string{"sorcerer", "wizard"}, Desc: "Двойная скорость, +2 КД, дополнительное действие."},
	{ID: "fly", Name: "Полёт", Level: 3, School: "Преобразование", Classes: []string{"sorcerer", "warlock", "wizard"}, Desc: "Скорость полёта 60 футов на 10 минут."},
	{ID: "revivify", Name: "Оживление", Level: 3, School: "Некромантия", Classes: []string{"cleric", "paladin"}, Desc: "Возвращает к жизни умершего не более минуты назад, 1 HP."},
}

func spellDef(id string) (SpellDef, bool) {
	i := slices.IndexFunc(spellDefs, func(s SpellDef) bool { return s.ID == id })
	if i < 0 {
		return SpellDef{}, false
	}
	return spellDefs[i], true
}

func init() {
	have := map[string]bool{}
	for _, s := range spellDefs {
		have[s.ID] = true
		have[s.Name] = true
	}
	for _, ln := range strings.Split(spellTable, "\n") {
		p := strings.Split(strings.TrimSpace(ln), "|")
		if len(p) < 6 || have[p[0]] || have[p[1]] {
			continue
		}
		lvl, _ := strconv.Atoi(p[2])
		s := SpellDef{ID: p[0], Name: p[1], Level: lvl, School: p[3], Desc: p[5]}
		if p[4] != "" {
			s.Classes = strings.Split(p[4], ",")
		}
		if len(p) >= 14 {
			s.Mode, s.Dmg, s.Type, s.Save = p[6], p[7], p[8], p[9]
			s.Half, s.Area, s.Up, s.Scale = p[10] == "1", p[11] == "1", p[12], p[13] == "1"
			if s.Mode == "heal" && s.Area {
				s.Area = true
			}
		}
		have[s.ID], have[s.Name] = true, true
		spellDefs = append(spellDefs, s)
	}
	slices.SortStableFunc(spellDefs, func(a, b SpellDef) int {
		if a.Level != b.Level {
			return a.Level - b.Level
		}
		return strings.Compare(a.Name, b.Name)
	})
}

// DamageTypeIDs – допустимые типы урона для собственных заклинаний и существ.
func DamageTypeIDs() []string {
	out := make([]string, len(dndDamage))
	for i, d := range dndDamage {
		out[i] = d.ID
	}
	return out
}
