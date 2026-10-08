package atlas

import (
	"image"
	"image/color"
	"math"
	"sort"
)

// Значки рельефа и поселений, нарисованные «от руки»: светлый склон слева, тень справа, тонкий контур тушью.

type icon struct {
	x, y, size float64
	kind       string
	snow       bool
	col        color.RGBA
}

var (
	treeDark  = rgb(52, 92, 50)
	treeLight = rgb(104, 146, 78)
	pineDark  = rgb(40, 76, 52)
	pineLight = rgb(78, 116, 78)
	jungleCol = rgb(46, 104, 56)
	duneCol   = rgb(196, 160, 98)
	reedCol   = rgb(70, 96, 70)
	wallColor = rgb(232, 220, 196)
	roofColor = rgb(160, 70, 52)
)

// relief – расставляет значки по сетке со случайным сдвигом, нижние рисуются поверх верхних.
func (p *painter) relief() {
	var icons []icon
	icons = append(icons, p.scatter(17*p.scale, p.mountainAt)...)
	icons = append(icons, p.scatter(9*p.scale, p.vegetationAt)...)
	sort.SliceStable(icons, func(a, b int) bool { return icons[a].y < icons[b].y })
	for _, ic := range icons {
		p.drawIcon(ic)
	}
}

func (p *painter) scatter(step float64, at func(x, y float64) (icon, bool)) []icon {
	var out []icon
	for y := step; y < float64(p.H)-step/2; y += step * 0.87 {
		row := int(y / step)
		for x := step/2 + float64(row%2)*step/2; x < float64(p.W); x += step {
			jx, jy := x+p.r.between(-0.35, 0.35)*step, y+p.r.between(-0.35, 0.35)*step
			ic, ok := at(jx, jy)
			if !ok {
				continue
			}
			box := image.Rect(int(ic.x-ic.size*0.6), int(ic.y-ic.size), int(ic.x+ic.size*0.6), int(ic.y+1))
			if !p.free(box) || p.blocked(ic.x, ic.y, ic.size*0.45) {
				continue
			}
			out = append(out, ic)
		}
	}
	return out
}

// blocked – у основания значка вода, река или дорога.
func (p *painter) blocked(x, y, r float64) bool {
	for _, d := range [][2]float64{{0, 0}, {-r, 0}, {r, 0}, {0, -r}, {-r * 0.7, -r}, {r * 0.7, -r}} {
		i := p.at(int(x+d[0]), int(y+d[1]))
		if p.kind[i] != pxLand || p.riverMask[i] || p.dWater[i] < 3 {
			return true
		}
	}
	return false
}

func (p *painter) mountainAt(x, y float64) (icon, bool) {
	i := p.at(int(x), int(y))
	h := float64(p.hp[i])
	cold := p.w.Temp[p.cell[i]] < 3
	switch {
	case h >= 64:
		if !p.r.chance(0.85) {
			return icon{}, false
		}
		return icon{x: x, y: y, size: (15 + (h-64)*0.55 + p.r.between(-2, 3)) * p.scale, kind: "mountain", snow: cold || h > 82}, true
	case h >= 50:
		if !p.r.chance(0.55) {
			return icon{}, false
		}
		return icon{x: x, y: y, size: (9 + (h-50)*0.25) * p.scale, kind: "hill", snow: cold && h > 56}, true
	}
	return icon{}, false
}

func (p *painter) vegetationAt(x, y float64) (icon, bool) {
	i := p.at(int(x), int(y))
	if float64(p.hp[i]) >= 50 {
		return icon{}, false
	}
	b := p.w.Biome[p.cell[i]]
	size := (6.5 + p.r.between(0, 2)) * p.scale
	chance := func(v float64) bool { return p.r.chance(v) }
	switch b {
	case bTemperateDeciduousForest, bTemperateRainforest, bTropicalSeasonalForest:
		if chance(0.72) {
			if b == bTemperateRainforest && chance(0.3) {
				return icon{x: x, y: y, size: size * 1.15, kind: "pine"}, true
			}
			return icon{x: x, y: y, size: size, kind: "tree"}, true
		}
	case bTaiga:
		if chance(0.75) {
			return icon{x: x, y: y, size: size * 1.2, kind: "pine", snow: p.w.Temp[p.cell[i]] < -2}, true
		}
	case bTropicalRainforest:
		if chance(0.8) {
			return icon{x: x, y: y, size: size * 1.1, kind: "palm"}, true
		}
	case bSavanna:
		if chance(0.1) {
			return icon{x: x, y: y, size: size, kind: "acacia"}, true
		}
	case bHotDesert, bColdDesert:
		if chance(0.08) {
			return icon{x: x, y: y, size: size * 1.8, kind: "dune"}, true
		}
	case bWetland:
		if chance(0.45) {
			return icon{x: x, y: y, size: size, kind: "reed"}, true
		}
	}
	return icon{}, false
}

func (p *painter) drawIcon(ic icon) {
	switch ic.kind {
	case "mountain":
		p.mountain(ic)
	case "hill":
		p.hill(ic)
	case "tree":
		p.tree(ic.x, ic.y, ic.size)
	case "pine":
		p.pine(ic.x, ic.y, ic.size, ic.snow)
	case "palm":
		p.palm(ic.x, ic.y, ic.size)
	case "acacia":
		p.acacia(ic.x, ic.y, ic.size)
	case "dune":
		p.dune(ic.x, ic.y, ic.size)
	case "reed":
		p.reed(ic.x, ic.y, ic.size)
	}
}

func (p *painter) mountain(ic icon) {
	x, y, s := ic.x, ic.y, ic.size
	half := s * p.r.between(0.62, 0.8)
	px := x + p.r.between(-0.12, 0.12)*s
	top := pt{px, y - s}
	left, right := pt{x - half, y}, pt{x + half, y}
	// неровный правый склон с уступом
	mid := pt{px + (right.x-px)*0.45, y - s*0.55 + p.r.between(-0.08, 0.08)*s}
	base := rgb(178, 160, 132)
	p.c.fill([][]pt{{left, top, mid, right}}, shade(base, 1.12), 1)
	p.c.fill([][]pt{{top, mid, right, {px + s*0.08, y}}}, shade(base, 0.68), 1)
	if ic.snow {
		cap1 := pt{px - (px-left.x)*0.32, y - s*0.68}
		cap2 := pt{px + (mid.x-px)*0.55, top.y + (mid.y-top.y)*0.55}
		p.c.fill([][]pt{{cap1, top, cap2, {px, y - s*0.62}}}, snowColor, 0.95)
	}
	// тень у подножия и контур
	p.c.stroke([]pt{{left.x + s*0.1, y + 0.5}, {right.x, y + 0.5}}, 1.2*p.scale, rgb(110, 92, 70), 0.35)
	p.c.stroke([]pt{left, top, mid, right}, 1.15*p.scale, inkColor, 0.9)
	p.c.stroke([]pt{top, {px + s*0.06, y - s*0.35}}, 0.8*p.scale, inkColor, 0.55)
}

func (p *painter) hill(ic icon) {
	x, y, s := ic.x, ic.y, ic.size
	var arc []pt
	for a := math.Pi; a <= 2*math.Pi+1e-9; a += math.Pi / 12 {
		arc = append(arc, pt{x + math.Cos(a)*s*0.75, y + math.Sin(a)*s*0.55})
	}
	col := rgb(170, 156, 110)
	if ic.snow {
		col = rgb(210, 208, 198)
	}
	p.c.fill([][]pt{arc}, shade(col, 1.08), 0.9)
	shadow := []pt{{x, y}}
	for a := 1.5 * math.Pi; a <= 2*math.Pi+1e-9; a += math.Pi / 12 {
		shadow = append(shadow, pt{x + math.Cos(a)*s*0.75, y + math.Sin(a)*s*0.55})
	}
	p.c.fill([][]pt{shadow}, shade(col, 0.72), 0.85)
	p.c.stroke(arc, 1.0*p.scale, inkColor, 0.8)
}

func (p *painter) tree(x, y, s float64) {
	p.c.stroke([]pt{{x, y}, {x, y - s*0.45}}, 1.2*p.scale, rgb(84, 60, 40), 0.9)
	p.c.fill([][]pt{ellipse(x+s*0.08, y-s*0.75, s*0.52, s*0.48)}, shade(treeDark, 0.85), 0.95)
	p.c.fill([][]pt{ellipse(x-s*0.1, y-s*0.85, s*0.36, s*0.32)}, treeLight, 0.9)
	p.c.stroke(circle(x+s*0.04, y-s*0.77, s*0.5), 0.7*p.scale, inkColor, 0.45)
}

func (p *painter) pine(x, y, s float64, snow bool) {
	p.c.stroke([]pt{{x, y}, {x, y - s*0.3}}, 1.1*p.scale, rgb(84, 60, 40), 0.9)
	for k, lv := range []float64{0.25, 0.55} {
		w := s * (0.42 - float64(k)*0.1)
		top := y - s*(lv+0.55)
		base := y - s*lv
		p.c.fill([][]pt{{{x - w, base}, {x, top}, {x + w, base}}}, pineDark, 0.95)
		p.c.fill([][]pt{{{x - w, base}, {x, top}, {x, base}}}, pineLight, 0.9)
		if snow {
			p.c.fill([][]pt{{{x - w*0.4, top + s*0.2}, {x, top}, {x + w*0.4, top + s*0.2}}}, snowColor, 0.9)
		}
	}
	p.c.stroke([]pt{{x - s*0.42, y - s*0.25}, {x, y - s*1.1}, {x + s*0.42, y - s*0.25}}, 0.7*p.scale, inkColor, 0.5)
}

func (p *painter) palm(x, y, s float64) {
	top := pt{x + s*0.15, y - s}
	p.c.stroke([]pt{{x, y}, {x + s*0.1, y - s*0.5}, top}, 1.2*p.scale, rgb(96, 70, 44), 0.9)
	for _, a := range []float64{-2.6, -2.0, -1.2, -0.5} {
		end := pt{top.x + math.Cos(a)*s*0.6, top.y + math.Sin(a)*s*0.35 + s*0.2}
		p.c.stroke([]pt{top, {(top.x + end.x) / 2, top.y - s*0.12}, end}, 1.6*p.scale, jungleCol, 0.95)
	}
}

func (p *painter) acacia(x, y, s float64) {
	p.c.stroke([]pt{{x, y}, {x, y - s*0.6}}, 1.0*p.scale, rgb(96, 70, 44), 0.85)
	p.c.fill([][]pt{ellipse(x, y-s*0.75, s*0.6, s*0.2)}, rgb(108, 128, 66), 0.9)
}

func (p *painter) dune(x, y, s float64) {
	var arc []pt
	for a := math.Pi * 1.1; a <= math.Pi*1.9; a += math.Pi / 16 {
		arc = append(arc, pt{x + math.Cos(a)*s*0.6, y + math.Sin(a)*s*0.25})
	}
	p.c.stroke(arc, 1.1*p.scale, duneCol, 0.85)
	p.c.stroke([]pt{{x + s*0.1, y - s*0.24}, {x + s*0.45, y - s*0.08}}, 0.8*p.scale, shade(duneCol, 0.7), 0.7)
}

func (p *painter) reed(x, y, s float64) {
	p.c.stroke([]pt{{x - s*0.5, y}, {x + s*0.5, y}}, 0.8*p.scale, reedCol, 0.7)
	for _, dx := range []float64{-0.25, 0, 0.25} {
		p.c.stroke([]pt{{x + dx*s, y}, {x + dx*s*1.4, y - s*0.6}}, 0.7*p.scale, reedCol, 0.8)
	}
}

// burgSize – размер значка поселения.
func (p *painter) burgSize(kind string) float64 {
	k := map[string]float64{"capital": 22, "city": 17, "castle": 15, "town": 13, "port": 13, "village": 10}[kind]
	return k * p.scale
}

func (p *painter) burgBox(b Burg) image.Rectangle {
	s := p.burgSize(b.Kind)
	return image.Rect(int(b.X-s*0.75), int(b.Y-s), int(b.X+s*0.75), int(b.Y+s*0.3))
}

func (p *painter) burgObstacles() {
	for _, b := range p.w.Burgs {
		p.obstacles = append(p.obstacles, p.burgBox(b).Inset(-2))
	}
}

func (p *painter) burgs() {
	for _, b := range p.w.Burgs {
		s := p.burgSize(b.Kind)
		col := p.w.States[b.State].Color
		switch b.Kind {
		case "capital":
			p.castle(b.X, b.Y, s, col, true)
		case "castle":
			p.castle(b.X, b.Y, s, col, false)
		case "city":
			p.house(b.X-s*0.3, b.Y, s*0.6)
			p.house(b.X+s*0.25, b.Y-s*0.05, s*0.7)
			p.tower(b.X, b.Y-s*0.1, s*0.75, col)
		case "town", "port":
			p.house(b.X-s*0.22, b.Y, s*0.7)
			p.house(b.X+s*0.25, b.Y+s*0.05, s*0.55)
		default:
			p.house(b.X, b.Y, s*0.75)
		}
		if b.Kind == "port" || b.Kind == "capital" && p.w.coastal(b.Cell) {
			p.anchor(b.X+s*0.85, b.Y-s*0.15, s*0.42)
		}
	}
}

func (p *painter) house(x, y, s float64) {
	w := s * 0.5
	body := []pt{{x - w, y}, {x - w, y - s*0.5}, {x + w, y - s*0.5}, {x + w, y}}
	roof := []pt{{x - w*1.2, y - s*0.48}, {x - w*0.2, y - s*1.0}, {x + w*0.2, y - s*1.0}, {x + w*1.2, y - s*0.48}}
	p.c.fill([][]pt{body}, wallColor, 1)
	p.c.fill([][]pt{roof}, roofColor, 1)
	p.c.stroke(append(body, body[0]), 0.8*p.scale, inkColor, 0.9)
	p.c.stroke(append(roof, roof[0]), 0.8*p.scale, inkColor, 0.9)
}

func (p *painter) tower(x, y, s float64, flag color.RGBA) {
	w := s * 0.22
	body := []pt{{x - w, y}, {x - w, y - s*0.8}, {x + w, y - s*0.8}, {x + w, y}}
	p.c.fill([][]pt{body}, wallColor, 1)
	p.c.fill([][]pt{{{x - w*1.3, y - s*0.78}, {x, y - s*1.25}, {x + w*1.3, y - s*0.78}}}, shade(flag, 0.9), 1)
	p.c.stroke(append(body, body[0]), 0.8*p.scale, inkColor, 0.9)
	p.c.stroke([]pt{{x - w*1.3, y - s*0.78}, {x, y - s*1.25}, {x + w*1.3, y - s*0.78}}, 0.8*p.scale, inkColor, 0.9)
}

// castle – стена с зубцами и две башни; у столицы ещё флаг цвета государства.
func (p *painter) castle(x, y, s float64, col color.RGBA, capital bool) {
	w := s * 0.55
	wall := []pt{{x - w, y}, {x - w, y - s*0.45}}
	for k := 0; k < 5; k++ {
		x0 := x - w + float64(k)*w*0.4
		wall = append(wall, pt{x0, y - s*0.52}, pt{x0 + w*0.2, y - s*0.52}, pt{x0 + w*0.2, y - s*0.45}, pt{x0 + w*0.4, y - s*0.45})
	}
	wall = append(wall, pt{x + w, y})
	p.c.fill([][]pt{wall}, wallColor, 1)
	p.c.stroke(append(wall, wall[0]), 0.9*p.scale, inkColor, 0.95)
	p.c.fill([][]pt{{{x - s*0.1, y}, {x - s*0.1, y - s*0.2}, {x + s*0.1, y - s*0.2}, {x + s*0.1, y}}}, inkColor, 0.8)
	p.tower(x-w, y, s*0.85, col)
	p.tower(x+w, y, s*0.85, col)
	if capital {
		p.tower(x, y-s*0.3, s, col)
		pole := pt{x, y - s*1.55}
		p.c.stroke([]pt{{x, y - s*1.28}, pole}, 0.9*p.scale, inkColor, 1)
		p.c.fill([][]pt{{pole, {pole.x + s*0.42, pole.y + s*0.1}, {pole.x, pole.y + s*0.22}}}, col, 1)
	}
}

func (p *painter) anchor(x, y, s float64) {
	ink := rgb(40, 52, 70)
	p.c.stroke([]pt{{x, y - s}, {x, y + s*0.6}}, 1.1*p.scale, ink, 0.9)
	p.c.stroke([]pt{{x - s*0.4, y - s*0.55}, {x + s*0.4, y - s*0.55}}, 1.0*p.scale, ink, 0.9)
	var arc []pt
	for a := 0.15 * math.Pi; a <= 0.85*math.Pi; a += math.Pi / 12 {
		arc = append(arc, pt{x + math.Cos(a)*s*0.55, y + math.Sin(a)*s*0.55})
	}
	p.c.stroke(arc, 1.1*p.scale, ink, 0.9)
	p.c.stroke(circle(x, y-s*1.15, s*0.18), 0.9*p.scale, ink, 0.9)
}

// ships – парусники в открытом море: на морских путях и вдали от берегов.
func (p *painter) ships() {
	var spots []pt
	for _, path := range p.w.Sea {
		if len(path) > 4 {
			spots = append(spots, p.cellPt(path[len(path)/2]))
		}
	}
	for range 60 {
		x, y := p.r.between(0.08, 0.92)*float64(p.W), p.r.between(0.08, 0.92)*float64(p.H)
		if i := p.at(int(x), int(y)); p.kind[i] == pxOcean && p.dLand[i] > 35*float32(p.scale) {
			spots = append(spots, pt{x, y})
		}
	}
	s := 13 * p.scale
	placed := 0
	for _, q := range spots {
		if placed >= 5 {
			break
		}
		box := image.Rect(int(q.x-s), int(q.y-s*1.4), int(q.x+s), int(q.y+s*0.4))
		if !p.free(box) || p.kind[p.at(int(q.x), int(q.y))] != pxOcean {
			continue
		}
		p.obstacles = append(p.obstacles, box.Inset(-int(s)))
		p.ship(q.x, q.y, s, p.r.chance(0.5))
		placed++
	}
}

func (p *painter) ship(x, y, s float64, left bool) {
	dir := 1.0
	if left {
		dir = -1
	}
	hull := []pt{{x - s*0.6, y - s*0.25}, {x + s*0.6, y - s*0.25}, {x + s*0.4*dir, y}, {x - s*0.4*dir, y}}
	p.c.fill([][]pt{hull}, rgb(110, 74, 44), 1)
	p.c.stroke(append(hull, hull[0]), 0.8*p.scale, inkColor, 0.9)
	p.c.stroke([]pt{{x, y - s*0.25}, {x, y - s*1.3}}, 0.9*p.scale, inkColor, 1)
	sail := []pt{{x - s*0.38, y - s*0.4}, {x - s*0.32, y - s*1.15}, {x + s*0.32, y - s*1.15}, {x + s*0.38, y - s*0.4}}
	p.c.fill([][]pt{sail}, rgb(246, 238, 220), 1)
	p.c.stroke(append(sail, sail[0]), 0.7*p.scale, inkColor, 0.8)
	p.c.fill([][]pt{{{x, y - s*1.3}, {x + s*0.3*dir, y - s*1.22}, {x, y - s*1.14}}}, rgb(168, 58, 48), 1)
	for k := range 3 {
		wx := x - s*0.7 + float64(k)*s*0.55
		p.c.stroke([]pt{{wx, y + s*0.2}, {wx + s*0.15, y + s*0.12}, {wx + s*0.3, y + s*0.2}}, 0.7*p.scale, rgb(230, 240, 240), 0.7)
	}
}
