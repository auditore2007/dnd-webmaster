package worldgen

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"heroesbook/internal/model"
)

// Импорт из Azgaar's Fantasy Map Generator (azgaar.github.io/Fantasy-Map-Generator): «Export → JSON» (Minimal или Full)
// вместе с картинкой карты (PNG или JPEG). Координаты городов и меток – в системе координат карты (info.width × info.height),
// картинка экспорта – та же карта в другом масштабе, поэтому достаточно пропорции.

type azgaarJSON struct {
	Info struct {
		MapName string  `json:"mapName"`
		Width   float64 `json:"width"`
		Height  float64 `json:"height"`
	} `json:"info"`
	Pack struct {
		Burgs []struct {
			I          int     `json:"i"`
			Name       string  `json:"name"`
			X          float64 `json:"x"`
			Y          float64 `json:"y"`
			Capital    int     `json:"capital"`
			Port       int     `json:"port"`
			Population float64 `json:"population"`
			Culture    int     `json:"culture"`
			Group      string  `json:"group"`
			Removed    bool    `json:"removed"`
		} `json:"burgs"`
		Markers []struct {
			Type string  `json:"type"`
			Icon string  `json:"icon"`
			X    float64 `json:"x"`
			Y    float64 `json:"y"`
		} `json:"markers"`
	} `json:"pack"`
}

var burgGroups = map[string]string{"capital": "capital", "city": "city", "town": "town", "village": "village", "hamlet": "village",
	"fort": "castle", "monastery": "temple", "caravanserai": "shop", "trading_post": "shop"}

var markerKinds = map[string]string{"inns": "tavern", "dungeons": "dungeon", "mines": "mine", "battlefields": "ruins", "volcanoes": "lair",
	"hot-springs": "shrine", "lake-monsters": "lair", "sea-monsters": "lair", "hill-monsters": "lair", "sacred-mountains": "shrine",
	"sacred-forests": "shrine", "sacred-pineries": "shrine", "sacred-palm-groves": "shrine", "brigands": "camp", "pirates": "camp",
	"statues": "shrine", "ruins": "ruins", "portals": "tower", "rifts": "lair", "disturbed-burials": "ruins", "necropolises": "dungeon",
	"encounters": "camp", "caves": "cave", "libraries": "tower", "lighthouses": "other", "waterfalls": "other", "bridges": "other",
	"water-sources": "other", "canoes": "other", "migration": "other", "dances": "other", "mirage": "other", "circuses": "other", "jousts": "other", "fairs": "other"}

// burgKind – вид поселения: по группе Azgaar, иначе по столичности, порту и населению (в тысячах).
func burgKind(group string, capital, port int, pop float64) string {
	if k, ok := burgGroups[group]; ok && capital == 0 {
		return k
	}
	switch {
	case capital == 1:
		return "capital"
	case port > 0:
		return "port"
	case pop >= 10:
		return "city"
	case pop >= 2:
		return "town"
	}
	return "village"
}

// Azgaar превращает экспорт Azgaar в места на картинке размером imgW×imgH. Каждой культуре карты достаётся своя раса.
func (g Gen) Azgaar(raw []byte, imgW, imgH int) (name string, places []model.Place, err error) {
	var a azgaarJSON
	if err := json.Unmarshal(raw, &a); err != nil {
		return "", nil, fmt.Errorf("это не JSON из Azgaar: %w", err)
	}
	if a.Info.Width <= 0 || a.Info.Height <= 0 {
		return "", nil, errors.New("в файле нет размеров карты (info.width/height) – нужен «Export → JSON» из Azgaar")
	}
	if len(a.Pack.Burgs) == 0 && len(a.Pack.Markers) == 0 {
		return "", nil, errors.New("в файле нет городов и меток – выберите экспорт Minimal или Full")
	}
	sx, sy := float64(imgW)/a.Info.Width, float64(imgH)/a.Info.Height
	races := shuffled(g, g.races())
	for _, b := range a.Pack.Burgs {
		if b.I == 0 || b.Removed || strings.TrimSpace(b.Name) == "" || len(places) >= maxPlaces {
			continue // нулевой элемент в Azgaar – пустая заглушка
		}
		race := races[max(0, b.Culture)%len(races)]
		places = append(places, model.Place{ID: g.NewID(), X: b.X * sx, Y: b.Y * sy, Kind: burgKind(b.Group, b.Capital, b.Port, b.Population),
			Name: Clip(strings.TrimSpace(b.Name), NameLen), Race: race})
	}
	for _, m := range a.Pack.Markers {
		if len(places) >= maxPlaces {
			break
		}
		kind, ok := markerKinds[m.Type]
		if !ok {
			kind = "other"
		}
		places = append(places, model.Place{ID: g.NewID(), X: m.X * sx, Y: m.Y * sy, Kind: kind,
			Name: Clip(strings.TrimSpace(Clip(m.Icon, 4)+" "+KindName(kind)+" "+pick(g.R, wildAdj)), NameLen)})
	}
	return Clip(strings.TrimSpace(a.Info.MapName), NameLen), g.Populate(places, false), nil
}

func shuffled(g Gen, xs []string) []string {
	out := append([]string(nil), xs...)
	for i := len(out) - 1; i > 0; i-- {
		j := g.R.Intn(i + 1)
		out[i], out[j] = out[j], out[i]
	}
	return out
}
