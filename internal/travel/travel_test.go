package travel

import (
	"image"
	"image/color"
	"strings"
	"testing"
)

// grid – сетка из строк: '~' вода, '=' дорога, '.' равнина, 'T' лес, '^' горы.
func grid(rows ...string) *Grid {
	g := &Grid{Cols: len(rows[0]), Rows: len(rows), Step: 10}
	code := map[rune]uint8{'~': Water, '=': Road, '.': Plain, 'T': Forest, '^': Mountain}
	for _, r := range rows {
		for _, c := range r {
			g.Cells = append(g.Cells, code[c])
		}
	}
	return g
}

func TestEncodeDecode(t *testing.T) {
	g := grid("~.=", "T^.")
	back, err := Decode(g.Encode())
	if err != nil || back.Cols != 3 || back.Rows != 2 || string(back.Cells) != string(g.Cells) {
		t.Fatalf("сетка не совпала: %+v %v", back, err)
	}
	for _, bad := range []string{"", "travel1;2;2;8;AAA", "other;1;1;1;AA=="} {
		if _, err := Decode(bad); err == nil {
			t.Errorf("испорченная сетка %q должна отклоняться", bad)
		}
	}
}

func TestRoadIsFasterThanForest(t *testing.T) {
	road := grid("==========")
	forest := grid("TTTTTTTTTT")
	from, to := Point{5, 5}, Point{95, 5}
	r1, err := Plan(road, from, to, "normal", 1)
	if err != nil {
		t.Fatal(err)
	}
	r2, _ := Plan(forest, from, to, "normal", 1)
	if r1.Miles != r2.Miles || r2.Days <= r1.Days {
		t.Errorf("по лесу дольше: дорога %.2f дн., лес %.2f дн.", r1.Days, r2.Days)
	}
	if r1.Legs[0].Terrain != "дорога" || r1.Path[0] != from || r1.Path[len(r1.Path)-1] != to {
		t.Errorf("маршрут: %+v", r1)
	}
	fast, _ := Plan(road, from, to, "fast", 1)
	if fast.Days > r1.Days {
		t.Error("быстрым темпом не дольше")
	}
	if _, err := Plan(road, from, to, "бегом", 1); err == nil {
		t.Error("неизвестный темп – ошибка")
	}
}

func TestPathGoesAroundWater(t *testing.T) {
	g := grid(
		".~~~~.",
		".~~~~.",
		"......",
	)
	r, err := Plan(g, Point{5, 5}, Point{55, 5}, "normal", 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range r.Path {
		if g.at(int(p.X)/10, int(p.Y)/10) == Water {
			t.Fatalf("путь идёт по воде: %+v", r.Path)
		}
	}
	if _, err := Plan(g, Point{5, 5}, Point{25, 5}, "normal", 1); err == nil || !strings.Contains(err.Error(), "вода") {
		t.Errorf("в воду не идём: %v", err)
	}
	island := grid(".~.")
	if _, err := Plan(island, Point{5, 5}, Point{25, 5}, "normal", 1); err == nil {
		t.Error("на остров по суше не попасть")
	}
	if at, terr := r.At(1); at != r.Path[len(r.Path)-1] || terr != Plain {
		t.Errorf("конец пути: %+v %d", at, terr)
	}
}

func TestFromImageClassifiesColors(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 40, 10))
	cols := []color.RGBA{{40, 90, 140, 255}, {240, 240, 240, 255}, {220, 190, 120, 255}, {60, 100, 50, 255}}
	for x := range 40 {
		for y := range 10 {
			img.Set(x, y, cols[x/10])
		}
	}
	g := FromImage(img, 10)
	want := []uint8{Water, Snow, Desert, Forest}
	for i, w := range want {
		if g.Cells[i] != w {
			t.Errorf("клетка %d: %d, ждали %d", i, g.Cells[i], w)
		}
	}
}
