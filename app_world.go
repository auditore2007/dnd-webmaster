package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg" // форматы картинок карт
	_ "image/png"
	"math"
	"slices"
	"strings"
	"time"

	"heroesbook/internal/atlas"
	"heroesbook/internal/model"
	"heroesbook/internal/rules"
	"heroesbook/internal/store"
	"heroesbook/internal/worldgen"
)

// ---------- карты: генерация мира, подземелий и локаций, места, население ----------

const (
	maxMapSide   = 100000 // координаты места не больше этого (картинки карт заметно меньше)
	maxPlaces    = 300
	placeNameLen = 80
	placeNoteLen = 2000
)

var worldSizes = map[string][2]int{"small": {2400, 1500}, "medium": {3200, 2000}, "large": {4000, 2500}}

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
	if len(a.state.Maps) >= maxMaps {
		return MapInfo{}, fmt.Errorf("карт не может быть больше %d", maxMaps)
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
	url, places, terrain, err := g.World(worldgen.WorldOpts{W: dim[0], H: dim[1], Title: title, Template: template})
	if err != nil {
		return MapInfo{}, err
	}
	m, err := a.addGeneratedMap(title, url, places)
	if err != nil {
		return MapInfo{}, err
	}
	if bs, ok := a.store.(blobStore); ok {
		_ = bs.SaveBlob("travel-"+m.ID, terrain) // без сетки путешествия всё равно работают – по цветам картинки
	}
	return m, nil
}

// GenerateDungeon рисует подземелье из rooms комнат с ловушками, тайниками и врагами под уровень группы.
func (a *App) GenerateDungeon(name string, seed, rooms, level int) (MapInfo, error) {
	g := a.gen(randomSeed(seed), level, nil)
	name = cmpStr(name, "Подземелье")
	url, places, grid, err := g.Dungeon(name, rooms)
	if err != nil {
		return MapInfo{}, err
	}
	m, err := a.addGeneratedMap(name, url, places)
	if err != nil {
		return MapInfo{}, err
	}
	return a.setGrid(m, grid)
}

// setGrid ставит боевую сетку (клетка 5 футов) только что созданной карте.
func (a *App) setGrid(m MapInfo, grid int) (MapInfo, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	i := a.mapIdx(m.ID)
	if i < 0 || grid <= 0 {
		return m, nil
	}
	a.state.Maps[i].Grid, a.state.Maps[i].Feet = grid, 5
	return a.mapInfo(a.state.Maps[i]), a.persist()
}

// ---------- карты локаций ----------

// OpenPlace открывает карту места: если её уже рисовали – ту же, иначе рисует новую по виду места и местности
// вокруг него на этой карте и населяет места внутри.
func (a *App) OpenPlace(mapID, placeID string, level int) (MapInfo, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	m, err := a.mapAt(mapID)
	if err != nil {
		return MapInfo{}, err
	}
	pi := slices.IndexFunc(m.Places, func(p model.Place) bool { return p.ID == placeID })
	if pi < 0 {
		return MapInfo{}, errors.New("место не найдено")
	}
	p := m.Places[pi]
	if j := a.mapIdx(p.Map); p.Map != "" && j >= 0 {
		return a.mapInfo(a.state.Maps[j]), nil
	}
	if !worldgen.HasLocation(p.Kind) {
		return MapInfo{}, errors.New("у комнаты нет своей карты – она уже на этой")
	}
	if len(a.state.Maps) >= maxMaps {
		return MapInfo{}, fmt.Errorf("карт не может быть больше %d – удалите ненужные", maxMaps)
	}
	bs, ok := a.store.(blobStore)
	if !ok {
		return MapInfo{}, errors.New("хранилище не поддерживает карты")
	}
	env := worldgen.Surroundings{Biome: "grass"}
	if img, err := bs.LoadBlob("map-" + mapID); err == nil {
		if px, err := pixels(img); err == nil {
			env = worldgen.SurroundingsAt(px, p.X, p.Y)
		}
	}
	url, places, grid, err := a.gen(randomSeed(0), level, nil).Location(p, env)
	if err != nil {
		return MapInfo{}, err
	}
	if len(url) > maxMapChars {
		return MapInfo{}, errors.New("карта места получилась слишком большой")
	}
	sub := store.MapMeta{ID: newID(), Name: cmpStr(clipRunes(p.Name, 60), "Локация"), Places: limitPlaces(places),
		Grid: grid, Feet: 5, Parent: mapID, ParentPlace: placeID}
	if err := bs.SaveBlob("map-"+sub.ID, url); err != nil {
		return MapInfo{}, err
	}
	a.checkpoint()
	m.Places[pi].Map = sub.ID
	a.state.Maps = append(a.state.Maps, sub)
	a.note("Карта места «%s»: мест %d", sub.Name, len(sub.Places))
	return a.mapInfo(sub), a.persist()
}

func clipRunes(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n])
	}
	return s
}

// decodeImage – картинка карты из data:URL.
func decodeImage(dataURL string) (image.Image, error) {
	_, body, ok := strings.Cut(dataURL, ",")
	if !ok {
		return nil, errors.New("у карты нет картинки")
	}
	raw, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	return img, err
}

// pixels – доступ к цветам картинки карты (data:URL) по координатам.
func pixels(dataURL string) (func(x, y int) (r, g, b uint8, ok bool), error) {
	img, err := decodeImage(dataURL)
	if err != nil {
		return nil, err
	}
	b := img.Bounds()
	return func(x, y int) (uint8, uint8, uint8, bool) {
		if !(image.Point{x, y}.In(b)) {
			return 0, 0, 0, false
		}
		r, g, bb, _ := img.At(x, y).RGBA()
		return uint8(r >> 8), uint8(g >> 8), uint8(bb >> 8), true
	}, nil
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
