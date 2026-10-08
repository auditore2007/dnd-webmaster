package worldgen

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"heroesbook/internal/atlas"
	"heroesbook/internal/model"
)

func testGen(seed int64, level int, races ...string) Gen {
	n := 0
	return Gen{R: Seeded(seed), Level: level, Races: races, NewID: func() string { n++; return fmt.Sprintf("p%d", n) }}
}

func TestWorldIsReproducibleAndSettlesEveryRace(t *testing.T) {
	races := []string{"human", "elf", "dwarf", "lizardfolk"}
	url1, places1, err := testGen(11, 3, races...).World(WorldOpts{W: 800, H: 500})
	if err != nil {
		t.Fatal(err)
	}
	url2, places2, _ := testGen(11, 3, races...).World(WorldOpts{W: 800, H: 500})
	if url1 != url2 || len(places1) != len(places2) {
		t.Error("один seed – одна и та же карта")
	}
	if !strings.HasPrefix(url1, "data:image/jpeg;base64,") {
		t.Fatalf("картинка: %.40s", url1)
	}
	for _, r := range races {
		if !slices.ContainsFunc(places1, func(p model.Place) bool { return p.Race == r && p.Kind == "capital" }) {
			t.Errorf("у расы %s нет столицы", r)
		}
	}
	for _, p := range places1 {
		if p.X < 0 || p.Y < 0 || p.X > 800 || p.Y > 500 {
			t.Errorf("%s вне карты: %.0f,%.0f", p.Name, p.X, p.Y)
		}
		if model.Settlement(p.Kind) && len(p.NPCs) == 0 {
			t.Errorf("в поселении %s никто не живёт", p.Name)
		}
	}
	if _, _, err := testGen(1, 1).World(WorldOpts{W: 10, H: 10}); err == nil {
		t.Error("слишком маленькая карта должна отклоняться")
	}
}

func TestQuestsLeadToRealPlaces(t *testing.T) {
	places := []model.Place{
		{ID: "town", Kind: "town", Name: "Брод", Race: "human"},
		{ID: "cave", Kind: "cave", Name: "Пещера", X: 100},
		{ID: "lair", Kind: "lair", Name: "Логово", X: 300},
	}
	out := testGen(3, 4).Populate(places, false)
	ids := map[string]bool{}
	for _, p := range out {
		ids[p.ID] = true
	}
	for _, p := range out {
		if Wild(p.Kind) && len(p.Foes) == 0 {
			t.Errorf("в диком месте %s нет врагов", p.Name)
		}
		for _, q := range p.Quests {
			if q.Target != "" && !ids[q.Target] {
				t.Errorf("задание «%s» ведёт в несуществующее место %s", q.Title, q.Target)
			}
			if q.Reward == "" || q.XP <= 0 {
				t.Errorf("у задания «%s» нет награды", q.Title)
			}
		}
	}
	if len(out[0].Quests) == 0 || len(out[0].NPCs) == 0 {
		t.Errorf("в городе нет заданий или жителей: %+v", out[0])
	}
	again := testGen(4, 4).Populate(out, true)
	if again[0].NPCs[0].Name != out[0].NPCs[0].Name {
		t.Error("onlyEmpty не должен переписывать уже населённые места")
	}
}

func TestLairsGetBossesUpToPartyLevel(t *testing.T) {
	out := testGen(9, 5).Populate([]model.Place{{ID: "l", Kind: "lair"}, {ID: "c", Kind: "camp"}}, false)
	if len(out[0].Foes) != 1 || out[0].Foes[0].Count != 1 {
		t.Errorf("в логове – один сильный враг: %+v", out[0].Foes)
	}
	for _, f := range out[1].Foes {
		if f.Count < 1 || f.Count > maxFoesGroup {
			t.Errorf("размер группы %d", f.Count)
		}
	}
}

func TestDungeonRoomsAndEntrance(t *testing.T) {
	url, places, err := testGen(5, 3).Dungeon(8)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, "data:image/jpeg") || len(places) < 4 {
		t.Fatalf("подземелье: %d комнат", len(places))
	}
	if places[0].Name != entryName || len(places[0].Foes) != 0 {
		t.Errorf("у входа тихо: %+v", places[0])
	}
}

func TestNamesForEveryRace(t *testing.T) {
	g := testGen(1, 1)
	for _, r := range AllRaces {
		if n := PersonName(g.R, r); len([]rune(n)) < 3 {
			t.Errorf("%s: имя %q", r, n)
		}
		if n := TownName(g.R, r); len([]rune(n)) < 3 {
			t.Errorf("%s: поселение %q", r, n)
		}
	}
}

func TestEveryRaceHasCastleSprite(t *testing.T) {
	for _, race := range AllRaces {
		if !atlas.HasCastle(race) {
			t.Errorf("нет картинки замка для расы %s", race)
		}
	}
}
