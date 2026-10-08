package atlas

import (
	"image"
	"image/color"
	"math"
	"slices"

	"heroesbook/internal/travel"
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
	snowColor   = rgb(244, 244, 240)
	riverColor  = rgb(64, 110, 146)
	roadColor   = rgb(112, 76, 44)
	seaWayColor = rgb(236, 230, 210)
)

// painter – поля карты попиксельно и всё, что нужно для слоёв.
type painter struct {
	w           *World
	c           canvas
	W, H        int
	hp          []float32 // высота пикселя
	kind        []uint8   // суша, море, озеро
	cell        []int32   // клетка пикселя (с искажением границ шумом)
	dLand       []float32 // расстояние от воды до суши
	dWater      []float32 // от суши до воды
	riverMask   []bool    // где реки, дороги и поля – туда не ставим значки рельефа
	roadOrRiver []bool    // только реки и дороги – туда не ставим постройки
	roadPx      []bool    // только дороги – для путешествий
	obstacles   []image.Rectangle
	labels      []label
	extras      []extra // мельницы у деревень
	r           rng
	warpX       *noise
	warpY       *noise
	detail      *noise
	scale       float64 // множитель размеров значков и шрифтов от размера карты
	cartouche   image.Rectangle
	compassBox  image.Rectangle
}

// Render рисует мир. Поселения, оказавшиеся из-за сглаживания берегов в воде, сдвигаются на ближайшую сушу.
func (w *World) Render(r Rand) *image.RGBA {
	p := &painter{w: w, W: w.Width, H: w.Height, r: rng{r}}
	p.c = canvas{image.NewRGBA(image.Rect(0, 0, p.W, p.H))}
	p.scale = clamp(math.Sqrt(float64(p.W*p.H))/1500, 0.7, 1.7)
	p.warpX, p.warpY, p.detail = newNoise(p.r), newNoise(p.r), newNoise(p.r)
	p.fields()
	p.moveBurgsAshore()
	p.base()
	p.states()
	p.layoutDecor()
	p.rivers()
	p.roads()
	p.farmland()
	p.burgObstacles()
	p.labelsLayout()
	p.relief()
	p.burgs()
	p.ships()
	p.drawLabels()
	p.decor()
	w.travel = p.travelGrid(travelStep)
	p.clouds()
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
// Веса раздельные (по x и по y), поэтому на пиксель – 8 экспонент, а не 16.
func (p *painter) gauss(u, v float64, val func(c int) float64) float64 {
	g := p.w.g
	i0, j0 := int(math.Floor(u)), int(math.Floor(v))
	var wx, wy [4]float64
	for k := range 4 {
		wx[k] = math.Exp(-sq(float64(i0-1+k)-u) / 0.9)
		wy[k] = math.Exp(-sq(float64(j0-1+k)-v) / 0.9)
	}
	var sum, wsum float64
	for b := range 4 {
		j := min(g.ny-1, max(0, j0-1+b))
		for a := range 4 {
			wt := wx[a] * wy[b]
			sum += wt * val(j*g.nx+min(g.nx-1, max(0, i0-1+a)))
			wsum += wt
		}
	}
	return sum / wsum
}

func (p *painter) fields() {
	w := p.w
	n := p.W * p.H
	p.hp, p.kind, p.cell = make([]float32, n), make([]uint8, n), make([]int32, n)
	height := func(c int) float64 { return float64(w.H[c]) }
	lake := func(c int) float64 {
		if w.Lake[c] {
			return 1
		}
		return 0
	}
	parallelRows(p.H, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := range p.W {
				p.fieldAt(x, y, height, lake)
			}
		}
	})
	land, water := make([]bool, n), make([]bool, n)
	for i, k := range p.kind {
		land[i], water[i] = k == pxLand, k != pxLand
	}
	p.dLand, p.dWater = distance(land, p.W, p.H), distance(water, p.W, p.H)
	p.riverMask, p.roadOrRiver, p.roadPx = make([]bool, n), make([]bool, n), make([]bool, n)
}

// fieldAt – высота, клетка и вид пикселя. В горах к сглаженной высоте добавляются хребты и долины,
// на холмах – мягкие увалы, везде – мелкая неровность берегов.
func (p *painter) fieldAt(x, y int, height, lake func(int) float64) {
	w, g := p.w, p.w.g
	i := y*p.W + x
	fx, fy := float64(x), float64(y)
	u, v := p.lattice(fx, fy)
	base := p.gauss(u, v, height)                                                          // гладко: билинейная интерполяция по квадратной решётке даёт прямые грани
	h := base + (p.detail.fbm(fx/16, fy/16, 4)-0.5)*(1.5+5.5*(1-smoothstep(22, 34, base))) // неровность – у берега
	if m := smoothstep(40, 72, base); m > 0 {
		h += m * (p.detail.ridged(fx/(72*p.scale)+31, fy/(72*p.scale)+7, 6) - 0.4) * 38
		h += m * (p.warpY.ridged(fx/(16*p.scale)+3, fy/(16*p.scale)+17, 3) - 0.45) * 9 // мелкие гребни и осыпи
	}
	if hl := smoothstep(28, 50, base) * (1 - smoothstep(55, 75, base)); hl > 0 {
		h += hl * (p.detail.fbm(fx/(30*p.scale)+3, fy/(30*p.scale)+9, 3) - 0.5) * 9
	}
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

func (p *painter) at(x, y int) int { return min(p.H-1, max(0, y))*p.W + min(p.W-1, max(0, x)) }

func (p *painter) moveBurgsAshore() {
	for i := range p.w.Burgs {
		p.ashore(&p.w.Burgs[i].X, &p.w.Burgs[i].Y)
	}
	for i := range p.w.Sites {
		p.ashore(&p.w.Sites[i].X, &p.w.Sites[i].Y)
	}
}

// ashore сдвигает точку, оказавшуюся в воде или у самой кромки, на ближайшую сушу.
func (p *painter) ashore(px, py *float64) {
	{
		x, y := int(*px), int(*py)
		if p.kind[p.at(x, y)] == pxLand && p.dWater[p.at(x, y)] > 3 {
			return
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
			*px, *py = float64(best%p.W), float64(best/p.W)
		}
	}
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
		a := 0.06 + 0.26*clamp(1-d/(9*p.scale), 0, 1)
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
		for _, run := range p.landRuns(smooth) { // через озёра и море русло не рисуется
			line, ws := smooth[run[0]:run[1]], widths[run[0]:run[1]]
			banks := make([]float64, len(ws))
			core := make([]float64, len(ws))
			for k, w := range ws {
				banks[k], core[k] = w+1.6*p.scale, w*0.45
			}
			p.c.fill(strokePolys(line, banks), rgb(70, 82, 60), 0.45) // влажные берега
			p.c.fill(strokePolys(line, ws), riverColor, 0.95)
			p.c.fill(strokePolys(line, core), rgb(120, 168, 190), 0.45) // светлая стремнина
		}
		p.markLine(smooth, 4)
		p.markRoad(smooth, 4)
	}
}

// landRuns – отрезки ломаной [от; до), лежащие над сушей; на шаг заходят в воду, чтобы река в неё впадала.
func (p *painter) landRuns(line []pt) [][2]int {
	var out [][2]int
	start := -1
	for i, q := range line {
		land := p.kind[p.at(int(q.x), int(q.y))] == pxLand
		switch {
		case land && start < 0:
			start = max(0, i-1)
		case !land && start >= 0:
			out = append(out, [2]int{start, min(len(line), i+1)})
			start = -1
		}
	}
	if start >= 0 && len(line)-start > 1 {
		out = append(out, [2]int{start, len(line)})
	}
	return out
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
		p.markRoad(smooth, 3)
		for _, q := range smooth {
			p.roadPx[p.at(int(q.x), int(q.y))] = true
		}
	}
}

func (p *painter) markRoad(line []pt, width float64) {
	r := int(math.Ceil(width/2 + 1))
	for _, q := range line {
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				p.roadOrRiver[p.at(int(q.x)+dx, int(q.y)+dy)] = true
			}
		}
	}
}

// clouds – лёгкие тени облаков: крупные мягкие пятна, чуть темнее на земле и на воде.
func (p *painter) clouds() {
	n := newNoise(p.r)
	img := p.c.img
	s := 520 * p.scale
	parallelRows(p.H, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := range p.W {
				c := smoothstep(0.56, 0.74, n.fbm(float64(x)/s, float64(y)/s, 4))
				if c == 0 {
					continue
				}
				k := 1 - 0.1*c
				o := img.PixOffset(x, y)
				img.Pix[o] = uint8(float64(img.Pix[o]) * k)
				img.Pix[o+1] = uint8(float64(img.Pix[o+1]) * k)
				img.Pix[o+2] = uint8(float64(img.Pix[o+2]) * (k + 0.03*c))
			}
		}
	})
}

// paper – фактура бумаги, тёплый тон и затемнение к краям.
func (p *painter) paper() {
	grain := newNoise(p.r)
	img := p.c.img
	cx, cy := float64(p.W)/2, float64(p.H)/2
	parallelRows(p.H, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := range p.W {
				fx, fy := float64(x), float64(y)
				k := 0.975 + 0.035*grain.fbm(fx/2.5, fy/2.5, 2) + 0.04*(grain.fbm(fx/80+40, fy/80, 3)-0.5)
				rx, ry := (fx-cx)/cx, (fy-cy)/cy
				k *= 1 - 0.2*clamp((math.Sqrt(rx*rx+ry*ry)-0.8)/0.6, 0, 1)
				o := img.PixOffset(x, y)
				for c := range 3 {
					v := float64(img.Pix[o+c])
					warm := [3]float64{240, 226, 196}[c]
					img.Pix[o+c] = uint8(clamp((v*0.94+warm*0.06)*k, 0, 255))
				}
			}
		}
	})
}

const travelStep = 8 // клетка сетки путешествий, пикселей

// MilesPerPx – масштаб карты мира (тот же, что у масштабной линейки).
const MilesPerPx = milesPerPx

// Travel – сетка местности для путешествий (есть после Render).
func (w *World) Travel() *travel.Grid { return w.travel }

// travelGrid – местность клеток для путешествий: вода, дорога или местность клетки мира с учётом высоты.
func (p *painter) travelGrid(step int) *travel.Grid {
	g := &travel.Grid{Cols: max(1, p.W/step), Rows: max(1, p.H/step), Step: step}
	g.Cells = make([]uint8, g.Cols*g.Rows)
	for j := range g.Rows {
		for i := range g.Cols {
			g.Cells[j*g.Cols+i] = p.travelCell(i*step, j*step, step)
		}
	}
	return g
}

func (p *painter) travelCell(x0, y0, step int) uint8 {
	water, road := 0, false
	for y := y0; y < y0+step; y += 2 {
		for x := x0; x < x0+step; x += 2 {
			i := p.at(x, y)
			if p.kind[i] != pxLand {
				water++
			}
			road = road || p.roadPx[i]
		}
	}
	if water*2 > step*step/4 {
		return travel.Water
	}
	if road {
		return travel.Road
	}
	i := p.at(x0+step/2, y0+step/2)
	switch h := float64(p.hp[i]); {
	case h >= 64:
		return travel.Mountain
	case h >= 48:
		return travel.Hills
	}
	switch p.w.TerrainOf(int(p.cell[i])) {
	case Forest, Jungle:
		return travel.Forest
	case Swamp:
		return travel.Swamp
	case Desert:
		return travel.Desert
	case Snow, Tundra:
		return travel.Snow
	case Mountain:
		return travel.Mountain
	case Hills:
		return travel.Hills
	}
	return travel.Plain
}
