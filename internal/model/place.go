package model

import "slices"

// Place – место на карте: поселение, таверна, логово, руины… Координаты – в пикселях картинки карты.
type Place struct {
	ID     string  `json:"id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Kind   string  `json:"kind"` // см. PlaceKinds
	Name   string  `json:"name"`
	Race   string  `json:"race"` // кто здесь живёт (id расы), "" – смешанно
	Note   string  `json:"note"`
	NPCs   []NPC   `json:"npcs"`
	Quests []Quest `json:"quests"`
	Foes   []Foe   `json:"foes"`
	Loot   string  `json:"loot"`
}

// NPC – персонаж места: кто он, чем занят и что скрывает.
type NPC struct {
	Name     string `json:"name"`
	Race     string `json:"race"`
	Role     string `json:"role"`
	Trait    string `json:"trait"`
	Want     string `json:"want"`
	Secret   string `json:"secret"`
	Attitude string `json:"attitude"` // дружелюбен | безразличен | враждебен
}

// Quest – задание, которое можно получить в этом месте. Target – id места, куда оно ведёт.
type Quest struct {
	Title  string `json:"title"`
	Text   string `json:"text"`
	Giver  string `json:"giver"`
	Target string `json:"target"`
	Reward string `json:"reward"`
	XP     int    `json:"xp"`
}

// Foe – существа из бестиария, которые встречаются в месте.
type Foe struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	CR    string `json:"cr"`
	Count int    `json:"count"`
}

// PlaceKinds – виды мест. Поселения населяют горожане, дикие места – враги.
var PlaceKinds = []string{"capital", "city", "town", "village", "tavern", "temple", "shop", "smithy", "castle", "port",
	"dungeon", "cave", "ruins", "lair", "camp", "tower", "shrine", "mine", "room", "forest", "mountain", "swamp", "lake", "other"}

// Settlement – обитаемое место (там живут NPC и раздают задания).
func Settlement(kind string) bool {
	return slices.Contains([]string{"capital", "city", "town", "village", "tavern", "temple", "shop", "smithy", "castle", "port", "tower", "shrine"}, kind)
}

// ClonePlaces – глубокая копия списка мест (для ответов интерфейсу).
func ClonePlaces(in []Place) []Place {
	out := slices.Clone(in)
	for i := range out {
		out[i].NPCs = slices.Clone(out[i].NPCs)
		out[i].Quests = slices.Clone(out[i].Quests)
		out[i].Foes = slices.Clone(out[i].Foes)
	}
	return out
}
