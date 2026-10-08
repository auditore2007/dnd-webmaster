package atlas

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Рельеф по шаблонам Azgaar's Fantasy Map Generator (MIT, см. LICENSE-AZGAAR): шаблон – сценарий из шагов
// «инструмент количество высота диапазонX диапазонY». Высоты 0–100, суша – от 20.

const seaLevel = 20

type heightmap struct {
	g          *grid
	h          []uint8
	r          rng
	blob, line float64
}

func lim(v float64) uint8 { return uint8(clamp(v, 0, 100)) }

// point – случайная координата из диапазона в процентах («15-85») от длины стороны.
func (m *heightmap) point(rangeStr string, length int) float64 {
	a, b, _ := strings.Cut(rangeStr, "-")
	lo, _ := strconv.ParseFloat(a, 64)
	hi, err := strconv.ParseFloat(b, 64)
	if err != nil {
		hi = lo
	}
	return m.r.between(lo/100*float64(length), hi/100*float64(length))
}

func (m *heightmap) addHill(count, height, rx, ry string) {
	for range int(m.r.number(count)) {
		change := make([]uint8, len(m.h))
		hgt := lim(m.r.number(height))
		start := 0
		for limit := 0; limit < 50; limit++ {
			start = m.g.find(m.point(rx, m.g.w), m.point(ry, m.g.h))
			if int(m.h[start])+int(hgt) <= 90 {
				break
			}
		}
		change[start] = hgt
		queue := []int32{int32(start)}
		for qi := 0; qi < len(queue); qi++ {
			q := queue[qi]
			for _, c := range m.g.nb[q] {
				if change[c] != 0 {
					continue
				}
				change[c] = uint8(math.Pow(float64(change[q]), m.blob) * (m.r.float()*0.2 + 0.9))
				if change[c] > 1 {
					queue = append(queue, c)
				}
			}
		}
		for i := range m.h {
			m.h[i] = lim(float64(m.h[i]) + float64(change[i]))
		}
	}
}

func (m *heightmap) addPit(count, height, rx, ry string) {
	for range int(m.r.number(count)) {
		used := make([]bool, len(m.h))
		h := float64(lim(m.r.number(height)))
		start := 0
		for limit := 0; limit < 50; limit++ {
			start = m.g.find(m.point(rx, m.g.w), m.point(ry, m.g.h))
			if m.h[start] >= seaLevel {
				break
			}
		}
		queue := []int32{int32(start)}
		for qi := 0; qi < len(queue); qi++ {
			h = math.Pow(h, m.blob) * (m.r.float()*0.2 + 0.9)
			if h < 1 {
				break
			}
			for _, c := range m.g.nb[queue[qi]] {
				if used[c] {
					continue
				}
				m.h[c] = lim(float64(m.h[c]) - h*(m.r.float()*0.2+0.9))
				used[c] = true
				queue = append(queue, c)
			}
		}
	}
}

// walk – путь клеток от start к end: каждый шаг к соседу, ближайшему к цели (иногда наугад – хребты извилистые).
func (m *heightmap) walk(start, end int, used []bool, randomness float64) []int {
	path := []int{start}
	used[start] = true
	cur := start
	for cur != end {
		best, bd := -1, math.MaxFloat64
		for _, e := range m.g.nb[cur] {
			if used[e] {
				continue
			}
			d := sq(m.g.px[end]-m.g.px[e]) + sq(m.g.py[end]-m.g.py[e])
			if m.r.float() > 1-randomness {
				d /= 2
			}
			if d < bd {
				best, bd = int(e), d
			}
		}
		if best < 0 {
			return path
		}
		cur = best
		path = append(path, cur)
		used[cur] = true
	}
	return path
}

// addLine – хребет (sign > 0) или впадина (sign < 0): путь между двумя точками, расходящийся волной в стороны.
func (m *heightmap) addLine(count, height, rx, ry string, sign float64) {
	trough := sign < 0
	for range int(m.r.number(count)) {
		used := make([]bool, len(m.h))
		h := float64(lim(m.r.number(height)))
		var sx, sy float64
		start := 0
		for limit := 0; limit < 50; limit++ {
			sx, sy = m.point(rx, m.g.w), m.point(ry, m.g.h)
			start = m.g.find(sx, sy)
			if !trough || m.h[start] >= seaLevel {
				break
			}
		}
		maxDist := float64(m.g.w) / 3
		if trough {
			maxDist = float64(m.g.w) / 2
		}
		var ex, ey float64
		for limit := 0; limit < 50; limit++ {
			ex = m.r.float()*float64(m.g.w)*0.8 + float64(m.g.w)*0.1
			ey = m.r.float()*float64(m.g.h)*0.7 + float64(m.g.h)*0.15
			d := math.Abs(ey-sy) + math.Abs(ex-sx)
			if d >= float64(m.g.w)/8 && d <= maxDist {
				break
			}
		}
		randomness := 0.15
		if trough {
			randomness = 0.2
		}
		path := m.walk(start, m.g.find(ex, ey), used, randomness)
		queue := append([]int(nil), path...)
		waves := 0
		for len(queue) > 0 {
			frontier := queue
			queue = nil
			waves++
			for _, c := range frontier {
				m.h[c] = lim(float64(m.h[c]) + sign*h*(m.r.float()*0.3+0.85))
			}
			h = math.Pow(h, m.line) - 1
			if h < 2 {
				break
			}
			for _, f := range frontier {
				for _, c := range m.g.nb[f] {
					if !used[c] {
						queue = append(queue, int(c))
						used[c] = true
					}
				}
			}
		}
		// от каждой шестой клетки хребта вниз по склону – отроги
		for d, cur := range path {
			if d%6 != 0 {
				continue
			}
			for range waves {
				low := int(m.g.nb[cur][0])
				for _, c := range m.g.nb[cur] {
					if m.h[c] < m.h[low] {
						low = int(c)
					}
				}
				m.h[low] = uint8((float64(m.h[cur])*2 + float64(m.h[low])) / 3)
				cur = low
			}
		}
	}
}

func (m *heightmap) addStrait(width, direction string) {
	desired := math.Min(m.r.number(width), float64(m.g.nx)/3)
	if desired < 1 && m.r.chance(desired) {
		return
	}
	w, h := float64(m.g.w), float64(m.g.h)
	vert := direction == "vertical"
	sx, sy := 5.0, math.Floor(m.r.float()*h*0.4+h*0.3)
	ex, ey := w-5, math.Floor(h-sy-h*0.1+m.r.float()*h*0.2)
	if vert {
		sx, sy = math.Floor(m.r.float()*w*0.4+w*0.3), 5
		ex, ey = math.Floor(w-sx-w*0.1+m.r.float()*w*0.2), h-5
	}
	cur, end := m.g.find(sx, sy), m.g.find(ex, ey)
	var path []int
	for steps := 0; cur != end && steps < m.g.size()*2; steps++ {
		best, bd := cur, math.MaxFloat64
		for _, e := range m.g.nb[cur] {
			d := sq(m.g.px[end]-m.g.px[e]) + sq(m.g.py[end]-m.g.py[e])
			if m.r.float() > 0.8 {
				d /= 2
			}
			if d < bd {
				best, bd = int(e), d
			}
		}
		cur = best
		path = append(path, cur)
	}
	used := make([]bool, len(m.h))
	var query []int
	step := 0.1 / desired
	for i := 0.0; i < desired; i++ {
		exp := 0.9 - step*(desired-i)
		for _, r := range path {
			for _, e := range m.g.nb[r] {
				if used[e] {
					continue
				}
				used[e] = true
				query = append(query, int(e))
				m.h[e] = uint8(math.Pow(float64(m.h[e]), exp))
			}
		}
		path = append([]int(nil), query...)
	}
}

// modify – прибавить или умножить высоты в диапазоне («land» – только суша, «all» – всё, «30-100»).
func (m *heightmap) modify(rangeStr string, add, mult float64) {
	lo, hi := 0.0, 100.0
	switch rangeStr {
	case "land":
		lo = seaLevel
	case "all":
	default:
		a, b, _ := strings.Cut(rangeStr, "-")
		lo, _ = strconv.ParseFloat(a, 64)
		hi, _ = strconv.ParseFloat(b, 64)
	}
	land := lo == seaLevel
	for i, v := range m.h {
		h := float64(v)
		if h < lo || h > hi {
			continue
		}
		if add != 0 {
			h += add
			if land {
				h = math.Max(h, seaLevel)
			}
		}
		if mult != 1 {
			if land {
				h = (h-seaLevel)*mult + seaLevel
			} else {
				h *= mult
			}
		}
		m.h[i] = lim(h)
	}
}

func (m *heightmap) smooth(fr float64) {
	out := make([]uint8, len(m.h))
	for i, v := range m.h {
		sum := float64(v)
		for _, c := range m.g.nb[i] {
			sum += float64(m.h[c])
		}
		mean := sum / float64(len(m.g.nb[i])+1)
		if fr == 1 {
			out[i] = lim(mean)
		} else {
			out[i] = lim((float64(v)*(fr-1) + mean) / fr)
		}
	}
	m.h = out
}

// mask – прижимает сушу к центру карты (power < 0 – наоборот, к краям).
func (m *heightmap) mask(power float64) {
	fr := math.Abs(power)
	if power == 0 {
		fr = 1
	}
	for i, v := range m.h {
		nx := 2*m.g.px[i]/float64(m.g.w) - 1
		ny := 2*m.g.py[i]/float64(m.g.h) - 1
		d := (1 - nx*nx) * (1 - ny*ny)
		if power < 0 {
			d = 1 - d
		}
		h := float64(v)
		m.h[i] = lim((h*(fr-1) + h*d) / fr)
	}
}

func (m *heightmap) invert(p float64, axes string) {
	if !m.r.chance(p) {
		return
	}
	out := make([]uint8, len(m.h))
	for i := range m.h {
		x, y := i%m.g.nx, i/m.g.nx
		if axes != "y" {
			x = m.g.nx - x - 1
		}
		if axes != "x" {
			y = m.g.ny - y - 1
		}
		out[i] = m.h[y*m.g.nx+x]
	}
	m.h = out
}

func (m *heightmap) step(line string) error {
	f := strings.Fields(line)
	if len(f) < 2 {
		return fmt.Errorf("шаг шаблона %q", line)
	}
	for len(f) < 5 {
		f = append(f, "0")
	}
	num := func(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }
	switch f[0] {
	case "Hill":
		m.addHill(f[1], f[2], f[3], f[4])
	case "Pit":
		m.addPit(f[1], f[2], f[3], f[4])
	case "Range":
		m.addLine(f[1], f[2], f[3], f[4], 1)
	case "Trough":
		m.addLine(f[1], f[2], f[3], f[4], -1)
	case "Strait":
		m.addStrait(f[1], f[2])
	case "Mask":
		m.mask(num(f[1]))
	case "Invert":
		m.invert(num(f[1]), f[2])
	case "Add":
		m.modify(f[2], num(f[1]), 1)
	case "Multiply":
		m.modify(f[2], 0, num(f[1]))
	case "Smooth":
		m.smooth(num(f[1]))
	default:
		return fmt.Errorf("неизвестный инструмент %q", f[0])
	}
	return nil
}

func buildHeights(g *grid, r rng, template string) ([]uint8, error) {
	t, ok := templates[template]
	if !ok {
		return nil, fmt.Errorf("нет шаблона %q", template)
	}
	m := &heightmap{g: g, h: make([]uint8, g.size()), r: r, blob: blobPower(g.size()), line: linePower(g.size())}
	for _, line := range strings.Split(t.steps, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if err := m.step(line); err != nil {
			return nil, err
		}
	}
	return m.h, nil
}
