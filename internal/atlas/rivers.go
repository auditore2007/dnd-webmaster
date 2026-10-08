package atlas

import (
	"container/heap"
	"math"
	"sort"
)

// Реки по Azgaar (MIT, см. LICENSE-AZGAAR): впадины заполняются, вода с каждой клетки стекает к самому низкому соседу,
// поток копится; где он больше minRiverFlux, начинается река. При слиянии имя остаётся у полноводной.

const (
	minRiverFlux = 30
	lakeDepth    = 3  // впадина глубже – озеро
	maxLakeCells = 40 // котловина больше – не озеро, а низина
)

// River – река: клетки от истока к устью; Out – уходит за край карты.
type River struct {
	Cells []int
	Out   bool
}

// fillDepressions – высоты, по которым вода гарантированно стекает к морю или краю (priority-flood).
func fillDepressions(w *World) []float64 {
	g := w.g
	alt := make([]float64, g.size())
	seen := make([]bool, g.size())
	pq := &cellQueue{}
	for i, h := range w.H {
		alt[i] = float64(h)
		if w.ocean[i] || g.border[i] {
			seen[i] = true
			heap.Push(pq, qItem{i, alt[i]})
		}
	}
	for pq.Len() > 0 {
		c := heap.Pop(pq).(qItem).cell
		for _, n := range g.nb[c] {
			if seen[n] {
				continue
			}
			seen[n] = true
			if alt[n] <= alt[c] {
				alt[n] = alt[c] + 0.01
			}
			heap.Push(pq, qItem{int(n), alt[n]})
		}
	}
	return alt
}

// markWater – океан (вода у края карты и большие внутренние моря) и озёра: малые замкнутые водоёмы и глубокие впадины.
func markWater(w *World) {
	g := w.g
	w.ocean = make([]bool, g.size())
	seen := make([]bool, g.size())
	inlandSea := g.size() / 60
	for i := range w.H {
		if seen[i] || w.H[i] >= seaLevel {
			continue
		}
		body := component(g, i, seen, func(c int) bool { return w.H[c] < seaLevel })
		edge := false
		for _, c := range body {
			edge = edge || g.border[c]
		}
		if edge || len(body) > inlandSea {
			for _, c := range body {
				w.ocean[c] = true
			}
		}
	}
	w.alt = fillDepressions(w)
	w.Lake = make([]bool, g.size())
	deep := func(c int) bool { return !w.ocean[c] && w.alt[c]-float64(w.H[c]) >= 1 }
	clear(seen)
	for i := range w.H {
		if seen[i] || !deep(i) {
			continue
		}
		pool := component(g, i, seen, deep)
		depth := 0.0
		for _, c := range pool {
			depth = math.Max(depth, w.alt[c]-float64(w.H[c]))
		}
		if depth < lakeDepth || len(pool) > maxLakeCells {
			continue // мелкая лужа или огромная котловина – суша
		}
		for _, c := range pool {
			w.Lake[c] = !g.border[c]
		}
	}
	for i, h := range w.H {
		if h < seaLevel && !w.ocean[i] {
			w.Lake[i] = true
		}
	}
}

// component – связная область клеток, подходящих под ok, начиная с start.
func component(g *grid, start int, seen []bool, ok func(int) bool) []int {
	seen[start] = true
	out := []int{start}
	for qi := 0; qi < len(out); qi++ {
		for _, n := range g.nb[out[qi]] {
			if !seen[n] && ok(int(n)) {
				seen[n] = true
				out = append(out, int(n))
			}
		}
	}
	return out
}

func rivers(w *World) {
	g := w.g
	n := g.size()
	w.Flux, w.River = make([]float64, n), make([]int32, n)
	conf := make([]float64, n)
	cells := make([]int, 0, n)
	for i := range n {
		if !w.ocean[i] {
			cells = append(cells, i)
		}
	}
	sort.SliceStable(cells, func(a, b int) bool { return w.alt[cells[a]] > w.alt[cells[b]] })
	modifier := math.Pow(float64(n)/10000, 0.25)
	data := map[int32][]int{}
	next := int32(1)
	out := map[int32]bool{}
	flowDown := func(to int, flux float64, river int32) {
		if r := w.River[to]; r != 0 {
			if flux > w.Flux[to]-conf[to] {
				conf[to] += w.Flux[to]
				w.River[to] = river
			} else {
				conf[to] += flux
			}
		} else {
			w.River[to] = river
		}
		data[river] = append(data[river], to)
		if !w.ocean[to] {
			w.Flux[to] += flux
		}
	}
	for _, i := range cells {
		w.Flux[i] += float64(w.Prec[i]) / modifier
		if g.border[i] {
			if w.River[i] != 0 {
				out[w.River[i]] = true
			}
			continue
		}
		low := int(g.nb[i][0])
		for _, c := range g.nb[i] {
			if w.alt[c] < w.alt[low] {
				low = int(c)
			}
		}
		if w.alt[i] <= w.alt[low] {
			continue
		}
		if w.Flux[i] < minRiverFlux {
			if !w.ocean[low] {
				w.Flux[low] += w.Flux[i]
			}
			continue
		}
		if w.River[i] == 0 {
			w.River[i] = next
			data[next] = []int{i}
			next++
		}
		flowDown(low, w.Flux[i], w.River[i])
	}
	for id := int32(1); id < next; id++ {
		if cs := data[id]; len(cs) >= 3 {
			w.Rivers = append(w.Rivers, River{Cells: cs, Out: out[id]})
		}
	}
	// клетки коротких ручьёв рекой не считаются
	keep := make([]bool, n)
	for _, r := range w.Rivers {
		for _, c := range r.Cells {
			keep[c] = true
		}
	}
	for i := range w.River {
		if !keep[i] {
			w.River[i] = 0
		}
	}
}

type qItem struct {
	cell int
	pri  float64
}

// cellQueue – очередь клеток с приоритетом (меньший – раньше; при равенстве – меньший номер, ради повторяемости).
type cellQueue []qItem

func (q cellQueue) Len() int { return len(q) }
func (q cellQueue) Less(i, j int) bool {
	if q[i].pri != q[j].pri {
		return q[i].pri < q[j].pri
	}
	return q[i].cell < q[j].cell
}
func (q cellQueue) Swap(i, j int) { q[i], q[j] = q[j], q[i] }
func (q *cellQueue) Push(x any)   { *q = append(*q, x.(qItem)) }
func (q *cellQueue) Pop() any {
	old := *q
	it := old[len(old)-1]
	*q = old[:len(old)-1]
	return it
}
