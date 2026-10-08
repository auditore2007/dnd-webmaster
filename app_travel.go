package main

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"

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

// groupColors – цвета групп по очереди.
var groupColors = []string{"#e0b040", "#4aa3df", "#d9534f", "#6cc070", "#b07cd8", "#e08a3c"}

const maxGroupName = 40

// SetGroup создаёт группу (пустой ID) или меняет её: название, состав, место на карте. Герой может быть только
// в одной группе на карте: взятые в эту группу уходят из прочих, опустевшие группы исчезают.
func (a *App) SetGroup(mapID string, in store.Group) (MapInfo, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	m, err := a.mapAt(mapID)
	if err != nil {
		return MapInfo{}, err
	}
	in.Name = strings.TrimSpace(in.Name)
	switch {
	case len([]rune(in.Name)) > maxGroupName:
		return MapInfo{}, fmt.Errorf("группа: название до %d символов", maxGroupName)
	case math.IsNaN(in.X) || math.IsNaN(in.Y) || in.X < 0 || in.Y < 0 || in.X > maxMapSide || in.Y > maxMapSide:
		return MapInfo{}, errors.New("группа: точка вне карты")
	case len(in.Members) == 0:
		return MapInfo{}, errors.New("в группе должен быть хотя бы один герой")
	}
	members := []string{}
	for _, id := range in.Members {
		if a.find(id) == nil {
			return MapInfo{}, fmt.Errorf("герой %q не найден", id)
		}
		if !slices.Contains(members, id) {
			members = append(members, id)
		}
	}
	a.checkpoint()
	gi := slices.IndexFunc(m.Groups, func(g store.Group) bool { return g.ID == in.ID })
	if in.ID == "" || gi < 0 {
		in.ID = newID()
		if in.Color == "" {
			in.Color = groupColors[len(m.Groups)%len(groupColors)]
		}
		if in.Name == "" {
			in.Name = fmt.Sprintf("Группа %d", len(m.Groups)+1)
		}
		m.Groups = append(m.Groups, store.Group{ID: in.ID})
		gi = len(m.Groups) - 1
	}
	if in.Name == "" {
		in.Name = m.Groups[gi].Name
	}
	if in.Color == "" {
		in.Color = m.Groups[gi].Color
	}
	in.Members = members
	m.Groups[gi] = in
	for k := range m.Groups { // герой – только в одной группе
		if m.Groups[k].ID != in.ID {
			m.Groups[k].Members = slices.DeleteFunc(m.Groups[k].Members, func(id string) bool { return slices.Contains(members, id) })
		}
	}
	m.Groups = slices.DeleteFunc(m.Groups, func(g store.Group) bool { return len(g.Members) == 0 })
	return a.mapInfo(*m), a.persist()
}

// DeleteGroup убирает группу с карты (герои остаются в игре).
func (a *App) DeleteGroup(mapID, groupID string) (MapInfo, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	m, err := a.mapAt(mapID)
	if err != nil {
		return MapInfo{}, err
	}
	i := slices.IndexFunc(m.Groups, func(g store.Group) bool { return g.ID == groupID })
	if i < 0 {
		return MapInfo{}, errors.New("группа не найдена")
	}
	a.checkpoint()
	m.Groups = slices.Delete(m.Groups, i, i+1)
	return a.mapInfo(*m), a.persist()
}

// leaveGroups убирает героя из всех групп на всех картах (герой удалён из игры).
func (a *App) leaveGroups(id string) {
	for i := range a.state.Maps {
		m := &a.state.Maps[i]
		for k := range m.Groups {
			m.Groups[k].Members = slices.DeleteFunc(m.Groups[k].Members, func(x string) bool { return x == id })
		}
		m.Groups = slices.DeleteFunc(m.Groups, func(g store.Group) bool { return len(g.Members) == 0 })
	}
}

func (a *App) groupAt(m *store.MapMeta, groupID string) (*store.Group, error) {
	i := slices.IndexFunc(m.Groups, func(g store.Group) bool { return g.ID == groupID })
	if i < 0 {
		return nil, errors.New("группа не найдена – поставьте группу на карту")
	}
	return &m.Groups[i], nil
}

// groupLevel – средний уровень героев группы (сила встреч в пути).
func (a *App) groupLevel(g *store.Group) int {
	sum, n := 0, 0
	for _, id := range g.Members {
		if c := a.find(id); c != nil && c.Kind != "monster" {
			sum += c.Level
			n++
		}
	}
	if n == 0 {
		return 1
	}
	return max(1, (sum+n/2)/n)
}

// PlanJourney прокладывает путь группы до точки (x, y) и считает дни в дороге при темпе pace (fast, normal, slow).
func (a *App) PlanJourney(mapID, groupID string, x, y float64, pace string) (Journey, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.plan(mapID, groupID, x, y, pace)
}

func (a *App) plan(mapID, groupID string, x, y float64, pace string) (Journey, error) {
	m, err := a.mapAt(mapID)
	if err != nil {
		return Journey{}, err
	}
	g, err := a.groupAt(m, groupID)
	if err != nil {
		return Journey{}, err
	}
	grid, err := a.travelGrid(mapID)
	if err != nil {
		return Journey{}, err
	}
	r, err := travel.Plan(grid, travel.Point{X: g.X, Y: g.Y}, travel.Point{X: x, Y: y}, pace, atlas.MilesPerPx)
	if err != nil {
		return Journey{}, err
	}
	return Journey{Route: r, Pace: pace, Encounters: []Encounter{}}, nil
}

// Travel ведёт группу до точки: за каждый день пути – бросок на встречу по местности, где группа в тот день идёт;
// сила врагов – по среднему уровню героев группы.
func (a *App) Travel(mapID, groupID string, x, y float64, pace string) (Journey, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	j, err := a.plan(mapID, groupID, x, y, pace)
	if err != nil {
		return Journey{}, err
	}
	m, _ := a.mapAt(mapID)
	grp, _ := a.groupAt(m, groupID)
	g := a.gen(randomSeed(0), a.groupLevel(grp), nil)
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
	a.checkpoint()
	grp.X, grp.Y = x, y
	a.note("%s: %.0f миль за %s, встреч в пути: %d", grp.Name, j.Miles, daysText(j.Days), len(j.Encounters))
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
