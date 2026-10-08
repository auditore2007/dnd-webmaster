package atlas

import (
	"fmt"
	"image"
	"math"
)

// Оформление атласа: картуш с названием, роза ветров и масштабная линейка – в самых «морских» углах; рамка с делениями.

const milesPerPx = 0.25 // средняя карта мира (3200 px) – около 800 миль в ширину

func (p *painter) frameWidth() float64 { return 16 * p.scale }

// layoutDecor выбирает углы: картуш – где больше всего воды, роза ветров – в следующем, масштаб – внизу в свободном.
func (p *painter) layoutDecor() {
	inset := int(p.frameWidth() + 12*p.scale)
	cw, ch := int(380*p.scale), int(96*p.scale)
	if loadFonts() == nil && p.w.Title != "" {
		l := label{text: p.w.Title, font: fonts.bold, size: 34 * p.scale, spacing: 2 * p.scale}
		tw, _, _ := l.measure()
		cw = max(cw, int(tw+110*p.scale))
	}
	cw = min(cw, p.W/2)
	ks := int(130 * p.scale)
	corner := func(k, w, h int) image.Rectangle {
		x, y := inset, inset
		if k%2 == 1 {
			x = p.W - inset - w
		}
		if k >= 2 {
			y = p.H - inset - h
		}
		return image.Rect(x, y, x+w, y+h)
	}
	wet := func(r image.Rectangle) float64 {
		n, wv := 0, 0
		for y := r.Min.Y; y < r.Max.Y; y += 4 {
			for x := r.Min.X; x < r.Max.X; x += 4 {
				n++
				if p.kind[p.at(x, y)] != pxLand {
					wv++
				}
			}
		}
		return float64(wv) / float64(max(1, n))
	}
	best := 0
	for k := range 4 {
		if wet(corner(k, cw, ch)) > wet(corner(best, cw, ch))+0.05 {
			best = k
		}
	}
	p.cartouche = corner(best, cw, ch)
	cb := -1
	for k := range 4 {
		if k != best && (cb < 0 || wet(corner(k, ks, ks)) > wet(corner(cb, ks, ks))) {
			cb = k
		}
	}
	p.compassBox = corner(cb, ks, ks)
	p.obstacles = append(p.obstacles, p.cartouche.Inset(-6), p.compassBox.Inset(-6), p.scaleBarBox().Inset(-6))
}

func (p *painter) scaleBarBox() image.Rectangle {
	inset := int(p.frameWidth() + 14*p.scale)
	w, h := int(250*p.scale), int(40*p.scale)
	x := inset
	if p.cartouche.Min.X < p.W/2 && p.cartouche.Min.Y > p.H/2 || p.compassBox.Min.X < p.W/2 && p.compassBox.Min.Y > p.H/2 {
		x = p.W - inset - w // левый нижний угол занят
		if p.cartouche.Min.X > p.W/2 && p.cartouche.Min.Y > p.H/2 || p.compassBox.Min.X > p.W/2 && p.compassBox.Min.Y > p.H/2 {
			x = p.W/2 - w/2
		}
	}
	return image.Rect(x, p.H-inset-h, x+w, p.H-inset)
}

func (p *painter) decor() {
	if loadFonts() != nil {
		return
	}
	if p.w.Title != "" {
		p.drawCartouche()
	}
	p.drawCompass()
	p.drawScale()
	p.drawFrame()
}

func rectPts(r image.Rectangle) []pt {
	x0, y0, x1, y1 := float64(r.Min.X), float64(r.Min.Y), float64(r.Max.X), float64(r.Max.Y)
	return []pt{{x0, y0}, {x1, y0}, {x1, y1}, {x0, y1}}
}

func closed(p []pt) []pt { return append(append([]pt(nil), p...), p[0]) }

func (p *painter) drawCartouche() {
	r := p.cartouche
	s := p.scale
	shadow := r.Add(image.Pt(int(4*s), int(5*s)))
	p.c.fill([][]pt{rectPts(shadow)}, rgb(40, 30, 20), 0.25)
	// свитки по краям
	cy := float64(r.Min.Y+r.Max.Y) / 2
	for _, x := range []float64{float64(r.Min.X), float64(r.Max.X)} {
		p.c.fill([][]pt{ellipse(x, cy, 14*s, float64(r.Dy())/2)}, rgb(214, 194, 150), 1)
		p.c.stroke(closed(ellipse(x, cy, 14*s, float64(r.Dy())/2)), 1.2*s, inkColor, 0.9)
		p.c.stroke(closed(ellipse(x, cy, 6*s, float64(r.Dy())/2-8*s)), 0.9*s, inkColor, 0.6)
	}
	p.c.fill([][]pt{rectPts(r)}, rgb(242, 230, 200), 1)
	p.c.stroke(closed(rectPts(r)), 2*s, inkColor, 0.95)
	p.c.stroke(closed(rectPts(r.Inset(int(5*s)))), 0.9*s, inkColor, 0.8)
	l := label{text: p.w.Title, font: fonts.bold, size: 34 * s, spacing: 2 * s, col: inkColor, alpha: 1}
	tw, a, _ := l.measure()
	l.x, l.y = float64(r.Min.X+r.Max.X)/2-tw/2, float64(r.Min.Y)+float64(r.Dy())*0.42+a*0.35
	p.drawLabel(l)
	sub := label{text: "карта земель и народов", font: fonts.italic, size: 15 * s, spacing: 1.5 * s, col: rgb(96, 70, 48), alpha: 1}
	sw, sa, _ := sub.measure()
	sub.x, sub.y = float64(r.Min.X+r.Max.X)/2-sw/2, float64(r.Max.Y)-float64(r.Dy())*0.2+sa*0.2
	p.drawLabel(sub)
	mid := float64(r.Min.X+r.Max.X) / 2
	y := float64(r.Max.Y) - float64(r.Dy())*0.36
	p.c.stroke([]pt{{mid - tw*0.4, y}, {mid - 8*s, y}}, 0.8*s, inkColor, 0.7)
	p.c.stroke([]pt{{mid + 8*s, y}, {mid + tw*0.4, y}}, 0.8*s, inkColor, 0.7)
	p.c.fill([][]pt{{{mid - 5*s, y}, {mid, y - 3*s}, {mid + 5*s, y}, {mid, y + 3*s}}}, inkColor, 0.8)
}

// drawCompass – роза ветров: 8 главных лучей (светлая и тёмная половины), 8 малых, кольцо и «С» на севере.
func (p *painter) drawCompass() {
	r := p.compassBox
	cx, cy := float64(r.Min.X+r.Max.X)/2, float64(r.Min.Y+r.Max.Y)/2+6*p.scale
	R := float64(r.Dx()) * 0.4
	ring := closed(circle(cx, cy, R*0.62))
	p.c.fill([][]pt{circle(cx, cy, R*0.66)}, rgb(242, 230, 200), 0.6)
	p.c.stroke(ring, 1.1*p.scale, inkColor, 0.8)
	p.c.stroke(closed(circle(cx, cy, R*0.55)), 0.7*p.scale, inkColor, 0.6)
	ray := func(angle, length, width float64) {
		tip := pt{cx + math.Sin(angle)*length, cy - math.Cos(angle)*length}
		l := pt{cx + math.Sin(angle-math.Pi/2)*width, cy - math.Cos(angle-math.Pi/2)*width}
		rr := pt{cx + math.Sin(angle+math.Pi/2)*width, cy - math.Cos(angle+math.Pi/2)*width}
		c := pt{cx, cy}
		p.c.fill([][]pt{{c, l, tip}}, rgb(244, 236, 214), 1)
		p.c.fill([][]pt{{c, tip, rr}}, inkColor, 0.9)
		p.c.stroke([]pt{l, tip, rr}, 0.8*p.scale, inkColor, 0.9)
	}
	for k := range 8 {
		ray(float64(k)*math.Pi/4+math.Pi/8, R*0.45, R*0.07)
	}
	for k := range 8 {
		length := R
		if k%2 == 1 {
			length = R * 0.68
		}
		ray(float64(k)*math.Pi/4, length, R*0.11)
	}
	p.c.fill([][]pt{circle(cx, cy, R*0.06)}, rgb(168, 58, 48), 1)
	n := label{text: "С", font: fonts.bold, size: 18 * p.scale, col: inkColor, alpha: 1, halo: paperColor, haloR: 2, haloA: 0.8}
	nw, _, _ := n.measure()
	n.x, n.y = cx-nw/2, cy-R-4*p.scale
	p.drawLabel(n)
}

func (p *painter) drawScale() {
	r := p.scaleBarBox()
	s := p.scale
	miles := 50.0
	for _, m := range []float64{50, 100, 200, 300, 500} {
		if m/milesPerPx <= float64(r.Dx())*0.85 {
			miles = m
		}
	}
	length := miles / milesPerPx
	x0, y0 := float64(r.Min.X)+8*s, float64(r.Max.Y)-14*s
	h := 6 * s
	for k := range 4 {
		seg := image.Rect(int(x0+length*float64(k)/4), int(y0), int(x0+length*float64(k+1)/4), int(y0+h))
		col := inkColor
		if k%2 == 1 {
			col = rgb(244, 236, 214)
		}
		p.c.fill([][]pt{rectPts(seg)}, col, 0.95)
	}
	p.c.stroke(closed([]pt{{x0, y0}, {x0 + length, y0}, {x0 + length, y0 + h}, {x0, y0 + h}}), 1*s, inkColor, 1)
	for k, txt := range []string{"0", fmt.Sprint(miles / 2), fmt.Sprintf("%g миль", miles)} {
		l := label{text: txt, font: fonts.regular, size: 12 * s, col: inkColor, alpha: 1, halo: paperColor, haloR: 2, haloA: 0.7}
		x := x0 + length*float64(k)/2
		if k < 2 {
			tw, _, _ := l.measure()
			x -= tw / 2
		} else {
			x -= 4 * s
		}
		l.x, l.y = x, y0-4*s
		p.drawLabel(l)
	}
}

// drawFrame – поле вокруг карты с двойной линией и чередующимися делениями, как у старинных карт.
func (p *painter) drawFrame() {
	f := p.frameWidth()
	W, H := float64(p.W), float64(p.H)
	band := rgb(212, 192, 150)
	p.c.fill([][]pt{{{0, 0}, {W, 0}, {W, f}, {0, f}}, {{0, H - f}, {W, H - f}, {W, H}, {0, H}},
		{{0, 0}, {f, 0}, {f, H}, {0, H}}, {{W - f, 0}, {W, 0}, {W, H}, {W - f, H}}}, band, 1)
	o := 3 * p.scale
	p.c.stroke(closed([]pt{{o, o}, {W - o, o}, {W - o, H - o}, {o, H - o}}), 1.6*p.scale, inkColor, 1)
	p.c.stroke(closed([]pt{{f, f}, {W - f, f}, {W - f, H - f}, {f, H - f}}), 1.6*p.scale, inkColor, 1)
	t0, t1 := f-7*p.scale, f-3*p.scale
	step := 40 * p.scale
	var ticks [][]pt
	for k, x := 0, f; x < W-f; k, x = k+1, x+step {
		if k%2 == 0 {
			x1 := math.Min(x+step, W-f)
			ticks = append(ticks, []pt{{x, t0}, {x1, t0}, {x1, t1}, {x, t1}}, []pt{{x, H - t1}, {x1, H - t1}, {x1, H - t0}, {x, H - t0}})
		}
	}
	for k, y := 0, f; y < H-f; k, y = k+1, y+step {
		if k%2 == 0 {
			y1 := math.Min(y+step, H-f)
			ticks = append(ticks, []pt{{t0, y}, {t1, y}, {t1, y1}, {t0, y1}}, []pt{{W - t1, y}, {W - t0, y}, {W - t0, y1}, {W - t1, y1}})
		}
	}
	p.c.fill(ticks, inkColor, 0.85)
	p.c.stroke(closed([]pt{{t0, t0}, {W - t0, t0}, {W - t0, H - t0}, {t0, H - t0}}), 0.7*p.scale, inkColor, 0.8)
	p.c.stroke(closed([]pt{{t1, t1}, {W - t1, t1}, {W - t1, H - t1}, {t1, H - t1}}), 0.7*p.scale, inkColor, 0.8)
}
