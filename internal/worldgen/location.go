package worldgen

import (
	"math"

	"heroesbook/internal/atlas"
	"heroesbook/internal/model"
)

// Карта локации – вид на одно место мира вблизи: поселение, замок, пещеру, руины, здание изнутри.
// Местность (биом, вода рядом) приходит с карты мира; места внутри населяются как обычно.

// Surroundings – какая местность вокруг места на карте мира.
type Surroundings struct {
	Biome    string  // grass, forest, snow, desert, swamp, jungle
	Water    bool    // рядом море или озеро
	WaterDir float64 // направление на воду, радианы
}

// HasLocation – можно ли открыть у места вида kind отдельную карту (у комнаты подземелья – нет).
func HasLocation(kind string) bool { return kind != "room" }

// Location рисует карту места p и возвращает картинку, места внутри и размер боевой клетки (0 – без сетки).
func (g Gen) Location(p model.Place, env Surroundings) (url string, places []model.Place, grid int, err error) {
	if p.Kind == "dungeon" {
		url, places, err = g.Dungeon(8 + g.R.Intn(6))
		return url, places, cellPx, err
	}
	m, err := atlas.Location(atlas.SiteOpts{Kind: p.Kind, Name: p.Name, Race: p.Race, Biome: env.Biome,
		Water: env.Water, WaterDir: env.WaterDir, R: g.R})
	if err != nil {
		return "", nil, 0, err
	}
	for _, sp := range m.Places {
		places = append(places, model.Place{ID: g.NewID(), X: sp.X, Y: sp.Y, Kind: sp.Kind, Name: sp.Name, Race: p.Race})
	}
	url, err = canvas{m.Img}.dataURL(92)
	if err != nil {
		return "", nil, 0, err
	}
	return url, g.Populate(places, false), m.Grid, nil
}

// SurroundingsAt – местность вокруг точки (x, y) картинки карты мира: пиксели на кольцах вокруг места
// (у самого места нарисован его значок) делятся на воду, снег, пески, лес и луга; вода – если её заметно много.
func SurroundingsAt(px func(x, y int) (r, g, b uint8, ok bool), x, y float64) Surroundings {
	counts := map[string]int{}
	var wx, wy float64
	water, land := 0, 0
	for _, rad := range []float64{45, 70, 100, 140} {
		for k := range 36 {
			a := float64(k) * math.Pi / 18
			r, g, b, ok := px(int(x+math.Cos(a)*rad), int(y+math.Sin(a)*rad))
			if !ok {
				continue
			}
			R, G, B := int(r), int(g), int(b)
			switch {
			case B > R+25 && B > G-10:
				water++
				wx, wy = wx+math.Cos(a), wy+math.Sin(a)
				continue
			case R > 200 && G > 200 && B > 200:
				counts["snow"]++
			case R > 175 && G > 150 && B < 150 && R-B > 45:
				counts["desert"]++
			case G > R && G < 110:
				counts["forest"]++
			case G > R && R < 95 && B < 80 && G > 115:
				counts["jungle"]++
			default:
				counts["grass"]++
			}
			land++
		}
	}
	env := Surroundings{Biome: "grass"}
	best := 0
	for _, b := range []string{"grass", "forest", "snow", "desert", "jungle"} {
		if counts[b] > best {
			env.Biome, best = b, counts[b]
		}
	}
	if total := water + land; total > 0 && float64(water)/float64(total) > 0.12 {
		env.Water, env.WaterDir = true, math.Atan2(wy, wx)
	}
	return env
}
