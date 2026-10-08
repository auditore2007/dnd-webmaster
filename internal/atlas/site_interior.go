package atlas

import (
	"image"
	"image/color"
	"math"
)

// Здание изнутри (таверна, храм, лавка, кузница) – боевая карта с клетками по 5 футов (40 px):
// пол из досок или плит, толстые наружные стены с окнами и дверью, внутренние стены делят дом на комнаты,
// в каждой – своя мебель. Каждая комната – место на карте.

const tile = 40.0

type room struct {
	r    image.Rectangle // в клетках
	name string
	fill func(s *site, r image.Rectangle)
}

func (s *site) interior() {
	s.ground()
	cols, rows := s.W/int(tile), s.H/int(tile)
	var bw, bh int
	switch s.o.Kind {
	case "temple":
		bw, bh = 26, 16
	case "shop":
		bw, bh = 18, 12
	case "smithy":
		bw, bh = 20, 13
	default:
		bw, bh = 24, 16
	}
	x0, y0 := (cols-bw)/2, (rows-bh)/2+1
	b := image.Rect(x0, y0, x0+bw, y0+bh)
	rooms := s.layoutRooms(b)
	stoneFloor := s.o.Kind == "temple" || s.o.Kind == "smithy"
	px := func(r image.Rectangle) []pt {
		return rectPts(image.Rect(r.Min.X*int(tile), r.Min.Y*int(tile), r.Max.X*int(tile), r.Max.Y*int(tile)))
	}
	outer := px(b)
	s.c.fill([][]pt{offset(outer, 14, 14)}, rgb(20, 22, 16), 0.35)
	if stoneFloor {
		s.texture([][]pt{outer}, 1, s.flagstone)
	} else {
		s.texture([][]pt{outer}, 1, s.planks)
	}
	for _, rm := range rooms {
		rm.fill(s, rm.r)
	}
	s.innerWalls(b, rooms)
	s.outerWall(b)
	for _, rm := range rooms {
		c := rm.r.Min.Add(rm.r.Max).Div(2)
		s.addPlace(float64(c.X)*tile, float64(c.Y)*tile, s.o.Kind, rm.name)
	}
	s.cartouche()
}

// layoutRooms – главный зал во всю ширину и ряд подсобных комнат с одной стороны.
func (s *site) layoutRooms(b image.Rectangle) []room {
	type side struct {
		name string
		fill func(*site, image.Rectangle)
	}
	var main room
	var extra []side
	switch s.o.Kind {
	case "temple":
		main = room{name: "Святилище", fill: (*site).nave}
		extra = []side{{"Ризница", (*site).storeroom}, {"Келья", (*site).bedroom}, {"Часовня", (*site).chapel}}
	case "shop":
		main = room{name: "Торговый зал", fill: (*site).shopFloor}
		extra = []side{{"Склад", (*site).storeroom}, {"Комната хозяина", (*site).bedroom}}
	case "smithy":
		main = room{name: "Кузня", fill: (*site).forge}
		extra = []side{{"Склад угля и руды", (*site).storeroom}, {"Жильё кузнеца", (*site).bedroom}}
	default:
		main = room{name: "Общий зал", fill: (*site).taproom}
		extra = []side{{"Кухня", (*site).kitchen}, {"Кладовая", (*site).storeroom}, {"Лестница наверх", (*site).stairs}}
	}
	depth := b.Dy() * 2 / 5
	main.r = image.Rect(b.Min.X, b.Min.Y+depth, b.Max.X, b.Max.Y)
	out := []room{main}
	x := b.Min.X
	for i, e := range extra {
		w := b.Dx() / len(extra)
		if i == len(extra)-1 {
			w = b.Max.X - x
		}
		out = append(out, room{r: image.Rect(x, b.Min.Y, x+w, b.Min.Y+depth), name: e.name, fill: e.fill})
		x += w
	}
	return out
}

func (s *site) planks(x, y int) color.RGBA {
	row := y / 14
	off := int(hash2(float64(row), 3) * 120)
	seam := (x+off)%120 < 2 || y%14 < 1
	base := shade(rgb(150, 108, 68), 0.82+0.3*hash2(float64(row), float64((x+off)/120)))
	base = shade(base, 0.93+0.12*s.n2.fbm(float64(x)/30, float64(y)/2, 2))
	if seam {
		return shade(base, 0.6)
	}
	return base
}

func (s *site) flagstone(x, y int) color.RGBA {
	const t = tile
	ix, iy := math.Floor(float64(x)/t), math.Floor(float64(y)/t)
	fx, fy := float64(x)-ix*t, float64(y)-iy*t
	base := shade(rgb(150, 144, 134), 0.85+0.25*hash2(ix, iy))
	base = shade(base, 0.94+0.12*s.n2.fbm(float64(x)/7, float64(y)/7, 2))
	if fx < 2 || fy < 2 {
		return shade(base, 0.62)
	}
	return base
}

func (s *site) outerWall(b image.Rectangle) {
	r := image.Rect(b.Min.X*int(tile), b.Min.Y*int(tile), b.Max.X*int(tile), b.Max.Y*int(tile))
	poly := closed(rectPts(r))
	s.texture(strokePolys(poly, []float64{18}), 1, func(x, y int) color.RGBA {
		return shade(rgb(120, 112, 104), 0.82+0.3*s.n2.fbm(float64(x)/5, float64(y)/5, 2))
	})
	s.c.stroke(offset(poly, -3, -3), 2, rgb(176, 168, 158), 0.6)
	// окна вдоль длинных стен и дверь посередине нижней
	for x := r.Min.X + 80; x < r.Max.X-60; x += 160 {
		for _, y := range []int{r.Min.Y, r.Max.Y} {
			s.c.fill([][]pt{rectPts(image.Rect(x, y-4, x+40, y+4))}, rgb(170, 210, 230), 1)
		}
	}
	mid := (r.Min.X + r.Max.X) / 2
	s.c.fill([][]pt{rectPts(image.Rect(mid-30, r.Max.Y-10, mid+30, r.Max.Y+10))}, rgb(150, 108, 68), 1)
	s.c.fill([][]pt{rectPts(image.Rect(mid-30, r.Max.Y-5, mid+30, r.Max.Y+5))}, rgb(98, 66, 38), 1)
	s.c.stroke(closed(rectPts(image.Rect(mid-30, r.Max.Y-5, mid+30, r.Max.Y+5))), 1.2, rgb(40, 26, 16), 0.8)
}

// innerWalls – стены между комнатами с проёмом-дверью в каждой.
func (s *site) innerWalls(b image.Rectangle, rooms []room) {
	wall := func(a, c pt) {
		s.c.stroke([]pt{a, c}, 10, rgb(112, 104, 96), 1)
		s.c.stroke([]pt{a, c}, 1, rgb(60, 54, 50), 0.6)
	}
	split := float64(rooms[0].r.Min.Y) * tile
	for _, rm := range rooms[1:] {
		x0, x1 := float64(rm.r.Min.X)*tile, float64(rm.r.Max.X)*tile
		mid := (x0 + x1) / 2
		wall(pt{x0, split}, pt{mid - 24, split})
		wall(pt{mid + 24, split}, pt{x1, split})
		if rm.r.Min.X > b.Min.X {
			wall(pt{x0, float64(rm.r.Min.Y) * tile}, pt{x0, split})
		}
	}
}

// ---------- мебель ----------

func (s *site) px(r image.Rectangle) (x0, y0, x1, y1 float64) {
	return float64(r.Min.X) * tile, float64(r.Min.Y) * tile, float64(r.Max.X) * tile, float64(r.Max.Y) * tile
}

// box – предмет мебели: тень, заливка, светлая кромка и контур.
func (s *site) box(r image.Rectangle, col color.RGBA) {
	p := rectPts(r)
	s.c.fill([][]pt{offset(p, 5, 5)}, rgb(10, 8, 6), 0.3)
	s.c.fill([][]pt{p}, col, 1)
	s.c.stroke([]pt{p[3], p[0], p[1]}, 2, shade(col, 1.3), 0.7)
	s.c.stroke(closed(p), 1.2, rgb(40, 28, 20), 0.85)
}

func (s *site) roundTable(x, y, r float64) {
	for k := range 4 {
		a := float64(k)*math.Pi/2 + s.r.between(0, 0.5)
		cx, cy := x+math.Cos(a)*(r+12), y+math.Sin(a)*(r+12)
		s.box(image.Rect(int(cx-7), int(cy-7), int(cx+7), int(cy+7)), rgb(120, 82, 50))
	}
	s.c.fill([][]pt{offset(circle(x, y, r), 5, 5)}, rgb(10, 8, 6), 0.3)
	s.c.fill([][]pt{circle(x, y, r)}, rgb(134, 92, 56), 1)
	s.c.fill([][]pt{circle(x-r*0.25, y-r*0.25, r*0.5)}, rgb(156, 112, 70), 0.6)
	s.c.stroke(closed(circle(x, y, r)), 1.2, rgb(40, 28, 20), 0.85)
	if s.r.chance(0.6) { // кружки
		s.c.fill([][]pt{circle(x+s.r.between(-8, 8), y+s.r.between(-8, 8), 3.5)}, rgb(200, 190, 170), 1)
	}
}

func (s *site) barrel(x, y, r float64) {
	s.c.fill([][]pt{offset(circle(x, y, r), 4, 4)}, rgb(10, 8, 6), 0.3)
	s.c.fill([][]pt{circle(x, y, r)}, rgb(118, 80, 46), 1)
	s.c.stroke(closed(circle(x, y, r*0.75)), 1.4, rgb(70, 66, 64), 0.9)
	s.c.stroke(closed(circle(x, y, r)), 1.2, rgb(40, 28, 20), 0.9)
}

func (s *site) hearth(x, y float64, w float64) {
	s.box(image.Rect(int(x-w/2), int(y-22), int(x+w/2), int(y+22)), rgb(120, 112, 104))
	for k := range 4 { // жар углей
		r := 16 - float64(k)*3.5
		s.c.fill([][]pt{ellipse(x, y, r*1.4, r)}, mix(rgb(120, 30, 10), rgb(255, 210, 90), float64(k)/3), 0.85)
	}
	for range 3 { // отблеск на полу
		s.c.fill([][]pt{ellipse(x, y+40, 60, 30)}, rgb(255, 170, 60), 0.05)
	}
}

func (s *site) taproom(r image.Rectangle) {
	x0, y0, x1, y1 := s.px(r)
	// стойка буквой Г у правой стены
	s.box(image.Rect(int(x1-200), int(y0+30), int(x1-170), int(y1-80)), rgb(110, 74, 44))
	s.box(image.Rect(int(x1-200), int(y1-110), int(x1-40), int(y1-80)), rgb(110, 74, 44))
	for y := y0 + 50; y < y1-120; y += 36 {
		s.c.fill([][]pt{circle(x1-220, y, 7)}, rgb(130, 90, 56), 1)
	}
	for y := y0 + 40; y < y1-130; y += 40 {
		s.barrel(x1-80, y, 16)
	}
	s.hearth(x0+120, y1-40, 90)
	for k := range 6 {
		cx := x0 + 120 + float64(k%3)*150
		cy := y0 + 110 + float64(k/3)*170
		if cx < x1-260 {
			s.roundTable(cx, cy, 24)
		}
	}
}

func (s *site) kitchen(r image.Rectangle) {
	x0, y0, x1, y1 := s.px(r)
	s.hearth((x0+x1)/2, y0+40, 120)
	s.box(image.Rect(int(x0+40), int(y1-110), int(x1-40), int(y1-70)), rgb(150, 116, 76))
	for x := x0 + 60; x < x1-60; x += 50 {
		s.c.fill([][]pt{circle(x, y1-90, 8)}, rgb(180, 170, 150), 1)
	}
}

func (s *site) storeroom(r image.Rectangle) {
	x0, y0, x1, y1 := s.px(r)
	for k := range 9 {
		x := x0 + 40 + s.r.between(0, x1-x0-80)
		y := y0 + 40 + s.r.between(0, y1-y0-120)
		if k%2 == 0 {
			s.barrel(x, y, 16)
		} else {
			s.box(image.Rect(int(x-16), int(y-16), int(x+16), int(y+16)), rgb(150, 112, 70))
			s.c.stroke([]pt{{x - 16, y - 16}, {x + 16, y + 16}}, 1.2, rgb(80, 56, 34), 0.8)
		}
	}
}

func (s *site) stairs(r image.Rectangle) {
	x0, y0, x1, y1 := s.px(r)
	sx := (x0 + x1) / 2
	for k := range 8 {
		y := y0 + 30 + float64(k)*16
		if y > y1-60 {
			break
		}
		s.box(image.Rect(int(sx-40), int(y), int(sx+40), int(y+14)), shade(rgb(140, 100, 62), 1-float64(k)*0.05))
	}
}

func (s *site) bedroom(r image.Rectangle) {
	x0, y0, _, y1 := s.px(r)
	s.box(image.Rect(int(x0+30), int(y0+30), int(x0+110), int(y0+150)), rgb(120, 82, 50))
	s.box(image.Rect(int(x0+36), int(y0+60), int(x0+104), int(y0+144)), rgb(150, 60, 50))
	s.box(image.Rect(int(x0+40), int(y0+36), int(x0+100), int(y0+58)), rgb(230, 224, 210))
	s.box(image.Rect(int(x0+140), int(y1-100), int(x0+220), int(y1-70)), rgb(130, 90, 56))
}

func (s *site) nave(r image.Rectangle) {
	x0, y0, x1, y1 := s.px(r)
	// алтарь на возвышении у левой стены, скамьи рядами, колонны по краям
	s.box(image.Rect(int(x0+30), int(y0+60), int(x0+130), int(y1-60)), rgb(170, 160, 146))
	s.box(image.Rect(int(x0+60), int((y0+y1)/2-30), int(x0+100), int((y0+y1)/2+30)), rgb(210, 200, 180))
	s.c.fill([][]pt{circle(x0+80, (y0+y1)/2, 8)}, rgb(230, 190, 80), 1)
	for x := x0 + 200; x < x1-80; x += 70 {
		for _, yy := range []float64{y0 + 70, (y0+y1)/2 + 20} {
			s.box(image.Rect(int(x), int(yy), int(x+24), int(yy+((y1-y0)/2-90))), rgb(124, 86, 54))
		}
	}
	for x := x0 + 180; x < x1-40; x += 140 {
		for _, yy := range []float64{y0 + 36, y1 - 36} {
			s.c.fill([][]pt{offset(circle(x, yy, 16), 4, 4)}, rgb(10, 8, 6), 0.3)
			s.c.fill([][]pt{circle(x, yy, 16)}, rgb(180, 172, 160), 1)
			s.c.fill([][]pt{circle(x-4, yy-4, 8)}, rgb(214, 206, 194), 1)
		}
	}
}

func (s *site) chapel(r image.Rectangle) {
	x0, y0, x1, y1 := s.px(r)
	cx, cy := (x0+x1)/2, (y0+y1)/2
	s.box(image.Rect(int(cx-30), int(cy-20), int(cx+30), int(cy+20)), rgb(196, 186, 170))
	for k := range 5 {
		s.c.fill([][]pt{circle(cx-40+float64(k)*20, cy-34, 4)}, rgb(255, 220, 120), 0.9)
	}
}

func (s *site) shopFloor(r image.Rectangle) {
	x0, y0, x1, y1 := s.px(r)
	s.box(image.Rect(int(x0+60), int((y0+y1)/2-20), int(x1-200), int((y0+y1)/2+20)), rgb(120, 82, 50))
	cols := []color.RGBA{rgb(180, 60, 50), rgb(60, 100, 170), rgb(210, 180, 70), rgb(90, 150, 80), rgb(200, 200, 190)}
	for x := x0 + 30; x < x1-60; x += 90 {
		s.box(image.Rect(int(x), int(y0+20), int(x+80), int(y0+44)), rgb(110, 76, 46))
		for k := range 5 {
			s.c.fill([][]pt{circle(x+10+float64(k)*15, y0+32, 4)}, cols[(k+int(x))%len(cols)], 1)
		}
	}
	for range 4 {
		x, y := x1-150+s.r.between(0, 100), y0+80+s.r.between(0, y1-y0-160)
		s.box(image.Rect(int(x-18), int(y-18), int(x+18), int(y+18)), rgb(150, 112, 70))
	}
}

func (s *site) forge(r image.Rectangle) {
	x0, y0, x1, y1 := s.px(r)
	s.hearth(x0+120, (y0+y1)/2, 110)
	ax, ay := x0+300, (y0+y1)/2
	anvil := []pt{{ax - 30, ay - 10}, {ax + 24, ay - 10}, {ax + 34, ay}, {ax + 24, ay + 10}, {ax - 30, ay + 10}}
	s.c.fill([][]pt{offset(anvil, 4, 4)}, rgb(10, 8, 6), 0.35)
	s.c.fill([][]pt{anvil}, rgb(70, 72, 78), 1)
	s.c.stroke([]pt{anvil[0], anvil[1]}, 2, rgb(150, 154, 160), 0.8)
	s.box(image.Rect(int(ax-60), int(y1-90), int(ax+60), int(y1-50)), rgb(150, 116, 76))
	trough := image.Rect(int(x0+60), int(y0+30), int(x0+180), int(y0+70))
	s.box(trough, rgb(110, 80, 50))
	s.c.fill([][]pt{rectPts(trough.Inset(6))}, rgb(60, 110, 140), 1)
	for x := x1 - 200; x < x1-40; x += 40 { // стойка с оружием
		s.c.stroke([]pt{{x, y0 + 30}, {x + 10, y0 + 110}}, 3, rgb(170, 174, 180), 1)
	}
	for range 40 { // кучка угля
		s.c.fill([][]pt{circle(x1-120+s.r.between(-40, 40), y1-80+s.r.between(-25, 25), s.r.between(3, 6))}, rgb(30, 28, 28), 1)
	}
}
