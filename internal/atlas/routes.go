package atlas

import (
	"container/heap"
	"math"
)

// Дороги: города каждого государства связаны деревом кратчайших связей, столицы – между собой;
// путь ищется по клеткам (A*), горы и болота дороже, уже проложенная дорога дешевле. Порты связаны морем.

const maxRoadSpan = 0.45 // дальше этой доли ширины карты столицы дорогой не связываются

func routes(w *World) {
	used := make([]bool, w.g.size())
	byState := map[int][]int{}
	for i, b := range w.Burgs {
		byState[b.State] = append(byState[b.State], i)
	}
	for s := range w.States {
		for _, e := range spanningTree(w, byState[s], math.Inf(1)) {
			w.addRoad(e, used)
		}
	}
	var caps, ports []int
	for i, b := range w.Burgs {
		if b.Kind == "capital" {
			caps = append(caps, i)
		}
		if b.Kind == "port" || b.Kind == "capital" && w.coastal(b.Cell) {
			ports = append(ports, i)
		}
	}
	for _, e := range spanningTree(w, caps, float64(w.Width)*maxRoadSpan) {
		w.addRoad(e, used)
	}
	for _, e := range spanningTree(w, ports, math.Inf(1)) {
		a, b := w.Burgs[e[0]], w.Burgs[e[1]]
		if a.State == b.State && w.connectedByRoad(e) {
			continue
		}
		if path := w.searchPath(seaStart(w, a.Cell), seaStart(w, b.Cell), true, nil); len(path) > 1 {
			w.Sea = append(w.Sea, append(append([]int{a.Cell}, path...), b.Cell))
		}
	}
}

func (w *World) addRoad(e [2]int, used []bool) {
	path := w.searchPath(w.Burgs[e[0]].Cell, w.Burgs[e[1]].Cell, false, used)
	if len(path) < 2 {
		return
	}
	for _, c := range path {
		used[c] = true
	}
	w.Roads = append(w.Roads, path)
}

func (w *World) connectedByRoad(e [2]int) bool {
	a, b := w.Burgs[e[0]].Cell, w.Burgs[e[1]].Cell
	for _, r := range w.Roads {
		if r[0] == a && r[len(r)-1] == b || r[0] == b && r[len(r)-1] == a {
			return true
		}
	}
	return false
}

func seaStart(w *World, c int) int {
	for _, n := range w.g.nb[c] {
		if w.ocean[n] {
			return int(n)
		}
	}
	return c
}

// spanningTree – рёбра минимального остовного дерева (Прим) по прямому расстоянию; длиннее maxLen не берутся.
func spanningTree(w *World, ids []int, maxLen float64) [][2]int {
	if len(ids) < 2 {
		return nil
	}
	in := make([]bool, len(ids))
	in[0] = true
	var out [][2]int
	for range len(ids) - 1 {
		bi, bj, bd := -1, -1, math.Inf(1)
		for i := range ids {
			if !in[i] {
				continue
			}
			for j := range ids {
				if in[j] {
					continue
				}
				a, b := w.Burgs[ids[i]], w.Burgs[ids[j]]
				if d := math.Hypot(a.X-b.X, a.Y-b.Y); d < bd {
					bi, bj, bd = i, j, d
				}
			}
		}
		if bj < 0 {
			break
		}
		in[bj] = true
		if bd <= maxLen {
			out = append(out, [2]int{ids[bi], ids[bj]})
		}
	}
	return out
}

// moveCost – множитель цены шага по суше.
func (w *World) moveCost(c int, used []bool) float64 {
	if used != nil && used[c] {
		return 0.5
	}
	k := 1.0
	switch w.TerrainOf(c) {
	case Mountain:
		k = 5
	case Snow:
		k = 7
	case Hills:
		k = 2
	case Forest:
		k = 1.4
	case Jungle:
		k = 2
	case Swamp:
		k = 3
	case Desert:
		k = 1.6
	case Lake:
		return math.Inf(1)
	}
	if w.River[c] != 0 {
		k += 0.5
	}
	return k
}

// searchPath – путь A* от a до b по суше (sea – по морю); nil, если пути нет.
func (w *World) searchPath(a, b int, sea bool, used []bool) []int {
	n := w.g.size()
	cost := make([]float64, n)
	from := make([]int32, n)
	for i := range cost {
		cost[i] = math.Inf(1)
		from[i] = -1
	}
	minK := 1.0
	if used != nil {
		minK = 0.5
	}
	cost[a] = 0
	pq := &cellQueue{{a, 0}}
	for pq.Len() > 0 {
		c := heap.Pop(pq).(qItem).cell
		if c == b {
			break
		}
		for _, e := range w.g.nb[c] {
			ei := int(e)
			if ei != b && sea != w.ocean[ei] || !sea && !w.land(ei) && ei != b {
				continue
			}
			k := 1.0
			if !sea {
				k = w.moveCost(ei, used)
			}
			nc := cost[c] + w.g.dist(c, ei)*k
			if nc < cost[ei] {
				cost[ei] = nc
				from[ei] = int32(c)
				heap.Push(pq, qItem{ei, nc + math.Hypot(w.g.px[ei]-w.g.px[b], w.g.py[ei]-w.g.py[b])*minK})
			}
		}
	}
	if math.IsInf(cost[b], 1) {
		return nil
	}
	var path []int
	for c := b; c != -1; c = int(from[c]) {
		path = append(path, c)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}
