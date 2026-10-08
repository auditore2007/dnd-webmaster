package atlas

import (
	"image/color"
	"math"
)

// Места под открытым небом: поляна в лесу и то, ради чего сюда идут, – руины, лагерь, каменный круг,
// башня; у озера – вода, у болота – топи и камыш, в горах – скалы. Подходы – тропы от краёв карты.

func (s *site) outdoor() {
	s.ground()
	cx, cy := float64(s.W)/2, float64(s.H)/2+30
	R := 380.0
	switch s.o.Kind {
	case "lake":
		s.lake(cx, cy, 300)
	case "swamp":
		s.swamp(cx, cy, R)
	case "mountain":
		s.crags(cx, cy, R)
	}
	for range 2 + s.r.r.Intn(2) {
		a := s.r.between(0, 2*math.Pi)
		far := pt{cx + math.Cos(a)*2000, cy + math.Sin(a)*2000}
		s.road(s.wiggle(pt{cx, cy}, far, 260), 16, false)
	}
	switch s.o.Kind {
	case "ruins":
		s.ruins(cx, cy)
	case "camp":
		s.camp(cx, cy)
	case "shrine":
		s.stoneCircle(cx, cy)
	case "tower":
		s.wizardTower(cx, cy)
	case "dungeon":
		s.ruins(cx, cy)
	default:
		s.addPlace(cx, cy, s.o.Kind, KindTitle(s.o.Kind))
	}
	clear := R
	if s.o.Kind == "forest" {
		clear = 160
	}
	s.clearing(cx, cy, clear)
	s.forestFill(0.9, 0.42)
	s.forestFill(0.5, 0.3)
	s.bushes(cx, cy, R)
	s.vignette()
	s.cartouche()
}

// KindTitle – подпись места по умолчанию для вида, у которого нет своих построек.
func KindTitle(kind string) string {
	return map[string]string{"forest": "Лесная поляна", "mountain": "Горный перевал", "swamp": "Топь", "lake": "Озеро", "other": "Место"}[kind]
}

func (s *site) lake(cx, cy, r float64) {
	var poly []pt
	seed := s.r.between(0, 50)
	for k := range 36 {
		a := float64(k) * 2 * math.Pi / 36
		rr := r * (0.7 + 0.6*s.n1.fbm(seed+math.Cos(a), seed+math.Sin(a), 3))
		poly = append(poly, pt{cx + math.Cos(a)*rr*1.4, cy + math.Sin(a)*rr})
	}
	s.water([][]pt{chaikin(append(poly, poly[0]), 2)})
	for range 40 {
		a := s.r.between(0, 2*math.Pi)
		s.reeds(cx+math.Cos(a)*r*1.3, cy+math.Sin(a)*r*0.95)
	}
}

func (s *site) swamp(cx, cy, r float64) {
	for range 18 {
		a, d := s.r.between(0, 2*math.Pi), s.r.between(0, r*1.6)
		x, y := cx+math.Cos(a)*d*1.3, cy+math.Sin(a)*d
		var poly []pt
		rr := s.r.between(40, 110)
		for k := range 16 {
			b := float64(k) * 2 * math.Pi / 16
			q := rr * (0.6 + 0.8*s.n2.fbm(x/50+math.Cos(b), y/50+math.Sin(b), 2))
			poly = append(poly, pt{x + math.Cos(b)*q*1.3, y + math.Sin(b)*q})
		}
		if s.freePoly(poly, 0) {
			s.texture([][]pt{chaikin(append(poly, poly[0]), 2)}, 0.9, func(px, py int) color.RGBA {
				return shade(rgb(70, 92, 70), 0.9+0.2*s.n1.fbm(float64(px)/9, float64(py)/9, 2))
			})
			s.occupy([][]pt{poly})
		}
	}
	for range 160 {
		s.reeds(s.r.between(0, float64(s.W)), s.r.between(0, float64(s.H)))
	}
}

func (s *site) reeds(x, y float64) {
	for range 6 {
		a := s.r.between(-2.4, -0.7)
		l := s.r.between(8, 16)
		s.c.stroke([]pt{{x, y}, {x + math.Cos(a)*l, y + math.Sin(a)*l}}, 1.4, rgb(110, 128, 64), 0.9)
	}
}

func (s *site) crags(cx, cy, r float64) {
	for range 26 {
		a, d := s.r.between(0, 2*math.Pi), s.r.between(r*0.6, r*2)
		x, y := cx+math.Cos(a)*d*1.3, cy+math.Sin(a)*d
		var poly []pt
		rr := s.r.between(40, 120)
		for k := range 9 {
			b := float64(k)*2*math.Pi/9 + s.r.between(-0.2, 0.2)
			q := rr * s.r.between(0.6, 1.1)
			poly = append(poly, pt{x + math.Cos(b)*q, y + math.Sin(b)*q})
		}
		if !s.freePoly(poly, 0) {
			continue
		}
		s.shadow(poly, rr*0.5)
		s.texture([][]pt{poly}, 1, func(px, py int) color.RGBA {
			dx, dy := float64(px)-x, float64(py)-y
			k := 1 + (-dx-dy)/(rr*3)
			return shade(rgb(136, 128, 118), k*(0.88+0.24*s.n2.fbm(float64(px)/6, float64(py)/14, 3)))
		})
		s.c.stroke(closed(poly), 1.4, rgb(50, 44, 40), 0.7)
		s.occupy([][]pt{poly})
	}
}

// bushes – кусты и камни по поляне.
func (s *site) bushes(cx, cy, r float64) {
	for range 120 {
		a, d := s.r.between(0, 2*math.Pi), s.r.between(r*0.3, r*1.2)
		x, y := cx+math.Cos(a)*d*1.3, cy+math.Sin(a)*d
		if s.busy(x, y) {
			continue
		}
		if s.r.chance(0.3) {
			s.rock(x, y, s.r.between(4, 9))
			continue
		}
		col, _ := s.treeColor()
		s.tree(x, y, s.r.between(5, 9), shade(col, 1.1), false)
		s.occupy([][]pt{circle(x, y, 6)})
	}
}

// ruins – остатки построек: стены с проломами, упавшие колонны, обломки; лестница вниз.
func (s *site) ruins(cx, cy float64) {
	stone := rgb(160, 152, 140)
	wallSeg := func(a, b pt) {
		s.shadowLine([]pt{a, b}, 16, 10)
		s.texture(strokePolys([]pt{a, b}, []float64{16}), 1, func(x, y int) color.RGBA {
			return shade(stone, 0.8+0.3*s.n2.fbm(float64(x)/4, float64(y)/4, 2))
		})
		s.c.stroke([]pt{a, b}, 1, rgb(60, 54, 48), 0.6)
	}
	names := []string{"Главный зал", "Обрушенная башня", "Двор", "Часовня", "Казарма"}
	for k := range 4 + s.r.r.Intn(2) {
		x, y := cx+s.r.between(-420, 420), cy+s.r.between(-260, 260)
		w, h := s.r.between(140, 260), s.r.between(110, 200)
		ang := s.r.between(-0.2, 0.2)
		poly := rotRect(x, y, w, h, ang)
		s.texture([][]pt{poly}, 0.7, s.flagstone)
		for i := range 4 {
			a, b := poly[i], poly[(i+1)%4]
			t0 := 0.0
			for t0 < 1 { // стена кусками с проломами
				t1 := math.Min(1, t0+s.r.between(0.15, 0.5))
				if s.r.chance(0.7) {
					wallSeg(pt{a.x + (b.x-a.x)*t0, a.y + (b.y-a.y)*t0}, pt{a.x + (b.x-a.x)*t1, a.y + (b.y-a.y)*t1})
				}
				t0 = t1 + s.r.between(0.05, 0.2)
			}
		}
		s.occupy([][]pt{poly})
		if k < len(names) {
			s.addPlace(x, y, "ruins", names[k])
		}
	}
	for range 14 { // колонны – стоят или лежат
		x, y := cx+s.r.between(-320, 320), cy+s.r.between(-220, 220)
		if s.r.chance(0.5) {
			s.c.fill([][]pt{offset(circle(x, y, 13), 9, 9)}, rgb(20, 20, 16), 0.3)
			s.c.fill([][]pt{circle(x, y, 13)}, stone, 1)
			s.c.fill([][]pt{circle(x-3, y-3, 7)}, shade(stone, 1.2), 1)
		} else {
			a := s.r.between(0, math.Pi)
			s.box2(rotRect(x, y, 80, 22, a), stone)
		}
	}
	for range 80 {
		x, y := cx+s.r.between(-360, 360), cy+s.r.between(-240, 240)
		s.rock(x, y, s.r.between(3, 8))
	}
	sx, sy := cx+s.r.between(-100, 100), cy+s.r.between(-60, 60)
	for k := range 6 { // ступени вниз, к подземелью
		w := 70 - float64(k)*6
		s.c.fill([][]pt{rectPts2(sx-w/2, sy+float64(k)*10, sx+w/2, sy+float64(k)*10+10)}, shade(stone, 0.9-float64(k)*0.12), 1)
	}
	s.addPlace(sx, sy+30, "dungeon", "Лестница вниз")
}

func rectPts2(x0, y0, x1, y1 float64) []pt { return []pt{{x0, y0}, {x1, y0}, {x1, y1}, {x0, y1}} }

func (s *site) box2(poly []pt, col color.RGBA) {
	s.c.fill([][]pt{offset(poly, 6, 6)}, rgb(20, 20, 16), 0.3)
	s.c.fill([][]pt{poly}, col, 1)
	s.c.stroke(closed(poly), 1.2, rgb(50, 44, 40), 0.7)
}

// camp – шатры по кругу, костёр в центре, частокол у военного лагеря.
func (s *site) camp(cx, cy float64) {
	s.texture([][]pt{ellipse(cx, cy, 300, 220)}, 0.6, func(x, y int) color.RGBA {
		return shade(s.pal.dirt[0], 0.9+0.2*s.n2.fbm(float64(x)/6, float64(y)/6, 2))
	})
	if s.r.chance(0.6) { // частокол
		var ring []pt
		for k := 0; k <= 60; k++ {
			a := float64(k) * 2 * math.Pi / 60
			ring = append(ring, pt{cx + math.Cos(a)*330, cy + math.Sin(a)*250})
		}
		for i, q := range ring {
			if i%10 == 0 {
				continue // проходы
			}
			s.c.fill([][]pt{offset(circle(q.x, q.y, 6), 5, 5)}, rgb(20, 20, 16), 0.3)
			s.c.fill([][]pt{circle(q.x, q.y, 6)}, rgb(120, 84, 50), 1)
			s.c.fill([][]pt{circle(q.x-2, q.y-2, 3)}, rgb(170, 130, 86), 1)
		}
	}
	s.c.fill([][]pt{circle(cx, cy, 40)}, rgb(255, 170, 60), 0.12)
	s.c.fill([][]pt{circle(cx, cy, 20)}, rgb(90, 86, 80), 1)
	s.c.fill([][]pt{circle(cx, cy, 13)}, rgb(255, 150, 40), 1)
	s.c.fill([][]pt{circle(cx, cy, 6)}, rgb(255, 230, 140), 1)
	s.addPlace(cx, cy+30, "camp", "Костёр")
	cols := []color.RGBA{rgb(170, 60, 50), rgb(196, 176, 130), rgb(110, 120, 80), rgb(150, 120, 90)}
	for k := range 9 {
		a := float64(k)*2*math.Pi/9 + s.r.between(-0.2, 0.2)
		x, y := cx+math.Cos(a)*200, cy+math.Sin(a)*150
		r := s.r.between(30, 46)
		col := cols[s.r.r.Intn(len(cols))]
		if k == 0 {
			r = 62
		}
		s.tent(x, y, r, col)
		if k == 0 {
			s.addPlace(x, y, "camp", "Шатёр вожака")
		}
	}
	for range 12 {
		x, y := cx+s.r.between(-240, 240), cy+s.r.between(-170, 170)
		if !s.busy(x, y) {
			s.barrel(x, y, 12)
		}
	}
}

// tent – круглый шатёр сверху: клинья ткани с разной освещённостью, шест в центре.
func (s *site) tent(x, y, r float64, col color.RGBA) {
	s.c.fill([][]pt{offset(circle(x, y, r), r*0.4, r*0.4)}, rgb(20, 20, 16), 0.3)
	n := 8
	for k := range n {
		a0, a1 := float64(k)*2*math.Pi/float64(n), float64(k+1)*2*math.Pi/float64(n)
		am := (a0 + a1) / 2
		s.c.fill([][]pt{{{x, y}, {x + math.Cos(a0)*r, y + math.Sin(a0)*r}, {x + math.Cos(a1)*r, y + math.Sin(a1)*r}}}, shade(col, light(math.Cos(am), math.Sin(am))), 1)
	}
	s.c.stroke(closed(circle(x, y, r)), 1.2, rgb(50, 36, 24), 0.8)
	s.c.fill([][]pt{circle(x, y, 3)}, rgb(80, 56, 34), 1)
	s.occupy([][]pt{circle(x, y, r)})
}

// stoneCircle – каменный круг с алтарём.
func (s *site) stoneCircle(cx, cy float64) {
	s.texture([][]pt{ellipse(cx, cy, 190, 150)}, 0.5, func(x, y int) color.RGBA { return shade(s.pal.grass[0], 1.15) })
	for k := range 12 {
		a := float64(k) * 2 * math.Pi / 12
		x, y := cx+math.Cos(a)*160, cy+math.Sin(a)*120
		s.box2(rotRect(x, y, 34, 18, a), rgb(150, 146, 140))
	}
	s.box2(rotRect(cx, cy, 70, 40, 0), rgb(170, 164, 154))
	s.c.fill([][]pt{circle(cx, cy, 30)}, rgb(120, 200, 255), 0.15)
	s.addPlace(cx, cy, "shrine", "Алтарь")
}

// wizardTower – высокая круглая башня с пристройкой и садиком.
func (s *site) wizardTower(cx, cy float64) {
	s.texture([][]pt{ellipse(cx, cy, 170, 130)}, 1, s.cobble)
	s.c.fill([][]pt{offset(circle(cx, cy, 90), 70, 70)}, rgb(20, 20, 16), 0.32)
	s.tower(cx, cy, 90, s.pal.stone[0])
	s.house(cx+150, cy+40, 90, 56, 0.3, rgb(84, 70, 120), true)
	s.addPlace(cx, cy, "tower", "Башня")
}
