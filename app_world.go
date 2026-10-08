package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg" // форматы картинок карт
	_ "image/png"
	"math"
	"regexp"
	"slices"
	"strings"
	"time"

	"heroesbook/internal/model"
	"heroesbook/internal/rules"
	"heroesbook/internal/store"
	"heroesbook/internal/vision"
	"heroesbook/internal/atlas"
	"heroesbook/internal/worldgen"
)

// ---------- карты: генерация мира и подземелий, импорт, места, население, распознавание ИИ ----------

// modelNameRe – имя модели попадает в путь запроса, поэтому только буквы, цифры, точка и дефис.
var modelNameRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.\-]*$`)

const (
	maxMapSide     = 100000   // координаты места не больше этого (картинки карт заметно меньше)
	maxImportChars = 80 << 20 // полный экспорт Azgaar бывает большим
	maxPlaces      = 300
	placeNameLen   = 80
	placeNoteLen   = 2000
	dupDistance    = 24 // найденное ИИ место ближе этого к такому же уже стоящему – дубликат
	settingsBlob   = "settings"
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

// ImportAzgaar: JSON из Azgaar's Fantasy Map Generator и картинка той же карты.
func (a *App) ImportAzgaar(name, jsonText, imageData string, level int) (MapInfo, error) {
	if len(jsonText) > maxImportChars {
		return MapInfo{}, errors.New("файл слишком большой – выберите в Azgaar экспорт Minimal")
	}
	if !strings.HasPrefix(imageData, "data:image/") || len(imageData) > maxMapChars {
		return MapInfo{}, errors.New("нужна картинка карты из Azgaar (PNG или JPEG, до 10 МБ)")
	}
	_, img, err := decodeDataURL(imageData)
	if err != nil {
		return MapInfo{}, err
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(img))
	if err != nil {
		return MapInfo{}, fmt.Errorf("не удалось прочитать картинку: %w", err)
	}
	g := a.gen(randomSeed(0), level, nil)
	mapName, places, err := g.Azgaar([]byte(jsonText), cfg.Width, cfg.Height)
	if err != nil {
		return MapInfo{}, err
	}
	return a.addGeneratedMap(cmpStr(name, cmpStr(mapName, "Карта Azgaar")), imageData, places)
}

// ImportOnePageDungeon: JSON из One Page Dungeon – подземелье рисуется заново, заметки становятся комнатами.
func (a *App) ImportOnePageDungeon(name, jsonText string, level int) (MapInfo, error) {
	if len(jsonText) > maxImportChars {
		return MapInfo{}, errors.New("файл слишком большой")
	}
	g := a.gen(randomSeed(0), level, nil)
	title, url, places, err := g.OnePageDungeon([]byte(jsonText))
	if err != nil {
		return MapInfo{}, err
	}
	return a.addGeneratedMap(cmpStr(name, cmpStr(title, "Подземелье")), url, places)
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
		return MapInfo{}, errors.New("на карте нет мест – поставьте их вручную, распознайте ИИ или создайте карту заново")
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

// ---------- распознавание карты (Gemini) ----------

type aiSettings struct {
	GeminiKey   string `json:"geminiKey"`
	GeminiModel string `json:"geminiModel"`
}

// AIView – настройки ИИ для интерфейса: сам ключ наружу не отдаётся.
type AIView struct {
	HasKey  bool   `json:"hasKey"`
	KeyHint string `json:"keyHint"`
	Model   string `json:"model"`
}

func (a *App) loadAI() aiSettings {
	var s aiSettings
	if bs, ok := a.store.(blobStore); ok {
		if raw, err := bs.LoadBlob(settingsBlob); err == nil && raw != "" {
			_ = json.Unmarshal([]byte(raw), &s)
		}
	}
	if s.GeminiModel == "" {
		s.GeminiModel = vision.DefaultModel
	}
	return s
}

func (a *App) AISettings() AIView {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.loadAI()
	v := AIView{HasKey: s.GeminiKey != "", Model: s.GeminiModel}
	if k := s.GeminiKey; len(k) > 4 {
		v.KeyHint = "…" + k[len(k)-4:]
	}
	return v
}

// SetAISettings сохраняет ключ и модель Gemini. Пустой ключ оставляет прежний; clear – удаляет ключ.
func (a *App) SetAISettings(key, modelName string, clear bool) (AIView, error) {
	a.mu.Lock()
	s := a.loadAI()
	if key = strings.TrimSpace(key); key != "" {
		if len(key) > 200 || strings.ContainsAny(key, " \t\r\n") {
			a.mu.Unlock()
			return AIView{}, errors.New("ключ API выглядит неверно")
		}
		s.GeminiKey = key
	}
	if clear {
		s.GeminiKey = ""
	}
	if modelName = strings.TrimSpace(modelName); modelName != "" {
		if len(modelName) > 80 || !modelNameRe.MatchString(modelName) {
			a.mu.Unlock()
			return AIView{}, errors.New("название модели выглядит неверно")
		}
		s.GeminiModel = modelName
	}
	bs, ok := a.store.(blobStore)
	if !ok {
		a.mu.Unlock()
		return AIView{}, errors.New("хранилище не поддерживает настройки")
	}
	raw, _ := json.Marshal(s)
	err := bs.SaveBlob(settingsBlob, string(raw))
	a.mu.Unlock()
	if err != nil {
		return AIView{}, err
	}
	return a.AISettings(), nil
}

// visionClient подменяется в тестах.
var visionClient = func(s aiSettings) interface {
	Locate(ctx context.Context, mime string, data []byte, w, h int) ([]vision.Found, error)
} {
	return vision.Client{Key: s.GeminiKey, Model: s.GeminiModel}
}

// RecognizeMap отправляет картинку карты в Gemini, ставит найденные места и населяет их.
// Запрос к нейросети идёт без блокировки: остальное приложение в это время работает.
func (a *App) RecognizeMap(mapID string, level int) (MapInfo, error) {
	a.mu.Lock()
	if a.mapIdx(mapID) < 0 {
		a.mu.Unlock()
		return MapInfo{}, errors.New("карта не найдена")
	}
	s := a.loadAI()
	var img string
	if bs, ok := a.store.(blobStore); ok {
		img, _ = bs.LoadBlob("map-" + mapID)
	}
	a.mu.Unlock()
	if s.GeminiKey == "" {
		return MapInfo{}, errors.New("не задан ключ Gemini API – откройте «Настройки ИИ» и вставьте ключ из aistudio.google.com")
	}
	mime, data, err := decodeDataURL(img)
	if err != nil {
		return MapInfo{}, err
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return MapInfo{}, fmt.Errorf("не удалось прочитать картинку карты: %w", err)
	}
	found, err := visionClient(s).Locate(context.Background(), mime, data, cfg.Width, cfg.Height)
	if err != nil {
		return MapInfo{}, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	m, err := a.mapAt(mapID)
	if err != nil {
		return MapInfo{}, err // карту удалили, пока шёл запрос
	}
	places := model.ClonePlaces(m.Places)
	added := 0
	for _, f := range found {
		if len(places) >= maxPlaces || slices.ContainsFunc(places, func(p model.Place) bool {
			return p.Kind == f.Kind && math.Hypot(p.X-f.X, p.Y-f.Y) < dupDistance
		}) {
			continue
		}
		p := model.Place{ID: newID(), X: f.X, Y: f.Y, Kind: f.Kind, Name: worldgen.Clip(f.Name, placeNameLen), Note: worldgen.Clip(f.Note, placeNoteLen)}
		if checkPlace(&p) == nil {
			places = append(places, p)
			added++
		}
	}
	if added == 0 {
		return a.mapInfo(*m), errors.New("нейросеть не нашла новых мест – все найденные уже стоят на карте")
	}
	a.checkpoint()
	m.Places = a.gen(randomSeed(0), level, nil).Populate(places, true)
	a.note("Карта «%s»: ИИ нашёл мест %d", m.Name, added)
	return a.mapInfo(*m), a.persist()
}

// decodeDataURL разбирает data:image/…;base64,… на тип и байты.
func decodeDataURL(s string) (string, []byte, error) {
	head, body, ok := strings.Cut(s, ",")
	if !ok || !strings.HasPrefix(head, "data:image/") || !strings.HasSuffix(head, ";base64") {
		return "", nil, errors.New("у карты нет картинки")
	}
	b, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return "", nil, fmt.Errorf("картинка карты повреждена: %w", err)
	}
	return strings.TrimSuffix(strings.TrimPrefix(head, "data:"), ";base64"), b, nil
}

func cmpStr(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
