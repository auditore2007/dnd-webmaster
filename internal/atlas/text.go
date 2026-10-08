package atlas

import (
	_ "embed"
	"image"
	"image/color"
	"math"
	"sort"
	"strings"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// Надписи шрифтом Alegreya SC (SIL OFL, см. fonts/OFL.txt) со светлым ореолом, чтобы читались на любом фоне.

//go:embed fonts/AlegreyaSC-Regular.ttf
var fontRegular []byte

//go:embed fonts/AlegreyaSC-Bold.ttf
var fontBold []byte

//go:embed fonts/AlegreyaSC-Italic.ttf
var fontItalic []byte

var fonts struct {
	once                  sync.Once
	regular, bold, italic *opentype.Font
	err                   error
}

func loadFonts() error {
	fonts.once.Do(func() {
		for _, f := range []struct {
			dst  **opentype.Font
			data []byte
		}{{&fonts.regular, fontRegular}, {&fonts.bold, fontBold}, {&fonts.italic, fontItalic}} {
			if *f.dst, fonts.err = opentype.Parse(f.data); fonts.err != nil {
				return
			}
		}
	})
	return fonts.err
}

type label struct {
	text    string
	font    *opentype.Font
	size    float64
	spacing float64
	x, y    float64 // начало базовой линии
	col     color.RGBA
	alpha   float64
	halo    color.RGBA
	haloR   int
	haloA   float64
}

func newFace(f *opentype.Font, size float64) font.Face {
	face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingNone})
	if err != nil {
		return nil
	}
	return face
}

// measure – ширина надписи и высота над/под базовой линией.
func (l label) measure() (w, ascent, descent float64) {
	face := newFace(l.font, l.size)
	if face == nil {
		return 0, 0, 0
	}
	defer face.Close()
	n := 0
	for _, r := range l.text {
		adv, _ := face.GlyphAdvance(r)
		w += float64(adv) / 64
		n++
	}
	m := face.Metrics()
	return w + l.spacing*float64(max(0, n-1)), float64(m.Ascent) / 64, float64(m.Descent) / 64
}

func (l label) box() image.Rectangle {
	w, a, d := l.measure()
	return image.Rect(int(l.x)-2, int(l.y-a*0.8)-2, int(l.x+w)+2, int(l.y+d)+2)
}

func (p *painter) drawLabel(l label) {
	face := newFace(l.font, l.size)
	if face == nil {
		return
	}
	defer face.Close()
	w, a, d := l.measure()
	pad := l.haloR + 2
	ox, oy := int(math.Floor(l.x))-pad, int(math.Floor(l.y-a))-pad
	mask := image.NewAlpha(image.Rect(0, 0, int(w)+2*pad+2, int(a+d)+2*pad+2))
	dr := font.Drawer{Dst: mask, Src: image.Opaque, Face: face}
	dot := l.x - float64(ox)
	for _, r := range l.text {
		dr.Dot = fixed.Point26_6{X: fixed.Int26_6(dot * 64), Y: fixed.Int26_6((l.y - float64(oy)) * 64)}
		dr.DrawString(string(r))
		adv, _ := face.GlyphAdvance(r)
		dot += float64(adv)/64 + l.spacing
	}
	b := mask.Bounds()
	if l.haloR > 0 {
		halo := dilate(mask, l.haloR)
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				if v := halo.AlphaAt(x, y).A; v > 0 {
					p.c.blend(ox+x, oy+y, l.halo, float64(v)/255*l.haloA)
				}
			}
		}
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if v := mask.AlphaAt(x, y).A; v > 0 {
				p.c.blend(ox+x, oy+y, l.col, float64(v)/255*l.alpha)
			}
		}
	}
}

// dilate – расширение маски на r пикселей (по кругу) с мягким краем.
func dilate(m *image.Alpha, r int) *image.Alpha {
	b := m.Bounds()
	out := image.NewAlpha(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			best := 0.0
			for dy := -r; dy <= r; dy++ {
				for dx := -r; dx <= r; dx++ {
					d := math.Hypot(float64(dx), float64(dy))
					if d > float64(r)+0.5 {
						continue
					}
					v := float64(m.AlphaAt(x+dx, y+dy).A) * clamp(float64(r)+1-d, 0, 1)
					best = math.Max(best, v)
				}
			}
			out.SetAlpha(x, y, color.Alpha{uint8(best)})
		}
	}
	return out
}

// labelsLayout – места для надписей: сначала названия государств и моря, потом поселений (у значка справа, слева, сверху или снизу).
func (p *painter) labelsLayout() {
	if loadFonts() != nil {
		return
	}
	p.seaLabel()
	p.stateLabels()
	for _, b := range p.w.Burgs {
		p.burgLabel(b)
	}
}

func (p *painter) place(l label) bool {
	box := l.box()
	if !p.free(box) {
		return false
	}
	p.obstacles = append(p.obstacles, box)
	p.labels = append(p.labels, l)
	return true
}

func (p *painter) burgLabel(b Burg) {
	size := map[string]float64{"capital": 17, "city": 14.5, "castle": 12.5, "town": 13, "port": 13, "village": 11.5}[b.Kind] * p.scale
	f := fonts.regular
	if b.Kind == "capital" || b.Kind == "city" {
		f = fonts.bold
	}
	l := label{text: b.Name, font: f, size: size, col: inkColor, alpha: 1, halo: paperColor, haloR: 2, haloA: 0.8}
	w, a, _ := l.measure()
	s := p.burgSize(b.Kind)
	for _, pos := range [][2]float64{
		{b.X + s*0.8 + 3, b.Y - s*0.3 + a*0.4},
		{b.X - s*0.8 - 3 - w, b.Y - s*0.3 + a*0.4},
		{b.X - w/2, b.Y - s*1.15 - 3},
		{b.X - w/2, b.Y + s*0.35 + a*0.85},
	} {
		l.x, l.y = pos[0], pos[1]
		if p.place(l) {
			return
		}
	}
	if b.Kind == "capital" { // столицу подписываем всегда
		l.x, l.y = b.X+s*0.8+3, b.Y-s*0.3+a*0.4
		if l.x+w > float64(p.W)-p.frameWidth()-4 {
			l.x = b.X - s*0.8 - 3 - w
		}
		p.labels = append(p.labels, l)
	}
}

// stateLabels – название государства крупно вразрядку в самой «глубокой» его клетке.
func (p *painter) stateLabels() {
	w := p.w
	depth := make([]int, w.g.size())
	var queue []int
	for c := range w.H {
		depth[c] = -1
		if !w.land(c) || w.State[c] < 0 {
			depth[c] = 0
			queue = append(queue, c)
			continue
		}
		for _, n := range w.g.nb[c] {
			if w.State[n] != w.State[c] || !w.land(int(n)) {
				depth[c] = 0
				queue = append(queue, c)
				break
			}
		}
	}
	for qi := 0; qi < len(queue); qi++ {
		for _, n := range w.g.nb[queue[qi]] {
			if depth[n] < 0 {
				depth[n] = depth[queue[qi]] + 1
				queue = append(queue, int(n))
			}
		}
	}
	for si, st := range w.States {
		if st.Cells < 12 {
			continue
		}
		var cells []int
		for c := range w.H {
			if int(w.State[c]) == si && depth[c] > 0 {
				cells = append(cells, c)
			}
		}
		sort.SliceStable(cells, func(a, b int) bool { return depth[cells[a]] > depth[cells[b]] })
		size := clamp(math.Sqrt(float64(st.Cells))*1.1, 15, 38) * p.scale
		l := label{text: strings.ToUpper(st.Name), font: fonts.bold, col: shade(st.Color, 0.55), alpha: 0.78,
			halo: paperColor, haloR: 2, haloA: 0.45}
		for ; size >= 13*p.scale; size *= 0.85 {
			l.size, l.spacing = size, size*0.3
			tw, a, _ := l.measure()
			if p.tryAround(&l, cells, tw, a) {
				break
			}
		}
	}
}

func (p *painter) tryAround(l *label, cells []int, tw, a float64) bool {
	for k, c := range cells {
		if k > 40 {
			break
		}
		for _, dy := range []float64{0, -1.2, 1.2} {
			l.x = p.w.g.px[c] - tw/2
			l.y = p.w.g.py[c] + a*0.35 + dy*a
			if p.place(*l) {
				return true
			}
		}
	}
	return false
}

// seaLabel – название моря курсивом там, где дальше всего от берегов.
func (p *painter) seaLabel() {
	if p.w.SeaName == "" {
		return
	}
	type spot struct {
		x, y int
		d    float32
	}
	var spots []spot
	for y := 0; y < p.H; y += 8 {
		for x := 0; x < p.W; x += 8 {
			if i := y*p.W + x; p.kind[i] == pxOcean {
				edge := min(x, y, p.W-x, p.H-y)
				spots = append(spots, spot{x, y, min(p.dLand[i], float32(edge)*0.8)})
			}
		}
	}
	sort.SliceStable(spots, func(a, b int) bool { return spots[a].d > spots[b].d })
	size := 24 * p.scale
	l := label{text: p.w.SeaName, font: fonts.italic, size: size, spacing: size * 0.25, col: rgb(30, 60, 88), alpha: 0.7}
	tw, a, _ := l.measure()
	for k, s := range spots {
		if k > 400 || float64(s.d) < tw/5 {
			return
		}
		l.x, l.y = float64(s.x)-tw/2, float64(s.y)+a*0.35
		if p.place(l) {
			return
		}
	}
}

func (p *painter) drawLabels() {
	for _, l := range p.labels {
		p.drawLabel(l)
	}
}
