package worldgen

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"heroesbook/internal/bestiary"
	"heroesbook/internal/dice"
	"heroesbook/internal/model"
	"heroesbook/internal/treasure"
)

// Gen – общие параметры генерации: генератор случайностей, уровень группы, расы мира и выдача ID.
type Gen struct {
	R     dice.Roller
	Level int      // средний уровень группы: от него зависит сила врагов и награды
	Races []string // расы мира (id); пусто – все основные
	NewID func() string
}

const (
	maxPlaces    = 300
	humanWeight  = 3 // люди встречаются чаще других
	maxFoesGroup = 6
)

// AllRaces – основные расы каталога (порядок важен: от него зависит воспроизводимость по seed).
var AllRaces = []string{"human", "elf", "dwarf", "halfling", "gnome", "halfelf", "halforc", "dragonborn", "tiefling", "aasimar", "genasi",
	"orc", "goblin", "hobgoblin", "bugbear", "kobold", "goliath", "firbolg", "aarakocra", "kenku", "tabaxi", "lizardfolk", "tortle",
	"triton", "minotaur", "warforged", "yuanti"}

func (g Gen) races() []string {
	if len(g.Races) > 0 {
		return g.Races
	}
	return AllRaces
}

// someRace – случайная раса мира; люди чаще.
func (g Gen) someRace() string {
	rs := g.races()
	if slices.Contains(rs, "human") && g.R.Intn(humanWeight) == 0 {
		return "human"
	}
	return pick(g.R, rs)
}

var roles = map[string][]string{
	"capital": {"правитель", "капитан стражи", "придворный маг", "богатый купец", "верховный жрец", "глава гильдии воров"},
	"city":    {"бургомистр", "капитан стражи", "торговец", "жрец", "информатор", "алхимик"},
	"town":    {"староста", "кузнец", "трактирщик", "стражник", "травница"},
	"village": {"староста", "охотник", "знахарка", "мельник", "пастух"},
	"tavern":  {"трактирщик", "бард", "наёмник", "торговец-путник", "картёжник"},
	"temple":  {"жрец", "послушник", "паломник"},
	"shrine":  {"отшельник", "паломник"},
	"shop":    {"торговец", "подмастерье"},
	"smithy":  {"кузнец", "подмастерье"},
	"castle":  {"лорд", "капитан гарнизона", "рыцарь", "советник"},
	"port":    {"капитан корабля", "контрабандист", "портовый грузчик", "начальник порта"},
	"tower":   {"волшебник", "ученик волшебника"},
}

var (
	traits = []string{"говорит загадками", "вечно жуёт яблоко", "громко смеётся", "никому не доверяет", "любит хвастаться", "молится по любому поводу",
		"шепелявит", "коллекционирует пуговицы", "боится темноты", "гордится старым шрамом", "ворчит без остановки", "раздаёт деньги направо и налево",
		"одержимо следит за чистотой", "пишет стихи", "не выпускает трубку изо рта", "злится, когда передразнивают"}
	wants = []string{"вернуть украденную семейную реликвию", "отомстить за брата", "разбогатеть любой ценой", "уехать отсюда навсегда",
		"найти пропавшую дочь", "получить место в гильдии", "искупить старый грех", "узнать, что скрывает правитель", "спокойно дожить до старости",
		"доказать всем свою храбрость"}
	secrets = []string{"шпионит на соседнее королевство", "задолжал бандитам крупную сумму", "на самом деле оборотень", "в прошлом – наёмный убийца",
		"знает тайный ход под городом", "подделывает печати", "скрывает встречу с драконом", "поклоняется запретному божеству",
		"прячет в подвале раненого беглеца", "ничего не скрывает – и это подозрительно"}
	attitudes = []string{"дружелюбен", "безразличен", "безразличен", "враждебен"}
)

// wildKinds – какие существа водятся в местах разного вида.
var wildKinds = map[string][]string{
	"dungeon":  {"undead", "monstrosity", "aberration", "humanoid", "ooze"},
	"cave":     {"beast", "monstrosity", "humanoid"},
	"ruins":    {"undead", "construct", "humanoid"},
	"lair":     {"dragon", "monstrosity", "giant", "fiend"},
	"camp":     {"humanoid"},
	"tower":    {"construct", "elemental", "humanoid"},
	"mine":     {"humanoid", "monstrosity", "elemental"},
	"forest":   {"beast", "fey", "plant", "humanoid"},
	"mountain": {"giant", "beast", "dragon"},
	"swamp":    {"beast", "plant", "undead", "monstrosity"},
	"lake":     {"beast", "monstrosity", "elemental"},
	"room":     {"undead", "humanoid", "monstrosity", "ooze", "construct", "aberration"},
	"other":    {"beast", "humanoid", "monstrosity"},
}

// Wild – место, где вместо горожан враги.
func Wild(kind string) bool {
	_, ok := wildKinds[kind]
	return ok && kind != "tower" // в башне живёт волшебник – это обитаемое место
}

// crXP – опыт за существо по уровню опасности (DMG).
var crXP = map[string]int{"0": 10, "1/8": 25, "1/4": 50, "1/2": 100, "1": 200, "2": 450, "3": 700, "4": 1100, "5": 1800, "6": 2300, "7": 2900,
	"8": 3900, "9": 5000, "10": 5900, "11": 7200, "12": 8400, "13": 10000, "14": 11500, "15": 13000, "16": 15000, "17": 18000, "18": 20000,
	"19": 22000, "20": 25000, "21": 33000, "22": 41000, "23": 50000, "24": 62000, "30": 155000}

// Populate заполняет места: в диких – враги и добыча, в обитаемых – NPC и задания, ведущие в дикие места карты.
// onlyEmpty – не трогать места, где уже есть содержимое.
func (g Gen) Populate(places []model.Place, onlyEmpty bool) []model.Place {
	out := model.ClonePlaces(places)
	for i := range out {
		p := &out[i]
		if !Wild(p.Kind) || (onlyEmpty && len(p.Foes) > 0) {
			continue
		}
		if p.Kind == "room" && g.R.Intn(3) == 0 { // треть комнат пустует – только ловушки и тайники
			p.Foes, p.Loot = nil, ""
			continue
		}
		p.Foes, p.Loot = g.foes(p.Kind), ""
		if len(p.Foes) > 0 {
			var crs []string
			for _, f := range p.Foes {
				for range f.Count {
					crs = append(crs, f.CR)
				}
			}
			p.Loot = treasure.ForCRs(g.R, crs, p.Kind == "lair" || p.Kind == "dungeon").Text
		}
	}
	for i := range out {
		p := &out[i]
		if !model.Settlement(p.Kind) || (onlyEmpty && (len(p.NPCs) > 0 || len(p.Quests) > 0)) {
			continue
		}
		p.NPCs = g.npcs(p)
		p.Quests = g.quests(p, out)
	}
	return out
}

func (g Gen) npcs(p *model.Place) []model.NPC {
	rs := roles[p.Kind]
	if len(rs) == 0 {
		rs = roles["village"]
	}
	n := min(len(rs), 2+g.R.Intn(2))
	if p.Kind == "capital" || p.Kind == "city" {
		n = min(len(rs), 3+g.R.Intn(2))
	}
	order := slices.Clone(rs)
	out := make([]model.NPC, 0, n)
	for i := 0; i < n; i++ {
		j := i + g.R.Intn(len(order)-i)
		order[i], order[j] = order[j], order[i]
		race := p.Race
		if race == "" || g.R.Intn(4) == 0 { // и в моноэтничном поселении бывают чужаки
			race = g.someRace()
		}
		out = append(out, model.NPC{Name: PersonName(g.R, race), Race: race, Role: order[i], Trait: pick(g.R, traits),
			Want: pick(g.R, wants), Secret: pick(g.R, secrets), Attitude: pick(g.R, attitudes)})
	}
	return out
}

// quests – 1–2 задания: в ближайшее дикое место с врагами, иначе бытовые поручения.
func (g Gen) quests(p *model.Place, all []model.Place) []model.Quest {
	var targets []*model.Place
	for i := range all {
		if Wild(all[i].Kind) && len(all[i].Foes) > 0 && all[i].ID != p.ID {
			targets = append(targets, &all[i])
		}
	}
	slices.SortStableFunc(targets, func(a, b *model.Place) int {
		return cmpF(dist(p, a), dist(p, b))
	})
	giver := ""
	if len(p.NPCs) > 0 {
		giver = p.NPCs[g.R.Intn(len(p.NPCs))].Name
	}
	n := 1 + g.R.Intn(2)
	var out []model.Quest
	for i := 0; i < n; i++ {
		if i < len(targets) && i < 3 {
			out = append(out, g.foeQuest(giver, targets[i]))
			continue
		}
		out = append(out, g.chore(giver, all))
	}
	return out
}

func (g Gen) foeQuest(giver string, t *model.Place) model.Quest {
	f := t.Foes[0]
	xp := 0
	for _, x := range t.Foes {
		xp += crXP[x.CR] * x.Count
	}
	name := t.Name
	if name == "" {
		name = "место неподалёку"
	}
	templates := []struct{ title, text string }{
		{"Очистить «%s»", "В «%[1]s» обосновались враги (%[2]s) и нападают на путников. Нужно положить этому конец."},
		{"Пропавшие у «%s»", "Несколько человек ушли к «%[1]s» и не вернулись. Говорят, там видели: %[2]s."},
		{"Реликвия из «%s»", "Древняя реликвия хранится в «%[1]s». Охрана: %[2]s. Заказчик щедро заплатит."},
		{"Награда за головы: «%s»", "Объявлена награда за врагов из «%[1]s»: %[2]s."},
	}
	tp := pick(g.R, templates)
	who := fmt.Sprintf("%s ×%d", f.Name, f.Count)
	if len(t.Foes) > 1 {
		who += " и другие"
	}
	gold := max(10, xp/2)
	return model.Quest{Title: fmt.Sprintf(tp.title, name), Text: fmt.Sprintf(tp.text, name, who), Giver: giver, Target: t.ID,
		Reward: fmt.Sprintf("%d зм", roundTo(gold, 10)), XP: xp}
}

func (g Gen) chore(giver string, all []model.Place) model.Quest {
	var towns []string
	for _, p := range all {
		if model.Settlement(p.Kind) && p.Name != "" {
			towns = append(towns, p.Name)
		}
	}
	dest := "соседний город"
	if len(towns) > 0 {
		dest = pick(g.R, towns)
	}
	level := max(1, g.Level)
	chores := []model.Quest{
		{Title: "Проводить караван", Text: fmt.Sprintf("Торговый караван идёт в «%s». Нужна охрана в пути – на дорогах неспокойно.", dest)},
		{Title: "Письмо без печати", Text: fmt.Sprintf("Доставить запечатанное письмо в «%s» и не задавать вопросов.", dest)},
		{Title: "Ночной вор", Text: "Кто-то по ночам ворует со складов. Выследить и поймать – живым."},
		{Title: "Редкие травы", Text: "Знахарке нужны травы, что растут только у старого кладбища. Собрать до полнолуния."},
		{Title: "Свадьба под угрозой", Text: "Две семьи враждуют, а свадьба через три дня. Помирить или хотя бы не дать пролиться крови."},
	}
	q := pick(g.R, chores)
	q.Giver, q.XP = giver, 50*level
	q.Reward = fmt.Sprintf("%d зм", roundTo(25*level+g.R.Intn(20*level), 5))
	return q
}

// foes – группа существ для дикого места, по силе группы героев.
func (g Gen) foes(kind string) []model.Foe {
	level := max(1, min(20, g.Level))
	kinds := wildKinds[kind]
	pool := bestiary.List()
	boss := kind == "lair"
	lo, hi := float64(level)/4, float64(level)
	if boss {
		lo, hi = float64(level), float64(level)+3
	}
	var cand []bestiary.Monster
	for _, m := range pool {
		cr := bestiary.CRValue(m.CR)
		if slices.Contains(kinds, m.Kind) && cr >= lo && cr <= hi {
			cand = append(cand, m)
		}
	}
	if len(cand) == 0 { // подходящих по виду нет – берём любых по силе
		for _, m := range pool {
			if cr := bestiary.CRValue(m.CR); cr >= lo && cr <= hi {
				cand = append(cand, m)
			}
		}
	}
	if len(cand) == 0 {
		return nil
	}
	m := pick(g.R, cand)
	n := 1
	if cr := bestiary.CRValue(m.CR); !boss && cr <= hi/3 {
		n = min(maxFoesGroup, 2+g.R.Intn(4))
	}
	out := []model.Foe{{ID: m.ID, Name: m.Name, CR: m.CR, Count: n}}
	if boss || g.R.Intn(3) != 0 {
		return out
	}
	// иногда – свита послабее
	for _, x := range cand {
		if bestiary.CRValue(x.CR) < bestiary.CRValue(m.CR) && x.ID != m.ID {
			return append(out, model.Foe{ID: x.ID, Name: x.Name, CR: x.CR, Count: 1 + g.R.Intn(3)})
		}
	}
	return out
}

func dist(a, b *model.Place) float64 { return math.Hypot(a.X-b.X, a.Y-b.Y) }

func cmpF(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func roundTo(n, step int) int { return max(step, (n+step/2)/step*step) }

// KindName – название вида места по-русски (для подписей без имени).
func KindName(kind string) string {
	names := map[string]string{"capital": "Столица", "city": "Город", "town": "Городок", "village": "Деревня", "tavern": "Таверна",
		"temple": "Храм", "shop": "Лавка", "smithy": "Кузница", "castle": "Замок", "port": "Порт", "dungeon": "Подземелье", "cave": "Пещера",
		"ruins": "Руины", "lair": "Логово", "camp": "Лагерь", "tower": "Башня", "shrine": "Святилище", "mine": "Шахта", "room": "Комната",
		"forest": "Лес", "mountain": "Горы", "swamp": "Болото", "lake": "Озеро", "other": "Место"}
	if n, ok := names[kind]; ok {
		return n
	}
	return strings.ToUpper(kind[:1]) + kind[1:]
}
