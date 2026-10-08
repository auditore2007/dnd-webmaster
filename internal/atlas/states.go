package atlas

import (
	"container/heap"
	"image/color"
	"math"
	"sort"
)

// Расселение по мотивам Azgaar (MIT, см. LICENSE-AZGAAR): клетки оцениваются по пригодности для жизни,
// у каждого народа – столица в любимой местности, государства растут от столиц, пока хватает «сил»
// на продвижение через чужие биомы, горы и реки; потом – города, замки и дороги.

var stateColors = []color.RGBA{
	{168, 58, 48, 255}, {52, 86, 156, 255}, {62, 124, 64, 255}, {196, 150, 46, 255}, {118, 66, 138, 255},
	{40, 126, 128, 255}, {198, 104, 44, 255}, {124, 84, 52, 255}, {150, 40, 82, 255}, {84, 98, 124, 255},
	{136, 150, 40, 255}, {176, 92, 120, 255}, {70, 70, 70, 255}, {30, 100, 90, 255}, {200, 80, 80, 255},
}

func settle(w *World, peoples []People, r rng) {
	w.State = make([]int16, w.g.size())
	for i := range w.State {
		w.State[i] = -1
	}
	rank(w)
	if len(peoples) == 0 {
		return
	}
	capitals(w, peoples, r)
	expand(w, peoples, r)
	towns(w, peoples, r)
	castles(w, peoples, r)
	routes(w)
}

// rank – пригодность клетки для поселения: биом, река рядом, низина, побережье, озеро.
func rank(w *World) {
	w.score = make([]float64, w.g.size())
	var sum, top float64
	var n int
	for c, f := range w.Flux {
		if w.River[c] != 0 {
			sum += f
			n++
			top = math.Max(top, f)
		}
	}
	mean := sum / math.Max(1, float64(n))
	for c := range w.H {
		if !w.land(c) {
			continue
		}
		s := habitability[w.Biome[c]]
		if s == 0 {
			continue
		}
		if w.River[c] != 0 && top > mean {
			s += clamp((w.Flux[c]-mean)/(top-mean), 0, 1) * 250
		}
		s -= (float64(w.H[c]) - 50) / 5
		if w.coastal(c) {
			s += 5
			if w.River[c] != 0 {
				s += 15 // устье – лучшее место для порта
			}
		}
		for _, nb := range w.g.nb[c] {
			if w.Lake[nb] {
				s += 10
				break
			}
		}
		w.score[c] = math.Max(0, s/5)
	}
}

// inner – клетка не у самого края карты (там рамка и надписи).
func (w *World) inner(c int) bool {
	m := 2.5 * w.g.s
	x, y := w.g.px[c], w.g.py[c]
	return x > m && y > m && x < float64(w.Width)-m && y < float64(w.Height)-m
}

func (w *World) farFromBurgs(c int, d float64) bool {
	for _, b := range w.Burgs {
		if math.Hypot(b.X-w.g.px[c], b.Y-w.g.py[c]) < d {
			return false
		}
	}
	return true
}

func (w *World) landCells() int {
	n := 0
	for c := range w.H {
		if w.land(c) {
			n++
		}
	}
	return n
}

// spacing – расстояние между поселениями, чтобы count штук разошлись по всей суше.
func (w *World) spacing(count int) float64 {
	area := float64(w.landCells()) * w.g.s * w.g.s
	return math.Sqrt(area / float64(max(1, count)))
}

func capitals(w *World, peoples []People, r rng) {
	minD := w.spacing(len(peoples)) * 0.7
	for i, p := range peoples {
		type cand struct {
			c int
			v float64
		}
		var cands []cand
		for c, s := range w.score {
			if s <= 0 || !w.inner(c) {
				continue
			}
			k := 0.5
			if l := likes(p, w.TerrainOf(c)); l > 0 {
				k = 1 + float64(l)
			}
			cands = append(cands, cand{c, (s + 2) * k * (0.6 + r.float()*0.8)})
		}
		sort.SliceStable(cands, func(a, b int) bool { return cands[a].v > cands[b].v })
		chosen := -1
		for d := minD; chosen < 0 && d > 1; d *= 0.75 {
			for _, cd := range cands {
				if w.farFromBurgs(cd.c, d) {
					chosen = cd.c
					break
				}
			}
		}
		if chosen < 0 {
			chosen = anyFreeLand(w)
		}
		if chosen < 0 {
			continue
		}
		name := p.TownName()
		w.States = append(w.States, State{Name: p.StateName(name), Race: p.Race, Capital: len(w.Burgs),
			Color: stateColors[i%len(stateColors)]})
		w.Burgs = append(w.Burgs, Burg{X: w.g.px[chosen], Y: w.g.py[chosen], Cell: chosen, Name: name, Race: p.Race,
			State: len(w.States) - 1, Kind: "capital"})
	}
}

func anyFreeLand(w *World) int {
	for c := range w.H {
		if w.land(c) && w.farFromBurgs(c, w.g.s) {
			return c
		}
	}
	return -1
}

func peopleOf(peoples []People, race string) People {
	for _, p := range peoples {
		if p.Race == race {
			return p
		}
	}
	return People{}
}

// stepCost – цена продвижения государства народа p в клетку c.
func stepCost(w *World, p People, c int) float64 {
	t := w.TerrainOf(c)
	cost := 10.0
	if likes(p, t) > 0 {
		cost += 10
	} else {
		cost += biomeCost[w.Biome[c]]
	}
	switch t {
	case Mountain, Snow:
		if likes(p, Mountain) > 0 || likes(p, Snow) > 0 {
			cost += 50
		} else {
			cost += 2200
		}
	case Hills:
		if likes(p, Hills) == 0 {
			cost += 300
		}
	case Lake:
		cost += 100
	}
	if w.River[c] != 0 {
		cost += clamp(w.Flux[c]/10, 20, 100)
	}
	return cost
}

// expand – государства растут от столиц одновременно: клетка достаётся тому, кому дешевле до неё дойти.
func expand(w *World, peoples []People, r rng) {
	n := w.g.size()
	cost := make([]float64, n)
	for i := range cost {
		cost[i] = math.Inf(1)
	}
	limit := float64(n) / 2.5
	power := make([]float64, len(w.States))
	pq := &stateQueue{}
	for i, s := range w.States {
		power[i] = r.between(1, 1.6)
		c := w.Burgs[s.Capital].Cell
		w.State[c] = int16(i)
		cost[c] = 0
		heap.Push(pq, stateItem{c, 0, i})
	}
	for pq.Len() > 0 {
		it := heap.Pop(pq).(stateItem)
		if it.pri > cost[it.cell] {
			continue
		}
		p := peopleOf(peoples, w.States[it.state].Race)
		for _, e := range w.g.nb[it.cell] {
			if w.ocean[e] {
				continue
			}
			c := it.pri + stepCost(w, p, int(e))/power[it.state]
			if c > limit || c >= cost[e] {
				continue
			}
			cost[e] = c
			w.State[e] = int16(it.state)
			heap.Push(pq, stateItem{int(e), c, it.state})
		}
	}
	for c, s := range w.State {
		if s >= 0 && w.land(c) {
			w.States[s].Cells++
		}
	}
}

// towns – города каждого государства, по числу его клеток; у моря – порты.
func towns(w *World, peoples []People, r rng) {
	land := w.landCells()
	target := min(70, max(len(w.States)*3, land/110))
	minD := w.spacing(target) * 0.6
	for si, st := range w.States {
		want := int(math.Round(float64(target)*float64(st.Cells)/float64(max(1, land)))) - 1
		want = min(6, max(1, want))
		p := peopleOf(peoples, st.Race)
		type cand struct {
			c int
			v float64
		}
		var cands []cand
		for c, s := range w.score {
			if int(w.State[c]) == si && s > 0 && w.inner(c) {
				cands = append(cands, cand{c, s * (0.5 + r.float())})
			}
		}
		sort.SliceStable(cands, func(a, b int) bool { return cands[a].v > cands[b].v })
		added := 0
		for _, cd := range cands {
			if added >= want {
				break
			}
			if !w.farFromBurgs(cd.c, minD) {
				continue
			}
			kind := "village"
			switch {
			case added == 0 && want >= 3:
				kind = "city"
			case w.coastal(cd.c) && r.chance(0.6):
				kind = "port"
			case added < want/2+1:
				kind = "town"
			}
			w.Burgs = append(w.Burgs, Burg{X: w.g.px[cd.c], Y: w.g.py[cd.c], Cell: cd.c, Name: p.TownName(), Race: st.Race,
				State: si, Kind: kind})
			added++
		}
	}
}

// castles – крепость на холмах у границы для крупных государств.
func castles(w *World, peoples []People, r rng) {
	minD := w.spacing(len(w.Burgs)+len(w.States)) * 0.5
	for si, st := range w.States {
		if st.Cells < 40 {
			continue
		}
		best, bv := -1, 0.0
		for c := range w.H {
			if int(w.State[c]) != si || !w.inner(c) || !w.land(c) {
				continue
			}
			t := w.TerrainOf(c)
			if t != Hills && t != Mountain {
				continue
			}
			v := r.float()
			for _, n := range w.g.nb[c] {
				if int(w.State[n]) != si && w.land(int(n)) {
					v += 1 // у границы
				}
			}
			if v > bv && w.farFromBurgs(c, minD) {
				best, bv = c, v
			}
		}
		if best >= 0 {
			name := "Замок " + peopleOf(peoples, st.Race).TownName()
			w.Burgs = append(w.Burgs, Burg{X: w.g.px[best], Y: w.g.py[best], Cell: best, Name: name, Race: st.Race,
				State: si, Kind: "castle"})
		}
	}
}

type stateItem struct {
	cell  int
	pri   float64
	state int
}

type stateQueue []stateItem

func (q stateQueue) Len() int { return len(q) }
func (q stateQueue) Less(i, j int) bool {
	if q[i].pri != q[j].pri {
		return q[i].pri < q[j].pri
	}
	return q[i].cell < q[j].cell
}
func (q stateQueue) Swap(i, j int) { q[i], q[j] = q[j], q[i] }
func (q *stateQueue) Push(x any)   { *q = append(*q, x.(stateItem)) }
func (q *stateQueue) Pop() any {
	old := *q
	it := old[len(old)-1]
	*q = old[:len(old)-1]
	return it
}
