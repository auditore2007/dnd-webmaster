package worldgen

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"heroesbook/internal/model"
)

// Подземелья: план на клетках (пол, стены, двери, колонны, вода), отрисовка и места-комнаты.
// Один рисовальщик для встроенного генератора и для импорта из One Page Dungeon.

const (
	cellPx    = 32
	marginC   = 2 // поля вокруг плана, клеток
	maxCells  = 120
	maxRooms  = 40
	entryName = "Вход"
)

type rect struct{ X, Y, W, H int }

func (r rect) center() (float64, float64) {
	return float64(r.X) + float64(r.W)/2, float64(r.Y) + float64(r.H)/2
}

type door struct {
	X, Y     int
	Vertical bool // дверь в вертикальной стене (проход влево-вправо)
}

type plan struct {
	W, H    int
	floor   []bool
	water   []bool
	doors   []door
	columns [][2]float64
}

func newPlan(w, h int) *plan {
	return &plan{W: w, H: h, floor: make([]bool, w*h), water: make([]bool, w*h)}
}

func (p *plan) in(x, y int) bool      { return x >= 0 && y >= 0 && x < p.W && y < p.H }
func (p *plan) isFloor(x, y int) bool { return p.in(x, y) && p.floor[y*p.W+x] }

func (p *plan) carve(r rect) {
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			if p.in(x, y) {
				p.floor[y*p.W+x] = true
			}
		}
	}
}

// trim обрезает план по занятой области и возвращает сдвиг (сколько клеток срезано слева и сверху).
func (p *plan) trim() (dx, dy int) {
	minX, minY, maxX, maxY := p.W, p.H, -1, -1
	for y := 0; y < p.H; y++ {
		for x := 0; x < p.W; x++ {
			if p.floor[y*p.W+x] {
				minX, minY, maxX, maxY = min(minX, x), min(minY, y), max(maxX, x), max(maxY, y)
			}
		}
	}
	if maxX < 0 {
		return 0, 0
	}
	q := newPlan(maxX-minX+1, maxY-minY+1)
	for y := 0; y < q.H; y++ {
		for x := 0; x < q.W; x++ {
			q.floor[y*q.W+x] = p.floor[(y+minY)*p.W+x+minX]
			q.water[y*q.W+x] = p.water[(y+minY)*p.W+x+minX]
		}
	}
	for _, d := range p.doors {
		q.doors = append(q.doors, door{d.X - minX, d.Y - minY, d.Vertical})
	}
	for _, c := range p.columns {
		q.columns = append(q.columns, [2]float64{c[0] - float64(minX), c[1] - float64(minY)})
	}
	*p = *q
	return minX, minY
}

// px – координаты клетки в пикселях картинки.
func px(c float64) float64 { return (c + marginC) * cellPx }

// render рисует план: камень, пол с сеткой, толстые стены по границе, двери, колонны, вода.
func (p *plan) render() (string, error) {
	W, H := (p.W+2*marginC)*cellPx, (p.H+2*marginC)*cellPx
	c := newCanvas(W, H, rgb(54, 49, 44))
	floor, gridC, wall := rgb(238, 230, 212), rgb(206, 196, 176), rgb(30, 26, 24)
	for y := 0; y < p.H; y++ {
		for x := 0; x < p.W; x++ {
			if !p.floor[y*p.W+x] {
				continue
			}
			col := floor
			if p.water[y*p.W+x] {
				col = rgb(120, 168, 200)
			}
			c.fillRect(int(px(float64(x))), int(px(float64(y))), cellPx, cellPx, col)
			c.line(px(float64(x)), px(float64(y)), px(float64(x+1)), px(float64(y)), 1, gridC, 1, 0, nil)
			c.line(px(float64(x)), px(float64(y)), px(float64(x)), px(float64(y+1)), 1, gridC, 1, 0, nil)
		}
	}
	// камень: штриховка вдоль стен
	for y := -1; y <= p.H; y++ {
		for x := -1; x <= p.W; x++ {
			if p.isFloor(x, y) || !(p.isFloor(x-1, y) || p.isFloor(x+1, y) || p.isFloor(x, y-1) || p.isFloor(x, y+1)) {
				continue
			}
			bx, by := px(float64(x)), px(float64(y))
			for k := 0.0; k < cellPx; k += 7 {
				c.line(bx+k, by, bx, by+k, 1, rgb(96, 88, 80), 0.8, 0, nil)
				c.line(bx+cellPx, by+k, bx+k, by+cellPx, 1, rgb(96, 88, 80), 0.8, 0, nil)
			}
		}
	}
	for y := 0; y < p.H; y++ {
		for x := 0; x < p.W; x++ {
			if !p.floor[y*p.W+x] {
				continue
			}
			X0, Y0, X1, Y1 := px(float64(x)), px(float64(y)), px(float64(x+1)), px(float64(y+1))
			if !p.isFloor(x, y-1) {
				c.line(X0, Y0, X1, Y0, 4, wall, 1, 0, nil)
			}
			if !p.isFloor(x, y+1) {
				c.line(X0, Y1, X1, Y1, 4, wall, 1, 0, nil)
			}
			if !p.isFloor(x-1, y) {
				c.line(X0, Y0, X0, Y1, 4, wall, 1, 0, nil)
			}
			if !p.isFloor(x+1, y) {
				c.line(X1, Y0, X1, Y1, 4, wall, 1, 0, nil)
			}
		}
	}
	for _, d := range p.doors {
		cx, cy := px(float64(d.X)+0.5), px(float64(d.Y)+0.5)
		if d.Vertical {
			c.fillRect(int(cx-4), int(cy-11), 8, 22, rgb(120, 78, 42))
		} else {
			c.fillRect(int(cx-11), int(cy-4), 22, 8, rgb(120, 78, 42))
		}
	}
	for _, k := range p.columns {
		c.disc(px(k[0]), px(k[1]), cellPx*0.28, wall, 1)
	}
	return c.dataURL(88)
}

var roomNotes = []string{"Ловушка: ядовитые иглы в замке сундука (Внимательность СЛ 13, яд 2d6)", "Ловушка: проваливающийся пол (Ловкость СЛ 12, 2d6 дробящего)",
	"Ловушка: огненная руна у порога (Ловкость СЛ 14, 3d6 огнём)", "Старый алтарь с выбитыми письменами", "Фонтан с мутной водой – пить опасно",
	"Разбитая статуя: в постаменте тайник", "Сундук под толстым слоем пыли", "Полки с истлевшими книгами – среди них одна целая",
	"Склеп: каменные гробы, один приоткрыт", "Оружейная: ржавые клинки и один блестящий", "Тюремные камеры, на стене чьи-то отметки дней",
	"Лаборатория алхимика: склянки, часть ещё цела", "Здесь недавно жгли костёр – угли тёплые", "Сквозняк из трещины в стене – тайный ход?"}

// Dungeon генерирует подземелье: комнаты, соединённые коридорами, с ловушками, тайниками и врагами.
func (g Gen) Dungeon(rooms int) (string, []model.Place, error) {
	rooms = max(4, min(maxRooms, rooms))
	side := 14 + rooms*2
	p := newPlan(side+8, side)
	var rs []rect
	for tries := 0; tries < rooms*40 && len(rs) < rooms; tries++ {
		r := rect{X: 1 + g.R.Intn(p.W-10), Y: 1 + g.R.Intn(p.H-10), W: 3 + g.R.Intn(6), H: 3 + g.R.Intn(5)}
		ok := true
		for _, o := range rs {
			if r.X-2 < o.X+o.W && o.X-2 < r.X+r.W && r.Y-2 < o.Y+o.H && o.Y-2 < r.Y+r.H {
				ok = false
				break
			}
		}
		if ok {
			rs = append(rs, r)
		}
	}
	if len(rs) < 2 {
		return "", nil, errors.New("не удалось разместить комнаты")
	}
	inRoom := make([]bool, p.W*p.H)
	for _, r := range rs {
		p.carve(r)
		for y := r.Y; y < r.Y+r.H; y++ {
			for x := r.X; x < r.X+r.W; x++ {
				inRoom[y*p.W+x] = true
			}
		}
		if r.W >= 5 && r.H >= 5 && g.R.Intn(3) == 0 { // колонны в больших залах
			p.columns = append(p.columns, [2]float64{float64(r.X) + 1.5, float64(r.Y) + 1.5}, [2]float64{float64(r.X+r.W) - 1.5, float64(r.Y) + 1.5},
				[2]float64{float64(r.X) + 1.5, float64(r.Y+r.H) - 1.5}, [2]float64{float64(r.X+r.W) - 1.5, float64(r.Y+r.H) - 1.5})
		}
	}
	// коридоры: каждая комната соединяется с ближайшей уже связанной (остовное дерево) и пара лишних петель
	linked := []int{0}
	connect := func(a, b rect) {
		ax, ay := a.center()
		bx, by := b.center()
		x, y := int(ax), int(ay)
		for x != int(bx) {
			p.carve(rect{x, y, 1, 1})
			x += sign(int(bx) - x)
		}
		for y != int(by) {
			p.carve(rect{x, y, 1, 1})
			y += sign(int(by) - y)
		}
	}
	for i := 1; i < len(rs); i++ {
		best, bd := 0, math.MaxFloat64
		ix, iy := rs[i].center()
		for _, j := range linked {
			jx, jy := rs[j].center()
			if d := math.Hypot(ix-jx, iy-jy); d < bd {
				best, bd = j, d
			}
		}
		connect(rs[i], rs[best])
		linked = append(linked, i)
	}
	for k := 0; k < len(rs)/4; k++ {
		connect(rs[g.R.Intn(len(rs))], rs[g.R.Intn(len(rs))])
	}
	// двери – там, где коридор входит в комнату
	for y := 0; y < p.H; y++ {
		for x := 0; x < p.W; x++ {
			i := y*p.W + x
			if !p.floor[i] || inRoom[i] {
				continue
			}
			switch {
			case p.in(x-1, y) && inRoom[i-1] || p.in(x+1, y) && inRoom[i+1]:
				if !p.isFloor(x, y-1) && !p.isFloor(x, y+1) {
					p.doors = append(p.doors, door{x, y, true})
				}
			case p.in(x, y-1) && inRoom[i-p.W] || p.in(x, y+1) && inRoom[i+p.W]:
				if !p.isFloor(x-1, y) && !p.isFloor(x+1, y) {
					p.doors = append(p.doors, door{x, y, false})
				}
			}
		}
	}
	dx, dy := p.trim()
	url, err := p.render()
	if err != nil {
		return "", nil, err
	}
	places := make([]model.Place, 0, len(rs))
	for i, r := range rs {
		cx, cy := r.center()
		cx, cy = cx-float64(dx), cy-float64(dy)
		pl := model.Place{ID: g.NewID(), X: px(cx), Y: px(cy), Kind: "room", Name: fmt.Sprintf("Комната %d", i+1), Note: pick(g.R, roomNotes)}
		if i == 0 {
			pl.Name, pl.Note = entryName, "Вход в подземелье"
		}
		places = append(places, pl)
	}
	places = g.Populate(places, false)
	places[0].Foes, places[0].Loot = nil, "" // у входа тихо
	return url, places, nil
}

func sign(v int) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	}
	return 0
}

// ---------- импорт One Page Dungeon (watabou.itch.io/one-page-dungeon, экспорт JSON) ----------

type opd struct {
	Title string `json:"title"`
	Story string `json:"story"`
	Rects []struct {
		X, Y, W, H float64
	} `json:"rects"`
	Doors []struct {
		X, Y float64
		Dir  struct{ X, Y float64 } `json:"dir"`
		Type int                    `json:"type"`
	} `json:"doors"`
	Notes []struct {
		Text string                 `json:"text"`
		Ref  string                 `json:"ref"`
		Pos  struct{ X, Y float64 } `json:"pos"`
	} `json:"notes"`
	Columns []struct{ X, Y float64 } `json:"columns"`
	Water   []struct{ X, Y float64 } `json:"water"`
}

// OnePageDungeon рисует подземелье из JSON One Page Dungeon и превращает его заметки в комнаты с врагами.
func (g Gen) OnePageDungeon(raw []byte) (title, url string, places []model.Place, err error) {
	var d opd
	if err := json.Unmarshal(raw, &d); err != nil {
		return "", "", nil, fmt.Errorf("это не JSON из One Page Dungeon: %w", err)
	}
	if len(d.Rects) == 0 {
		return "", "", nil, errors.New("в файле нет комнат (rects) – нужен экспорт JSON из One Page Dungeon")
	}
	minX, minY, maxX, maxY := math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64
	for _, r := range d.Rects {
		minX, minY = math.Min(minX, r.X), math.Min(minY, r.Y)
		maxX, maxY = math.Max(maxX, r.X+r.W), math.Max(maxY, r.Y+r.H)
	}
	w, h := int(maxX-minX), int(maxY-minY)
	if w <= 0 || h <= 0 || w > maxCells || h > maxCells {
		return "", "", nil, fmt.Errorf("подземелье %d×%d клеток: поддерживается до %d×%d", w, h, maxCells, maxCells)
	}
	ox, oy := -minX, -minY
	p := newPlan(w, h)
	for _, r := range d.Rects {
		p.carve(rect{int(r.X + ox), int(r.Y + oy), int(r.W), int(r.H)})
	}
	for _, wt := range d.Water {
		if x, y := int(wt.X+ox), int(wt.Y+oy); p.in(x, y) {
			p.water[y*p.W+x] = true
		}
	}
	for _, dr := range d.Doors {
		p.doors = append(p.doors, door{int(dr.X + ox), int(dr.Y + oy), dr.Dir.X != 0})
	}
	for _, c := range d.Columns {
		p.columns = append(p.columns, [2]float64{c.X + ox + 0.5, c.Y + oy + 0.5})
	}
	url, err = p.render()
	if err != nil {
		return "", "", nil, err
	}
	for _, n := range d.Notes {
		if len(places) >= maxPlaces {
			break
		}
		name := Clip("Комната "+n.Ref, NameLen)
		if n.Ref == "" {
			name = "Комната"
		}
		places = append(places, model.Place{ID: g.NewID(), X: px(n.Pos.X + ox), Y: px(n.Pos.Y + oy), Kind: "room", Name: name, Note: Clip(strings.TrimSpace(n.Text), NoteLen)})
	}
	if len(places) == 0 { // заметок нет – комнатами считаем крупные прямоугольники
		for i, r := range d.Rects {
			if r.W >= 2 && r.H >= 2 && len(places) < maxPlaces {
				places = append(places, model.Place{ID: g.NewID(), X: px(r.X + ox + r.W/2), Y: px(r.Y + oy + r.H/2), Kind: "room", Name: fmt.Sprintf("Комната %d", i+1)})
			}
		}
	}
	places = g.Populate(places, false)
	if d.Story != "" && len(places) > 0 {
		places[0].Note = Clip(strings.TrimSpace(d.Story+"\n\n"+places[0].Note), NoteLen)
	}
	return Clip(strings.TrimSpace(d.Title), NameLen), url, places, nil
}
