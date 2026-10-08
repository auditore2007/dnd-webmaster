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

func rgb(r, g, b uint8) color.RGBA { return color.RGBA{r, g, b, 255} }

// seeded – воспроизводимый генератор: один seed – одна и та же карта.
type seeded struct{ r *rand.Rand }

func (s seeded) Intn(n int) int { return s.r.IntN(n) }

// Seeded возвращает генератор случайностей с заданным зерном.
func Seeded(seed int64) dice.Roller {
	return seeded{rand.New(rand.NewPCG(uint64(seed), uint64(seed)^0x9e3779b97f4a7c15))}
}
