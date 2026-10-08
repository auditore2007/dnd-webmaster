package atlas

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"testing"
)

type pcg struct{ r *rand.Rand }

func (p pcg) Intn(n int) int { return p.r.IntN(n) }

func seeded(seed uint64) Rand { return pcg{rand.New(rand.NewPCG(seed, seed^0x9e37))} }

func testPeoples() []People {
	n := 0
	name := func() string { n++; return fmt.Sprintf("Город%d", n) }
	state := func(c string) string { return "Королевство " + c }
	return []People{
		{Race: "human", Likes: []Terrain{Grass, Coast}, TownName: name, StateName: state},
		{Race: "elf", Likes: []Terrain{Forest, Jungle}, TownName: name, StateName: state},
		{Race: "dwarf", Likes: []Terrain{Mountain, Hills}, TownName: name, StateName: state},
		{Race: "lizardfolk", Likes: []Terrain{Swamp}, TownName: name, StateName: state},
	}
}

func generate(t *testing.T, seed uint64, template string) *World {
	t.Helper()
	w, err := Generate(Opts{W: 900, H: 560, Template: template, Title: "Тест", SeaName: "Море", Peoples: testPeoples(), R: seeded(seed)})
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func TestGenerateIsReproducible(t *testing.T) {
	a, b := generate(t, 3, ""), generate(t, 3, "")
	if !slices.Equal(a.H, b.H) || !slices.Equal(a.Burgs, b.Burgs) || a.Template != b.Template {
		t.Error("одно зерно – один и тот же мир")
	}
}

func TestEveryPeopleGetsCapitalOnLand(t *testing.T) {
	for seed := uint64(1); seed <= 5; seed++ {
		w := generate(t, seed, "")
		if len(w.States) != len(testPeoples()) {
			t.Fatalf("seed %d: государств %d", seed, len(w.States))
		}
		for _, s := range w.States {
			if c := w.Burgs[s.Capital]; c.Kind != "capital" || c.Race != s.Race {
				t.Errorf("seed %d: столица %s – %+v", seed, s.Race, c)
			}
		}
		for _, b := range w.Burgs {
			if !w.land(b.Cell) {
				t.Errorf("seed %d: %s стоит в воде", seed, b.Name)
			}
		}
	}
}

func TestAllTemplatesBuild(t *testing.T) {
	for _, id := range templateOrder {
		g := newGrid(900, 560, 3000, rng{seeded(7)})
		h, err := buildHeights(g, rng{seeded(7)}, id)
		if err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		land := 0
		for _, v := range h {
			if v > 100 {
				t.Fatalf("%s: высота %d", id, v)
			}
			if v >= seaLevel {
				land++
			}
		}
		if land == 0 {
			t.Errorf("%s: нет суши", id)
		}
	}
	if _, err := buildHeights(newGrid(900, 560, 3000, rng{seeded(1)}), rng{seeded(1)}, "нет"); err == nil {
		t.Error("неизвестный шаблон должен отклоняться")
	}
}

func TestRiversFlowDownhill(t *testing.T) {
	w := generate(t, 4, "continents")
	if len(w.Rivers) == 0 {
		t.Fatal("нет рек")
	}
	for _, r := range w.Rivers {
		for i := 1; i < len(r.Cells); i++ {
			if w.alt[r.Cells[i]] > w.alt[r.Cells[i-1]] {
				t.Fatalf("река течёт вверх: %v", r.Cells)
			}
		}
	}
}

func TestRenderKeepsSizeAndBurgsAshore(t *testing.T) {
	w := generate(t, 2, "")
	img := w.Render(seeded(9))
	if b := img.Bounds(); b.Dx() != 900 || b.Dy() != 560 {
		t.Fatalf("размер %v", b)
	}
	if !slices.ContainsFunc(img.Pix, func(v uint8) bool { return v != 0 }) {
		t.Error("пустая картинка")
	}
	for _, b := range w.Burgs {
		if b.X < 0 || b.Y < 0 || b.X >= 900 || b.Y >= 560 {
			t.Errorf("%s вне карты", b.Name)
		}
	}
}

func TestDashesAndDistance(t *testing.T) {
	d := dashes([]pt{{0, 0}, {30, 0}}, 5, 5)
	if len(d) != 3 || math.Abs(d[0][len(d[0])-1].x-5) > 1e-9 || math.Abs(d[1][0].x-10) > 1e-9 {
		t.Errorf("пунктир: %v", d)
	}
	src := make([]bool, 25)
	src[12] = true // центр 5×5
	dist := distance(src, 5, 5)
	if dist[12] != 0 || dist[13] != 1 || dist[0] < 2.6 || dist[0] > 3 {
		t.Errorf("расстояния: %v", dist)
	}
}
