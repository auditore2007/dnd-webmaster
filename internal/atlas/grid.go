package atlas

import "math"

// grid – сетка клеток как у Azgaar: квадратная решётка со случайно сдвинутыми точками, клетки по строкам.
// Клетка Вороного – ближайшая точка среди соседних 3×3 квадратов, поэтому триангуляция не нужна: соседи – 8 клеток вокруг.
type grid struct {
	w, h   int
	s      float64 // сторона квадрата в пикселях
	nx, ny int
	px, py []float64
	nb     [][]int32
	border []bool
}

func newGrid(w, h, cells int, r rng) *grid {
	s := math.Sqrt(float64(w*h) / float64(cells))
	nx, ny := int(math.Ceil(float64(w)/s)), int(math.Ceil(float64(h)/s))
	g := &grid{w: w, h: h, s: s, nx: nx, ny: ny}
	n := nx * ny
	g.px, g.py = make([]float64, n), make([]float64, n)
	g.nb, g.border = make([][]int32, n), make([]bool, n)
	jitter := s * 0.45
	for j := range ny {
		for i := range nx {
			c := j*nx + i
			g.px[c] = clamp((float64(i)+0.5)*s+r.between(-jitter, jitter), 0, float64(w-1))
			g.py[c] = clamp((float64(j)+0.5)*s+r.between(-jitter, jitter), 0, float64(h-1))
			g.border[c] = i == 0 || j == 0 || i == nx-1 || j == ny-1
			for dj := -1; dj <= 1; dj++ {
				for di := -1; di <= 1; di++ {
					x, y := i+di, j+dj
					if (di != 0 || dj != 0) && x >= 0 && y >= 0 && x < nx && y < ny {
						g.nb[c] = append(g.nb[c], int32(y*nx+x))
					}
				}
			}
		}
	}
	return g
}

func (g *grid) size() int { return g.nx * g.ny }

// find – клетка квадрата, в который попадает точка.
func (g *grid) find(x, y float64) int {
	i := min(g.nx-1, max(0, int(x/g.s)))
	j := min(g.ny-1, max(0, int(y/g.s)))
	return j*g.nx + i
}

// nearest – клетка Вороного: ближайшая точка среди 3×3 квадратов вокруг.
func (g *grid) nearest(x, y float64) int {
	c := g.find(x, y)
	best, bd := c, math.MaxFloat64
	for _, n := range append(g.nb[c], int32(c)) {
		if d := sq(g.px[n]-x) + sq(g.py[n]-y); d < bd {
			best, bd = int(n), d
		}
	}
	return best
}

func (g *grid) dist(a, b int) float64 { return math.Hypot(g.px[a]-g.px[b], g.py[a]-g.py[b]) }

func sq(v float64) float64 { return v * v }

// blobPower и linePower – как быстро гаснут холмы и хребты: чем больше клеток, тем медленнее (таблицы Azgaar).
func blobPower(cells int) float64 {
	return interpolate(cells, []int{1000, 2000, 5000, 10000, 20000, 30000, 40000, 50000},
		[]float64{0.93, 0.95, 0.97, 0.98, 0.99, 0.991, 0.993, 0.994})
}

func linePower(cells int) float64 {
	return interpolate(cells, []int{1000, 2000, 5000, 10000, 20000, 30000, 40000, 50000},
		[]float64{0.75, 0.77, 0.79, 0.81, 0.82, 0.83, 0.84, 0.86})
}

func interpolate(x int, xs []int, ys []float64) float64 {
	if x <= xs[0] {
		return ys[0]
	}
	for i := 1; i < len(xs); i++ {
		if x <= xs[i] {
			t := float64(x-xs[i-1]) / float64(xs[i]-xs[i-1])
			return ys[i-1] + (ys[i]-ys[i-1])*t
		}
	}
	return ys[len(ys)-1]
}
