// Package treasure — генератор добычи по мотивам таблиц сокровищ DMG: монеты, драгоценные камни, предметы искусства
// и магические предметы. Таблицы упрощены и набраны по памяти; мастер всегда может подкрутить результат.
package treasure

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"

	"heroesbook/internal/dice"
	"heroesbook/internal/model"
	"heroesbook/internal/srd"
)

// Treasure — результат броска: монеты и предметы.
type Treasure struct {
	CR    string         `json:"cr"`
	Hoard bool           `json:"hoard"`
	Coins map[string]int `json:"coins"`
	Items []model.Item   `json:"items"`
	Text  string         `json:"text"`
}

type coin struct {
	kind string
	expr string
	mult int
}

func roll(r dice.Roller, expr string) int {
	ex, err := dice.Parse(expr)
	if err != nil {
		return 0
	}
	return ex.Roll(r, false).Total
}

func d100(r dice.Roller) int { return 1 + r.Intn(100) }

// Tier по уровню опасности: 0 — CR 0–4, 1 — 5–10, 2 — 11–16, 3 — 17+.
func Tier(cr string) int {
	v := 0.0
	if n, d, ok := strings.Cut(cr, "/"); ok {
		a, _ := strconv.ParseFloat(n, 64)
		b, _ := strconv.ParseFloat(d, 64)
		if b > 0 {
			v = a / b
		}
	} else {
		v, _ = strconv.ParseFloat(cr, 64)
	}
	switch {
	case v <= 4:
		return 0
	case v <= 10:
		return 1
	case v <= 16:
		return 2
	}
	return 3
}

type indRow struct {
	upTo  int
	coins []coin
}

var individual = [4][]indRow{
	{{30, []coin{{"cp", "5d6", 1}}}, {60, []coin{{"sp", "4d6", 1}}}, {70, []coin{{"ep", "3d6", 1}}}, {95, []coin{{"gp", "3d6", 1}}}, {100, []coin{{"pp", "1d6", 1}}}},
	{{30, []coin{{"cp", "4d6", 100}, {"ep", "1d6", 10}}}, {60, []coin{{"sp", "6d6", 10}, {"gp", "2d6", 10}}}, {70, []coin{{"ep", "3d6", 10}, {"gp", "2d6", 10}}},
		{95, []coin{{"gp", "4d6", 10}}}, {100, []coin{{"gp", "2d6", 10}, {"pp", "3d6", 1}}}},
	{{20, []coin{{"sp", "4d6", 100}, {"gp", "1d6", 100}}}, {35, []coin{{"ep", "1d6", 100}, {"gp", "1d6", 100}}}, {75, []coin{{"gp", "2d6", 100}, {"pp", "1d6", 10}}}, {100, []coin{{"gp", "2d6", 100}, {"pp", "2d6", 10}}}},
	{{15, []coin{{"ep", "2d6", 1000}, {"gp", "8d6", 100}}}, {55, []coin{{"gp", "1d6", 1000}, {"pp", "1d6", 100}}}, {100, []coin{{"gp", "1d6", 1000}, {"pp", "2d6", 100}}}},
}

var hoardCoins = [4][]coin{
	{{"cp", "6d6", 100}, {"sp", "3d6", 100}, {"gp", "2d6", 10}},
	{{"cp", "2d6", 100}, {"sp", "2d6", 1000}, {"gp", "6d6", 100}, {"pp", "3d6", 10}},
	{{"gp", "4d6", 1000}, {"pp", "5d6", 100}},
	{{"gp", "12d6", 1000}, {"pp", "8d6", 1000}},
}

type magic struct {
	table byte
	dice  string
}

type hoardRow struct {
	upTo  int
	gem   int // цена камня в зм, 0 — нет
	art   int
	n     string // сколько камней или статуэток
	magic []magic
}

func g(upTo, gem int, n string, m ...magic) hoardRow {
	return hoardRow{upTo: upTo, gem: gem, n: n, magic: m}
}
func a(upTo, art int, n string, m ...magic) hoardRow {
	return hoardRow{upTo: upTo, art: art, n: n, magic: m}
}
func m(t byte, d string) magic { return magic{t, d} }

var hoardRows = [4][]hoardRow{
	{g(6, 0, ""), g(16, 10, "2d6"), a(26, 25, "2d4"), g(36, 50, "2d6"),
		g(44, 10, "2d6", m('A', "1d6")), a(52, 25, "2d4", m('A', "1d6")), g(60, 50, "2d6", m('A', "1d6")),
		g(65, 10, "2d6", m('B', "1d4")), a(70, 25, "2d4", m('B', "1d4")), g(75, 50, "2d6", m('B', "1d4")),
		g(78, 10, "2d6", m('C', "1d4")), a(80, 25, "2d4", m('C', "1d4")), g(85, 50, "2d6", m('C', "1d4")),
		a(92, 25, "2d4", m('F', "1d4")), g(97, 50, "2d6", m('F', "1d4")), a(99, 25, "2d4", m('G', "1")), g(100, 50, "2d6", m('G', "1"))},
	{g(4, 0, ""), a(10, 25, "2d4"), g(16, 50, "3d6"), g(22, 100, "3d6"), a(28, 250, "2d4"),
		a(32, 25, "2d4", m('A', "1d6")), g(36, 50, "3d6", m('A', "1d6")), g(40, 100, "3d6", m('A', "1d6")), a(44, 250, "2d4", m('A', "1d6")),
		a(49, 25, "2d4", m('B', "1d4")), g(54, 50, "3d6", m('B', "1d4")), g(59, 100, "3d6", m('B', "1d4")), a(63, 250, "2d4", m('B', "1d4")),
		a(66, 25, "2d4", m('C', "1d4")), g(69, 50, "3d6", m('C', "1d4")), g(72, 100, "3d6", m('C', "1d4")), a(74, 250, "2d4", m('C', "1d4")),
		a(76, 25, "2d4", m('D', "1")), g(78, 50, "3d6", m('D', "1")), g(79, 100, "3d6", m('D', "1")), a(80, 250, "2d4", m('D', "1")),
		a(84, 25, "2d4", m('F', "1d4")), g(88, 50, "3d6", m('F', "1d4")), g(91, 100, "3d6", m('F', "1d4")), a(94, 250, "2d4", m('F', "1d4")),
		a(96, 25, "2d4", m('G', "1d4")), g(98, 50, "3d6", m('G', "1d4")), g(99, 100, "3d6", m('H', "1")), a(100, 250, "2d4", m('H', "1"))},
	{g(3, 0, ""), a(6, 250, "2d4"), a(9, 750, "2d4"), g(12, 500, "3d6"), g(15, 1000, "3d6"),
		a(19, 250, "2d4", m('C', "1d4")), a(23, 750, "2d4", m('C', "1d4")), g(26, 500, "3d6", m('C', "1d4")), g(29, 1000, "3d6", m('C', "1d4")),
		a(35, 250, "2d4", m('D', "1d4")), a(40, 750, "2d4", m('D', "1d4")), g(45, 500, "3d6", m('D', "1d4")), g(50, 1000, "3d6", m('D', "1d4")),
		a(54, 250, "2d4", m('F', "1d4")), a(58, 750, "2d4", m('F', "1d4")), g(62, 500, "3d6", m('F', "1d4")), g(66, 1000, "3d6", m('F', "1d4")),
		a(68, 250, "2d4", m('G', "1d4")), a(70, 750, "2d4", m('G', "1d4")), g(72, 500, "3d6", m('G', "1d4")), g(74, 1000, "3d6", m('G', "1d4")),
		a(76, 250, "2d4", m('H', "1d4")), a(78, 750, "2d4", m('H', "1d4")), g(80, 500, "3d6", m('H', "1d4")), g(82, 1000, "3d6", m('H', "1d4")),
		a(85, 250, "2d4", m('I', "1")), a(88, 750, "2d4", m('I', "1")), g(90, 500, "3d6", m('I', "1")), g(100, 1000, "3d6", m('I', "1"))},
	{g(2, 0, ""), g(5, 1000, "3d8"), a(8, 2500, "1d10"), a(11, 7500, "1d4"), g(14, 5000, "1d8"),
		g(22, 1000, "3d8", m('C', "1d8")), a(28, 2500, "1d10", m('C', "1d8")), a(35, 7500, "1d4", m('C', "1d8")), g(42, 5000, "1d8", m('C', "1d8")),
		g(49, 1000, "3d8", m('D', "1d6")), a(56, 2500, "1d10", m('D', "1d6")), a(63, 7500, "1d4", m('D', "1d6")), g(70, 5000, "1d8", m('D', "1d6")),
		g(74, 1000, "3d8", m('E', "1d6")), a(77, 2500, "1d10", m('E', "1d6")), a(80, 7500, "1d4", m('E', "1d6")), g(83, 5000, "1d8", m('E', "1d6")),
		g(85, 1000, "3d8", m('G', "1d4")), a(88, 2500, "1d10", m('G', "1d4")), a(90, 7500, "1d4", m('G', "1d4")), g(92, 5000, "1d8", m('G', "1d4")),
		g(94, 1000, "3d8", m('H', "1d4")), a(96, 2500, "1d10", m('H', "1d4")), a(98, 7500, "1d4", m('I', "1d4")), g(100, 5000, "1d8", m('I', "1d4"))},
}

type weight struct {
	rar string
	w   int
}

// Таблицы магических предметов A–I сведены к редкости предмета.
var tableRarity = map[byte][]weight{
	'A': {{"обычный", 70}, {"необычный", 30}},
	'B': {{"необычный", 70}, {"обычный", 15}, {"редкий", 15}},
	'C': {{"редкий", 70}, {"необычный", 30}},
	'D': {{"очень редкий", 70}, {"редкий", 30}},
	'E': {{"легендарный", 40}, {"очень редкий", 60}},
	'F': {{"необычный", 60}, {"редкий", 40}},
	'G': {{"редкий", 60}, {"очень редкий", 40}},
	'H': {{"очень редкий", 70}, {"легендарный", 30}},
	'I': {{"легендарный", 100}},
}

var gems = map[int][]string{
	10:   {"азурит", "полосатый агат", "синий кварц", "глазчатый агат", "гематит", "лазурит", "малахит", "мшистый агат", "обсидиан", "родохрозит", "тигровый глаз", "бирюза"},
	50:   {"гелиотроп", "сердолик", "халцедон", "хризопраз", "цитрин", "яшма", "лунный камень", "оникс", "горный хрусталь", "сардоникс", "звёздчатый розовый кварц", "циркон"},
	100:  {"янтарь", "аметист", "хризоберилл", "коралл", "гранат", "нефрит", "гагат", "жемчуг", "шпинель", "турмалин"},
	500:  {"александрит", "аквамарин", "чёрный жемчуг", "синяя шпинель", "перидот", "топаз"},
	1000: {"чёрный опал", "синий сапфир", "изумруд", "огненный опал", "опал", "звёздный рубин", "звёздный сапфир", "жёлтый сапфир"},
	5000: {"чёрный сапфир", "алмаз", "гиацинт", "рубин"},
}

var art = map[int][]string{
	25:   {"серебряный кувшин", "резная костяная статуэтка", "маленький золотой браслет", "облачение из золотой парчи", "чёрная бархатная маска с серебряной нитью", "медный кубок с серебряной филигранью", "пара резных костяных кубиков", "зеркальце в расписной деревянной раме", "шёлковый платок с вышивкой", "золотой медальон с портретом"},
	250:  {"золотое кольцо с гелиотропами", "резная статуэтка из слоновой кости", "большой золотой браслет", "серебряное ожерелье с самоцветом", "бронзовая корона", "шёлковая мантия с золотым шитьём", "большой гобелен тонкой работы", "латунная кружка с нефритовой инкрустацией", "шкатулка с бирюзовыми фигурками зверей", "золотая птичья клетка с филигранью из электрума"},
	750:  {"серебряный кубок с лунными камнями", "посеребрённый стальной длинный меч с гагатом", "резная арфа из редкого дерева с слоновой костью", "маленький золотой идол", "золотой гребень в виде дракона с гранатовыми глазами", "пробка с золотистым аметистом", "парадный кинжал из электрума с чёрной жемчужиной", "серебряно-золотая брошь", "статуэтка из обсидиана с золотом", "расписная золотая боевая маска"},
	2500: {"тонкая золотая цепь с огненным опалом", "картина старого мастера", "мантия из шёлка и бархата с лунными камнями", "платиновый браслет с сапфиром", "расшитая перчатка с самоцветами", "ножной браслет с самоцветами", "золотая музыкальная шкатулка", "золотой венец с четырьмя аквамаринами", "повязка на глаз с ложным глазом из сапфира", "ожерелье из мелкого розового жемчуга"},
	7500: {"золотая корона с самоцветами", "платиновое кольцо с самоцветами", "золотая статуэтка с рубинами", "золотая чаша с изумрудами", "золотая шкатулка с платиновой филигранью", "расписной золотой детский саркофаг", "нефритовая игральная доска с золотыми фигурами", "рог из слоновой кости с золотой филигранью и самоцветами"},
}

var catalog = sync.OnceValue(srd.Items)

func pick[T any](r dice.Roller, xs []T) T { return xs[r.Intn(len(xs))] }

func magicItem(r dice.Roller, table byte) (model.Item, bool) {
	ws := tableRarity[table]
	tot := 0
	for _, w := range ws {
		tot += w.w
	}
	n := r.Intn(tot)
	rar := ws[0].rar
	for _, w := range ws {
		if n < w.w {
			rar = w.rar
			break
		}
		n -= w.w
	}
	all := catalog()
	order := []string{rar, "необычный", "редкий", "очень редкий", "обычный"}
	for _, want := range order {
		var pool []model.Item
		for _, it := range all {
			if it.Rarity == want && it.Cat != "gear" && it.Cat != "tool" {
				pool = append(pool, it)
			}
		}
		if len(pool) > 0 {
			it := pick(r, pool)
			it.ID, it.Qty = "", 1
			return it, true
		}
	}
	return model.Item{}, false
}

func addItem(t *Treasure, it model.Item) {
	for i := range t.Items {
		if t.Items[i].Name == it.Name && t.Items[i].Price == it.Price {
			t.Items[i].Qty += it.Qty
			return
		}
	}
	t.Items = append(t.Items, it)
}

func addCoins(t *Treasure, r dice.Roller, cs []coin) {
	for _, c := range cs {
		t.Coins[c.kind] += roll(r, c.expr) * c.mult
	}
}

// Individual — добыча с одного существа.
func Individual(r dice.Roller, cr string) Treasure {
	t := Treasure{CR: cr, Coins: map[string]int{}}
	n := d100(r)
	for _, row := range individual[Tier(cr)] {
		if n <= row.upTo {
			addCoins(&t, r, row.coins)
			break
		}
	}
	t.Text = describe(t)
	return t
}

// Hoard — сокровищница: монеты, камни, предметы искусства, магические предметы.
func Hoard(r dice.Roller, cr string) Treasure {
	tier := Tier(cr)
	t := Treasure{CR: cr, Hoard: true, Coins: map[string]int{}}
	addCoins(&t, r, hoardCoins[tier])
	n := d100(r)
	for _, row := range hoardRows[tier] {
		if n > row.upTo {
			continue
		}
		cnt := 0
		if row.n != "" {
			cnt = roll(r, row.n)
		}
		switch {
		case row.gem > 0:
			for i := 0; i < cnt; i++ {
				addItem(&t, model.Item{Name: "Драгоценный камень: " + pick(r, gems[row.gem]), Qty: 1, Cat: "gem", Price: fmt.Sprintf("%d зм", row.gem), Desc: "Драгоценность, продаётся за указанную цену."})
			}
		case row.art > 0:
			for i := 0; i < cnt; i++ {
				addItem(&t, model.Item{Name: "Предмет искусства: " + pick(r, art[row.art]), Qty: 1, Cat: "art", Price: fmt.Sprintf("%d зм", row.art), Weight: 1, Desc: "Предмет искусства, продаётся за указанную цену."})
			}
		}
		for _, mg := range row.magic {
			for i, k := 0, roll(r, mg.dice); i < k; i++ {
				if it, ok := magicItem(r, mg.table); ok {
					addItem(&t, it)
				}
			}
		}
		break
	}
	t.Text = describe(t)
	return t
}

// Merge складывает несколько добыч в одну.
func Merge(ts ...Treasure) Treasure {
	out := Treasure{Coins: map[string]int{}}
	for _, t := range ts {
		for k, v := range t.Coins {
			out.Coins[k] += v
		}
		for _, it := range t.Items {
			addItem(&out, it)
		}
		out.Hoard = out.Hoard || t.Hoard
	}
	out.Text = describe(out)
	return out
}

// ForCRs: добыча за побеждённых существ. Для сокровищницы берётся самое опасное существо.
func ForCRs(r dice.Roller, crs []string, hoard bool) Treasure {
	if len(crs) == 0 {
		return Treasure{Coins: map[string]int{}}
	}
	if hoard {
		best := crs[0]
		for _, c := range crs {
			if Tier(c) > Tier(best) || (Tier(c) == Tier(best) && cmpCR(c, best) > 0) {
				best = c
			}
		}
		return Hoard(r, best)
	}
	ts := make([]Treasure, len(crs))
	for i, c := range crs {
		ts[i] = Individual(r, c)
	}
	out := Merge(ts...)
	out.CR = "бой"
	return out
}

func cmpCR(a, b string) int {
	va, vb := crValue(a), crValue(b)
	switch {
	case va < vb:
		return -1
	case va > vb:
		return 1
	}
	return 0
}

func crValue(cr string) float64 {
	if n, d, ok := strings.Cut(cr, "/"); ok {
		a, _ := strconv.ParseFloat(n, 64)
		b, _ := strconv.ParseFloat(d, 64)
		if b > 0 {
			return a / b
		}
	}
	v, _ := strconv.ParseFloat(cr, 64)
	return v
}

var coinNames = map[string]string{"pp": "пм", "gp": "зм", "ep": "эм", "sp": "см", "cp": "мм"}

func describe(t Treasure) string {
	var p []string
	for _, k := range model.CoinOrder {
		if t.Coins[k] > 0 {
			p = append(p, fmt.Sprintf("%d %s", t.Coins[k], coinNames[k]))
		}
	}
	for _, it := range t.Items {
		if it.Qty > 1 {
			p = append(p, fmt.Sprintf("%s ×%d", it.Name, it.Qty))
		} else {
			p = append(p, it.Name)
		}
	}
	if len(p) == 0 {
		return "ничего ценного"
	}
	return strings.Join(p, ", ")
}

// GoldValue оценивает цену добычи в золотых монетах (монеты и предметы с ценой в зм).
func GoldValue(t Treasure) float64 {
	cp := 0
	for k, v := range t.Coins {
		cp += v * model.CoinValueCP[k]
	}
	gp := float64(cp) / 100
	for _, it := range t.Items {
		if f, _, ok := strings.Cut(it.Price, " зм"); ok {
			f = strings.ReplaceAll(strings.ReplaceAll(f, " ", ""), ",", "")
			if v, err := strconv.ParseFloat(f, 64); err == nil {
				gp += v * float64(max(1, it.Qty))
			}
		}
	}
	return gp
}

var _ = slices.Contains[[]string]
