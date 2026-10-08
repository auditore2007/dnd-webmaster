package atlas

import (
	"image"
	"image/color"
	"math"

	"golang.org/x/image/vector"
)

// Векторное рисование со сглаживанием поверх image.RGBA: многоугольники, линии переменной толщины, пунктир.

type pt struct{ x, y float64 }

type canvas struct{ img *image.RGBA }

func (c canvas) bounds() image.Rectangle { return c.img.Bounds() }

// fill закрашивает объединение многоугольников (все приводятся к одному направлению обхода, чтобы не вычитались).
func (c canvas) fill(polys [][]pt, col color.RGBA, alpha float64) {
	if len(polys) == 0 || alpha <= 0 {
		return
	}
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, p := range polys {
		for _, q := range p {
			minX, minY = math.Min(minX, q.x), math.Min(minY, q.y)
			maxX, maxY = math.Max(maxX, q.x), math.Max(maxY, q.y)
		}
	}
	r := image.Rect(int(math.Floor(minX))-1, int(math.Floor(minY))-1, int(math.Ceil(maxX))+1, int(math.Ceil(maxY))+1).Intersect(c.bounds())
	if r.Empty() {
		return
	}
	z := vector.NewRasterizer(r.Dx(), r.Dy())
	ox, oy := float64(r.Min.X), float64(r.Min.Y)
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
				z.MoveTo(float32(q.x-ox), float32(q.y-oy))
			} else {
				z.LineTo(float32(q.x-ox), float32(q.y-oy))
			}
		}
		z.ClosePath()
	}
	src := image.NewUniform(color.NRGBA{col.R, col.G, col.B, uint8(clamp(alpha, 0, 1) * 255)})
	z.Draw(c.img, r, src, image.Point{})
}

func area(p []pt) float64 {
	s := 0.0
	for i := range p {
		a, b := p[i], p[(i+1)%len(p)]
		s += a.x*b.y - b.x*a.y
	}
	return s / 2
}

func circle(x, y, r float64) []pt {
	n := max(8, min(32, int(r*3)))
	out := make([]pt, n)
	for i := range out {
		a := 2 * math.Pi * float64(i) / float64(n)
		out[i] = pt{x + math.Cos(a)*r, y + math.Sin(a)*r}
	}
	return out
}

func ellipse(x, y, rx, ry float64) []pt {
	out := circle(0, 0, 1)
	for i := range out {
		out[i] = pt{x + out[i].x*rx, y + out[i].y*ry}
	}
	return out
}

// strokePolys – линия с толщиной width[i] в точке i (одна толщина – если width длины 1), скруглённые стыки.
func strokePolys(line []pt, width []float64) [][]pt {
	wAt := func(i int) float64 {
		if len(width) == 1 {
			return width[0]
		}
		return width[i]
	}
	var polys [][]pt
	for i := 0; i+1 < len(line); i++ {
		a, b := line[i], line[i+1]
		dx, dy := b.x-a.x, b.y-a.y
		l := math.Hypot(dx, dy)
		if l < 1e-6 {
			continue
		}
		nx, ny := -dy/l, dx/l
		wa, wb := wAt(i)/2, wAt(i+1)/2
		polys = append(polys, []pt{{a.x + nx*wa, a.y + ny*wa}, {b.x + nx*wb, b.y + ny*wb}, {b.x - nx*wb, b.y - ny*wb}, {a.x - nx*wa, a.y - ny*wa}})
	}
	for i, p := range line {
		if w := wAt(i) / 2; w > 0.6 {
			polys = append(polys, circle(p.x, p.y, w))
		}
	}
	return polys
}

func (c canvas) stroke(line []pt, width float64, col color.RGBA, alpha float64) {
	c.fill(strokePolys(line, []float64{width}), col, alpha)
}

// dashes режет линию на штрихи длиной on с промежутками off.
func dashes(line []pt, on, off float64) [][]pt {
	if len(line) < 2 {
		return nil
	}
	var out [][]pt
	cur := []pt{line[0]}
	drawing, left := true, on
	for i := 0; i+1 < len(line); i++ {
		a, b := line[i], line[i+1]
		l := math.Hypot(b.x-a.x, b.y-a.y)
		t := 0.0
		for l-t > left {
			t += left
			p := pt{a.x + (b.x-a.x)*t/l, a.y + (b.y-a.y)*t/l}
			if drawing {
				out = append(out, append(cur, p))
				cur, left = nil, off
			} else {
				cur, left = []pt{p}, on
			}
			drawing = !drawing
		}
		left -= l - t
		if drawing {
			cur = append(cur, b)
		}
	}
	if drawing && len(cur) > 1 {
		out = append(out, cur)
	}
	return out
}

// chaikin сглаживает ломаную, сохраняя концы.
func chaikin(line []pt, iterations int) []pt {
	for range iterations {
		if len(line) < 3 {
			return line
		}
		out := []pt{line[0]}
		for i := 0; i+1 < len(line); i++ {
			a, b := line[i], line[i+1]
			out = append(out, pt{a.x*0.75 + b.x*0.25, a.y*0.75 + b.y*0.25}, pt{a.x*0.25 + b.x*0.75, a.y*0.25 + b.y*0.75})
		}
		line = append(out, line[len(line)-1])
	}
	return line
}

func rgb(r, g, b uint8) color.RGBA { return color.RGBA{r, g, b, 255} }

func shade(c color.RGBA, k float64) color.RGBA {
	f := func(v uint8) uint8 { return uint8(clamp(float64(v)*k, 0, 255)) }
	return color.RGBA{f(c.R), f(c.G), f(c.B), 255}
}

func mix(a, b color.RGBA, t float64) color.RGBA {
	t = clamp(t, 0, 1)
	m := func(p, q uint8) uint8 { return uint8(float64(p) + (float64(q)-float64(p))*t) }
	return color.RGBA{m(a.R, b.R), m(a.G, b.G), m(a.B, b.B), 255}
}

func (c canvas) blend(x, y int, col color.RGBA, a float64) {
	if !(image.Point{x, y}.In(c.bounds())) || a <= 0 {
		return
	}
	i := c.img.PixOffset(x, y)
	p := c.img.Pix[i : i+3 : i+3]
	p[0] = uint8(float64(p[0]) + (float64(col.R)-float64(p[0]))*a)
	p[1] = uint8(float64(p[1]) + (float64(col.G)-float64(p[1]))*a)
	p[2] = uint8(float64(p[2]) + (float64(col.B)-float64(p[2]))*a)
}

// distance – расстояние (в пикселях, фаска 3-4) от каждого пикселя до ближайшего, где src истинно.
func distance(src []bool, w, h int) []float32 {
	const inf = 1 << 20
	d := make([]int32, w*h)
	for i, s := range src {
		if !s {
			d[i] = inf
		}
	}
	relax := func(i, j int, k int32) {
		if v := d[j] + k; v < d[i] {
			d[i] = v
		}
	}
	for y := range h {
		for x := range w {
			i := y*w + x
			if x > 0 {
				relax(i, i-1, 3)
			}
			if y > 0 {
				relax(i, i-w, 3)
				if x > 0 {
					relax(i, i-w-1, 4)
				}
				if x < w-1 {
					relax(i, i-w+1, 4)
				}
			}
		}
	}
	for y := h - 1; y >= 0; y-- {
		for x := w - 1; x >= 0; x-- {
			i := y*w + x
			if x < w-1 {
				relax(i, i+1, 3)
			}
			if y < h-1 {
				relax(i, i+w, 3)
				if x < w-1 {
					relax(i, i+w+1, 4)
				}
				if x > 0 {
					relax(i, i+w-1, 4)
				}
			}
		}
	}
	out := make([]float32, w*h)
	for i, v := range d {
		out[i] = float32(v) / 3
	}
	return out
}
