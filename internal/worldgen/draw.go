package worldgen

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"math/rand/v2"

	"heroesbook/internal/dice"
)

// Простое рисование поверх image.RGBA: заливки, линии, круги, кодирование в data:URL.

type canvas struct{ img *image.RGBA }

func newCanvas(w, h int, bg color.RGBA) canvas {
	c := canvas{image.NewRGBA(image.Rect(0, 0, w, h))}
	c.fillRect(0, 0, w, h, bg)
	return c
}

func (c canvas) w() int { return c.img.Bounds().Dx() }
func (c canvas) h() int { return c.img.Bounds().Dy() }

func (c canvas) set(x, y int, col color.RGBA) {
	if x >= 0 && y >= 0 && x < c.w() && y < c.h() {
		c.img.SetRGBA(x, y, col)
	}
}

// blend смешивает цвет пикселя с col (a – доля нового цвета, 0..1).
func (c canvas) blend(x, y int, col color.RGBA, a float64) {
	if x < 0 || y < 0 || x >= c.w() || y >= c.h() {
		return
	}
	o := c.img.RGBAAt(x, y)
	mix := func(p, q uint8) uint8 { return uint8(float64(p)*(1-a) + float64(q)*a) }
	c.img.SetRGBA(x, y, color.RGBA{mix(o.R, col.R), mix(o.G, col.G), mix(o.B, col.B), 255})
}

func (c canvas) fillRect(x, y, w, h int, col color.RGBA) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			c.set(i, j, col)
		}
	}
}

func (c canvas) disc(cx, cy, r float64, col color.RGBA, a float64) {
	for y := int(cy - r - 1); y <= int(cy+r+1); y++ {
		for x := int(cx - r - 1); x <= int(cx+r+1); x++ {
			d := math.Hypot(float64(x)-cx, float64(y)-cy)
			if d <= r {
				c.blend(x, y, col, a*math.Min(1, r-d+0.5))
			}
		}
	}
}

// line – толстая линия; dash > 0 – пунктир с таким шагом; skip – где не рисовать (например, по воде).
func (c canvas) line(x0, y0, x1, y1, width float64, col color.RGBA, a float64, dash float64, skip func(x, y int) bool) {
	n := int(math.Hypot(x1-x0, y1-y0)) + 1
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		if dash > 0 && int(float64(i)/dash)%2 == 1 {
			continue
		}
		x, y := x0+(x1-x0)*t, y0+(y1-y0)*t
		if skip != nil && skip(int(x), int(y)) {
			continue
		}
		c.disc(x, y, width/2, col, a)
	}
}

// dataURL кодирует картинку в JPEG (карты большие – PNG вышел бы в разы тяжелее).
func (c canvas) dataURL(quality int) (string, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, c.img, &jpeg.Options{Quality: quality}); err != nil {
		return "", err
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// mountain – горный значок: левый склон освещён, правый в тени, контур тёмный.
func (c canvas) mountain(cx, by, size float64, base color.RGBA) {
	top := by - size
	for y := int(top); y <= int(by); y++ {
		t := (float64(y) - top) / size
		half := t * size * 0.75
		for x := int(cx - half); x <= int(cx+half); x++ {
			col := shade(base, 1.12)
			if float64(x) > cx {
				col = shade(base, 0.72)
			}
			c.blend(x, y, col, 0.95)
		}
	}
	ink := rgb(62, 48, 36)
	c.line(cx-size*0.75, by, cx, top, 1.3, ink, 0.85, 0, nil)
	c.line(cx, top, cx+size*0.75, by, 1.3, ink, 0.85, 0, nil)
}

// hill – холмик: дуга с тенью.
func (c canvas) hill(cx, cy, r float64) {
	ink := rgb(92, 82, 52)
	for a := math.Pi; a <= 2*math.Pi; a += 0.08 {
		c.disc(cx+math.Cos(a)*r, cy+math.Sin(a)*r*0.6, 0.7, ink, 0.8)
	}
}

func rgb(r, g, b uint8) color.RGBA { return color.RGBA{r, g, b, 255} }

func shade(c color.RGBA, k float64) color.RGBA {
	f := func(v uint8) uint8 { return uint8(math.Max(0, math.Min(255, float64(v)*k))) }
	return color.RGBA{f(c.R), f(c.G), f(c.B), 255}
}

func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	t = math.Max(0, math.Min(1, t))
	m := func(p, q uint8) uint8 { return uint8(float64(p) + (float64(q)-float64(p))*t) }
	return color.RGBA{m(a.R, b.R), m(a.G, b.G), m(a.B, b.B), 255}
}

// seeded – воспроизводимый генератор: один seed – одна и та же карта.
type seeded struct{ r *rand.Rand }

func (s seeded) Intn(n int) int { return s.r.IntN(n) }

// Seeded возвращает генератор случайностей с заданным зерном.
func Seeded(seed int64) dice.Roller {
	return seeded{rand.New(rand.NewPCG(uint64(seed), uint64(seed)^0x9e3779b97f4a7c15))}
}

// noise – значение шума в точке: решётка случайных чисел с плавной интерполяцией, несколько октав (fbm).
type noise struct {
	perm [512]int32
	val  [256]float64
}

func newNoise(r dice.Roller) *noise {
	n := &noise{}
	p := make([]int32, 256)
	for i := range p {
		p[i] = int32(i)
		n.val[i] = float64(r.Intn(1<<20)) / float64(1<<20)
	}
	for i := 255; i > 0; i-- {
		j := r.Intn(i + 1)
		p[i], p[j] = p[j], p[i]
	}
	for i := 0; i < 512; i++ {
		n.perm[i] = p[i&255]
	}
	return n
}

func (n *noise) lattice(x, y int) float64 { return n.val[n.perm[int(n.perm[x&255])+y&255]&255] }

func smooth(t float64) float64 { return t * t * (3 - 2*t) }

func (n *noise) at(x, y float64) float64 {
	x0, y0 := math.Floor(x), math.Floor(y)
	tx, ty := smooth(x-x0), smooth(y-y0)
	ix, iy := int(x0), int(y0)
	a, b := n.lattice(ix, iy), n.lattice(ix+1, iy)
	c, d := n.lattice(ix, iy+1), n.lattice(ix+1, iy+1)
	return (a*(1-tx)+b*tx)*(1-ty) + (c*(1-tx)+d*tx)*ty
}

// fbm – сумма октав: крупные материки и мелкие детали берегов.
func (n *noise) fbm(x, y float64, octaves int) float64 {
	sum, amp, norm := 0.0, 1.0, 0.0
	for i := 0; i < octaves; i++ {
		sum += amp * n.at(x, y)
		norm += amp
		amp *= 0.5
		x, y = x*2.03+17.1, y*2.03+9.7
	}
	return sum / norm
}
