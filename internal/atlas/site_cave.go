package atlas

import (
	"image/color"
	"math"
	"sort"
)

// Пещера, шахта или логово сверху: залы-«пузыри», соединённые извилистыми ходами, сглаженные клеточным
// автоматом; края неровные попиксельно; у стен – тень, на полу – камешки, сталагмиты и подземные озёра.
// В шахте – крепь и рельсы, в логове – кости и груда сокровищ. Вход – с края карты.

const caveCell = 8.0

type cavePlan struct {
	w, h  int
	open  []bool
	halls []pt // центры залов в клетках
	sizes []float64
	paths [][]pt
}

func (c *cavePlan) at(x, y int) bool {
	return x >= 0 && y >= 0 && x < c.w && y < c.h && c.open[y*c.w+x]
}

func (s *site) cave() {
	plan := s.cavePlan()
	s.caveRender(plan)
	s.caveDetails(plan)
	s.cavePlaces(plan)
	s.vignette()
	s.cartouche()
}

func (s *site) cavePlan() *cavePlan {
	c := &cavePlan{w: int(float64(s.W) / caveCell), h: int(float64(s.H) / caveCell)}
	c.open = make([]bool, c.w*c.h)
	n := 8 + s.r.r.Intn(4)
	for len(c.halls) < n {
		p := pt{s.r.between(0.12, 0.88) * float64(c.w), s.r.between(0.15, 0.85) * float64(c.h)}
		ok := true
		for _, q := range c.halls {
			if math.Hypot(p.x-q.x, p.y-q.y) < 34 {
				ok = false
			}
		}
		if ok || s.r.chance(0.05) {
			c.halls = append(c.halls, p)
			c.sizes = append(c.sizes, s.r.between(12, 22))
		}
	}
	// вход – ход от края карты к ближайшему залу
	entry := pt{0, s.r.between(0.3, 0.7) * float64(c.h)}
	c.halls = append([]pt{entry}, c.halls...)
	c.sizes = append([]float64{4}, c.sizes...)
	for i, h := range c.halls {
		s.blob(c, h, c.sizes[i])
	}
	// ходы: дерево ближайших связей и пара лишних петель
	in := map[int]bool{0: true}
	for len(in) < len(c.halls) {
		bi, bj, bd := -1, -1, math.Inf(1)
		for i := range c.halls {
			if !in[i] {
				continue
			}
			for j := range c.halls {
				if in[j] {
					continue
				}
				if d := math.Hypot(c.halls[i].x-c.halls[j].x, c.halls[i].y-c.halls[j].y); d < bd {
					bi, bj, bd = i, j, d
				}
			}
		}
		in[bj] = true
		s.tunnel(c, c.halls[bi], c.halls[bj])
	}
	for range 2 {
		i, j := s.r.r.Intn(len(c.halls)), s.r.r.Intn(len(c.halls))
		if i != j {
			s.tunnel(c, c.halls[i], c.halls[j])
		}
	}
	for range 2 { // сглаживание клеточным автоматом
		next := make([]bool, len(c.open))
		for y := range c.h {
			for x := range c.w {
				n := 0
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if c.at(x+dx, y+dy) {
							n++
						}
					}
				}
				next[y*c.w+x] = n >= 5 || c.open[y*c.w+x] && n >= 4
			}
		}
		c.open = next
	}
	return c
}

func (s *site) blob(c *cavePlan, p pt, r float64) {
	seed := s.r.between(0, 100)
	for y := int(p.y - r*1.6); y <= int(p.y+r*1.6); y++ {
		for x := int(p.x - r*1.6); x <= int(p.x+r*1.6); x++ {
			if x < 1 || y < 1 || x >= c.w-1 || y >= c.h-1 {
				if x == 0 && y > 0 && y < c.h && r <= 4 { // вход открыт к краю
					c.open[y*c.w] = math.Hypot(float64(x)-p.x, float64(y)-p.y) < r
				}
				continue
			}
			a := math.Atan2(float64(y)-p.y, float64(x)-p.x)
			rr := r * (0.75 + 0.5*s.n1.fbm(seed+math.Cos(a)*1.5, seed+math.Sin(a)*1.5, 3))
			if math.Hypot(float64(x)-p.x, (float64(y)-p.y)*1.15) < rr {
				c.open[y*c.w+x] = true
			}
		}
	}
}

// tunnel – извилистый ход переменной ширины между двумя точками.
func (s *site) tunnel(c *cavePlan, a, b pt) {
	path := s.wiggle(a, b, math.Hypot(b.x-a.x, b.y-a.y)*0.35)
	path = chaikin(path, 2)
	c.paths = append(c.paths, path)
	for i := 0; i+1 < len(path); i++ {
		p, q := path[i], path[i+1]
		l := math.Hypot(q.x-p.x, q.y-p.y)
		for t := 0.0; t <= l; t += 0.5 {
			x, y := p.x+(q.x-p.x)*t/l, p.y+(q.y-p.y)*t/l
			w := 2 + 2*s.n2.fbm(x/10, y/10, 2)
			for dy := -3; dy <= 3; dy++ {
				for dx := -3; dx <= 3; dx++ {
					xx, yy := int(x)+dx, int(y)+dy
					if xx >= 1 && yy >= 1 && xx < c.w-1 && yy < c.h-1 && math.Hypot(float64(dx), float64(dy)) < w {
						c.open[yy*c.w+xx] = true
					}
				}
			}
		}
	}
}

// caveRender – скала и пол попиксельно: неровная кромка, тень у стен, светлый гребень скалы.
func (s *site) caveRender(c *cavePlan) {
	floor := rgb(124, 112, 96)
	rock := rgb(70, 64, 58)
	switch s.o.Kind {
	case "mine":
		floor = rgb(132, 108, 80)
	case "lair":
		floor = rgb(110, 96, 84)
	}
	if s.pal.snow {
		rock = rgb(96, 104, 116)
	}
	openPx := make([]bool, s.W*s.H)
	val := func(x, y int) float64 {
		if c.at(x, y) {
			return 1
		}
		return 0
	}
	parallelRows(s.H, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := range s.W {
				u, v := float64(x)/caveCell-0.5, float64(y)/caveCell-0.5
				i0, j0 := int(math.Floor(u)), int(math.Floor(v))
				tx, ty := u-float64(i0), v-float64(j0)
				a := (val(i0, j0)*(1-tx)+val(i0+1, j0)*tx)*(1-ty) + (val(i0, j0+1)*(1-tx)+val(i0+1, j0+1)*tx)*ty
				a += (s.n1.fbm(float64(x)/9, float64(y)/9, 3) - 0.5) * 0.5
				openPx[y*s.W+x] = a > 0.5
			}
		}
	})
	walls := make([]bool, len(openPx))
	for i, o := range openPx {
		walls[i] = !o
	}
	dWall := distance(walls, s.W, s.H)   // от пола до стены
	dFloor := distance(openPx, s.W, s.H) // от скалы до пола
	img := s.c.img
	parallelRows(s.H, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := range s.W {
				i := y*s.W + x
				fx, fy := float64(x), float64(y)
				var col color.RGBA
				if openPx[i] {
					d := float64(dWall[i])
					col = shade(floor, 0.86+0.2*s.n2.fbm(fx/14, fy/14, 3))
					col = shade(col, 0.92+0.12*s.n1.fbm(fx/2.5, fy/2.5, 2))
					col = shade(col, 0.45+0.55*smoothstep(0, 26, d)) // тень у стен
				} else {
					d := float64(dFloor[i])
					col = shade(rock, 0.8+0.3*s.n2.fbm(fx/22, fy/22, 4))
					col = shade(col, 1.35-0.6*smoothstep(0, 40, d)) // у кромки светлее, вглубь темнее
					crack := s.n1.ridged(fx/40, fy/40, 3)
					col = mix(col, shade(rock, 0.45), smoothstep(0.85, 0.97, crack)*0.6)
				}
				o := img.PixOffset(x, y)
				img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = col.R, col.G, col.B, 255
			}
		}
	})
	// контур кромки
	parallelRows(s.H, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := range s.W {
				if d := float64(dFloor[y*s.W+x]); d > 0 && d < 2.2 {
					s.c.blend(x, y, rgb(30, 26, 22), (2.2-d)/2.2*0.8)
				}
			}
		}
	})
	s.caveOpen = openPx
}

// caveDetails – озёра, сталагмиты, камни; крепь и рельсы в шахте; кости и сокровища в логове.
func (s *site) caveDetails(c *cavePlan) {
	open := func(x, y float64) bool {
		xi, yi := int(x), int(y)
		return xi >= 0 && yi >= 0 && xi < s.W && yi < s.H && s.caveOpen[yi*s.W+xi]
	}
	// подземное озеро в одном из залов
	if len(c.halls) > 3 && s.o.Kind != "mine" {
		h := c.halls[2+s.r.r.Intn(len(c.halls)-2)]
		cx, cy := h.x*caveCell, h.y*caveCell
		var pool []pt
		for k := range 20 {
			a := float64(k) * 2 * math.Pi / 20
			r := 50 * (0.7 + 0.6*s.n1.fbm(math.Cos(a)+5, math.Sin(a)+5, 2))
			pool = append(pool, pt{cx + math.Cos(a)*r*1.3, cy + math.Sin(a)*r})
		}
		s.water([][]pt{chaikin(append(pool, pool[0]), 2)})
	}
	for range 260 {
		x, y := s.r.between(0, float64(s.W)), s.r.between(0, float64(s.H))
		if !open(x, y) || s.busy(x, y) {
			continue
		}
		r := s.r.between(3, 9)
		s.rock(x, y, r)
	}
	switch s.o.Kind {
	case "mine":
		for _, path := range c.paths {
			line := make([]pt, len(path))
			for i, q := range path {
				line[i] = pt{q.x * caveCell, q.y * caveCell}
			}
			s.rails(line)
		}
	case "lair":
		big := 1
		for i := range c.sizes {
			if c.sizes[i] > c.sizes[big] {
				big = i
			}
		}
		h := c.halls[big]
		s.hoard(h.x*caveCell, h.y*caveCell)
		for range 60 {
			x, y := h.x*caveCell+s.r.between(-160, 160), h.y*caveCell+s.r.between(-120, 120)
			if open(x, y) {
				s.bone(x, y)
			}
		}
	}
}

// rock – камень или сталагмит сверху: тень, тёмное основание, светлая макушка.
func (s *site) rock(x, y, r float64) {
	s.c.fill([][]pt{ellipse(x+r*0.5, y+r*0.5, r, r*0.9)}, rgb(10, 8, 6), 0.35)
	s.c.fill([][]pt{ellipse(x, y, r, r*0.85)}, rgb(96, 88, 78), 1)
	s.c.fill([][]pt{circle(x-r*0.25, y-r*0.25, r*0.5)}, rgb(150, 140, 126), 1)
	s.occupy([][]pt{circle(x, y, r)})
}

// rails – рельсы со шпалами и крепь поперёк хода.
func (s *site) rails(line []pt) {
	line = chaikin(line, 2)
	for i := 0; i+1 < len(line); i += 3 {
		a, b := line[i], line[i+1]
		l := math.Hypot(b.x-a.x, b.y-a.y)
		if l == 0 {
			continue
		}
		nx, ny := -(b.y-a.y)/l, (b.x-a.x)/l
		s.c.stroke([]pt{{a.x - nx*9, a.y - ny*9}, {a.x + nx*9, a.y + ny*9}}, 3, rgb(96, 70, 44), 0.95)
		if i%15 == 0 { // крепь
			s.c.stroke([]pt{{a.x - nx*22, a.y - ny*22}, {a.x + nx*22, a.y + ny*22}}, 6, rgb(120, 86, 52), 0.95)
			s.c.stroke([]pt{{a.x - nx*22, a.y - ny*22}, {a.x + nx*22, a.y + ny*22}}, 1, rgb(60, 40, 24), 0.8)
		}
	}
	for _, sgn := range []float64{-1, 1} {
		var rail []pt
		for i := 0; i+1 < len(line); i++ {
			a, b := line[i], line[i+1]
			l := math.Max(1e-6, math.Hypot(b.x-a.x, b.y-a.y))
			nx, ny := -(b.y-a.y)/l, (b.x-a.x)/l
			rail = append(rail, pt{a.x + nx*6*sgn, a.y + ny*6*sgn})
		}
		s.c.stroke(rail, 1.6, rgb(70, 70, 76), 0.95)
	}
}

func (s *site) hoard(x, y float64) {
	for range 220 {
		a, d := s.r.between(0, 2*math.Pi), s.r.between(0, 55)*s.r.float()
		px, py := x+math.Cos(a)*d*1.3, y+math.Sin(a)*d
		col := rgb(232, 190, 70)
		if s.r.chance(0.15) {
			col = []color.RGBA{rgb(200, 60, 60), rgb(60, 160, 220), rgb(90, 200, 120)}[s.r.r.Intn(3)]
		}
		s.c.fill([][]pt{circle(px, py, s.r.between(1.5, 3.5))}, shade(col, s.r.between(0.8, 1.2)), 1)
	}
	s.occupy([][]pt{circle(x, y, 70)})
}

func (s *site) bone(x, y float64) {
	a := s.r.between(0, math.Pi)
	l := s.r.between(6, 14)
	p, q := pt{x - math.Cos(a)*l, y - math.Sin(a)*l}, pt{x + math.Cos(a)*l, y + math.Sin(a)*l}
	s.c.stroke([]pt{p, q}, 2.2, rgb(226, 218, 196), 0.95)
	s.c.fill([][]pt{circle(p.x, p.y, 2.4), circle(q.x, q.y, 2.4)}, rgb(236, 228, 208), 1)
}

// cavePlaces – залы становятся местами: вход, гроты, у логова – логово, у шахты – забой.
func (s *site) cavePlaces(c *cavePlan) {
	idx := make([]int, len(c.halls)-1)
	for i := range idx {
		idx[i] = i + 1
	}
	sort.SliceStable(idx, func(a, b int) bool { return c.sizes[idx[a]] > c.sizes[idx[b]] })
	names := map[string][]string{
		"cave": {"Большой грот", "Зал сталактитов", "Подземное озеро", "Узкий лаз", "Гулкий зал", "Грибная пещера", "Колодец тьмы", "Хрустальный грот", "Нижний ярус"},
		"mine": {"Главный забой", "Рудный склад", "Обвал", "Старая штольня", "Кузня рудокопов", "Затопленный ход", "Жила", "Подъёмник", "Нижний горизонт"},
		"lair": {"Логово", "Кладовая костей", "Сторожевой зал", "Гнездо", "Узкий лаз", "Нижний грот", "Пещера эха", "Тайник", "Обрыв"},
	}[s.o.Kind]
	if names == nil {
		names = []string{"Зал"}
	}
	e := c.halls[0]
	s.addPlace(e.x*caveCell+30, e.y*caveCell, "room", "Вход")
	for k, i := range idx {
		h := c.halls[i]
		s.addPlace(h.x*caveCell, h.y*caveCell, "room", names[k%len(names)])
	}
}
