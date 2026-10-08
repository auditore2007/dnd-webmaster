package atlas

import (
	"image/color"
	"math"
)

// Замок сверху: холм, ров с водой и подъёмный мост, зубчатая стена неправильным многоугольником с круглыми
// башнями, надвратные башни; во дворе донжон, казармы, часовня, кузница, конюшни, колодец и плац;
// снаружи – дорога, хутор, поля и лес.

func (s *site) castle() {
	s.ground()
	cx, cy := float64(s.W)/2+s.r.between(-60, 60), float64(s.H)/2+30
	R := 330.0
	s.hill(cx, cy, R*1.6)
	n := 6 + s.r.r.Intn(3)
	ring := make([]pt, n)
	for k := range n {
		a := float64(k)*2*math.Pi/float64(n) + s.r.between(-0.15, 0.15)
		rr := R * s.r.between(0.85, 1.1)
		ring[k] = pt{cx + math.Cos(a)*rr*1.2, cy + math.Sin(a)*rr}
	}
	gateA := s.r.between(0, 2*math.Pi)
	gate := pt{cx + math.Cos(gateA)*R*1.25, cy + math.Sin(gateA)*R}
	far := pt{cx + math.Cos(gateA)*3000, cy + math.Sin(gateA)*3000}
	s.road(s.wiggle(gate, far, 200), 26, false)
	if s.o.Biome != "desert" && s.r.chance(0.7) {
		var moat []pt
		for k := 0; k <= 40; k++ {
			a := float64(k) * 2 * math.Pi / 40
			moat = append(moat, pt{cx + math.Cos(a)*(R+90)*1.2, cy + math.Sin(a)*(R+90)})
		}
		s.water(strokePolys(chaikin(moat, 1), []float64{46}))
		mid := pt{cx + math.Cos(gateA)*(R+90)*1.2, cy + math.Sin(gateA)*(R+90)}
		bridge := rotRect(mid.x, mid.y, 90, 30, gateA)
		s.shadow(bridge, 6)
		s.texture([][]pt{bridge}, 1, func(x, y int) color.RGBA {
			v := float64(x)*math.Cos(gateA) + float64(y)*math.Sin(gateA)
			return shade(rgb(140, 104, 66), 0.82+0.22*math.Abs(math.Sin(v/3)))
		})
		s.c.stroke(closed(bridge), 1.2, rgb(60, 40, 24), 0.8)
	}
	s.texture([][]pt{ring}, 1, func(x, y int) color.RGBA {
		return shade(s.pal.dirt[0], 0.9+0.18*s.n2.fbm(float64(x)/5, float64(y)/5, 2))
	})
	keepAt := pt{cx - math.Cos(gateA)*R*0.35*1.2, cy - math.Sin(gateA)*R*0.35}
	s.road([]pt{gate, {(gate.x + keepAt.x) / 2, (gate.y + keepAt.y) / 2}, keepAt}, 30, true)
	s.castleBuildings(cx, cy, R, ring, gateA)
	s.curtain(ring, gate)
	s.outskirts(cx, cy, R*1.9, gateA)
	s.clearing(cx, cy, R*1.5)
	s.forestFill(0.85, 0.5)
	s.vignette()
	s.cartouche()
}

// hill – мягкий холм: освещённый северо-западный склон, тень на юго-востоке.
func (s *site) hill(cx, cy, r float64) {
	img := s.c.img
	parallelRows(s.H, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := range s.W {
				dx, dy := (float64(x)-cx)/r, (float64(y)-cy)/r
				d := math.Hypot(dx, dy)
				if d >= 1 {
					continue
				}
				k := 1 + (-dx-dy)*0.12*(1-d)
				o := img.PixOffset(x, y)
				for c := range 3 {
					img.Pix[o+c] = uint8(clamp(float64(img.Pix[o+c])*k, 0, 255))
				}
			}
		}
	})
}

// curtain – стена по кольцу с проёмом у ворот, башни на углах, надвратные башни.
func (s *site) curtain(ring []pt, gate pt) {
	stone := s.pal.stone[0]
	n := len(ring)
	for k := range n {
		a, b := ring[k], ring[(k+1)%n]
		var seg []pt
		steps := int(math.Hypot(b.x-a.x, b.y-a.y) / 5)
		flush := func() {
			if len(seg) < 2 {
				seg = nil
				return
			}
			s.shadowLine(seg, 26, 22)
			s.texture(strokePolys(seg, []float64{26}), 1, func(x, y int) color.RGBA {
				return shade(stone, 0.82+0.26*s.n2.fbm(float64(x)/4, float64(y)/4, 2))
			})
			for _, d := range dashes(seg, 6, 5) { // зубцы
				s.c.stroke(d, 6, shade(stone, 1.2), 0.9)
			}
			s.c.stroke(seg, 1, shade(stone, 0.55), 0.6)
			seg = nil
		}
		for i := 0; i <= steps; i++ {
			t := float64(i) / float64(max(1, steps))
			q := pt{a.x + (b.x-a.x)*t, a.y + (b.y-a.y)*t}
			if math.Hypot(q.x-gate.x, q.y-gate.y) < 60 {
				flush()
				continue
			}
			seg = append(seg, q)
		}
		flush()
	}
	for _, q := range ring {
		s.tower(q.x, q.y, 34, stone)
	}
	// надвратные башни по бокам проёма
	best, bd := 0, math.Inf(1)
	for k := range n {
		a, b := ring[k], ring[(k+1)%n]
		m := pt{(a.x + b.x) / 2, (a.y + b.y) / 2}
		if d := math.Hypot(m.x-gate.x, m.y-gate.y); d < bd {
			best, bd = k, d
		}
	}
	a, b := ring[best], ring[(best+1)%n]
	l := math.Hypot(b.x-a.x, b.y-a.y)
	ux, uy := (b.x-a.x)/l, (b.y-a.y)/l
	// точка стены, ближайшая к воротам
	t := clamp((gate.x-a.x)*ux+(gate.y-a.y)*uy, 70, l-70)
	gx, gy := a.x+ux*t, a.y+uy*t
	for _, sgn := range []float64{-1, 1} {
		s.tower(gx+ux*sgn*62, gy+uy*sgn*62, 26, stone)
	}
}

// castleBuildings – донжон напротив ворот и постройки вдоль стен.
func (s *site) castleBuildings(cx, cy, R float64, ring []pt, gateA float64) {
	slate := shade(s.pal.stone[0], 0.62)
	kx, ky := cx-math.Cos(gateA)*R*0.35*1.2, cy-math.Sin(gateA)*R*0.35
	ang := gateA + s.r.between(-0.2, 0.2)
	keep := rotRect(kx, ky, 190, 150, ang)
	s.shadow(keep, 30)
	s.house(kx, ky, 170, 130, ang, slate, true)
	for _, q := range keep {
		s.tower(q.x, q.y, 26, s.pal.stone[0])
	}
	s.occupy([][]pt{rotRect(kx, ky, 230, 190, ang)})
	s.addPlace(kx, ky, "castle", "Донжон")
	s.addPlace(kx+math.Cos(ang)*70, ky+math.Sin(ang)*70, "dungeon", "Темница")
	// плац и колодец посреди двора
	yard := rotRect(cx+math.Cos(gateA)*R*0.25, cy+math.Sin(gateA)*R*0.2, 150, 100, gateA)
	if s.freePoly(yard, 10) {
		s.texture([][]pt{yard}, 0.9, func(x, y int) color.RGBA {
			return shade(rgb(196, 176, 132), 0.9+0.16*s.n2.fbm(float64(x)/3, float64(y)/3, 2))
		})
		s.c.stroke(closed(yard), 3, rgb(120, 90, 60), 0.6)
		s.occupy([][]pt{yard})
	}
	named := []struct{ kind, name string }{{"temple", "Часовня"}, {"smithy", "Кузница"}, {"castle", "Казармы"}, {"castle", "Конюшни"}, {"tavern", "Пиршественный зал"}}
	ni := 0
	n := len(ring)
	for k := range n {
		a, b := ring[k], ring[(k+1)%n]
		l := math.Hypot(b.x-a.x, b.y-a.y)
		ang := math.Atan2(b.y-a.y, b.x-a.x)
		inx, iny := cx-(a.x+b.x)/2, cy-(a.y+b.y)/2
		il := math.Hypot(inx, iny)
		inx, iny = inx/il, iny/il
		for t := 70.0; t < l-70; t += 100 {
			w, d := s.r.between(70, 110), s.r.between(36, 48)
			x, y := a.x+(b.x-a.x)*t/l+inx*(d/2+22), a.y+(b.y-a.y)*t/l+iny*(d/2+22)
			if !s.freePoly(rotRect(x, y, w+6, d+6, ang), 20) {
				continue
			}
			s.house(x, y, w, d, ang, s.pal.roof[s.r.r.Intn(len(s.pal.roof))], s.r.chance(0.5))
			if ni < len(named) {
				s.addPlace(x, y, named[ni].kind, named[ni].name)
				ni++
			}
		}
	}
	s.c.fill([][]pt{circle(cx, cy, 11)}, shade(s.pal.stone[0], 0.9), 1)
	s.c.fill([][]pt{circle(cx, cy, 6)}, rgb(40, 80, 110), 1)
}

// outskirts – хутор у дороги, поля вокруг.
func (s *site) outskirts(cx, cy, r, gateA float64) {
	for k := range 6 {
		d := r + float64(k)*45
		x, y := cx+math.Cos(gateA)*d*1.2+s.r.between(-60, 60), cy+math.Sin(gateA)*d+s.r.between(-60, 60)
		s.tryHouse(x, y, s.r.between(40, 60), s.r.between(30, 38), gateA+math.Pi/2)
	}
	s.fieldsAround(cx, cy, r*0.85)
}
