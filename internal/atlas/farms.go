package atlas

import (
	"image/color"
	"math"
)

// Обжитые земли: вокруг поселений – лоскуты полей с межами и бороздами, хутора и мастерские рядом с городами,
// фермы вдоль дорог. Всё это – декор: местами на карте приложения они не становятся.

var fieldColors = []color.RGBA{rgb(214, 190, 108), rgb(196, 178, 96), rgb(150, 168, 84), rgb(128, 152, 76), rgb(166, 130, 86), rgb(184, 166, 112)}

// fieldsAround – сколько полей у поселения и на каком расстоянии.
var fieldsAround = map[string]int{"capital": 16, "city": 12, "town": 8, "port": 6, "village": 6}

// hamletsAround – сколько домиков-хуторов рядом с поселением и какие постройки там бывают.
var hamletsAround = map[string]struct {
	n     int
	kinds []string
}{
	"capital": {8, []string{"house", "tavern", "smithy", "market", "temple", "house", "stable"}},
	"city":    {6, []string{"house", "tavern", "smithy", "house", "stable"}},
	"town":    {4, []string{"house", "house", "smithy", "stable", "farm"}},
	"port":    {3, []string{"house", "house", "lighthouse"}},
	"village": {2, []string{"house", "farm", "stable"}},
}

func (p *painter) farmland() {
	for _, b := range p.w.Burgs {
		n := fieldsAround[b.Kind]
		if n == 0 || p.w.Temp[p.cell[p.at(int(b.X), int(b.Y))]] < -2 {
			continue
		}
		angle := p.r.between(0, math.Pi)
		for k := 0; k < n*3 && n > 0; k++ {
			d := p.r.between(12, 26+float64(n)*2.2) * p.scale
			a := p.r.between(0, 2*math.Pi)
			cx, cy := b.X+math.Cos(a)*d, b.Y+math.Sin(a)*d*0.8
			fw, fh := p.r.between(12, 22)*p.scale, p.r.between(8, 14)*p.scale
			if p.field(cx, cy, fw, fh, angle+p.r.between(-0.25, 0.25)) {
				n--
			}
		}
	}
}

// field – прямоугольное поле, повёрнутое на angle; не ложится на воду, горы, реки и дороги.
func (p *painter) field(cx, cy, fw, fh, angle float64) bool {
	ca, sa := math.Cos(angle), math.Sin(angle)
	corner := func(u, v float64) pt { return pt{cx + u*ca - v*sa, cy + u*sa + v*ca} }
	poly := []pt{corner(-fw/2, -fh/2), corner(fw/2, -fh/2), corner(fw/2, fh/2), corner(-fw/2, fh/2)}
	for _, q := range append(poly, pt{cx, cy}) {
		i := p.at(int(q.x), int(q.y))
		if q.x < 0 || q.y < 0 || q.x >= float64(p.W) || q.y >= float64(p.H) ||
			p.kind[i] != pxLand || p.riverMask[i] || p.hp[i] > 50 || p.dWater[i] < 3 {
			return false
		}
	}
	col := fieldColors[p.r.r.Intn(len(fieldColors))]
	p.c.fill([][]pt{poly}, col, 0.85)
	for k := 1; k < 4; k++ { // борозды вдоль длинной стороны
		v := -fh/2 + fh*float64(k)/4
		p.c.stroke([]pt{corner(-fw/2+1, v), corner(fw/2-1, v)}, 0.6*p.scale, shade(col, 0.78), 0.35)
	}
	p.c.stroke(closed(poly), 0.9*p.scale, rgb(86, 98, 56), 0.45) // межа
	p.markLine(poly, 2)
	return true
}

// hamlets – домики вокруг поселений и фермы вдоль дорог.
func (p *painter) hamlets() {
	for _, b := range p.w.Burgs {
		h, ok := hamletsAround[b.Kind]
		if !ok {
			continue
		}
		placed := 0
		for k := 0; k < h.n*6 && placed < h.n; k++ {
			d := p.r.between(16, 30+float64(h.n)*3) * p.scale
			a := p.r.between(0, 2*math.Pi)
			name := "buildings/" + h.kinds[p.r.r.Intn(len(h.kinds))]
			if p.addExtra(name, b.X+math.Cos(a)*d, b.Y+math.Sin(a)*d*0.75, p.r.between(14, 17)*p.scale) {
				placed++
			}
		}
	}
	for _, path := range p.w.Roads {
		line := make([]pt, len(path))
		for k, c := range path {
			line[k] = p.cellPt(c)
		}
		step := 70 * p.scale
		for _, d := range dashes(chaikin(line, 2), 1, step) {
			q, prev := d[0], d[len(d)-1]
			if !p.r.chance(0.45) {
				continue
			}
			dx, dy := prev.x-q.x, prev.y-q.y
			l := math.Max(1e-6, math.Hypot(dx, dy))
			side := 12 * p.scale
			if p.r.chance(0.5) {
				side = -side
			}
			kind := []string{"farm", "house", "stable", "farm"}[p.r.r.Intn(4)]
			p.addExtra("buildings/"+kind, q.x-dy/l*side, q.y+dx/l*side, p.r.between(13, 16)*p.scale)
		}
	}
}

// addExtra ставит декоративную постройку, если место на суше и свободно.
func (p *painter) addExtra(name string, x, y, h float64) bool {
	box := spriteBox(name, x, y, h)
	i := p.at(int(x), int(y))
	if p.kind[i] != pxLand || p.dWater[i] < 4 || p.hp[i] > 60 || p.riverMaskAt(x, y) || !p.free(box) {
		return false
	}
	p.obstacles = append(p.obstacles, box.Inset(-1))
	p.extras = append(p.extras, extra{name, x, y, h})
	return true
}

// riverMaskAt – река или дорога у основания постройки (поля не мешают).
func (p *painter) riverMaskAt(x, y float64) bool { return p.roadOrRiver[p.at(int(x), int(y))] }
