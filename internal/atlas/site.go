package atlas

import (
	"errors"
	"image"
	"image/color"
	"math"

	"golang.org/x/image/vector"
)

// Карты локаций – вид сверху на одно место мира: поселение с улицами и домами, замок, пещеру, руины,
// лагерь, внутренности здания. Рисуется «как на макете»: земля с фактурой, тени от построек и деревьев
// (свет с северо-запада), крыши со скатами разной освещённости, мощёные улицы, вода с отмелями.

// SiteOpts – что за место и в какой местности оно стоит (местность берётся с карты мира).
type SiteOpts struct {
	Kind     string  // вид места (model.PlaceKinds)
	Name     string  // название – на картуше
	Race     string  // народ: от него цвет крыш и стиль
	Biome    string  // grass, forest, snow, desert, swamp, jungle
	WaterDir float64 // с какой стороны вода (радианы, 0 – восток), если Water
	Water    bool    // рядом море или озеро
	R        Rand
}

// SitePlace – место внутри локации (таверна в городе, зал в пещере…), координаты – в пикселях картинки.
type SitePlace struct {
	X, Y       float64
	Kind, Name string
}

// SiteMap – готовая карта локации; Grid – размер клетки боевой сетки в пикселях (0 – без сетки).
type SiteMap struct {
	Img    *image.RGBA
	Places []SitePlace
	Grid   int
}

// site – холст локации и всё, что нужно генераторам.
type site struct {
	o        SiteOpts
	c        canvas
	W, H     int
	r        rng
	n1, n2   *noise
	occ      []bool // занятость по клеткам occStep×occStep: дома не налезают друг на друга и на дороги
	ow, oh   int
	pal      groundPalette
	places   []SitePlace
	clear    func(x, y float64) bool // где лес не растёт (поляна под поселение)
	caveOpen []bool                  // пол пещеры попиксельно
}

const occStep = 4

// Location рисует карту локации.
func Location(o SiteOpts) (*SiteMap, error) {
	if o.R == nil {
		return nil, errors.New("atlas: нет генератора случайностей")
	}
	gen, w, h, grid := siteKind(o.Kind)
	s := &site{o: o, W: w, H: h, r: rng{o.R}}
	s.c = canvas{image.NewRGBA(image.Rect(0, 0, w, h))}
	s.n1, s.n2 = newNoise(s.r), newNoise(s.r)
	s.ow, s.oh = w/occStep+1, h/occStep+1
	s.occ = make([]bool, s.ow*s.oh)
	s.pal = palettes[o.Biome]
	if len(s.pal.grass) == 0 {
		s.pal = palettes["grass"]
	}
	gen(s)
	return &SiteMap{Img: s.c.img, Places: s.places, Grid: grid}, nil
}

// siteKind – генератор, размер картинки и боевая сетка для вида места.
func siteKind(kind string) (func(*site), int, int, int) {
	switch kind {
	case "capital":
		return (*site).town, 3200, 2200, 0
	case "city":
		return (*site).town, 2800, 1900, 0
	case "town", "port":
		return (*site).town, 2400, 1650, 0
	case "village":
		return (*site).town, 2000, 1400, 0
	case "castle":
		return (*site).castle, 2400, 1700, 0
	case "cave", "mine", "lair":
		return (*site).cave, 2000, 1500, 40
	case "tavern", "temple", "shop", "smithy":
		return (*site).interior, 1400, 1000, 40
	}
	return (*site).outdoor, 2000, 1400, 0 // руины, лагерь, святилище, башня, лес, горы, болото, озеро, подземелье-вход
}

func (s *site) addPlace(x, y float64, kind, name string) {
	s.places = append(s.places, SitePlace{X: math.Round(x), Y: math.Round(y), Kind: kind, Name: name})
}

// ---------- местность ----------

type groundPalette struct {
	grass, grass2, dirt, road, stone, roof []color.RGBA
	snow                                   bool
}

func pal(c ...color.RGBA) []color.RGBA { return c }

var palettes = map[string]groundPalette{
	"grass":  {grass: pal(rgb(104, 138, 66), rgb(126, 152, 78)), grass2: pal(rgb(88, 122, 58)), dirt: pal(rgb(146, 124, 88)), road: pal(rgb(170, 148, 110)), stone: pal(rgb(150, 144, 132)), roof: pal(rgb(164, 82, 58), rgb(150, 96, 62), rgb(118, 102, 92), rgb(176, 140, 84))},
	"forest": {grass: pal(rgb(84, 120, 58), rgb(100, 132, 64)), grass2: pal(rgb(66, 100, 50)), dirt: pal(rgb(128, 108, 76)), road: pal(rgb(156, 134, 98)), stone: pal(rgb(140, 136, 126)), roof: pal(rgb(124, 92, 64), rgb(148, 106, 70), rgb(110, 104, 96))},
	"jungle": {grass: pal(rgb(70, 116, 52), rgb(86, 132, 58)), grass2: pal(rgb(54, 98, 44)), dirt: pal(rgb(130, 104, 70)), road: pal(rgb(158, 128, 92)), stone: pal(rgb(132, 130, 116)), roof: pal(rgb(176, 150, 92), rgb(150, 118, 70))},
	"swamp":  {grass: pal(rgb(92, 106, 68), rgb(106, 116, 74)), grass2: pal(rgb(74, 90, 60)), dirt: pal(rgb(110, 98, 72)), road: pal(rgb(136, 120, 90)), stone: pal(rgb(128, 128, 116)), roof: pal(rgb(132, 112, 76), rgb(112, 98, 70))},
	"desert": {grass: pal(rgb(214, 190, 140), rgb(222, 200, 152)), grass2: pal(rgb(196, 170, 120)), dirt: pal(rgb(188, 156, 110)), road: pal(rgb(200, 176, 132)), stone: pal(rgb(196, 176, 140)), roof: pal(rgb(212, 196, 160), rgb(190, 150, 104), rgb(176, 112, 76))},
	"snow":   {grass: pal(rgb(232, 236, 240), rgb(222, 228, 236)), grass2: pal(rgb(206, 214, 226)), dirt: pal(rgb(150, 140, 128)), road: pal(rgb(184, 178, 170)), stone: pal(rgb(150, 150, 154)), roof: pal(rgb(120, 100, 86), rgb(96, 92, 96), rgb(140, 84, 64)), snow: true},
}

// ground – земля с крупными и мелкими пятнами, мягким рельефом и проплешинами.
func (s *site) ground() {
	img := s.c.img
	g := s.pal
	parallelRows(s.H, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := range s.W {
				fx, fy := float64(x), float64(y)
				big := s.n1.fbm(fx/260, fy/260, 3)
				mid := s.n2.fbm(fx/50+9, fy/50, 3)
				col := mix(g.grass[0], g.grass[len(g.grass)-1], big)
				col = mix(col, g.grass2[0], smoothstep(0.55, 0.75, mid)*0.6)
				col = mix(col, g.dirt[0], smoothstep(0.7, 0.82, s.n1.fbm(fx/70+40, fy/70, 3))*0.5)
				fine := s.n2.fbm(fx/2.2, fy/2.2, 2)
				col = shade(col, 0.9+0.2*fine)
				// мягкий рельеф: светлее склоны на северо-запад
				h1, h2 := s.n1.fbm((fx+3)/300, fy/300, 3), s.n1.fbm(fx/300, (fy+3)/300, 3)
				h0 := s.n1.fbm(fx/300, fy/300, 3)
				col = shade(col, clamp(1-((h1-h0)+(h2-h0))*40, 0.85, 1.15))
				o := img.PixOffset(x, y)
				img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = col.R, col.G, col.B, 255
			}
		}
	})
}

// texture закрашивает объединение многоугольников цветом fn(x, y) с прозрачностью alpha (сглаженный край).
func (s *site) texture(polys [][]pt, alpha float64, fn func(x, y int) color.RGBA) {
	r, mask := s.rasterize(polys)
	if mask == nil {
		return
	}
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if a := mask.AlphaAt(x-r.Min.X, y-r.Min.Y).A; a > 0 {
				s.c.blend(x, y, fn(x, y), float64(a)/255*alpha)
			}
		}
	}
}

// rasterize – маска объединения многоугольников в пределах их рамки.
func (s *site) rasterize(polys [][]pt) (image.Rectangle, *image.Alpha) {
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, p := range polys {
		for _, q := range p {
			minX, minY = math.Min(minX, q.x), math.Min(minY, q.y)
			maxX, maxY = math.Max(maxX, q.x), math.Max(maxY, q.y)
		}
	}
	r := image.Rect(int(minX)-1, int(minY)-1, int(maxX)+2, int(maxY)+2).Intersect(s.c.bounds())
	if r.Empty() {
		return r, nil
	}
	z := vector.NewRasterizer(r.Dx(), r.Dy())
	for _, p := range polys {
		if len(p) < 3 {
			continue
		}
		rev := area(p) < 0
		for i := range p {
			q := p[i]
			if rev {
				q = p[len(p)-1-i]
			}
			if i == 0 {
				z.MoveTo(float32(q.x-float64(r.Min.X)), float32(q.y-float64(r.Min.Y)))
			} else {
				z.LineTo(float32(q.x-float64(r.Min.X)), float32(q.y-float64(r.Min.Y)))
			}
		}
		z.ClosePath()
	}
	mask := image.NewAlpha(image.Rect(0, 0, r.Dx(), r.Dy()))
	z.Draw(mask, mask.Bounds(), image.Opaque, image.Point{})
	return r, mask
}

// occupy отмечает занятые клетки под многоугольниками; free – свободна ли рамка многоугольника.
func (s *site) occupy(polys [][]pt) {
	r, mask := s.rasterize(polys)
	if mask == nil {
		return
	}
	for y := r.Min.Y; y < r.Max.Y; y += 2 {
		for x := r.Min.X; x < r.Max.X; x += 2 {
			if mask.AlphaAt(x-r.Min.X, y-r.Min.Y).A > 60 {
				s.occ[(y/occStep)*s.ow+x/occStep] = true
			}
		}
	}
}

func (s *site) freePoly(poly []pt, margin float64) bool {
	for _, q := range poly {
		if q.x < margin || q.y < margin || q.x > float64(s.W)-margin || q.y > float64(s.H)-margin {
			return false
		}
	}
	r, mask := s.rasterize([][]pt{poly})
	if mask == nil {
		return false
	}
	for y := r.Min.Y; y < r.Max.Y; y += 3 {
		for x := r.Min.X; x < r.Max.X; x += 3 {
			if mask.AlphaAt(x-r.Min.X, y-r.Min.Y).A > 30 && s.occ[(y/occStep)*s.ow+x/occStep] {
				return false
			}
		}
	}
	return true
}

func (s *site) busy(x, y float64) bool {
	if x < 0 || y < 0 || x >= float64(s.W) || y >= float64(s.H) {
		return true
	}
	return s.occ[(int(y)/occStep)*s.ow+int(x)/occStep]
}

// rotRect – прямоугольник w×d с центром (cx, cy), повёрнутый на angle.
func rotRect(cx, cy, w, d, angle float64) []pt {
	ca, sa := math.Cos(angle), math.Sin(angle)
	at := func(u, v float64) pt { return pt{cx + u*ca - v*sa, cy + u*sa + v*ca} }
	return []pt{at(-w/2, -d/2), at(w/2, -d/2), at(w/2, d/2), at(-w/2, d/2)}
}

func offset(poly []pt, dx, dy float64) []pt {
	out := make([]pt, len(poly))
	for i, q := range poly {
		out[i] = pt{q.x + dx, q.y + dy}
	}
	return out
}

// shadow – тень от предмета высотой h: многоугольник, сдвинутый на юго-восток.
func (s *site) shadow(poly []pt, h float64) {
	sh := offset(poly, h*0.55, h*0.55)
	hull := append(append([]pt{}, poly...), sh...)
	s.c.fill([][]pt{convexHull(hull)}, rgb(20, 24, 16), 0.32)
}

// convexHull – выпуклая оболочка (монотонная цепь Эндрю).
func convexHull(ps []pt) []pt {
	pts := append([]pt(nil), ps...)
	for i := 1; i < len(pts); i++ {
		for j := i; j > 0 && (pts[j].x < pts[j-1].x || pts[j].x == pts[j-1].x && pts[j].y < pts[j-1].y); j-- {
			pts[j], pts[j-1] = pts[j-1], pts[j]
		}
	}
	cross := func(o, a, b pt) float64 { return (a.x-o.x)*(b.y-o.y) - (a.y-o.y)*(b.x-o.x) }
	var lower, upper []pt
	for _, p := range pts {
		for len(lower) >= 2 && cross(lower[len(lower)-2], lower[len(lower)-1], p) <= 0 {
			lower = lower[:len(lower)-1]
		}
		lower = append(lower, p)
	}
	for i := len(pts) - 1; i >= 0; i-- {
		p := pts[i]
		for len(upper) >= 2 && cross(upper[len(upper)-2], upper[len(upper)-1], p) <= 0 {
			upper = upper[:len(upper)-1]
		}
		upper = append(upper, p)
	}
	return append(lower[:len(lower)-1], upper[:len(upper)-1]...)
}

// light – освещённость грани с внешней нормалью (nx, ny): свет с северо-запада.
func light(nx, ny float64) float64 {
	l := math.Hypot(nx, ny)
	if l == 0 {
		return 1
	}
	return 0.8 + 0.34*(-nx-ny)/(l*math.Sqrt2)
}

// house – дом сверху: тень, вальмовая (hip) или двускатная крыша, у каждого ската свой свет, черепица рядами.
func (s *site) house(cx, cy, w, d, angle float64, roof color.RGBA, hip bool) []pt {
	if d > w {
		w, d = d, w
		angle += math.Pi / 2
	}
	foot := rotRect(cx, cy, w, d, angle)
	s.shadow(foot, d*0.55)
	ca, sa := math.Cos(angle), math.Sin(angle)
	at := func(u, v float64) pt { return pt{cx + u*ca - v*sa, cy + u*sa + v*ca} }
	inset := 0.0
	if hip {
		inset = d * 0.45
	}
	r1, r2 := at(-w/2+inset, 0), at(w/2-inset, 0)
	a, b, c, dd := foot[0], foot[1], foot[2], foot[3]
	faces := []struct {
		poly   []pt
		nx, ny float64
	}{
		{[]pt{a, b, r2, r1}, sa, -ca},  // скат к -v
		{[]pt{dd, c, r2, r1}, -sa, ca}, // скат к +v
		{[]pt{a, dd, r1}, -ca, -sa},    // торец -u
		{[]pt{b, c, r2}, ca, sa},       // торец +u
	}
	for k, f := range faces {
		if !hip && k >= 2 {
			continue
		}
		s.c.fill([][]pt{f.poly}, shade(roof, light(f.nx, f.ny)), 1)
	}
	// ряды черепицы вдоль карниза
	for v := d / 2 * 0.25; v < d/2; v += math.Max(2.5, d/9) {
		for _, sign := range []float64{-1, 1} {
			u0 := -w/2 + inset*(1-v/(d/2))
			s.c.stroke([]pt{at(u0, sign*v), at(-u0, sign*v)}, 0.7, shade(roof, 0.6), 0.22)
		}
	}
	s.c.stroke([]pt{r1, r2}, 1.2, shade(roof, 0.5), 0.7)
	if hip {
		s.c.stroke([]pt{a, r1, dd}, 0.8, shade(roof, 0.55), 0.45)
		s.c.stroke([]pt{b, r2, c}, 0.8, shade(roof, 0.55), 0.45)
	}
	s.c.stroke(closed(foot), 1.1, rgb(40, 30, 24), 0.65)
	if s.pal.snow { // снег на крыше
		s.c.fill([][]pt{faces[0].poly}, rgb(240, 244, 248), 0.55)
	}
	if s.r.chance(0.6) { // труба
		u := s.r.between(-w/3, w/3)
		ch := rotRect(at(u, -d*0.2).x, at(u, -d*0.2).y, 4, 4, angle)
		s.c.fill([][]pt{offset(ch, 2, 2)}, rgb(20, 20, 20), 0.3)
		s.c.fill([][]pt{ch}, rgb(110, 96, 88), 1)
	}
	s.occupy([][]pt{foot})
	return foot
}

// tree – дерево сверху: тень, тёмная крона из нескольких «облаков», освещённые бока.
func (s *site) tree(x, y, r float64, col color.RGBA, conifer bool) {
	s.c.fill([][]pt{ellipse(x+r*0.5, y+r*0.45, r*1.05, r*0.95)}, rgb(16, 22, 12), 0.3)
	if conifer {
		star := func(rr, rot float64) []pt {
			var out []pt
			for k := range 18 {
				a := rot + float64(k)*math.Pi/9
				l := rr
				if k%2 == 1 {
					l = rr * 0.62
				}
				out = append(out, pt{x + math.Cos(a)*l, y + math.Sin(a)*l})
			}
			return out
		}
		rot := s.r.between(0, math.Pi)
		s.c.fill([][]pt{star(r, rot)}, shade(col, 0.62), 1)
		s.c.fill([][]pt{offset(star(r*0.7, rot+0.2), -r*0.12, -r*0.12)}, shade(col, 0.85), 1)
		s.c.fill([][]pt{offset(star(r*0.38, rot+0.4), -r*0.18, -r*0.18)}, shade(col, 1.08), 1)
		if s.pal.snow {
			s.c.fill([][]pt{offset(star(r*0.5, rot), -r*0.15, -r*0.15)}, rgb(236, 240, 246), 0.7)
		}
		return
	}
	s.c.fill([][]pt{circle(x, y, r)}, shade(col, 0.6), 1)
	for k := range 6 {
		a := float64(k)*math.Pi/3 + s.r.between(0, 0.8)
		s.c.fill([][]pt{circle(x+math.Cos(a)*r*0.45, y+math.Sin(a)*r*0.45, r*0.5)}, shade(col, 0.78), 1)
	}
	s.c.fill([][]pt{circle(x-r*0.22, y-r*0.22, r*0.55)}, col, 1)
	s.c.fill([][]pt{circle(x-r*0.36, y-r*0.36, r*0.26)}, shade(col, 1.22), 0.9)
	if s.pal.snow {
		s.c.fill([][]pt{circle(x-r*0.25, y-r*0.25, r*0.4)}, rgb(236, 240, 246), 0.55)
	}
}

func (s *site) treeColor() (color.RGBA, bool) {
	switch s.o.Biome {
	case "snow":
		return s.vary(rgb(52, 84, 64)), true
	case "desert":
		return s.vary(rgb(110, 128, 62)), false
	case "jungle":
		return s.vary(rgb(46, 104, 48)), false
	case "forest":
		return s.vary(rgb(52, 94, 48)), s.r.chance(0.35)
	}
	return s.vary(rgb(70, 112, 52)), s.r.chance(0.15)
}

func (s *site) vary(c color.RGBA) color.RGBA {
	return mix(shade(c, s.r.between(0.85, 1.12)), rgb(150, 150, 60), s.r.between(0, 0.12))
}

// forestFill – деревья там, где свободно и шум «лесистости» выше порога.
func (s *site) forestFill(density, threshold float64) {
	step := 26.0
	for y := step / 2; y < float64(s.H); y += step * 0.86 {
		for x := step / 2; x < float64(s.W); x += step {
			jx, jy := x+s.r.between(-0.45, 0.45)*step, y+s.r.between(-0.45, 0.45)*step
			if s.n2.fbm(jx/220+70, jy/220, 3) < threshold || !s.r.chance(density) || s.busy(jx, jy) || s.clear != nil && s.clear(jx, jy) {
				continue
			}
			col, con := s.treeColor()
			r := s.r.between(11, 19)
			s.tree(jx, jy, r, col, con)
			s.occupy([][]pt{circle(jx, jy, r*0.6)})
		}
	}
}

// road – дорога: мягкие края и колеи; cobble – мощёная улица.
func (s *site) road(line []pt, width float64, cobble bool) {
	line = chaikin(line, 3)
	edge := strokePolys(line, []float64{width + 6})
	s.texture(edge, 0.35, func(x, y int) color.RGBA { return shade(s.pal.dirt[0], 0.8) })
	body := strokePolys(line, []float64{width})
	if cobble {
		s.texture(body, 0.95, s.cobble)
	} else {
		s.texture(body, 0.85, func(x, y int) color.RGBA {
			n := s.n2.fbm(float64(x)/5, float64(y)/5, 2)
			return shade(s.pal.road[0], 0.86+0.24*n)
		})
	}
	s.occupy(strokePolys(line, []float64{width + 8}))
}

// cobble – булыжник: ячейки Вороного с тёмными швами.
func (s *site) cobble(x, y int) color.RGBA {
	const cell = 7.0
	fx, fy := float64(x)/cell, float64(y)/cell
	ix, iy := math.Floor(fx), math.Floor(fy)
	d1, d2, id := 9.0, 9.0, 0.0
	for j := -1.0; j <= 1; j++ {
		for i := -1.0; i <= 1; i++ {
			h := hash2(ix+i, iy+j)
			px, py := ix+i+h, iy+j+hash2(iy+j+31, ix+i+17)
			d := sq(px-fx) + sq(py-fy)
			if d < d1 {
				d1, d2, id = d, d1, h
			} else if d < d2 {
				d2 = d
			}
		}
	}
	base := shade(s.pal.stone[0], 0.86+0.26*id)
	seam := smoothstep(0.0, 0.12, math.Sqrt(d2)-math.Sqrt(d1))
	return mix(shade(base, 0.55), base, seam)
}

// hash2 – псевдослучайное число 0..1 для пары координат.
func hash2(a, b float64) float64 {
	v := math.Sin(a*127.1+b*311.7) * 43758.5453
	return v - math.Floor(v)
}

// water – водоём: глубина по расстоянию до берега, отмель, пена и песок у кромки.
func (s *site) water(polys [][]pt) {
	r, mask := s.rasterize(polys)
	if mask == nil {
		return
	}
	w, h := r.Dx(), r.Dy()
	in := make([]bool, w*h)
	for i := range in {
		in[i] = mask.Pix[i] < 128
	}
	d := distance(in, w, h)
	for y := range h {
		for x := range w {
			a := float64(mask.Pix[y*w+x]) / 255
			if a == 0 {
				continue
			}
			dd := float64(d[y*w+x])
			col := mix(rgb(118, 176, 172), rgb(40, 94, 128), clamp(dd/90, 0, 1))
			col = mix(col, rgb(26, 64, 100), clamp((dd-90)/200, 0, 1))
			fx, fy := float64(x+r.Min.X), float64(y+r.Min.Y)
			col = shade(col, 0.95+0.1*s.n1.fbm(fx/30, fy/12, 3))
			col = mix(col, rgb(226, 240, 236), 0.5*clamp(1-dd/6, 0, 1))
			s.c.blend(x+r.Min.X, y+r.Min.Y, col, a)
		}
	}
	// песок по берегу
	s.texture(polysOutline(polys, 10), 0.5, func(x, y int) color.RGBA {
		return shade(rgb(214, 196, 150), 0.95+0.1*s.n2.fbm(float64(x)/3, float64(y)/3, 2))
	})
	s.occupy(polys)
}

// polysOutline – полосы ширины width вдоль контуров многоугольников.
func polysOutline(polys [][]pt, width float64) [][]pt {
	var out [][]pt
	for _, p := range polys {
		out = append(out, strokePolys(closed(p), []float64{width})...)
	}
	return out
}

// shore – многоугольник моря с неровным берегом со стороны dir (вода занимает долю part картинки).
func (s *site) shore(dir, part float64) []pt {
	cx, cy := float64(s.W)/2, float64(s.H)/2
	nx, ny := math.Cos(dir), math.Sin(dir)
	tx, ty := -ny, nx
	reach := math.Abs(nx)*float64(s.W)/2 + math.Abs(ny)*float64(s.H)/2
	base := reach * (1 - 2*part)
	var line []pt
	span := math.Hypot(float64(s.W), float64(s.H))
	for t := -span; t <= span; t += 20 {
		off := base + (s.n1.fbm(t/260+5, 3, 3)-0.5)*220 + (s.n2.fbm(t/60, 7, 2)-0.5)*50
		line = append(line, pt{cx + tx*t + nx*off, cy + ty*t + ny*off})
	}
	far := 2 * span
	line = append(line, pt{cx + tx*span + nx*far, cy + ty*span + ny*far}, pt{cx - tx*span + nx*far, cy - ty*span + ny*far})
	return line
}

// river – извилистая река через всю карту.
func (s *site) river(width float64) []pt {
	a := s.r.between(0, math.Pi)
	cx, cy := float64(s.W)/2+s.r.between(-0.2, 0.2)*float64(s.W), float64(s.H)/2+s.r.between(-0.2, 0.2)*float64(s.H)
	dx, dy := math.Cos(a), math.Sin(a)
	span := math.Hypot(float64(s.W), float64(s.H)) * 0.6
	var line []pt
	for t := -span; t <= span; t += 30 {
		off := (s.n1.fbm(t/400+11, 2, 3) - 0.5) * 500
		line = append(line, pt{cx + dx*t - dy*off, cy + dy*t + dx*off})
	}
	line = chaikin(line, 2)
	widths := make([]float64, len(line))
	for i := range widths {
		widths[i] = width * (0.85 + 0.3*s.n2.fbm(float64(i)/20, 1, 2))
	}
	s.water(strokePolys(line, widths))
	return line
}

// cartouche – название локации в рамке вверху.
func (s *site) cartouche() {
	if loadFonts() != nil || s.o.Name == "" {
		return
	}
	p := &painter{c: s.c, W: s.W, H: s.H, scale: 1}
	l := label{text: s.o.Name, font: fonts.bold, size: 40, spacing: 2, col: inkColor, alpha: 1}
	tw, a, _ := l.measure()
	r := image.Rect(int(float64(s.W)/2-tw/2-40), 24, int(float64(s.W)/2+tw/2+40), 24+int(a)+40)
	s.c.fill([][]pt{rectPts(r.Add(image.Pt(4, 5)))}, rgb(20, 16, 10), 0.3)
	s.c.fill([][]pt{rectPts(r)}, rgb(242, 230, 200), 0.95)
	s.c.stroke(closed(rectPts(r)), 2, inkColor, 0.9)
	s.c.stroke(closed(rectPts(r.Inset(5))), 0.8, inkColor, 0.7)
	l.x, l.y = float64(s.W)/2-tw/2, float64(r.Min.Y)+20+a*0.8
	p.drawLabel(l)
}

// vignette – лёгкое затемнение к краям, чтобы взгляд держался на центре.
func (s *site) vignette() {
	img := s.c.img
	cx, cy := float64(s.W)/2, float64(s.H)/2
	parallelRows(s.H, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := range s.W {
				rx, ry := (float64(x)-cx)/cx, (float64(y)-cy)/cy
				k := 1 - 0.28*clamp((math.Sqrt(rx*rx+ry*ry)-0.7)/0.7, 0, 1)
				o := img.PixOffset(x, y)
				for c := range 3 {
					img.Pix[o+c] = uint8(float64(img.Pix[o+c]) * k)
				}
			}
		}
	})
}
