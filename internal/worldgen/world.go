package worldgen

import (
	"errors"
	"image/color"
	"math"
	"slices"
	"sort"

	"heroesbook/internal/model"
)

// Генератор карты мира: высоты и влажность из шума → биомы → рельеф с тенями, реки, дороги;
// каждая раса селится в подходящей ей местности; в диких землях – логова, руины, пещеры.

type biome uint8

const (
	deep biome = iota
	shallow
	beach
	grass
	forest
	jungle
	swamp
	desert
	tundra
	hills
	mountain
	snow
)

var biomeColor = map[biome]color.RGBA{
	deep: rgb(28, 62, 104), shallow: rgb(52, 104, 150), beach: rgb(222, 205, 152), grass: rgb(124, 166, 82), forest: rgb(58, 112, 58),
	jungle: rgb(36, 102, 52), swamp: rgb(84, 104, 70), desert: rgb(214, 186, 120), tundra: rgb(170, 176, 150), hills: rgb(138, 140, 88),
	mountain: rgb(128, 114, 98), snow: rgb(236, 238, 242),
}

// racePrefs – где селится раса (по убыванию предпочтения).
var racePrefs = map[string][]biome{
	"human": {grass, beach, forest}, "halfelf": {grass, forest}, "aasimar": {grass, hills}, "genasi": {desert, beach, mountain},
	"elf": {forest, jungle}, "firbolg": {forest}, "dwarf": {mountain, hills}, "gnome": {hills, forest}, "halfling": {grass, hills},
	"halforc": {hills, grass}, "orc": {hills, tundra, mountain}, "goblin": {hills, forest}, "hobgoblin": {hills, grass}, "bugbear": {forest, hills},
	"kobold": {mountain, hills}, "dragonborn": {hills, desert}, "tiefling": {desert, grass}, "goliath": {mountain, snow},
	"aarakocra": {mountain}, "kenku": {forest, grass}, "tabaxi": {jungle, grass}, "lizardfolk": {swamp, jungle}, "tortle": {beach, swamp},
	"triton": {beach}, "minotaur": {hills, desert}, "warforged": {grass, hills}, "yuanti": {jungle, swamp, desert},
}

// wildFor – какие дикие места бывают в биоме.
var wildFor = map[biome][]string{
	grass: {"ruins", "camp", "tower"}, forest: {"camp", "ruins", "shrine", "lair"}, jungle: {"ruins", "lair"}, swamp: {"lair", "ruins"},
	desert: {"ruins", "dungeon"}, tundra: {"camp", "ruins"}, hills: {"dungeon", "camp", "mine"}, mountain: {"cave", "lair", "mine"},
	snow: {"cave", "lair"}, beach: {"cave"},
}

var wildAdj = []string{"Кривого Зуба", "Забытых Королей", "Чёрной Воды", "Шёпота", "Последнего Рыцаря", "Красной Луны", "Пепла",
	"Безымянного", "Ржавых Цепей", "Слепого Ворона", "Костяного Трона", "Туманов"}

// WorldOpts – параметры карты мира.
type WorldOpts struct {
	W, H int
}

type cell struct {
	x, y int
	b    biome
}

// World рисует карту мира и возвращает её картинку (data:URL) и места (поселения рас и дикие места) с населением.
func (g Gen) World(o WorldOpts) (string, []model.Place, error) {
	if o.W < 400 || o.H < 300 || o.W > 4000 || o.H > 3000 {
		return "", nil, errors.New("размер карты: от 400×300 до 4000×3000")
	}
	w, h := o.W, o.H
	hn, mn, warp := newNoise(g.R), newNoise(g.R), newNoise(g.R)
	height := make([]float64, w*h)
	moist := make([]float64, w*h)
	scale := 3.2 / float64(max(w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			nx, ny := float64(x)/float64(w)-0.5, float64(y)/float64(h)-0.5
			d := math.Sqrt(nx*nx*1.2 + ny*ny*1.6) // материк к центру, океан по краям
			fx, fy := float64(x)*scale, float64(y)*scale
			wx, wy := warp.fbm(fx*0.7, fy*0.7, 3), warp.fbm(fx*0.7+31, fy*0.7+17, 3)
			height[y*w+x] = hn.fbm(fx+wx*1.1, fy+wy*1.1, 6) + 0.38 - 0.95*d*d
			moist[y*w+x] = mn.fbm(float64(x)*scale*0.8+50, float64(y)*scale*0.8+50, 4)
		}
	}
	sea, hillsLvl, mountLvl := levels(height)
	biomes := make([]biome, w*h)
	for y := 0; y < h; y++ {
		lat := math.Abs(float64(y)/float64(h)-0.5) * 1.6 // 0 – экватор; карта – регион, а не весь мир
		for x := 0; x < w; x++ {
			i := y*w + x
			biomes[i] = classify(height[i], moist[i], lat, sea, hillsLvl, mountLvl)
		}
	}
	c := newCanvas(w, h, biomeColor[deep])
	g.paint(c, height, biomes, sea)
	g.rivers(c, height, biomes, mountLvl)
	g.glyphs(c, biomes)
	isWater := func(x, y int) bool {
		if x < 0 || y < 0 || x >= w || y >= h {
			return true
		}
		b := biomes[y*w+x]
		return b == deep || b == shallow
	}
	places := g.settle(w, h, biomes)
	g.roads(c, places, isWater)
	places = append(places, g.wilds(w, h, biomes, places)...)
	url, err := c.dataURL(86)
	if err != nil {
		return "", nil, err
	}
	return url, g.Populate(places, false), nil
}

// levels – уровни моря, холмов и гор по распределению высот: воды всегда около 45%, гор – верхние 6% суши.
func levels(height []float64) (sea, hillsLvl, mountLvl float64) {
	s := slices.Clone(height)
	sort.Float64s(s)
	at := func(q float64) float64 { return s[int(q*float64(len(s)-1))] }
	return at(0.45), at(0.83), at(0.94)
}

func classify(h, m, lat, sea, hillsLvl, mountLvl float64) biome {
	temp := 1 - lat*1.1 - math.Max(0, h-sea)*0.6
	switch {
	case h < sea-0.06:
		return deep
	case h < sea:
		return shallow
	case h < sea+0.01:
		return beach
	case h > mountLvl:
		if temp < 0.22 {
			return snow
		}
		return mountain
	case temp < 0.08:
		return snow
	case h > hillsLvl:
		return hills
	case temp < 0.2:
		return tundra
	case m < 0.4 && temp > 0.55:
		return desert
	case m > 0.62 && h < sea+0.05:
		return swamp
	case m > 0.53 && temp > 0.72:
		return jungle
	case m > 0.5:
		return forest
	}
	return grass
}

// paint – цвет биома с тенями от рельефа (свет с северо-запада), глубиной воды и мелкой «бумажной» фактурой.
func (g Gen) paint(c canvas, height []float64, biomes []biome, sea float64) {
	w, h := c.w(), c.h()
	grain := newNoise(g.R)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			b := biomes[i]
			col := biomeColor[b]
			gr := 0.94 + 0.12*grain.at(float64(x)*0.35, float64(y)*0.35)
			if b == deep || b == shallow {
				depth := math.Min(1, (sea-height[i])*6)
				col = lerpColor(biomeColor[shallow], biomeColor[deep], depth)
				c.img.SetRGBA(x, y, shade(col, gr))
				continue
			}
			l, u := height[max(0, i-3)], height[max(0, i-3*w)]
			slope := (height[i] - l) + (height[i] - u)
			relief := 7.0 // на равнинах тень мягкая, иначе проступают «волны» шума
			if b == hills || b == mountain || b == snow {
				relief = 22
			}
			k := 1 + slope*relief
			c.img.SetRGBA(x, y, shade(col, math.Max(0.6, math.Min(1.35, k))*gr))
		}
	}
	// береговая линия
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			i := y*w + x
			land := biomes[i] != deep && biomes[i] != shallow
			if land && (biomes[i-1] == shallow || biomes[i+1] == shallow || biomes[i-w] == shallow || biomes[i+w] == shallow) {
				c.blend(x, y, rgb(70, 60, 40), 0.55)
			}
		}
	}
}

// rivers – реки стекают с гор к морю по самому крутому спуску.
func (g Gen) rivers(c canvas, height []float64, biomes []biome, mountLvl float64) {
	w, h := c.w(), c.h()
	const step = 4
	water := rgb(58, 112, 160)
	var sources []cell
	for _, cl := range landCells(w, h, biomes, 12) {
		if height[cl.y*w+cl.x] > mountLvl-0.03 {
			sources = append(sources, cl)
		}
	}
	for k := 0; k < 12 && len(sources) > 0; k++ {
		src := sources[g.R.Intn(len(sources))]
		x, y := src.x, src.y
		width := 1.8
		for n := 0; n < 1500; n++ {
			bx, by, best := x, y, height[y*w+x]
			for _, d := range [][2]int{{step, 0}, {-step, 0}, {0, step}, {0, -step}, {step, step}, {-step, step}, {step, -step}, {-step, -step}} {
				nx, ny := x+d[0], y+d[1]
				if nx < 0 || ny < 0 || nx >= w || ny >= h {
					continue
				}
				if v := height[ny*w+nx]; v < best {
					bx, by, best = nx, ny, v
				}
			}
			if bx == x && by == y { // низина – маленькое озеро
				c.disc(float64(x), float64(y), 5, water, 0.9)
				break
			}
			c.line(float64(x), float64(y), float64(bx), float64(by), width, water, 0.85, 0, nil)
			x, y = bx, by
			width = math.Min(5, width+0.03)
			if biomes[y*w+x] == shallow || biomes[y*w+x] == deep {
				break
			}
		}
	}
}

// landCells – клетки суши с шагом step (кандидаты для поселений и диких мест).
func landCells(w, h int, biomes []biome, step int) []cell {
	var out []cell
	for y := step; y < h-step; y += step {
		for x := step; x < w-step; x += step {
			b := biomes[y*w+x]
			if b != deep && b != shallow {
				out = append(out, cell{x, y, b})
			}
		}
	}
	return out
}

func farFrom(x, y int, places []model.Place, d float64) bool {
	for _, p := range places {
		if math.Hypot(p.X-float64(x), p.Y-float64(y)) < d {
			return false
		}
	}
	return true
}

func nearWater(x, y, w, h int, biomes []biome) bool {
	for _, d := range [][2]int{{12, 0}, {-12, 0}, {0, 12}, {0, -12}} {
		nx, ny := x+d[0], y+d[1]
		if nx >= 0 && ny >= 0 && nx < w && ny < h && (biomes[ny*w+nx] == shallow || biomes[ny*w+nx] == deep) {
			return true
		}
	}
	return false
}

// settle – поселения: у каждой расы столица и 1–2 поселения в любимой местности.
func (g Gen) settle(w, h int, biomes []biome) []model.Place {
	cells := landCells(w, h, biomes, 8)
	races := g.races()
	minD := float64(max(w, h)) / (6 + float64(len(races))*0.9)
	var out []model.Place
	for _, race := range races {
		villages := 1 + g.R.Intn(2)
		for n := 0; n <= villages && len(out) < maxPlaces/2; n++ {
			c, ok := g.spot(cells, racePrefs[race], out, minD)
			if !ok {
				break
			}
			kind := "village"
			switch {
			case n == 0:
				kind = "capital"
			case nearWater(c.x, c.y, w, h, biomes) && g.R.Intn(2) == 0:
				kind = "port"
			case g.R.Intn(2) == 0:
				kind = "town"
			}
			out = append(out, model.Place{ID: g.NewID(), X: float64(c.x), Y: float64(c.y), Kind: kind, Race: race, Name: TownName(g.R, race)})
		}
	}
	return out
}

// spot – случайная подходящая клетка: сначала любимые биомы расы, потом любая суша.
func (g Gen) spot(cells []cell, prefs []biome, taken []model.Place, minD float64) (cell, bool) {
	for _, pass := range []func(c cell) bool{
		func(c cell) bool { return len(prefs) > 0 && c.b == prefs[0] },
		func(c cell) bool { return slices.Contains(prefs, c.b) },
		func(c cell) bool { return c.b != snow && c.b != beach },
	} {
		for _, d := range []float64{minD, minD * 0.6} {
			var cand []cell
			for _, c := range cells {
				if pass(c) && farFrom(c.x, c.y, taken, d) {
					cand = append(cand, c)
				}
			}
			if len(cand) > 0 {
				return cand[g.R.Intn(len(cand))], true
			}
		}
	}
	return cell{}, false
}

// wilds – логова, руины, пещеры и лагеря вдали от поселений.
func (g Gen) wilds(w, h int, biomes []biome, settled []model.Place) []model.Place {
	cells := landCells(w, h, biomes, 10)
	if len(cells) == 0 {
		return nil
	}
	n := max(6, len(settled)*4/5)
	minD := float64(max(w, h)) / 18
	var out []model.Place
	for i := 0; i < n*4 && len(out) < n; i++ {
		c := cells[g.R.Intn(len(cells))]
		kinds := wildFor[c.b]
		if len(kinds) == 0 || !farFrom(c.x, c.y, settled, minD) || !farFrom(c.x, c.y, out, minD*0.8) {
			continue
		}
		kind := pick(g.R, kinds)
		out = append(out, model.Place{ID: g.NewID(), X: float64(c.x), Y: float64(c.y), Kind: kind, Name: KindName(kind) + " " + pick(g.R, wildAdj)})
	}
	return out
}

// roads – дороги между поселениями (минимальное остовное дерево), по воде не рисуются.
func (g Gen) roads(c canvas, places []model.Place, water func(x, y int) bool) {
	if len(places) < 2 {
		return
	}
	in := map[int]bool{0: true}
	road := rgb(120, 86, 48)
	for len(in) < len(places) {
		bi, bj, bd := -1, -1, math.MaxFloat64
		for i := range in {
			for j := range places {
				if in[j] {
					continue
				}
				if d := dist(&places[i], &places[j]); d < bd {
					bi, bj, bd = i, j, d
				}
			}
		}
		in[bj] = true
		if bd > float64(max(c.w(), c.h()))/3 { // слишком далеко – морской путь, не дорога
			continue
		}
		a, b := places[bi], places[bj]
		c.line(a.X, a.Y, b.X, b.Y, 2.2, road, 0.75, 6, water)
	}
}

// glyphs – значки рельефа поверх заливки: горы-треугольники со светлым и тёмным склоном, холмики, кроны деревьев.
func (g Gen) glyphs(c canvas, biomes []biome) {
	w, h := c.w(), c.h()
	at := func(x, y int) biome { return biomes[max(0, min(h-1, y))*w+max(0, min(w-1, x))] }
	for y := 6; y < h; y += 11 {
		for x := 6 + (y/11%2)*5; x < w; x += 11 {
			jx, jy := x+g.R.Intn(5)-2, y+g.R.Intn(5)-2
			switch at(jx, jy) {
			case forest, jungle:
				crown := rgb(34, 78, 38)
				if at(jx, jy) == jungle {
					crown = rgb(22, 82, 44)
				}
				c.disc(float64(jx), float64(jy), 3.6, shade(crown, 0.8), 0.9)
				c.disc(float64(jx)-1, float64(jy)-1, 2.4, shade(crown, 1.25), 0.8)
			case swamp:
				if g.R.Intn(2) == 0 {
					c.line(float64(jx-3), float64(jy), float64(jx+3), float64(jy), 1, rgb(60, 84, 70), 0.8, 0, nil)
				}
			}
		}
	}
	for y := 10; y < h; y += 20 {
		for x := 10 + (y/20%2)*10; x < w; x += 20 {
			jx, jy := x+g.R.Intn(7)-3, y+g.R.Intn(7)-3
			switch b := at(jx, jy); b {
			case mountain, snow:
				if g.R.Intn(5) < 2 { // не сплошной сеткой – хребты выглядят живее
					continue
				}
				size := 10 + float64(g.R.Intn(8))
				peak := rgb(240, 240, 244)
				if b == mountain {
					peak = rgb(150, 134, 116)
				}
				c.mountain(float64(jx), float64(jy), size, peak)
			case hills:
				if g.R.Intn(2) == 0 {
					c.hill(float64(jx), float64(jy), 6+float64(g.R.Intn(3)))
				}
			}
		}
	}
}
