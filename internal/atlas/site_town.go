package atlas

import (
	"image/color"
	"math"
)

// Поселение сверху: главные дороги расходятся от площади к краям карты, у города – кольцевая улица и переулки;
// дома стоят вдоль улиц фасадом к ним, к окраинам реже; на площади – рынок и колодец; храм, таверна, кузница;
// у города и столицы – стена с башнями и воротами, у столицы – замок; у воды – пристани; за стенами – поля и лес.

type townSpec struct {
	roads, radius    float64
	walls, keep      bool
	houseW, houseGap float64
	second           float64 // вероятность второго ряда домов за первым
}

var townSpecs = map[string]townSpec{
	"village": {roads: 2, radius: 330, houseW: 46, houseGap: 26, second: 0.15},
	"town":    {roads: 3, radius: 520, houseW: 42, houseGap: 8, second: 0.6},
	"port":    {roads: 3, radius: 520, houseW: 42, houseGap: 8, second: 0.6},
	"city":    {roads: 4, radius: 700, walls: true, houseW: 38, houseGap: 3, second: 0.9},
	"capital": {roads: 5, radius: 880, walls: true, keep: true, houseW: 38, houseGap: 2, second: 0.95},
}

type street struct {
	line  []pt
	width float64
	main  bool
}

func (s *site) town() {
	spec := townSpecs[s.o.Kind]
	water := s.o.Water || s.o.Kind == "port"
	s.ground()
	cx, cy := float64(s.W)/2, float64(s.H)/2+20
	var sea []pt
	if water {
		sea = s.shore(s.o.WaterDir, 0.24)
		s.water([][]pt{sea})
		cx += math.Cos(s.o.WaterDir) * float64(s.W) * 0.1 // город жмётся к воде
		cy += math.Sin(s.o.WaterDir) * float64(s.H) * 0.1
	} else if s.r.chance(0.55) {
		s.river(46)
	}
	plaza := s.plazaPoly(cx, cy, spec.radius*0.1+50)
	streets := s.townStreets(cx, cy, spec, water)
	for _, st := range streets {
		s.townRoad(cx, cy, spec, st)
	}
	s.texture([][]pt{plaza}, 1, s.cobble)
	s.occupy([][]pt{plaza})
	if spec.walls {
		s.townWall(cx, cy, spec.radius*0.82, streets)
	}
	s.specialBuildings(cx, cy, spec)
	if water {
		s.docks(cx, cy, sea, spec)
	}
	s.houses(cx, cy, spec, streets)
	s.fieldsAround(cx, cy, spec.radius)
	s.clearing(cx, cy, spec.radius*1.25)
	s.forestFill(0.8, 0.56)
	s.forestFill(0.3, 0.42)
	s.gardens(cx, cy, spec.radius)
	s.vignette()
	s.cartouche()
}

func (s *site) plazaPoly(cx, cy, r float64) []pt {
	var out []pt
	for k := range 14 {
		a := float64(k) * 2 * math.Pi / 14
		rr := r * (0.85 + 0.3*s.r.float())
		out = append(out, pt{cx + math.Cos(a)*rr*1.2, cy + math.Sin(a)*rr})
	}
	return chaikin(append(out, out[0]), 2)
}

// wiggle – ломаная от a к b с плавными изгибами.
func (s *site) wiggle(a, b pt, amp float64) []pt {
	n := int(math.Hypot(b.x-a.x, b.y-a.y)/60) + 2
	dx, dy := b.x-a.x, b.y-a.y
	l := math.Hypot(dx, dy)
	nx, ny := -dy/l, dx/l
	seed := s.r.between(0, 100)
	out := make([]pt, 0, n+1)
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		off := (s.n1.fbm(seed+t*3, seed, 2) - 0.5) * amp * math.Sin(t*math.Pi)
		out = append(out, pt{a.x + dx*t + nx*off, a.y + dy*t + ny*off})
	}
	return out
}

func (s *site) townStreets(cx, cy float64, spec townSpec, water bool) []street {
	var out []street
	base := s.r.between(0, 2*math.Pi)
	span := math.Hypot(float64(s.W), float64(s.H))
	for k := range int(spec.roads) {
		a := base + float64(k)*2*math.Pi/spec.roads + s.r.between(-0.3, 0.3)
		if water && math.Cos(a-s.o.WaterDir) > 0.5 {
			a += math.Pi * 0.5 // в море дороги не ведут
		}
		end := pt{cx + math.Cos(a)*span, cy + math.Sin(a)*span}
		out = append(out, street{s.wiggle(pt{cx, cy}, end, 160), 26, true})
	}
	if spec.radius < 400 {
		return out
	}
	r := spec.radius * 0.5 // кольцевая улица и переулки
	var ring []pt
	for k := 0; k <= 24; k++ {
		a := float64(k) * 2 * math.Pi / 24
		rr := r * (0.9 + 0.2*s.n1.fbm(float64(k%24)/4, 3, 2))
		ring = append(ring, pt{cx + math.Cos(a)*rr*1.15, cy + math.Sin(a)*rr})
	}
	out = append(out, street{ring, 18, false})
	for range int(spec.roads * 2.5) {
		a := s.r.between(0, 2*math.Pi)
		r0, r1 := spec.radius*s.r.between(0.15, 0.35), spec.radius*s.r.between(0.6, 0.95)
		from := pt{cx + math.Cos(a)*r0*1.15, cy + math.Sin(a)*r0}
		b := a + s.r.between(-0.4, 0.4)
		to := pt{cx + math.Cos(b)*r1*1.15, cy + math.Sin(b)*r1}
		out = append(out, street{s.wiggle(from, to, 60), 14, false})
	}
	return out
}

// townWall – стена кольцом с башнями; где её пересекают главные дороги – ворота.
func (s *site) townWall(cx, cy, r float64, streets []street) {
	n := 16
	ring := make([]pt, n)
	for k := range n {
		a := float64(k) * 2 * math.Pi / float64(n)
		rr := r * (0.92 + 0.16*s.r.float())
		ring[k] = pt{cx + math.Cos(a)*rr*1.18, cy + math.Sin(a)*rr}
	}
	wallCol := s.pal.stone[0]
	gate := func(q pt) bool {
		for _, st := range streets {
			if !st.main {
				continue
			}
			for _, sp := range chaikin(st.line, 2) {
				if math.Hypot(sp.x-q.x, sp.y-q.y) < 30 {
					return true
				}
			}
		}
		return false
	}
	for k := range n {
		a, b := ring[k], ring[(k+1)%n]
		steps := int(math.Hypot(b.x-a.x, b.y-a.y) / 6)
		var seg []pt
		flush := func() {
			if len(seg) > 1 {
				s.shadowLine(seg, 22, 16)
				s.texture(strokePolys(seg, []float64{22}), 1, func(x, y int) color.RGBA {
					return shade(wallCol, 0.85+0.25*s.n2.fbm(float64(x)/4, float64(y)/4, 2))
				})
				s.c.stroke(offset(seg, -5, -5), 3, shade(wallCol, 1.25), 0.5) // освещённый гребень
				s.c.stroke(seg, 1, shade(wallCol, 0.6), 0.5)
				s.occupy(strokePolys(seg, []float64{34}))
			}
			seg = nil
		}
		for i := 0; i <= steps; i++ {
			t := float64(i) / float64(max(1, steps))
			q := pt{a.x + (b.x-a.x)*t, a.y + (b.y-a.y)*t}
			if gate(q) {
				flush()
				continue
			}
			seg = append(seg, q)
		}
		flush()
	}
	for _, q := range ring {
		s.tower(q.x, q.y, 24, wallCol)
	}
}

// shadowLine – тень от стены высотой h вдоль линии.
func (s *site) shadowLine(line []pt, width, h float64) {
	s.c.fill(strokePolys(offset(line, h*0.5, h*0.5), []float64{width}), rgb(20, 24, 16), 0.3)
}

// tower – круглая башня: тень, стена, коническая крыша из секторов разной освещённости.
func (s *site) tower(x, y, r float64, stone color.RGBA) {
	s.c.fill([][]pt{offset(circle(x, y, r), r*0.9, r*0.9)}, rgb(20, 24, 16), 0.32)
	s.c.fill([][]pt{circle(x, y, r)}, shade(stone, 0.8), 1)
	roof := s.pal.roof[0]
	for k := range 12 {
		a0, a1 := float64(k)*math.Pi/6, float64(k+1)*math.Pi/6
		am := (a0 + a1) / 2
		s.c.fill([][]pt{{{x, y}, {x + math.Cos(a0)*r*0.86, y + math.Sin(a0)*r*0.86}, {x + math.Cos(a1)*r*0.86, y + math.Sin(a1)*r*0.86}}},
			shade(roof, light(math.Cos(am), math.Sin(am))), 1)
	}
	s.c.stroke(closed(circle(x, y, r)), 1.2, rgb(40, 30, 24), 0.7)
	s.occupy([][]pt{circle(x, y, r+2)})
}

// specialBuildings – рынок, храм, таверна, кузница, замок; каждое становится местом на карте.
func (s *site) specialBuildings(cx, cy float64, spec townSpec) {
	roof := s.pal.roof
	for range 6 + int(spec.radius/100) {
		a, d := s.r.between(0, 2*math.Pi), s.r.between(20, spec.radius*0.12+20)
		s.stall(cx+math.Cos(a)*d*1.2, cy+math.Sin(a)*d, s.r.between(0, math.Pi))
	}
	s.c.fill([][]pt{offset(circle(cx, cy, 9), 4, 4)}, rgb(20, 20, 20), 0.3) // колодец
	s.c.fill([][]pt{circle(cx, cy, 9)}, shade(s.pal.stone[0], 0.9), 1)
	s.c.fill([][]pt{circle(cx, cy, 5)}, rgb(40, 80, 110), 1)
	s.occupy([][]pt{circle(cx, cy, 12)})
	s.addPlace(cx+24, cy+18, "shop", "Рынок")
	place := func(kind, name string, w, d float64, hip bool, col color.RGBA, dist float64) bool {
		for range 60 {
			a := s.r.between(0, 2*math.Pi)
			r := s.r.between(dist, dist+70)
			x, y := cx+math.Cos(a)*r*1.15, cy+math.Sin(a)*r
			ang := math.Atan2(cy-y, cx-x) + math.Pi/2
			if !s.freePoly(rotRect(x, y, w+16, d+16, ang), 40) {
				continue
			}
			s.house(x, y, w, d, ang, col, hip)
			s.addPlace(x, y, kind, name)
			return true
		}
		return false
	}
	if spec.keep {
		s.keep(cx, cy, spec.radius)
	}
	if place("temple", "Храм", 74, 40, false, shade(s.pal.stone[0], 0.75), spec.radius*0.18+40) {
		p := s.places[len(s.places)-1]
		s.tower(p.X, p.Y, 12, s.pal.stone[0]) // колокольня
	}
	place("tavern", "Таверна", 54, 36, true, roof[0], spec.radius*0.15+30)
	place("smithy", "Кузница", 40, 30, true, roof[len(roof)-1], spec.radius*0.35)
	if spec.radius >= 600 {
		place("tower", "Башня мага", 30, 30, true, rgb(84, 70, 120), spec.radius*0.4)
	}
}

// keep – замок правителя: стена-квадрат с башнями по углам, двор, донжон.
func (s *site) keep(cx, cy, radius float64) {
	a := s.r.between(0, 2*math.Pi)
	x, y := cx+math.Cos(a)*radius*0.32*1.15, cy+math.Sin(a)*radius*0.32
	ang := s.r.between(-0.2, 0.2)
	wall := rotRect(x, y, 170, 150, ang)
	s.shadow(wall, 12)
	s.texture([][]pt{wall}, 1, s.cobble)
	s.c.stroke(closed(wall), 12, shade(s.pal.stone[0], 0.95), 1)
	s.c.stroke(closed(wall), 2, shade(s.pal.stone[0], 1.25), 0.7)
	for _, q := range wall {
		s.tower(q.x, q.y, 18, s.pal.stone[0])
	}
	s.house(x, y, 80, 66, ang, shade(s.pal.stone[0], 0.7), true)
	s.tower(x+28, y-24, 14, s.pal.stone[0])
	s.occupy([][]pt{rotRect(x, y, 200, 180, ang)})
	s.addPlace(x, y, "castle", "Замок")
}

// stall – рыночный навес в полоску.
func (s *site) stall(x, y, ang float64) {
	r := rotRect(x, y, 16, 11, ang)
	if !s.freePoly(r, 10) {
		return
	}
	s.shadow(r, 5)
	cols := []color.RGBA{rgb(186, 60, 50), rgb(60, 90, 160), rgb(200, 170, 60), rgb(70, 130, 70)}
	c1 := cols[s.r.r.Intn(len(cols))]
	for k := range 4 {
		c := c1
		if k%2 == 1 {
			c = rgb(236, 228, 210)
		}
		ca, sa := math.Cos(ang), math.Sin(ang)
		u := -6 + float64(k)*4
		s.c.fill([][]pt{rotRect(x+u*ca, y+u*sa, 4, 11, ang)}, c, 1)
	}
	s.c.stroke(closed(r), 0.8, rgb(40, 30, 24), 0.6)
	s.occupy([][]pt{r})
}

// houses – дома вдоль улиц с обеих сторон, фасадом к улице; к окраине реже.
func (s *site) houses(cx, cy float64, spec townSpec, streets []street) {
	for _, st := range streets {
		line := chaikin(st.line, 2)
		for i := 0; i+1 < len(line); i++ {
			a, b := line[i], line[i+1]
			seg := math.Hypot(b.x-a.x, b.y-a.y)
			if seg < 1 {
				continue
			}
			ang := math.Atan2(b.y-a.y, b.x-a.x)
			nx, ny := -math.Sin(ang), math.Cos(ang)
			for t := 0.0; t < seg; t += spec.houseW*0.7 + spec.houseGap*s.r.float() {
				x, y := a.x+(b.x-a.x)*t/seg, a.y+(b.y-a.y)*t/seg
				dist := math.Hypot((x-cx)/1.15, y-cy)
				if dist > spec.radius || !s.r.chance(1.15-dist/spec.radius*0.8) {
					continue
				}
				for _, side := range []float64{-1, 1} {
					w := spec.houseW * s.r.between(0.75, 1.5)
					d := spec.houseW * s.r.between(0.65, 0.95)
					off := st.width/2 + d/2 + 9
					hx, hy := x+nx*side*off, y+ny*side*off
					s.tryHouse(hx, hy, w, d, ang)
					if s.r.chance(spec.second) {
						s.tryHouse(hx+nx*side*(d+5), hy+ny*side*(d+5), w*s.r.between(0.7, 1.1), d*s.r.between(0.8, 1.1), ang)
					}
				}
			}
		}
	}
}

func (s *site) tryHouse(x, y, w, d, ang float64) {
	ang += s.r.between(-0.06, 0.06)
	if !s.freePoly(rotRect(x, y, w+4, d+4, ang), 24) {
		return
	}
	roof := s.pal.roof[s.r.r.Intn(len(s.pal.roof))]
	s.house(x, y, w, d, ang, shade(roof, s.r.between(0.88, 1.1)), s.r.chance(0.55))
	if s.r.chance(0.3) { // огород за домом
		ca, sa := math.Cos(ang), math.Sin(ang)
		g := rotRect(x-sa*(d/2+9), y+ca*(d/2+9), w*0.8, 12, ang)
		if s.freePoly(g, 10) {
			s.field(g, ang)
		}
	}
}

// field – грядки или поле с бороздами вдоль angle и живой изгородью по краю.
func (s *site) field(poly []pt, angle float64) {
	cols := []color.RGBA{rgb(150, 120, 74), rgb(176, 160, 88), rgb(120, 140, 66), rgb(200, 178, 100)}
	if s.pal.snow {
		cols = []color.RGBA{rgb(214, 214, 214), rgb(196, 194, 190)}
	}
	base := cols[s.r.r.Intn(len(cols))]
	ca, sa := math.Cos(angle), math.Sin(angle)
	s.texture([][]pt{poly}, 0.92, func(x, y int) color.RGBA {
		v := -float64(x)*sa + float64(y)*ca
		k := 0.88 + 0.16*math.Abs(math.Sin(v/2.2))
		return shade(base, k*(0.95+0.1*s.n2.fbm(float64(x)/6, float64(y)/6, 2)))
	})
	s.c.stroke(closed(poly), 1.4, rgb(78, 98, 52), 0.55)
	s.occupy([][]pt{poly})
}

// fieldsAround – лоскутные поля за околицей.
func (s *site) fieldsAround(cx, cy, radius float64) {
	for range int(radius / 6) {
		a := s.r.between(0, 2*math.Pi)
		d := s.r.between(radius*0.85, radius*1.6)
		x, y := cx+math.Cos(a)*d*1.2, cy+math.Sin(a)*d
		ang := a + s.r.between(-0.3, 0.3)
		poly := rotRect(x, y, s.r.between(90, 180), s.r.between(60, 110), ang)
		if s.freePoly(poly, 6) {
			s.field(poly, ang)
		}
	}
}

// docks – пристани от берега в воду и лодки у них; место «Пристань».
func (s *site) docks(cx, cy float64, sea []pt, spec townSpec) {
	dir := s.o.WaterDir
	nx, ny := math.Cos(dir), math.Sin(dir)
	x, y := cx, cy
	for range 600 { // от центра к воде, пока не упрёмся в берег
		if pointInPoly(pt{x, y}, sea) {
			break
		}
		x, y = x+nx*4, y+ny*4
	}
	s.road(s.wiggle(pt{cx, cy}, pt{x - nx*30, y - ny*30}, 60), 22, spec.radius >= 500)
	s.house(x-nx*70+ny*60, y-ny*70-nx*60, 70, 40, dir+math.Pi/2, s.pal.roof[len(s.pal.roof)-1], false) // склад
	tx, ty := -ny, nx
	for k := -2; k <= 2; k++ {
		bx, by := x+tx*float64(k)*70-nx*10, y+ty*float64(k)*70-ny*10
		pier := rotRect(bx+nx*60, by+ny*60, 130, 14, dir)
		s.shadow(pier, 4)
		s.texture([][]pt{pier}, 1, func(px, py int) color.RGBA {
			v := float64(px)*nx + float64(py)*ny
			return shade(rgb(140, 104, 66), 0.85+0.2*math.Abs(math.Sin(v/3)))
		})
		s.c.stroke(closed(pier), 1, rgb(60, 40, 24), 0.7)
		s.boat(bx+nx*70+tx*22, by+ny*70+ty*22, dir)
	}
	s.addPlace(x-nx*20, y-ny*20, "port", "Пристань")
}

func pointInPoly(q pt, poly []pt) bool {
	in := false
	for i, j := 0, len(poly)-1; i < len(poly); j, i = i, i+1 {
		a, b := poly[i], poly[j]
		if (a.y > q.y) != (b.y > q.y) && q.x < (b.x-a.x)*(q.y-a.y)/(b.y-a.y)+a.x {
			in = !in
		}
	}
	return in
}

// boat – лодка сверху: корпус, настил, скамьи, тень на воде.
func (s *site) boat(x, y, ang float64) {
	ca, sa := math.Cos(ang), math.Sin(ang)
	at := func(u, v float64) pt { return pt{x + u*ca - v*sa, y + u*sa + v*ca} }
	hull := []pt{at(-22, 0), at(-14, -8), at(14, -8), at(24, 0), at(14, 8), at(-14, 8)}
	s.c.fill([][]pt{offset(hull, 5, 5)}, rgb(10, 30, 40), 0.3)
	s.c.fill([][]pt{hull}, rgb(122, 84, 50), 1)
	s.c.fill([][]pt{{at(-18, 0), at(-12, -5), at(12, -5), at(18, 0), at(12, 5), at(-12, 5)}}, rgb(160, 120, 78), 1)
	for _, u := range []float64{-6, 6} {
		s.c.stroke([]pt{at(u, -5), at(u, 5)}, 2, rgb(100, 70, 40), 1)
	}
	s.c.stroke(closed(hull), 1, rgb(50, 32, 20), 0.8)
}

// townRoad – улица: внутри города мощёная, за его пределами – грунтовка.
func (s *site) townRoad(cx, cy float64, spec townSpec, st street) {
	var inside, outside []pt
	for _, q := range st.line {
		if math.Hypot((q.x-cx)/1.15, q.y-cy) < spec.radius*0.85 {
			inside = append(inside, q)
		} else {
			outside = append(outside, q)
		}
	}
	cobble := spec.walls || spec.radius >= 500
	if len(outside) > 1 && st.main {
		s.road(append([]pt{inside[len(inside)-1]}, outside...), st.width, false)
	}
	if len(inside) > 1 {
		s.road(inside, st.width, cobble && st.main || spec.walls)
	}
}

// clearing – вокруг поселения лес вырублен: клетки помечаются занятыми только для деревьев.
func (s *site) clearing(cx, cy, r float64) {
	s.clear = func(x, y float64) bool { return math.Hypot((x-cx)/1.15, y-cy) < r }
}

// gardens – отдельные деревья и сады между домами.
func (s *site) gardens(cx, cy, radius float64) {
	for range int(radius) {
		a, d := s.r.between(0, 2*math.Pi), s.r.between(0, radius*1.1)
		x, y := cx+math.Cos(a)*d*1.15, cy+math.Sin(a)*d
		if s.busy(x, y) || s.busy(x+8, y+8) || s.busy(x-8, y-8) {
			continue
		}
		col, con := s.treeColor()
		r := s.r.between(8, 14)
		s.tree(x, y, r, col, con)
		s.occupy([][]pt{circle(x, y, r*0.7)})
	}
}
