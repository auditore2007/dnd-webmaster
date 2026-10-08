// Package travel – путешествия по карте мира: сетка местности, поиск пути, мили и дни в дороге.
package travel

import (
	"container/heap"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"math"
	"strconv"
	"strings"
)

// Местность клетки сетки путешествий.
const (
	Water uint8 = iota
	Road
	Plain
	Forest
	Hills
	Mountain
	Swamp
	Desert
	Snow
)

// Names – название местности по-русски.
var Names = []string{"вода", "дорога", "равнина", "лес", "холмы", "горы", "болото", "пустыня", "снега"}

// slow – во сколько раз медленнее идти по местности, чем по дороге (труднопроходимая местность вдвое медленнее).
var slow = []float64{math.Inf(1), 1, 1.2, 2, 1.8, 3, 2.5, 1.6, 2.2}

// Paces – миль в день при темпе (8 часов в пути, по правилам D&D).
var Paces = map[string]float64{"fast": 30, "normal": 24, "slow": 18}

// Grid – сетка местности: клетка step×step пикселей картинки карты.
type Grid struct {
	Cols, Rows, Step int
	Cells            []uint8
}

const header = "travel1"

// Encode – сетка строкой для хранения рядом с картой.
func (g *Grid) Encode() string {
	return fmt.Sprintf("%s;%d;%d;%d;%s", header, g.Cols, g.Rows, g.Step, base64.StdEncoding.EncodeToString(g.Cells))
}

// Decode разбирает строку Encode.
func Decode(s string) (*Grid, error) {
	f := strings.Split(s, ";")
	if len(f) != 5 || f[0] != header {
		return nil, errors.New("сетка местности повреждена")
	}
	cols, e1 := strconv.Atoi(f[1])
	rows, e2 := strconv.Atoi(f[2])
	step, e3 := strconv.Atoi(f[3])
	cells, e4 := base64.StdEncoding.DecodeString(f[4])
	if err := errors.Join(e1, e2, e3, e4); err != nil || cols <= 0 || rows <= 0 || step <= 0 || len(cells) != cols*rows {
		return nil, errors.New("сетка местности повреждена")
	}
	return &Grid{cols, rows, step, cells}, nil
}

// FromImage угадывает местность по цветам картинки (для своих карт и миров без сетки).
func FromImage(img image.Image, step int) *Grid {
	b := img.Bounds()
	g := &Grid{Cols: max(1, b.Dx()/step), Rows: max(1, b.Dy()/step), Step: step}
	g.Cells = make([]uint8, g.Cols*g.Rows)
	for j := range g.Rows {
		for i := range g.Cols {
			r, gg, bb, _ := img.At(b.Min.X+i*step+step/2, b.Min.Y+j*step+step/2).RGBA()
			g.Cells[j*g.Cols+i] = classify(int(r>>8), int(gg>>8), int(bb>>8))
		}
	}
	return g
}

func classify(r, g, b int) uint8 {
	mx, mn := max(r, g, b), min(r, g, b)
	switch {
	case b > r+25 && b > g-10:
		return Water
	case r > 205 && g > 205 && b > 205:
		return Snow
	case r > 175 && g > 150 && b < 150 && r-b > 45:
		return Desert
	case g > r && g < 110:
		return Forest
	case mx-mn < 28 && mx < 175:
		return Mountain
	}
	return Plain
}

func (g *Grid) at(i, j int) uint8 { return g.Cells[j*g.Cols+i] }

func (g *Grid) cellOf(x, y float64) (int, int) {
	return min(g.Cols-1, max(0, int(x)/g.Step)), min(g.Rows-1, max(0, int(y)/g.Step))
}

// Point – точка на картинке карты.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Leg – сколько миль пути пришлось на местность.
type Leg struct {
	Terrain string  `json:"terrain"`
	Miles   float64 `json:"miles"`
}

// Route – путь по карте: точки, мили, дни и что за местность по дороге.
type Route struct {
	Path  []Point `json:"path"`
	Miles float64 `json:"miles"`
	Days  float64 `json:"days"`
	Legs  []Leg   `json:"legs"`
	// Along[i] – местность в точке Path[i]
	Along []uint8 `json:"-"`
}

// Plan ищет путь от from до to (A* по клеткам, восемь направлений) и считает дни при темпе pace.
// milesPerPx – масштаб карты.
func Plan(g *Grid, from, to Point, pace string, milesPerPx float64) (Route, error) {
	perDay, ok := Paces[pace]
	if !ok {
		return Route{}, fmt.Errorf("неизвестный темп %q", pace)
	}
	si, sj := g.cellOf(from.X, from.Y)
	ti, tj := g.cellOf(to.X, to.Y)
	if g.at(ti, tj) == Water {
		return Route{}, errors.New("туда по суше не пройти – это вода")
	}
	n := g.Cols * g.Rows
	cost := make([]float64, n)
	prev := make([]int32, n)
	for i := range cost {
		cost[i], prev[i] = math.Inf(1), -1
	}
	start, goal := sj*g.Cols+si, tj*g.Cols+ti
	cost[start] = 0
	q := &pq{{start, 0}}
	for q.Len() > 0 {
		c := heap.Pop(q).(item).cell
		if c == goal {
			break
		}
		ci, cj := c%g.Cols, c/g.Cols
		for dj := -1; dj <= 1; dj++ {
			for di := -1; di <= 1; di++ {
				ni, nj := ci+di, cj+dj
				if (di == 0 && dj == 0) || ni < 0 || nj < 0 || ni >= g.Cols || nj >= g.Rows {
					continue
				}
				t := g.at(ni, nj)
				if t == Water && !(ni == si && nj == sj) {
					continue
				}
				step := math.Hypot(float64(di), float64(dj)) * (slow[t] + slow[g.at(ci, cj)]) / 2
				nc := cost[c] + step
				k := nj*g.Cols + ni
				if nc < cost[k] {
					cost[k], prev[k] = nc, int32(c)
					heap.Push(q, item{k, nc + math.Hypot(float64(ni-ti), float64(nj-tj))})
				}
			}
		}
	}
	if math.IsInf(cost[goal], 1) {
		return Route{}, errors.New("по суше туда не добраться – нужен корабль")
	}
	var cells []int
	for c := goal; c != -1; c = int(prev[c]) {
		cells = append(cells, c)
	}
	r := Route{}
	miles := map[uint8]float64{}
	for k := len(cells) - 1; k >= 0; k-- {
		c := cells[k]
		p := Point{(float64(c%g.Cols) + 0.5) * float64(g.Step), (float64(c/g.Cols) + 0.5) * float64(g.Step)}
		if k == len(cells)-1 {
			p = from
		} else if k == 0 {
			p = to
		}
		t := g.Cells[c]
		if len(r.Path) > 0 {
			last := r.Path[len(r.Path)-1]
			m := math.Hypot(p.X-last.X, p.Y-last.Y) * milesPerPx
			r.Miles += m
			miles[t] += m
			r.Days += m * slow[t] / perDay
		}
		r.Path = append(r.Path, p)
		r.Along = append(r.Along, t)
	}
	for t := Road; t <= Snow; t++ {
		if miles[t] > 0.05 {
			r.Legs = append(r.Legs, Leg{Names[t], math.Round(miles[t]*10) / 10})
		}
	}
	r.Miles = math.Round(r.Miles*10) / 10
	r.Days = math.Ceil(r.Days*4) / 4 // с точностью до четверти дня
	return r, nil
}

// At – точка и местность на доле пути t (0..1) по пройденным милям.
func (r Route) At(t float64) (Point, uint8) {
	if len(r.Path) == 0 {
		return Point{}, Plain
	}
	total := 0.0
	for i := 1; i < len(r.Path); i++ {
		total += math.Hypot(r.Path[i].X-r.Path[i-1].X, r.Path[i].Y-r.Path[i-1].Y)
	}
	want, run := t*total, 0.0
	for i := 1; i < len(r.Path); i++ {
		run += math.Hypot(r.Path[i].X-r.Path[i-1].X, r.Path[i].Y-r.Path[i-1].Y)
		if run >= want {
			return r.Path[i], r.Along[i]
		}
	}
	return r.Path[len(r.Path)-1], r.Along[len(r.Along)-1]
}

type item struct {
	cell int
	pri  float64
}

type pq []item

func (q pq) Len() int { return len(q) }
func (q pq) Less(i, j int) bool {
	if q[i].pri != q[j].pri {
		return q[i].pri < q[j].pri
	}
	return q[i].cell < q[j].cell
}
func (q pq) Swap(i, j int) { q[i], q[j] = q[j], q[i] }
func (q *pq) Push(x any)   { *q = append(*q, x.(item)) }
func (q *pq) Pop() any {
	old := *q
	it := old[len(old)-1]
	*q = old[:len(old)-1]
	return it
}
