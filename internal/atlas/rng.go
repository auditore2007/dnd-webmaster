package atlas

import (
	"math"
	"strconv"
	"strings"
)

// Rand – источник случайностей (подходит dice.Roller). От него зависит вся карта: одно зерно – один мир.
type Rand interface{ Intn(n int) int }

type rng struct{ r Rand }

func (g rng) float() float64 { return float64(g.r.Intn(1<<30)) / (1 << 30) }

func (g rng) between(a, b float64) float64 { return a + (b-a)*g.float() }

// intIn – целое из [a; b] включительно, как rand(min, max) у Azgaar.
func (g rng) intIn(a, b int) int {
	if b <= a {
		return a
	}
	return a + g.r.Intn(b-a+1)
}

func (g rng) chance(p float64) bool { return g.float() < p }

// number читает число из шаблона: «3», «1.5» (дробная часть – вероятность ещё одной единицы) или «2-4» (случайное из диапазона).
func (g rng) number(s string) float64 {
	if a, b, ok := strings.Cut(s, "-"); ok && a != "" {
		lo, _ := strconv.ParseFloat(a, 64)
		hi, _ := strconv.ParseFloat(b, 64)
		return float64(g.intIn(int(lo), int(hi)))
	}
	v, _ := strconv.ParseFloat(s, 64)
	whole := math.Trunc(v)
	if g.chance(v - whole) {
		whole++
	}
	return whole
}

// noise – сглаженный шум на решётке случайных чисел; fbm складывает октавы.
type noise struct {
	perm [512]int32
	val  [256]float64
}

func newNoise(r rng) *noise {
	n := &noise{}
	p := make([]int32, 256)
	for i := range p {
		p[i] = int32(i)
		n.val[i] = r.float()
	}
	for i := 255; i > 0; i-- {
		j := r.r.Intn(i + 1)
		p[i], p[j] = p[j], p[i]
	}
	for i := range n.perm {
		n.perm[i] = p[i&255]
	}
	return n
}

func (n *noise) lattice(x, y int) float64 { return n.val[n.perm[int(n.perm[x&255])+y&255]&255] }

func (n *noise) at(x, y float64) float64 {
	x0, y0 := math.Floor(x), math.Floor(y)
	tx, ty := x-x0, y-y0
	tx, ty = tx*tx*(3-2*tx), ty*ty*(3-2*ty)
	ix, iy := int(x0), int(y0)
	a, b := n.lattice(ix, iy), n.lattice(ix+1, iy)
	c, d := n.lattice(ix, iy+1), n.lattice(ix+1, iy+1)
	return (a*(1-tx)+b*tx)*(1-ty) + (c*(1-tx)+d*tx)*ty
}

func (n *noise) fbm(x, y float64, octaves int) float64 {
	sum, amp, norm := 0.0, 1.0, 0.0
	for range octaves {
		sum += amp * n.at(x, y)
		norm += amp
		amp *= 0.5
		x, y = x*2.03+17.1, y*2.03+9.7
	}
	return sum / norm
}

func clamp(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

// ridged – «гребневый» шум: острые хребты и долины между ними (0..1).
func (n *noise) ridged(x, y float64, octaves int) float64 {
	sum, amp, norm := 0.0, 1.0, 0.0
	for range octaves {
		r := 1 - math.Abs(2*n.at(x, y)-1)
		sum += amp * r * r
		norm += amp
		amp *= 0.5
		x, y = x*2.07+5.3, y*2.07+11.9
	}
	return sum / norm
}
