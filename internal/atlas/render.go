package atlas

import (
	"image"
	"image/color"
	"math"
	"slices"
)

// Отрисовка в стиле иллюстрированного атласа: мягкая заливка биомов с отмывкой рельефа, береговые «волны»,
// тонированные государства с каймой у границ, реки, дороги, значки гор и лесов, города, надписи, картуш,
// роза ветров, масштаб и рамка, поверх всего – фактура бумаги.

const (
	pxLand  uint8 = 0
	pxOcean uint8 = 1
	pxLake  uint8 = 2
)

var biomePaint = [13]color.RGBA{
	rgb(46, 88, 120),   // море
	rgb(228, 204, 146), // жаркая пустыня
	rgb(200, 190, 150), // холодная пустыня
	rgb(206, 194, 126), // саванна
	rgb(178, 186, 114), // луга
	rgb(146, 166, 90),  // сезонный тропический лес
	rgb(118, 150, 86),  // лиственный лес
	rgb(86, 132, 74),   // джунгли
	rgb(98, 136, 88),   // влажный лес умеренного пояса
	rgb(104, 124, 94),  // тайга
	rgb(182, 176, 146), // тундра
	rgb(236, 238, 236), // ледник
	rgb(112, 134, 100), // болото
}

var (
	inkColor    = rgb(58, 42, 30)
	paperColor  = rgb(240, 228, 198)
	rockColor   = rgb(156, 136, 110)
	snowColor   = rgb(244, 244, 240)
	shallowSea  = rgb(118, 166, 172)
	deepSea     = rgb(52, 96, 128)
	lakeColor   = rgb(110, 160, 176)
	riverColor  = rgb(64, 110, 146)
	roadColor   = rgb(112, 76, 44)
	seaWayColor = rgb(236, 230, 210)
)

// painter – поля карты попиксельно и всё, что нужно для слоёв.
type painter struct {
	w          *World
	c          canvas
	W, H       int
	hp         []float32 // высота пикселя
	kind       []uint8   // суша, море, озеро
	cell       []int32   // клетка пикселя (с искажением границ шумом)
	dLand      []float32 // расстояние от воды до суши
	dWater     []float32 // от суши до воды
	riverMask  []bool    // где реки и дороги – туда не ставим значки
	obstacles  []image.Rectangle
	labels     []label
	r          rng
	warpX      *noise
	warpY      *noise
	detail     *noise
	scale      float64 // множитель размеров значков и шрифтов от размера карты
	cartouche  image.Rectangle
	compassBox image.Rectangle
}

// Render рисует мир. Поселения, оказавшиеся из-за сглаживания берегов в воде, сдвигаются на ближайшую сушу.
func (w *World) Render(r Rand) *image.RGBA {
	p := &painter{w: w, W: w.Width, H: w.Height, r: rng{r}}
	p.c = canvas{image.NewRGBA(image.Rect(0, 0, p.W, p.H))}
	p.scale = clamp(math.Sqrt(float64(p.W*p.H))/1265, 0.7, 1.5)
	p.warpX, p.warpY, p.detail = newNoise(p.r), newNoise(p.r), newNoise(p.r)
	p.fields()
	p.moveBurgsAshore()
	p.base()
	p.states()
	p.layoutDecor()
	p.rivers()
	p.roads()
	p.burgObstacles()
	p.labelsLayout()
	p.relief()
	p.burgs()
	p.ships()
	p.drawLabels()
	p.decor()
	p.paper()
	return p.c.img
}

// lattice – координаты пикселя на решётке клеток, со смещением шумом для живых берегов.
func (p *painter) lattice(x, y float64) (float64, float64) {
	s := p.w.g.s
	wx := (p.warpX.fbm(x/110, y/110, 3) - 0.5) * s * 1.6
	wy := (p.warpY.fbm(x/110, y/110, 3) - 0.5) * s * 1.6
	return (x+wx)/s - 0.5, (y+wy)/s - 0.5
}

// bilinear – значение поля клеток в точке решётки (u, v).
func (p *painter) bilinear(u, v float64, val func(c int) float64) float64 {
	g := p.w.g
	i0, j0 := int(math.Floor(u)), int(math.Floor(v))
	tx, ty := u-float64(i0), v-float64(j0)
	at := func(i, j int) float64 {
		i = min(g.nx-1, max(0, i))
		j = min(g.ny-1, max(0, j))
		return val(j*g.nx + i)
	}
	a, b := at(i0, j0), at(i0+1, j0)
	c, d := at(i0, j0+1), at(i0+1, j0+1)
	return (a*(1-tx)+b*tx)*(1-ty) + (c*(1-tx)+d*tx)*ty
}

// gauss – гладкое значение поля по 4×4 клеткам вокруг (для озёр: у бинарного поля билинейная интерполяция даёт квадраты).
func (p *painter) gauss(u, v float64, val func(c int) float64) float64 {
	g := p.w.g
	i0, j0 := int(math.Floor(u)), int(math.Floor(v))
	var sum, wsum float64
	for j := j0 - 1; j <= j0+2; j++ {
		for i := i0 - 1; i <= i0+2; i++ {
			wt := math.Exp(-(sq(float64(i)-u) + sq(float64(j)-v)) / 0.9)
			sum += wt * val(min(g.ny-1, max(0, j))*g.nx+min(g.nx-1, max(0, i)))
			wsum += wt
		}
	}
	return sum / wsum
}

func (p *painter) fields() {
	w, g := p.w, p.w.g
	n := p.W * p.H
	p.hp, p.kind, p.cell = make([]float32, n), make([]uint8, n), make([]int32, n)
	height := func(c int) float64 { return float64(w.H[c]) }
	lake := func(c int) float64 {
		if w.Lake[c] {
			return 1
		}
		return 0
	}
	for y := range p.H {
		for x := range p.W {
			i := y*p.W + x
			fx, fy := float64(x), float64(y)
			u, v := p.lattice(fx, fy)
			h := p.bilinear(u, v, height) + (p.detail.fbm(fx/16, fy/16, 4)-0.5)*8
			p.hp[i] = float32(h)
			c := g.nearest((u+0.5)*g.s, (v+0.5)*g.s)
			p.cell[i] = int32(c)
			switch {
			case p.bilinear(u, v, lake) > 0 && p.gauss(u, v, lake)+(p.detail.fbm(fx/10+50, fy/10, 3)-0.5)*0.4 > 0.5:
				p.kind[i] = pxLake
			case h < seaLevel && w.ocean[c]:
				p.kind[i] = pxOcean
			case h < seaLevel:
				p.kind[i] = pxLake
			}
		}
	}
	land, water := make([]bool, n), make([]bool, n)
	for i, k := range p.kind {
		land[i], water[i] = k == pxLand, k != pxLand
	}
	p.dLand, p.dWater = distance(land, p.W, p.H), distance(water, p.W, p.H)
	p.riverMask = make([]bool, n)
}

func (p *painter) at(x, y int) int { return min(p.H-1, max(0, y))*p.W + min(p.W-1, max(0, x)) }

func (p *painter) moveBurgsAshore() {
	for i := range p.w.Burgs {
		b := &p.w.Burgs[i]
		x, y := int(b.X), int(b.Y)
		if p.kind[p.at(x, y)] == pxLand && p.dWater[p.at(x, y)] > 3 {
			continue
		}
		best, bd := -1, math.MaxFloat64
		r := int(p.w.g.s * 2.5)
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				j := p.at(x+dx, y+dy)
				if p.kind[j] != pxLand || p.dWater[j] < 4 {
					continue
				}
				if d := float64(dx*dx + dy*dy); d < bd {
					best, bd = j, d
				}
			}
		}
		if best >= 0 {
			b.X, b.Y = float64(best%p.W), float64(best/p.W)
		}
	}
}

// base – суша (биом, высота, отмывка) и вода (глубина, береговые волны, контур берега).
func (p *painter) base() {
	img := p.c.img
	for y := range p.H {
		for x := range p.W {
			i := y*p.W + x
			var col color.RGBA
			switch p.kind[i] {
			case pxLand:
				col = p.landColor(x, y, i)
			case pxOcean:
				col = p.seaColor(i)
			default:
				col = mix(lakeColor, shallowSea, 0.2)
				col = shade(col, 1-0.06*clamp(float64(p.dLand[i])/8, 0, 1))
			}
			if p.kind[i] != pxLand {
				ink := clamp(2.2-float64(p.dLand[i]), 0, 1) * 0.85
				col = mix(col, inkColor, ink)
			}
			o := img.PixOffset(x, y)
			img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = col.R, col.G, col.B, 255
		}
	}
}

func (p *painter) landColor(x, y, i int) color.RGBA {
	w := p.w
	u, v := p.lattice(float64(x), float64(y))
	var r, g, b, sum float64
	i0, j0 := int(math.Floor(u)), int(math.Floor(v))
	tx, ty := u-float64(i0), v-float64(j0)
	for _, k := range [4][3]float64{{0, 0, (1 - tx) * (1 - ty)}, {1, 0, tx * (1 - ty)}, {0, 1, (1 - tx) * ty}, {1, 1, tx * ty}} {
		ci := min(w.g.nx-1, max(0, i0+int(k[0])))
		cj := min(w.g.ny-1, max(0, j0+int(k[1])))
		c := cj*w.g.nx + ci
		if !w.land(c) || k[2] <= 0 {
			continue
		}
		bc := biomePaint[w.Biome[c]]
		r, g, b, sum = r+float64(bc.R)*k[2], g+float64(bc.G)*k[2], b+float64(bc.B)*k[2], sum+k[2]
	}
	col := biomePaint[bGrassland]
	if sum > 0 {
		col = color.RGBA{uint8(r / sum), uint8(g / sum), uint8(b / sum), 255}
	}
	h := float64(p.hp[i])
	cold := float64(w.Temp[p.cell[i]]) < 4
	col = mix(col, rockColor, (h-58)/30)
	if cold || h > 88 {
		col = mix(col, snowColor, (h-72)/14)
	}
	// у самой воды – светлая полоска пляжа
	col = mix(col, rgb(226, 212, 168), 0.45*(1-float64(p.dWater[i])/4))
	gx := float64(p.hp[p.at(x+1, y)] - p.hp[p.at(x-1, y)])
	gy := float64(p.hp[p.at(x, y+1)] - p.hp[p.at(x, y-1)])
	relief := 0.022 + 0.02*clamp((h-40)/40, 0, 1)
	return shade(col, clamp(1+(gx+gy)*relief, 0.62, 1.3))
}

func (p *painter) seaColor(i int) color.RGBA {
	h := float64(p.hp[i])
	d := float64(p.dLand[i])
	col := mix(shallowSea, deepSea, clamp((seaLevel-h)/16, 0, 1)*0.7+clamp(d/140, 0, 1)*0.3)
	// береговые волны – тонкие линии, повторяющие берег
	for k, ring := range []float64{5, 10, 16, 23, 31} {
		on := clamp(1-math.Abs(d-ring)/0.85, 0, 1)
		col = mix(col, rgb(30, 64, 92), on*0.32*(1-float64(k)*0.17))
	}
	return mix(col, rgb(170, 204, 200), 0.35*clamp(1-d/7, 0, 1))
}

// states – тон государства на суше, гуще у границы; граница – пунктир.
func (p *painter) states() {
	w := p.w
	n := p.W * p.H
	state := func(i int) int16 {
		if p.kind[i] != pxLand {
			return -2
		}
		return w.State[p.cell[i]]
	}
	edge := make([]bool, n)
	for y := range p.H {
		for x := range p.W {
			i := y*p.W + x
			s := state(i)
			if s == -2 {
				continue
			}
			if x+1 < p.W && state(i+1) != s && state(i+1) != -2 || y+1 < p.H && state(i+p.W) != s && state(i+p.W) != -2 {
				edge[i] = true
			}
		}
	}
	dB := distance(edge, p.W, p.H)
	for i := range n {
		s := state(i)
		if s < 0 {
			continue
		}
		x, y := i%p.W, i/p.W
		col := w.States[s].Color
		d := float64(dB[i])
		a := 0.1 + 0.32*clamp(1-d/(9*p.scale), 0, 1)
		p.c.blend(x, y, col, a)
		if d < 1.1 && ((x+y)/4)%2 == 0 {
			p.c.blend(x, y, shade(col, 0.45), 0.85)
		}
	}
}

func (p *painter) free(r image.Rectangle) bool {
	if !r.In(image.Rect(int(24*p.scale), int(24*p.scale), p.W-int(24*p.scale), p.H-int(24*p.scale))) {
		return false
	}
	return !slices.ContainsFunc(p.obstacles, r.Overlaps)
}

func (p *painter) markLine(line []pt, width float64) {
	r := int(math.Ceil(width/2 + 2))
	for _, q := range line {
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				p.riverMask[p.at(int(q.x)+dx, int(q.y)+dy)] = true
			}
		}
	}
}

func (p *painter) cellPt(c int) pt { return pt{p.w.g.px[c], p.w.g.py[c]} }

// rivers – русла с меандрами, от истока к устью шире.
func (p *painter) rivers() {
	w := p.w
	for _, rv := range w.Rivers {
		var line []pt
		var flux []float64
		for k, c := range rv.Cells {
			q := p.cellPt(c)
			if k > 0 {
				prev := line[len(line)-1]
				dx, dy := q.x-prev.x, q.y-prev.y
				off := p.r.between(-0.22, 0.22)
				line = append(line, pt{(prev.x+q.x)/2 - dy*off, (prev.y+q.y)/2 + dx*off})
				flux = append(flux, w.Flux[c])
			}
			line = append(line, q)
			flux = append(flux, w.Flux[c])
		}
		if rv.Out {
			last := line[len(line)-1]
			line = append(line, p.toEdge(last))
			flux = append(flux, flux[len(flux)-1])
		}
		smooth := chaikin(line, 2)
		widths := make([]float64, len(smooth))
		for k := range smooth {
			f := flux[min(len(flux)-1, k*len(flux)/len(smooth))]
			taper := clamp(float64(k)/6, 0.3, 1)
			widths[k] = clamp(0.5+math.Sqrt(f)/14, 0.6, 5.5) * taper * p.scale
		}
		p.c.fill(strokePolys(smooth, widths), riverColor, 0.95)
		p.markLine(smooth, 4)
	}
}

func (p *painter) toEdge(q pt) pt {
	dl, dr, dt, db := q.x, float64(p.W)-q.x, q.y, float64(p.H)-q.y
	switch math.Min(math.Min(dl, dr), math.Min(dt, db)) {
	case dl:
		return pt{0, q.y}
	case dr:
		return pt{float64(p.W), q.y}
	case dt:
		return pt{q.x, 0}
	}
	return pt{q.x, float64(p.H)}
}

// roads – дороги пунктиром, морские пути – светлым точечным пунктиром.
func (p *painter) roads() {
	for _, path := range p.w.Sea {
		line := make([]pt, len(path))
		for k, c := range path {
			line[k] = p.cellPt(c)
		}
		for _, d := range dashes(chaikin(line, 3), 3*p.scale, 6*p.scale) {
			p.c.stroke(d, 1.6*p.scale, seaWayColor, 0.75)
		}
	}
	for _, path := range p.w.Roads {
		line := make([]pt, len(path))
		for k, c := range path {
			line[k] = p.cellPt(c)
		}
		smooth := chaikin(line, 3)
		for _, d := range dashes(smooth, 7*p.scale, 4*p.scale) {
			p.c.stroke(d, 1.8*p.scale, roadColor, 0.8)
		}
		p.markLine(smooth, 3)
	}
}

// paper – фактура бумаги, тёплый тон и затемнение к краям.
func (p *painter) paper() {
	grain := newNoise(p.r)
	img := p.c.img
	cx, cy := float64(p.W)/2, float64(p.H)/2
	for y := range p.H {
		for x := range p.W {
			fx, fy := float64(x), float64(y)
			k := 0.95 + 0.07*grain.fbm(fx/3, fy/3, 2) + 0.05*(grain.fbm(fx/60+40, fy/60, 3)-0.5)
			rx, ry := (fx-cx)/cx, (fy-cy)/cy
			k *= 1 - 0.22*clamp((math.Sqrt(rx*rx+ry*ry)-0.75)/0.6, 0, 1)
			o := img.PixOffset(x, y)
			for c := range 3 {
				v := float64(img.Pix[o+c])
				warm := [3]float64{240, 226, 196}[c]
				img.Pix[o+c] = uint8(clamp((v*0.9+warm*0.1)*k, 0, 255))
			}
		}
	}
}
