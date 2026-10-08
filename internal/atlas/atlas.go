// Package atlas – генератор карты мира по алгоритмам Azgaar's Fantasy Map Generator (MIT, см. LICENSE-AZGAAR),
// перенесённым на Go: рельеф по шаблонам, климат, биомы, реки, государства рас, города и дороги;
// рисуется в стиле иллюстрированного атласа (render*.go).
package atlas

import (
	"errors"
	"image/color"
	"math"
	"slices"
)

// Terrain – местность клетки для расселения рас и диких мест.
type Terrain uint8

const (
	Sea Terrain = iota
	Lake
	Coast
	Grass
	Forest
	Jungle
	Swamp
	Desert
	Tundra
	Hills
	Mountain
	Snow
)

// People – народ мира: раса, любимая местность (по убыванию) и как он называет города и государство.
type People struct {
	Race      string
	Likes     []Terrain
	TownName  func() string
	StateName func(capital string) string
}

// Opts – параметры мира.
type Opts struct {
	W, H     int
	Template string // id шаблона рельефа; пусто – случайный
	Title    string // название карты на картуше
	SeaName  string
	Peoples  []People
	R        Rand
}

// World – готовый мир: клетки с высотами, климатом и биомами, реки, государства, города и дороги.
type World struct {
	Title, SeaName string
	Template       string
	Width, Height  int

	g          *grid
	H          []uint8 // высота клетки 0–100, суша от 20
	Temp       []int8
	Prec       []uint8
	Biome      []uint8
	Flux       []float64
	River      []int32
	Lake       []bool
	ocean      []bool
	alt        []float64
	Rivers     []River
	latN, latT float64

	State  []int16 // государство клетки, -1 – ничьи земли
	States []State
	Burgs  []Burg
	Sites  []Site  // дикие места; заполняет вызывающий до Render
	Roads  [][]int // пути по клеткам
	Sea    [][]int // морские пути
	score  []float64
}

// State – государство одной расы.
type State struct {
	Name    string
	Race    string
	Capital int // номер города в Burgs
	Color   color.RGBA
	Cells   int
}

// Burg – поселение. Kind: capital, city, town, village, port, castle.
type Burg struct {
	X, Y  float64
	Cell  int
	Name  string
	Race  string
	State int
	Kind  string
}

const (
	pxPerCell   = 160 // площадь клетки в пикселях: 1600×1000 – 10 000 клеток
	minLandPart = 0.22
	maxLandPart = 0.85 // без моря атлас скучный
	maxAttempts = 4
)

// Generate строит мир. Если суши слишком мало для всех народов, шаблон перебрасывается.
func Generate(o Opts) (*World, error) {
	if o.W < 400 || o.H < 300 || o.R == nil {
		return nil, errors.New("atlas: размер карты от 400×300")
	}
	r := rng{o.R}
	var w *World
	for attempt := range maxAttempts {
		template := o.Template
		if template == "" || attempt > 0 {
			template = pickTemplate(r)
		}
		var err error
		w, err = build(o, r, template)
		if err != nil {
			return nil, err
		}
		if part := w.landPart(); part >= minLandPart && part <= maxLandPart || o.Template != "" && attempt == 0 && part > 0.08 {
			break
		}
	}
	settle(w, o.Peoples, r)
	return w, nil
}

func build(o Opts, r rng, template string) (*World, error) {
	g := newGrid(o.W, o.H, max(500, o.W*o.H/pxPerCell), r)
	h, err := buildHeights(g, r, template)
	if err != nil {
		return nil, err
	}
	w := &World{Title: o.Title, SeaName: o.SeaName, Template: template, Width: o.W, Height: o.H, g: g, H: h}
	// карта – область северного полушария: от холодного севера до субтропиков на юге
	w.latT = r.between(35, 50)
	w.latN = r.between(math.Max(w.latT+5, 40), 56)
	markWater(w)
	w.Temp = temperatures(w)
	w.Prec = precipitation(w, r)
	rivers(w)
	w.Biome = biomes(w)
	return w, nil
}

func (w *World) land(c int) bool { return w.H[c] >= seaLevel && !w.Lake[c] }

func (w *World) landPart() float64 {
	n := 0
	for i := range w.H {
		if w.H[i] >= seaLevel {
			n++
		}
	}
	return float64(n) / float64(len(w.H))
}

func (w *World) coastal(c int) bool {
	for _, n := range w.g.nb[c] {
		if w.ocean[n] {
			return true
		}
	}
	return false
}

// TerrainOf – местность клетки.
func (w *World) TerrainOf(c int) Terrain {
	switch {
	case w.ocean[c]:
		return Sea
	case w.Lake[c]:
		return Lake
	case w.H[c] >= 70:
		if w.Temp[c] < 0 {
			return Snow
		}
		return Mountain
	case w.Biome[c] == bGlacier:
		return Snow
	case w.H[c] >= 50:
		return Hills
	case w.coastal(c) && w.H[c] < 24:
		return Coast
	}
	switch w.Biome[c] {
	case bHotDesert, bColdDesert:
		return Desert
	case bSavanna, bGrassland:
		return Grass
	case bTropicalRainforest:
		return Jungle
	case bWetland:
		return Swamp
	case bTundra:
		return Tundra
	}
	return Forest
}

// Spot – клетка суши: где на карте и какая там местность.
type Spot struct {
	X, Y    float64
	Terrain Terrain
	State   int
}

// LandSpots – клетки суши вдали от краёв (кандидаты для диких мест).
func (w *World) LandSpots() []Spot {
	var out []Spot
	margin := 40.0
	for c := range w.H {
		x, y := w.g.px[c], w.g.py[c]
		if !w.land(c) || x < margin || y < margin || x > float64(w.Width)-margin || y > float64(w.Height)-margin {
			continue
		}
		out = append(out, Spot{x, y, w.TerrainOf(c), int(w.State[c])})
	}
	return out
}

func likes(p People, t Terrain) int {
	if i := slices.Index(p.Likes, t); i >= 0 {
		return len(p.Likes) - i
	}
	return 0
}
