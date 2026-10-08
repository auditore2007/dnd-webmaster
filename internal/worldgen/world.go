package worldgen

import (
	"errors"
	"math"

	"heroesbook/internal/atlas"
	"heroesbook/internal/model"
)

// Карта мира рисуется пакетом atlas (алгоритмы Azgaar): рельеф, климат, реки, государства рас с городами и дорогами.
// Здесь – кто где любит селиться, как называются государства и какие дикие места бывают в разной местности.

type terrain = atlas.Terrain

// racePrefs – где селится раса (по убыванию предпочтения).
var racePrefs = map[string][]terrain{
	"human": {atlas.Grass, atlas.Coast, atlas.Forest}, "halfelf": {atlas.Grass, atlas.Forest}, "aasimar": {atlas.Grass, atlas.Hills},
	"genasi": {atlas.Desert, atlas.Coast, atlas.Mountain}, "elf": {atlas.Forest, atlas.Jungle}, "firbolg": {atlas.Forest},
	"dwarf": {atlas.Mountain, atlas.Hills}, "gnome": {atlas.Hills, atlas.Forest}, "halfling": {atlas.Grass, atlas.Hills},
	"halforc": {atlas.Hills, atlas.Grass}, "orc": {atlas.Hills, atlas.Tundra, atlas.Mountain}, "goblin": {atlas.Hills, atlas.Forest},
	"hobgoblin": {atlas.Hills, atlas.Grass}, "bugbear": {atlas.Forest, atlas.Hills}, "kobold": {atlas.Mountain, atlas.Hills},
	"dragonborn": {atlas.Hills, atlas.Desert}, "tiefling": {atlas.Desert, atlas.Grass}, "goliath": {atlas.Mountain, atlas.Snow},
	"aarakocra": {atlas.Mountain}, "kenku": {atlas.Forest, atlas.Grass}, "tabaxi": {atlas.Jungle, atlas.Grass},
	"lizardfolk": {atlas.Swamp, atlas.Jungle}, "tortle": {atlas.Coast, atlas.Swamp}, "triton": {atlas.Coast},
	"minotaur": {atlas.Hills, atlas.Desert}, "warforged": {atlas.Grass, atlas.Hills}, "yuanti": {atlas.Jungle, atlas.Swamp, atlas.Desert},
}

// wildFor – какие дикие места бывают в местности.
var wildFor = map[terrain][]string{
	atlas.Grass: {"ruins", "camp", "tower"}, atlas.Forest: {"camp", "ruins", "shrine", "lair"}, atlas.Jungle: {"ruins", "lair"},
	atlas.Swamp: {"lair", "ruins"}, atlas.Desert: {"ruins", "dungeon"}, atlas.Tundra: {"camp", "ruins"},
	atlas.Hills: {"dungeon", "camp", "mine"}, atlas.Mountain: {"cave", "lair", "mine"}, atlas.Snow: {"cave", "lair"}, atlas.Coast: {"cave"},
}

var wildAdj = []string{"Кривого Зуба", "Забытых Королей", "Чёрной Воды", "Шёпота", "Последнего Рыцаря", "Красной Луны", "Пепла",
	"Безымянного", "Ржавых Цепей", "Слепого Ворона", "Костяного Трона", "Туманов"}

var seaNames = []string{"Море Туманов", "Море Бурь", "Море Шёпота", "Янтарное море", "Море Забытых Королей", "Студёное море",
	"Море Тысячи Островов", "Серебряное море", "Море Безмолвия", "Закатное море"}

// stateForms – как называется государство культуры.
var stateForms = map[string]string{
	"human": "Королевство", "elf": "Владения", "dwarf": "Твердыня", "halfling": "Графство", "gnome": "Земли", "orc": "Орда",
	"goblinoid": "Ханство", "dragon": "Империя", "infernal": "Доминион", "beast": "Племена", "giant": "Предел",
	"construct": "Протекторат", "serpent": "Царство",
}

// WorldOpts – параметры карты мира.
type WorldOpts struct {
	W, H     int
	Title    string // название на картуше
	Template string // шаблон рельефа (atlas.Templates); пусто – случайный
}

// World рисует карту мира и возвращает её картинку (data:URL) и места (поселения рас и дикие места) с населением.
func (g Gen) World(o WorldOpts) (string, []model.Place, error) {
	if o.W < 400 || o.H < 300 || o.W > 4000 || o.H > 3000 {
		return "", nil, errors.New("размер карты: от 400×300 до 4000×3000")
	}
	names := map[string]bool{}
	var peoples []atlas.People
	for _, race := range g.races() {
		peoples = append(peoples, atlas.People{
			Race:      race,
			Likes:     racePrefs[race],
			TownName:  func() string { return g.uniqueTown(race, names) },
			StateName: func(capital string) string { return stateForms[raceCulture[race]] + " " + capital },
		})
	}
	w, err := atlas.Generate(atlas.Opts{W: o.W, H: o.H, Template: o.Template, Title: o.Title, SeaName: pick(g.R, seaNames),
		Peoples: peoples, R: g.R})
	if err != nil {
		return "", nil, err
	}
	img := w.Render(g.R)
	places := g.burgPlaces(w)
	places = append(places, g.wilds(w, places)...)
	url, err := canvas{img}.dataURL(88)
	if err != nil {
		return "", nil, err
	}
	return url, g.Populate(places, false), nil
}

// uniqueTown – название поселения, не повторяющее уже выданные (после 10 попыток – какое есть).
func (g Gen) uniqueTown(race string, used map[string]bool) string {
	name := TownName(g.R, race)
	for range 10 {
		if !used[name] {
			break
		}
		name = TownName(g.R, race)
	}
	used[name] = true
	return name
}

func (g Gen) burgPlaces(w *atlas.World) []model.Place {
	var out []model.Place
	for _, b := range w.Burgs {
		if len(out) >= maxPlaces/2 {
			break
		}
		out = append(out, model.Place{ID: g.NewID(), X: math.Round(b.X), Y: math.Round(b.Y), Kind: b.Kind, Race: b.Race, Name: b.Name,
			Note: w.States[b.State].Name})
	}
	return out
}

// wilds – логова, руины, пещеры и лагеря вдали от поселений.
func (g Gen) wilds(w *atlas.World, settled []model.Place) []model.Place {
	spots := w.LandSpots()
	if len(spots) == 0 {
		return nil
	}
	n := max(6, len(settled)*4/5)
	minD := float64(max(w.Width, w.Height)) / 18
	var out []model.Place
	for i := 0; i < n*6 && len(out) < n; i++ {
		s := spots[g.R.Intn(len(spots))]
		kinds := wildFor[s.Terrain]
		if len(kinds) == 0 || !farFrom(s.X, s.Y, settled, minD) || !farFrom(s.X, s.Y, out, minD*0.8) {
			continue
		}
		kind := pick(g.R, kinds)
		out = append(out, model.Place{ID: g.NewID(), X: math.Round(s.X), Y: math.Round(s.Y), Kind: kind,
			Name: KindName(kind) + " " + pick(g.R, wildAdj)})
	}
	return out
}

func farFrom(x, y float64, places []model.Place, d float64) bool {
	for _, p := range places {
		if math.Hypot(p.X-x, p.Y-y) < d {
			return false
		}
	}
	return true
}
