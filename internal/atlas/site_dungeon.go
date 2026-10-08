package atlas

import (
	"image"
	"image/color"
	"math"
)

// Подземелье – боевая карта с клетками по 5 футов (40 px): комнаты разного назначения, коридоры с изломом,
// двери на входах. Пол – плиты с трещинами и мхом, стены – кладка с освещённым краем, толща скалы тонет во мраке;
// свет дают факелы на стенах и жаровни: тёплые круги света в темноте. В каждой комнате – своё убранство,
// в некоторых – ловушки (нарисованы на полу и описаны в заметке).

const (
	dungeonMinRooms = 4
	dungeonMaxRooms = 40
)

type dRoom struct {
	r    image.Rectangle // в клетках
	kind string
	trap string
}

type dungeonPlan struct {
	cols, rows int
	floor      []uint8 // 0 – скала, 1 – комната, 2 – коридор, 3 – вода
	rooms      []dRoom
	doors      []dDoor
}

type dDoor struct {
	x, y     int
	vertical bool // дверь в вертикальной стене (проход слева направо)
}

func (d *dungeonPlan) at(x, y int) uint8 {
	if x < 0 || y < 0 || x >= d.cols || y >= d.rows {
		return 0
	}
	return d.floor[y*d.cols+x]
}

func (d *dungeonPlan) set(x, y int, v uint8) {
	if x > 0 && y > 0 && x < d.cols-1 && y < d.rows-1 && d.floor[y*d.cols+x] == 0 {
		d.floor[y*d.cols+x] = v
	}
}

// dungeonSize – размер картинки под число комнат.
func dungeonSize(rooms int) (int, int) {
	rooms = max(dungeonMinRooms, min(dungeonMaxRooms, rooms))
	cols := min(96, 22+rooms*2)
	rows := min(68, 16+rooms*3/2)
	return cols * int(tile), rows * int(tile)
}

var dungeonRoomKinds = []string{"crypt", "armory", "library", "shrine", "prison", "barracks", "storage", "flooded", "treasure"}

var dungeonNames = map[string]string{
	"entrance": "Вход", "crypt": "Склеп", "armory": "Оружейная", "library": "Библиотека", "shrine": "Святилище",
	"prison": "Темница", "barracks": "Казарма", "storage": "Кладовая", "flooded": "Затопленный зал",
	"treasure": "Сокровищница", "throne": "Тронный зал",
}

var dungeonNotes = map[string]string{
	"entrance": "Вход в подземелье: ступени наверх, сквозняк гасит свечи",
	"crypt":    "Каменные саркофаги вдоль стен, один приоткрыт",
	"armory":   "Стойки с ржавым оружием, среди них один клинок блестит",
	"library":  "Полки с истлевшими книгами – среди них одна целая",
	"shrine":   "Алтарь забытого бога, свечи горят сами собой",
	"prison":   "Камеры за решётками, на стене чьи-то отметки дней",
	"barracks": "Нары и сундуки стражи, на столе недоеденный ужин",
	"storage":  "Бочки и ящики, часть разбита – здесь кто-то рылся",
	"flooded":  "Пол залит тёмной водой, глубина – по колено, местами с головой",
	"treasure": "Сундуки на возвышении, монеты рассыпаны по полу",
	"throne":   "Трон на возвышении, жаровни по сторонам – здесь сидит хозяин подземелья",
}

var dungeonTraps = []string{"Ловушка: нажимная плита – из стен бьют дротики (Ловкость СЛ 13, 2d6 колющего)",
	"Ловушка: яма с кольями под ложным полом (Ловкость СЛ 12, 3d6 колющего)",
	"Ловушка: огненная руна у порога (Ловкость СЛ 14, 3d6 огнём)",
	"Ловушка: ядовитый газ из трещин (Телосложение СЛ 13, 2d8 ядом)"}

func (s *site) dungeon() {
	d := s.dungeonPlan()
	lights := s.dungeonRender(d)
	for _, rm := range d.rooms {
		s.furnish(rm, &lights)
	}
	s.dungeonDoors(d)
	s.dungeonLight(d, lights)
	for _, rm := range d.rooms {
		c := rm.r.Min.Add(rm.r.Max)
		note := dungeonNotes[rm.kind]
		if rm.trap != "" {
			note += ". " + rm.trap
		}
		s.places = append(s.places, SitePlace{X: float64(c.X) * tile / 2, Y: float64(c.Y) * tile / 2, Kind: "room",
			Name: dungeonNames[rm.kind], Note: note})
	}
	s.cartouche()
}

func (s *site) dungeonPlan() *dungeonPlan {
	d := &dungeonPlan{cols: s.W / int(tile), rows: s.H / int(tile)}
	d.floor = make([]uint8, d.cols*d.rows)
	want := max(dungeonMinRooms, min(dungeonMaxRooms, s.o.Rooms))
	for tries := 0; tries < want*60 && len(d.rooms) < want; tries++ {
		w, h := 4+s.r.r.Intn(6), 4+s.r.r.Intn(5)
		x, y := 2+s.r.r.Intn(max(1, d.cols-w-4)), 3+s.r.r.Intn(max(1, d.rows-h-5))
		r := image.Rect(x, y, x+w, y+h)
		ok := true
		for _, o := range d.rooms {
			if r.Inset(-2).Overlaps(o.r) {
				ok = false
				break
			}
		}
		if ok {
			d.rooms = append(d.rooms, dRoom{r: r})
		}
	}
	// назначение: первая – вход, самая большая – тронный зал, остальные – вперемешку
	big := 1
	for i, rm := range d.rooms {
		if i > 0 && rm.r.Dx()*rm.r.Dy() > d.rooms[big].r.Dx()*d.rooms[big].r.Dy() {
			big = i
		}
	}
	for i := range d.rooms {
		switch {
		case i == 0:
			d.rooms[i].kind = "entrance"
		case i == big && len(d.rooms) > 2:
			d.rooms[i].kind = "throne"
		default:
			d.rooms[i].kind = dungeonRoomKinds[(i+s.r.r.Intn(len(dungeonRoomKinds)))%len(dungeonRoomKinds)]
		}
		if i > 0 && s.r.chance(0.3) {
			d.rooms[i].trap = dungeonTraps[s.r.r.Intn(len(dungeonTraps))]
		}
		r := d.rooms[i].r
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x < r.Max.X; x++ {
				d.floor[y*d.cols+x] = 1
			}
		}
	}
	// коридоры: остовное дерево по ближайшим и пара петель; ширина 1–2 клетки
	linked := []int{0}
	for i := 1; i < len(d.rooms); i++ {
		best, bd := 0, math.MaxFloat64
		for _, j := range linked {
			if dd := roomDist(d.rooms[i].r, d.rooms[j].r); dd < bd {
				best, bd = j, dd
			}
		}
		s.corridor(d, d.rooms[i].r, d.rooms[best].r)
		linked = append(linked, i)
	}
	for range len(d.rooms) / 4 {
		a, b := s.r.r.Intn(len(d.rooms)), s.r.r.Intn(len(d.rooms))
		if a != b {
			s.corridor(d, d.rooms[a].r, d.rooms[b].r)
		}
	}
	s.findDoors(d)
	return d
}

func roomDist(a, b image.Rectangle) float64 {
	ca, cb := a.Min.Add(a.Max), b.Min.Add(b.Max)
	return math.Hypot(float64(ca.X-cb.X), float64(ca.Y-cb.Y)) / 2
}

// corridor – ход из центра одной комнаты в центр другой с одним изломом.
func (s *site) corridor(d *dungeonPlan, a, b image.Rectangle) {
	ax, ay := (a.Min.X+a.Max.X)/2, (a.Min.Y+a.Max.Y)/2
	bx, by := (b.Min.X+b.Max.X)/2, (b.Min.Y+b.Max.Y)/2
	wide := s.r.chance(0.3)
	carve := func(x, y int) {
		d.set(x, y, 2)
		if wide {
			d.set(x+1, y, 2)
			d.set(x, y+1, 2)
		}
	}
	horizontalFirst := s.r.chance(0.5)
	x, y := ax, ay
	step := func(tx, ty int) {
		for x != tx {
			carve(x, y)
			x += sgn(tx - x)
		}
		for y != ty {
			carve(x, y)
			y += sgn(ty - y)
		}
	}
	if horizontalFirst {
		step(bx, ay)
		step(bx, by)
	} else {
		step(ax, by)
		step(bx, by)
	}
}

func sgn(v int) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	}
	return 0
}

// findDoors – двери там, где коридор шириной в клетку упирается в комнату.
func (s *site) findDoors(d *dungeonPlan) {
	for y := 1; y < d.rows-1; y++ {
		for x := 1; x < d.cols-1; x++ {
			if d.at(x, y) != 2 {
				continue
			}
			switch {
			case (d.at(x-1, y) == 1 || d.at(x+1, y) == 1) && d.at(x, y-1) == 0 && d.at(x, y+1) == 0:
				d.doors = append(d.doors, dDoor{x, y, true})
			case (d.at(x, y-1) == 1 || d.at(x, y+1) == 1) && d.at(x-1, y) == 0 && d.at(x+1, y) == 0:
				d.doors = append(d.doors, dDoor{x, y, false})
			}
		}
	}
}

type lamp struct {
	x, y, r float64
	col     color.RGBA
}

// dungeonRender – пол, стены, скала; возвращает факелы на стенах комнат.
func (s *site) dungeonRender(d *dungeonPlan) []lamp {
	open := make([]bool, s.W*s.H)
	for y := range s.H {
		for x := range s.W {
			open[y*s.W+x] = d.at(x/int(tile), y/int(tile)) != 0
		}
	}
	dFloor := distance(open, s.W, s.H) // от скалы до пола
	walls := make([]bool, len(open))
	for i, o := range open {
		walls[i] = !o
	}
	dWall := distance(walls, s.W, s.H)
	img := s.c.img
	parallelRows(s.H, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := range s.W {
				i := y*s.W + x
				var col color.RGBA
				if open[i] {
					col = s.flagstone(x, y)
					if d.at(x/int(tile), y/int(tile)) == 2 {
						col = shade(col, 0.88)
					}
					col = mix(col, rgb(84, 104, 60), smoothstep(0.66, 0.8, s.n1.fbm(float64(x)/30, float64(y)/30, 3))*0.5) // мох
					crack := s.n2.ridged(float64(x)/26, float64(y)/26, 3)
					col = mix(col, rgb(50, 46, 42), smoothstep(0.9, 0.98, crack)*0.6)
					col = shade(col, 0.55+0.45*smoothstep(0, 22, float64(dWall[i])))
				} else {
					dd := float64(dFloor[i])
					if dd < 14 { // кладка стены
						col = s.bricks(x, y)
						col = shade(col, 1.25-0.35*dd/14)
					} else {
						col = shade(rgb(52, 48, 46), (0.8+0.3*s.n2.fbm(float64(x)/18, float64(y)/18, 3))*(1-0.75*smoothstep(14, 60, dd)))
					}
				}
				o := img.PixOffset(x, y)
				img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = col.R, col.G, col.B, 255
			}
		}
	})
	var lights []lamp
	for _, rm := range d.rooms {
		x0, y0, x1, y1 := s.px(rm.r)
		for x := x0 + tile*1.5; x < x1-tile; x += tile * 3 {
			for _, y := range []float64{y0 - 4, y1 + 4} {
				s.torch(x, y)
				lights = append(lights, lamp{x, y, 230, rgb(255, 190, 110)})
			}
		}
	}
	return lights
}

// bricks – кладка: ряды камней со сдвигом, швы темнее.
func (s *site) bricks(x, y int) color.RGBA {
	const bw, bh = 22, 11
	row := y / bh
	off := (row % 2) * bw / 2
	base := shade(rgb(118, 110, 102), 0.82+0.3*hash2(float64((x+off)/bw), float64(row)))
	if (x+off)%bw < 2 || y%bh < 2 {
		return shade(base, 0.55)
	}
	return base
}

func (s *site) torch(x, y float64) {
	s.c.fill([][]pt{circle(x, y, 5)}, rgb(70, 50, 30), 1)
	s.c.fill([][]pt{circle(x, y, 3.5)}, rgb(255, 170, 60), 1)
	s.c.fill([][]pt{circle(x, y, 1.8)}, rgb(255, 240, 180), 1)
}

// dungeonDoors – деревянные двери с железными полосами и каменными косяками.
func (s *site) dungeonDoors(d *dungeonPlan) {
	for _, dr := range d.doors {
		cx, cy := (float64(dr.x)+0.5)*tile, (float64(dr.y)+0.5)*tile
		w, h := tile*0.9, 10.0
		ang := 0.0
		if dr.vertical {
			ang = math.Pi / 2
		}
		door := rotRect(cx, cy, w, h, ang)
		s.c.fill([][]pt{offset(door, 3, 3)}, rgb(10, 8, 6), 0.4)
		s.c.fill([][]pt{door}, rgb(120, 80, 44), 1)
		for _, u := range []float64{-w / 4, w / 4} {
			s.c.fill([][]pt{rotRect(cx+math.Cos(ang)*u, cy+math.Sin(ang)*u, 3, h, ang)}, rgb(70, 70, 76), 1)
		}
		s.c.stroke(closed(door), 1.2, rgb(40, 26, 16), 0.9)
		for _, u := range []float64{-w/2 - 3, w/2 + 3} {
			s.c.fill([][]pt{rotRect(cx+math.Cos(ang)*u, cy+math.Sin(ang)*u, 7, 16, ang)}, rgb(140, 132, 122), 1)
		}
	}
}

// furnish – убранство комнаты по её назначению; жаровни добавляют свет.
func (s *site) furnish(rm dRoom, lights *[]lamp) {
	x0, y0, x1, y1 := s.px(rm.r)
	cx, cy := (x0+x1)/2, (y0+y1)/2
	cell := func(i, j int) (float64, float64) { return x0 + (float64(i)+0.5)*tile, y0 + (float64(j)+0.5)*tile }
	w, h := rm.r.Dx(), rm.r.Dy()
	switch rm.kind {
	case "entrance":
		for k := range 5 { // ступени наверх
			s.box(image.Rect(int(cx-50), int(y0+8+float64(k)*12), int(cx+50), int(y0+20+float64(k)*12)), shade(rgb(150, 144, 134), 1.1-float64(k)*0.08))
		}
	case "crypt":
		for i := 1; i < w-1; i += 2 {
			for _, j := range []int{0, h - 1} {
				x, y := cell(i, j)
				s.box(image.Rect(int(x-16), int(y-34), int(x+16), int(y+34)), rgb(150, 146, 138))
				s.c.fill([][]pt{rectPts2(x-10, y-26, x+10, y+26)}, rgb(170, 166, 158), 1)
			}
		}
	case "armory":
		for i := 0; i < w; i++ {
			x, _ := cell(i, 0)
			s.c.stroke([]pt{{x - 8, y0 + 6}, {x + 8, y0 + 34}}, 3, rgb(170, 174, 180), 1)
			s.c.stroke([]pt{{x + 8, y0 + 6}, {x - 8, y0 + 34}}, 3, rgb(150, 154, 160), 1)
		}
		s.box(image.Rect(int(cx-50), int(cy-14), int(cx+50), int(cy+14)), rgb(120, 82, 50))
	case "library":
		for j := 0; j < h; j += 2 {
			_, y := cell(0, j)
			s.box(image.Rect(int(x0+6), int(y-14), int(x1-tile*1.5), int(y+14)), rgb(100, 68, 40))
			for x := x0 + 12; x < x1-tile*1.5-8; x += 9 {
				s.c.fill([][]pt{rectPts2(x, y-10, x+6, y+10)}, []color.RGBA{rgb(150, 50, 40), rgb(50, 80, 130), rgb(70, 110, 60), rgb(160, 130, 70)}[int(x)%4], 0.9)
			}
		}
	case "shrine":
		s.box(image.Rect(int(cx-40), int(y0+20), int(cx+40), int(y0+60)), rgb(190, 182, 168))
		s.brazier(cx-70, y0+40, lights)
		s.brazier(cx+70, y0+40, lights)
		s.c.fill([][]pt{circle(cx, cy+20, 46)}, rgb(150, 40, 40), 0.25) // ритуальный круг
		s.c.stroke(closed(circle(cx, cy+20, 46)), 2, rgb(170, 40, 40), 0.8)
	case "prison":
		for i := 0; i+2 <= w; i += 2 {
			x, _ := cell(i, 0)
			for b := x - tile/2; b <= x+tile*1.5; b += 8 {
				s.c.stroke([]pt{{b, y0}, {b, y0 + tile*1.5}}, 2, rgb(80, 80, 86), 1)
			}
			s.c.stroke([]pt{{x - tile/2, y0 + tile*1.5}, {x + tile*1.5, y0 + tile*1.5}}, 3, rgb(80, 80, 86), 1)
			s.bone(x+10, y0+20)
		}
	case "barracks":
		for i := 0; i < w; i += 2 {
			x, y := cell(i, 0)
			s.box(image.Rect(int(x-14), int(y-16), int(x+14), int(y+44)), rgb(110, 76, 46))
			s.box(image.Rect(int(x-11), int(y-12), int(x+11), int(y+2)), rgb(200, 192, 176))
		}
		s.box(image.Rect(int(cx-60), int(cy+10), int(cx+60), int(cy+40)), rgb(130, 90, 56))
	case "storage":
		for range w * h / 3 {
			x, y := x0+s.r.between(20, x1-x0-20), y0+s.r.between(20, y1-y0-20)
			if s.r.chance(0.5) {
				s.barrel(x, y, 14)
			} else {
				s.box(image.Rect(int(x-15), int(y-15), int(x+15), int(y+15)), rgb(150, 112, 70))
			}
		}
	case "flooded":
		var pool []pt
		for k := range 20 {
			a := float64(k) * 2 * math.Pi / 20
			rr := 0.75 + 0.3*s.n1.fbm(math.Cos(a)+3, math.Sin(a)+3, 2)
			pool = append(pool, pt{cx + math.Cos(a)*(x1-x0)*0.42*rr, cy + math.Sin(a)*(y1-y0)*0.42*rr})
		}
		s.water([][]pt{chaikin(append(pool, pool[0]), 2)})
	case "treasure":
		s.hoard(cx, cy)
		for _, dx := range []float64{-70, 70} {
			s.chest(cx+dx, y0+40)
		}
	case "throne":
		s.texture([][]pt{rectPts2(cx-30, y0+10, cx+30, y1-10)}, 0.9, func(x, y int) color.RGBA { // ковёр
			return shade(rgb(130, 30, 34), 0.9+0.15*s.n2.fbm(float64(x)/3, float64(y)/3, 2))
		})
		s.box(image.Rect(int(cx-26), int(y0+12), int(cx+26), int(y0+56)), rgb(150, 110, 40))
		s.box(image.Rect(int(cx-16), int(y0+22), int(cx+16), int(y0+50)), rgb(140, 40, 40))
		for _, dx := range []float64{-80, 80} {
			s.brazier(cx+dx, y0+60, lights)
			s.brazier(cx+dx, y1-60, lights)
		}
		for x := x0 + tile; x < x1-tile/2; x += tile * 2 {
			for _, y := range []float64{y0 + tile, y1 - tile} {
				s.pillar(x, y)
			}
		}
	}
	if rm.trap != "" { // нажимная плита или яма с кольями
		tx, ty := cell(1+s.r.r.Intn(max(1, w-2)), 1+s.r.r.Intn(max(1, h-2)))
		if s.r.chance(0.5) {
			s.c.fill([][]pt{rectPts2(tx-16, ty-16, tx+16, ty+16)}, rgb(120, 116, 108), 0.9)
			s.c.stroke(closed(rectPts2(tx-16, ty-16, tx+16, ty+16)), 1, rgb(40, 36, 32), 0.6)
		} else {
			s.c.fill([][]pt{rectPts2(tx-18, ty-18, tx+18, ty+18)}, rgb(16, 14, 12), 0.95)
			for k := range 9 {
				px, py := tx-12+float64(k%3)*12, ty-12+float64(k/3)*12
				s.c.fill([][]pt{{{px - 3, py + 3}, {px, py - 4}, {px + 3, py + 3}}}, rgb(150, 150, 156), 1)
			}
		}
	}
	for range w * h / 4 { // мусор и кости
		x, y := x0+s.r.between(10, x1-x0-10), y0+s.r.between(10, y1-y0-10)
		if s.r.chance(0.5) {
			s.rock(x, y, s.r.between(2, 5))
		} else if s.r.chance(0.3) {
			s.bone(x, y)
		}
	}
}

func (s *site) brazier(x, y float64, lights *[]lamp) {
	s.c.fill([][]pt{offset(circle(x, y, 14), 4, 4)}, rgb(10, 8, 6), 0.4)
	s.c.fill([][]pt{circle(x, y, 14)}, rgb(80, 70, 64), 1)
	s.c.fill([][]pt{circle(x, y, 10)}, rgb(255, 150, 50), 1)
	s.c.fill([][]pt{circle(x, y, 5)}, rgb(255, 236, 160), 1)
	*lights = append(*lights, lamp{x, y, 260, rgb(255, 170, 90)})
}

func (s *site) pillar(x, y float64) {
	s.c.fill([][]pt{offset(circle(x, y, 15), 8, 8)}, rgb(10, 8, 6), 0.45)
	s.c.fill([][]pt{circle(x, y, 15)}, rgb(140, 134, 126), 1)
	s.c.fill([][]pt{circle(x-4, y-4, 8)}, rgb(180, 174, 164), 1)
	s.c.stroke(closed(circle(x, y, 15)), 1, rgb(40, 36, 32), 0.7)
}

func (s *site) chest(x, y float64) {
	s.box(image.Rect(int(x-20), int(y-13), int(x+20), int(y+13)), rgb(120, 80, 40))
	s.c.fill([][]pt{rectPts2(x-20, y-2, x+20, y+2)}, rgb(200, 170, 70), 1)
	s.c.fill([][]pt{circle(x, y, 3)}, rgb(230, 200, 90), 1)
}

// dungeonLight – свет: везде полумрак, у факелов и жаровен – тёплое пятно; толща скалы почти чёрная.
func (s *site) dungeonLight(d *dungeonPlan, lights []lamp) {
	img := s.c.img
	parallelRows(s.H, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := range s.W {
				fx, fy := float64(x), float64(y)
				lr, lg, lb := 0.42, 0.42, 0.5 // холодный полумрак
				for _, l := range lights {
					dx, dy := fx-l.x, fy-l.y
					if math.Abs(dx) > l.r || math.Abs(dy) > l.r {
						continue
					}
					k := 1 - math.Hypot(dx, dy)/l.r
					if k <= 0 {
						continue
					}
					k *= k * 0.95
					lr += k * float64(l.col.R) / 255
					lg += k * float64(l.col.G) / 255
					lb += k * float64(l.col.B) / 255
				}
				o := img.PixOffset(x, y)
				img.Pix[o] = uint8(clamp(float64(img.Pix[o])*math.Min(lr, 1.35), 0, 255))
				img.Pix[o+1] = uint8(clamp(float64(img.Pix[o+1])*math.Min(lg, 1.3), 0, 255))
				img.Pix[o+2] = uint8(clamp(float64(img.Pix[o+2])*math.Min(lb, 1.2), 0, 255))
			}
		}
	})
}
