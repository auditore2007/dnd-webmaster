package main

import (
	"errors"
	"fmt"
	"math"

	"heroesbook/internal/atlas"
	"heroesbook/internal/model"
	"heroesbook/internal/store"
	"heroesbook/internal/travel"
	"heroesbook/internal/worldgen"
)

// ---------- путешествия по карте мира ----------

const travelStep = 8 // клетка сетки местности, пикселей (для карт без сохранённой сетки)

// Journey – маршрут отряда: путь, мили, дни, местность по дороге и (после похода) встречи.
type Journey struct {
	travel.Route
	Pace       string      `json:"pace"`
	Encounters []Encounter `json:"encounters"`
}

// Encounter – случайная встреча в пути.
type Encounter struct {
	Day     int         `json:"day"`
	X       float64     `json:"x"`
	Y       float64     `json:"y"`
	Terrain string      `json:"terrain"`
	Text    string      `json:"text"`
	Foes    []model.Foe `json:"foes"`
}

// SetParty ставит фишку отряда на карту.
func (a *App) SetParty(mapID string, x, y float64) (MapInfo, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	m, err := a.mapAt(mapID)
	if err != nil {
		return MapInfo{}, err
	}
	if math.IsNaN(x) || math.IsNaN(y) || x < 0 || y < 0 || x > maxMapSide || y > maxMapSide {
		return MapInfo{}, errors.New("отряд: точка вне карты")
	}
	a.checkpoint()
	m.Party = &store.Point{X: x, Y: y}
	return a.mapInfo(*m), a.persist()
}

// PlanJourney прокладывает путь отряда до точки (x, y) и считает дни в дороге при темпе pace (fast, normal, slow).
func (a *App) PlanJourney(mapID string, x, y float64, pace string) (Journey, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.plan(mapID, x, y, pace)
}

func (a *App) plan(mapID string, x, y float64, pace string) (Journey, error) {
	m, err := a.mapAt(mapID)
	if err != nil {
		return Journey{}, err
	}
	if m.Party == nil {
		return Journey{}, errors.New("сначала поставьте отряд на карту")
	}
	g, err := a.travelGrid(mapID)
	if err != nil {
		return Journey{}, err
	}
	r, err := travel.Plan(g, travel.Point{X: m.Party.X, Y: m.Party.Y}, travel.Point{X: x, Y: y}, pace, atlas.MilesPerPx)
	if err != nil {
		return Journey{}, err
	}
	return Journey{Route: r, Pace: pace, Encounters: []Encounter{}}, nil
}

// Travel ведёт отряд до точки: за каждый день пути – бросок на встречу по местности, где отряд в тот день идёт.
func (a *App) Travel(mapID string, x, y float64, pace string, level int) (Journey, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	j, err := a.plan(mapID, x, y, pace)
	if err != nil {
		return Journey{}, err
	}
	g := a.gen(randomSeed(0), level, nil)
	days := max(1, int(math.Ceil(j.Days)))
	for d := 1; d <= days; d++ {
		at, t := j.At((float64(d) - 0.5) / float64(days))
		if float64(g.R.Intn(1000))/1000 >= worldgen.EncounterChance[t] {
			continue
		}
		text, foes := g.Encounter(t)
		if len(foes) == 0 {
			continue
		}
		j.Encounters = append(j.Encounters, Encounter{Day: d, X: at.X, Y: at.Y, Terrain: travel.Names[t], Text: text, Foes: foes})
	}
	m, _ := a.mapAt(mapID)
	a.checkpoint()
	m.Party = &store.Point{X: x, Y: y}
	a.note("Отряд прошёл %.0f миль за %s, встреч в пути: %d", j.Miles, daysText(j.Days), len(j.Encounters))
	return j, a.persist()
}

// travelGrid – сохранённая при создании мира сетка местности или угаданная по цветам картинки.
func (a *App) travelGrid(mapID string) (*travel.Grid, error) {
	bs, ok := a.store.(blobStore)
	if !ok {
		return nil, errors.New("хранилище не поддерживает карты")
	}
	if s, err := bs.LoadBlob("travel-" + mapID); err == nil && s != "" {
		if g, err := travel.Decode(s); err == nil {
			return g, nil
		}
	}
	data, err := bs.LoadBlob("map-" + mapID)
	if err != nil {
		return nil, err
	}
	img, err := decodeImage(data)
	if err != nil {
		return nil, err
	}
	return travel.FromImage(img, travelStep), nil
}

func daysText(d float64) string {
	whole := int(d)
	switch {
	case d < 1:
		return fmt.Sprintf("%.0f ч", d*8)
	case float64(whole) == d:
		return fmt.Sprintf("%d дн.", whole)
	}
	return fmt.Sprintf("%.2g дн.", d)
}
