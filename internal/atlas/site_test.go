package atlas

import (
	"slices"
	"testing"
)

func TestEveryPlaceKindHasLocationMap(t *testing.T) {
	kinds := []string{"capital", "city", "town", "village", "port", "castle", "cave", "mine", "lair", "tavern", "temple",
		"shop", "smithy", "ruins", "camp", "shrine", "tower", "forest", "mountain", "swamp", "lake", "other", "dungeon"}
	for i, k := range kinds {
		m, err := Location(SiteOpts{Kind: k, Name: "Место", Biome: []string{"grass", "snow", "desert", "forest", "swamp", "jungle"}[i%6],
			Water: i%3 == 0, WaterDir: 1, R: seeded(uint64(i + 1))})
		if err != nil {
			t.Fatalf("%s: %v", k, err)
		}
		b := m.Img.Bounds()
		if len(m.Places) == 0 {
			t.Errorf("%s: внутри нет мест", k)
		}
		for _, p := range m.Places {
			if p.X < 0 || p.Y < 0 || p.X >= float64(b.Dx()) || p.Y >= float64(b.Dy()) || p.Name == "" {
				t.Errorf("%s: место вне карты или без имени: %+v", k, p)
			}
		}
		battle := slices.Contains([]string{"cave", "mine", "lair", "tavern", "temple", "shop", "smithy", "dungeon"}, k)
		if battle != (m.Grid > 0) {
			t.Errorf("%s: сетка %d", k, m.Grid)
		}
	}
	if _, err := Location(SiteOpts{Kind: "town"}); err == nil {
		t.Error("без генератора случайностей – ошибка")
	}
}

func TestLocationIsReproducible(t *testing.T) {
	a, _ := Location(SiteOpts{Kind: "town", Name: "Брод", Biome: "grass", R: seeded(4)})
	b, _ := Location(SiteOpts{Kind: "town", Name: "Брод", Biome: "grass", R: seeded(4)})
	if !slices.Equal(a.Img.Pix, b.Img.Pix) || !slices.Equal(a.Places, b.Places) {
		t.Error("одно зерно – одна и та же карта места")
	}
}
