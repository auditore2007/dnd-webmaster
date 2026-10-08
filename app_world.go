package main

import (
	"errors"
	"fmt"
	_ "image/jpeg" // форматы картинок карт
	_ "image/png"
	"math"
	"slices"
	"strings"
	"time"

	"heroesbook/internal/model"
	"heroesbook/internal/rules"
	"heroesbook/internal/store"
	"heroesbook/internal/atlas"
	"heroesbook/internal/worldgen"
)

// ---------- карты: генерация мира, подземелий и локаций, места, население ----------

const (
	maxMapSide     = 100000   // координаты места не больше этого (картинки карт заметно меньше)
	maxPlaces      = 300
	placeNameLen   = 80
	placeNoteLen   = 2000
)

var worldSizes = map[string][2]int{"small": {1800, 1125}, "medium": {2400, 1500}, "large": {3200, 2000}}

func (a *App) gen(seed int64, level int, races []string) worldgen.Gen {
	r := worldgen.Seeded(seed)
	return worldgen.Gen{R: r, Level: max(1, min(20, level)), Races: races, NewID: newID}
}

func randomSeed(seed int) int64 {
	if seed != 0 {
		return int64(seed)
	}
	return time.Now().UnixNano()
}

// checkRaces оставляет только расы из каталога; пусто – все.
func checkRaces(races []string) ([]string, error) {
	rs, err := rules.Get("dnd5e")
	if err != nil {
		return nil, err
	}
	var out []string
	for _, id := range races {
		if !slices.ContainsFunc(rs.Catalog().Races, func(r rules.Race) bool { return r.ID == id }) {
			return nil, fmt.Errorf("неизвестная раса: %q", id)
		}
		if !slices.Contains(out, id) {
			out = append(out, id)
		}
	}
	return out, nil
}

// addGeneratedMap сохраняет новую карту с местами (картинка – data:URL).
func (a *App) addGeneratedMap(name, data string, places []model.Place) (MapInfo, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	name = strings.TrimSpace(name)
	if r := []rune(name); len(r) > 60 {
		name = string(r[:60])
	}
	if name == "" {
		name = "Новая карта"
	}
	if len(data) > maxMapChars {
		return MapInfo{}, errors.New("карта получилась слишком большой – выберите размер поменьше")
	}
	if len(a.state.Maps) >= 30 {
		return MapInfo{}, errors.New("карт не может быть больше 30")
	}
	bs, ok := a.store.(blobStore)
	if !ok {
		return MapInfo{}, errors.New("хранилище не поддерживает карты")
	}
	m := store.MapMeta{ID: newID(), Name: name, Places: limitPlaces(places)}
	if err := bs.SaveBlob("map-"+m.ID, data); err != nil {
		return MapInfo{}, err
	}
	a.checkpoint()
	a.state.Maps = append(a.state.Maps, m)
	a.note("Новая карта «%s»: мест %d", name, len(m.Places))
	return a.mapInfo(m), a.persist()
}

func limitPlaces(ps []model.Place) []model.Place {
	if len(ps) > maxPlaces {
		return ps[:maxPlaces]
	}
	return ps
}

// GenerateWorld рисует карту мира в стиле атласа, расселяет на ней выбранные расы (пусто – все) и населяет места.
// seed 0 – случайный; template – шаблон рельефа (пусто – случайный).
func (a *App) GenerateWorld(name string, seed int, races []string, level int, size, template string) (MapInfo, error) {
	races, err := checkRaces(races)
	if err != nil {
		return MapInfo{}, err
	}
	dim, ok := worldSizes[size]
	if !ok {
		dim = worldSizes["medium"]
	}
	g := a.gen(randomSeed(seed), level, races)
	if _, ok := atlas.Templates()[template]; template != "" && !ok {
		return MapInfo{}, fmt.Errorf("нет шаблона рельефа %q", template)
	}
	title := cmpStr(name, "Новый мир")
	url, places, err := g.World(worldgen.WorldOpts{W: dim[0], H: dim[1], Title: title, Template: template})
	if err != nil {
		return MapInfo{}, err
	}
	return a.addGeneratedMap(title, url, places)
}

// GenerateDungeon рисует подземелье из rooms комнат с ловушками, тайниками и врагами под уровень группы.
func (a *App) GenerateDungeon(name string, seed, rooms, level int) (MapInfo, error) {
	g := a.gen(randomSeed(seed), level, nil)
	url, places, err := g.Dungeon(rooms)
	if err != nil {
		return MapInfo{}, err
	}
	return a.addGeneratedMap(cmpStr(name, "Подземелье"), url, places)
}

// ---------- места ----------

func (a *App) mapAt(id string) (*store.MapMeta, error) {
	i := a.mapIdx(id)
	if i < 0 {
		return nil, errors.New("карта не найдена")
	}
	return &a.state.Maps[i], nil
}

func checkPlace(p *model.Place) error {
	p.Name, p.Note = strings.TrimSpace(p.Name), strings.TrimSpace(p.Note)
	switch {
	case !slices.Contains(model.PlaceKinds, p.Kind):
		return fmt.Errorf("неизвестный вид места: %q", p.Kind)
	case len([]rune(p.Name)) > placeNameLen || len([]rune(p.Note)) > placeNoteLen:
		return fmt.Errorf("место: название до %d символов, заметка до %d", placeNameLen, placeNoteLen)
	case math.IsNaN(p.X) || math.IsNaN(p.Y) || p.X < 0 || p.Y < 0 || p.X > maxMapSide || p.Y > maxMapSide:
		return errors.New("место: координаты вне карты")
	}
	if p.Name == "" {
		p.Name = worldgen.KindName(p.Kind)
	}
	return nil
}

// AddPlace ставит место на карту (разметка вручную).
func (a *App) AddPlace(mapID string, x, y float64, kind, name string) (MapInfo, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	m, err := a.mapAt(mapID)
	if err != nil {
		return MapInfo{}, err
	}
	if len(m.Places) >= maxPlaces {
		return MapInfo{}, fmt.Errorf("мест на карте не может быть больше %d", maxPlaces)
	}
	p := model.Place{ID: newID(), X: x, Y: y, Kind: kind, Name: name}
	if err := checkPlace(&p); err != nil {
		return MapInfo{}, err
	}
	a.checkpoint()
	m.Places = append(m.Places, p)
	return a.mapInfo(*m), a.persist()
}

// UpdatePlace меняет название, вид, заметку и народ места (содержимое – через население).
func (a *App) UpdatePlace(mapID string, in model.Place) (MapInfo, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	m, err := a.mapAt(mapID)
	if err != nil {
		return MapInfo{}, err
	}
	i := slices.IndexFunc(m.Places, func(p model.Place) bool { return p.ID == in.ID })
	if i < 0 {
		return MapInfo{}, errors.New("место не найдено")
	}
	p := m.Places[i]
	p.Name, p.Kind, p.Note = in.Name, in.Kind, in.Note
	if in.Race != "" {
		if _, err := checkRaces([]string{in.Race}); err != nil {
			return MapInfo{}, err
		}
	}
	p.Race = in.Race
	if err := checkPlace(&p); err != nil {
		return MapInfo{}, err
	}
	a.checkpoint()
	m.Places[i] = p
	return a.mapInfo(*m), a.persist()
}

func (a *App) RemovePlace(mapID, placeID string) (MapInfo, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	m, err := a.mapAt(mapID)
	if err != nil {
		return MapInfo{}, err
	}
	i := slices.IndexFunc(m.Places, func(p model.Place) bool { return p.ID == placeID })
	if i < 0 {
		return MapInfo{}, errors.New("место не найдено")
	}
	a.checkpoint()
	m.Places = slices.Delete(m.Places, i, i+1)
	return a.mapInfo(*m), a.persist()
}

// PopulateMap населяет места карты: onlyEmpty – только те, где ещё пусто; иначе всё заново.
func (a *App) PopulateMap(mapID string, level int, onlyEmpty bool) (MapInfo, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	m, err := a.mapAt(mapID)
	if err != nil {
		return MapInfo{}, err
	}
	if len(m.Places) == 0 {
		return MapInfo{}, errors.New("на карте нет мест – поставьте их вручную или создайте карту заново")
	}
	a.checkpoint()
	m.Places = a.gen(randomSeed(0), level, nil).Populate(m.Places, onlyEmpty)
	a.note("Карта «%s» населена: мест %d", m.Name, len(m.Places))
	return a.mapInfo(*m), a.persist()
}

// PopulatePlace заново населяет одно место.
func (a *App) PopulatePlace(mapID, placeID string, level int) (MapInfo, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	m, err := a.mapAt(mapID)
	if err != nil {
		return MapInfo{}, err
	}
	i := slices.IndexFunc(m.Places, func(p model.Place) bool { return p.ID == placeID })
	if i < 0 {
		return MapInfo{}, errors.New("место не найдено")
	}
	a.checkpoint()
	cleared := model.ClonePlaces(m.Places)
	cleared[i].NPCs, cleared[i].Quests, cleared[i].Foes, cleared[i].Loot = nil, nil, nil, ""
	m.Places = a.gen(randomSeed(0), level, nil).Populate(cleared, true)
	return a.mapInfo(*m), a.persist()
}

func cmpStr(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
