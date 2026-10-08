package main

import (
	"fmt"
	"math/rand/v2"
	"testing"
)

// pcgRoller – настоящие случайные броски для стресс-теста (у newApp кубики всегда выдают одно число).
type pcgRoller struct{ r *rand.Rand }

func (p pcgRoller) Intn(n int) int { return p.r.IntN(n) }

// checkEncounter – целостность боя: ход в пределах очереди, у всех в очереди есть инициатива и они существуют.
func checkEncounter(t *testing.T, a *App, step string) {
	t.Helper()
	e := a.Encounter()
	if e == nil {
		return
	}
	if len(e.Order) == 0 || e.Turn < 0 || e.Turn >= len(e.Order) {
		t.Fatalf("%s: ход %d при очереди %d", step, e.Turn, len(e.Order))
	}
	seen := map[string]bool{}
	for _, id := range e.Order {
		if seen[id] {
			t.Fatalf("%s: %s в очереди дважды", step, id)
		}
		seen[id] = true
		if _, ok := e.Init[id]; !ok {
			t.Fatalf("%s: у %s нет инициативы", step, id)
		}
		if a.find(id) == nil {
			t.Fatalf("%s: в бою несуществующий участник %s", step, id)
		}
	}
}

// TestRandomBattlesStayConsistent – сотни боёв со случайными действиями: ничего не падает, бой остаётся целостным.
func TestRandomBattlesStayConsistent(t *testing.T) {
	classes := []string{"fighter", "wizard", "cleric", "rogue", "barbarian", "paladin", "druid", "warlock"}
	monsters := []string{"goblin", "orc", "wolf", "skeleton", "ogre", "young-red-dragon", "bandit-captain"}
	for seed := uint64(1); seed <= 400; seed++ {
		r := rand.New(rand.NewPCG(seed, seed*7))
		a := &App{store: &memStore{}, rng: pcgRoller{r}}
		var ids []string
		for k := range 2 + r.IntN(3) {
			v, err := a.CreateCharacter("dnd5e", fmt.Sprintf("Герой%d", k), "human", "", classes[r.IntN(len(classes))])
			if err != nil {
				t.Fatal(err)
			}
			ids = append(ids, v.ID)
		}
		for range 1 + r.IntN(3) {
			ms, err := a.AddMonster(monsters[r.IntN(len(monsters))], 1+r.IntN(2))
			if err != nil {
				t.Fatal(err)
			}
			for _, m := range ms {
				ids = append(ids, m.ID)
			}
		}
		if _, err := a.StartEncounter(ids); err != nil {
			t.Fatal(err)
		}
		for step := 0; step < 200 && a.Encounter() != nil; step++ {
			e := a.Encounter()
			name := fmt.Sprintf("seed %d step %d", seed, step)
			var others []string
			for _, id := range e.Order {
				if id != e.Current() {
					others = append(others, id)
				}
			}
			target := ""
			if len(others) > 0 {
				target = others[r.IntN(len(others))]
			}
			switch r.IntN(13) {
			case 0, 1, 2:
				_, _ = a.Attack(target, r.IntN(3)-1, "")
			case 3:
				_, _ = a.AttackAll(target, 0, "adv")
			case 4, 5:
				_, _ = a.MonsterTurn()
			case 6, 12:
				if c := a.find(e.Current()); c != nil && len(c.Spells) > 0 {
					_, _ = a.CastSpell(r.IntN(len(c.Spells)), r.IntN(3), []string{target}, "")
				}
			case 7:
				_, _ = a.UseSpecial("breath")
			case 8:
				if r.IntN(4) == 0 {
					_, _ = a.RemoveFromEncounter(target)
				}
			case 9:
				if r.IntN(5) == 0 {
					_, _ = a.SpawnInBattle(monsters[r.IntN(len(monsters))], 1)
				}
			case 10:
				if r.IntN(6) == 0 {
					_ = a.Undo()
				}
			default:
				_, _ = a.NextTurn()
			}
			checkEncounter(t, a, name)
			if e := a.Encounter(); e != nil && e.Won != "" {
				_, _ = a.EndEncounter(false)
			}
		}
		_, _ = a.EndEncounter(true)
		for _, c := range a.state.Characters {
			if c.Conc != "" || len(c.Effects) > 0 {
				t.Errorf("seed %d: после боя у %s остались эффекты %v / концентрация %q", seed, c.Name, c.Effects, c.Conc)
			}
		}
	}
}

func TestHeroesDuelAndTeams(t *testing.T) {
	a := newApp()
	x, _ := a.CreateCharacter("dnd5e", "Лира", "human", "", "fighter")
	y, _ := a.CreateCharacter("dnd5e", "Торг", "human", "", "fighter")
	z, _ := a.CreateCharacter("dnd5e", "Эльм", "human", "", "wizard")
	e, err := a.StartEncounter([]string{x.ID, y.ID, z.ID})
	if err != nil {
		t.Fatal(err)
	}
	if e.Team[x.ID] == e.Team[y.ID] || e.Team[y.ID] == e.Team[z.ID] {
		t.Fatalf("одни герои – каждый сам за себя: %v", e.Team)
	}
	// Лира и Эльм против Торга
	if _, err := a.SetTeam(z.ID, e.Team[x.ID]); err != nil {
		t.Fatal(err)
	}
	if _, err := a.SetTeam(z.ID, 9); err == nil {
		t.Error("команды только 1–4")
	}
	a.state.Characters[1].HP = 0 // Торг упал
	a.state.Characters[1].Stable = true
	e, _ = a.NextTurn()
	if want := fmt.Sprintf("team%d", e.Team[x.ID]); e.Won != want {
		t.Errorf("победа команды: %q, ждали %q", e.Won, want)
	}
}

func TestLeavingBattleClearsEffects(t *testing.T) {
	a := newApp()
	h, _ := a.CreateCharacter("dnd5e", "Лира", "human", "", "fighter")
	ms, _ := a.AddMonster("goblin", 1)
	if _, err := a.StartEncounter([]string{h.ID, ms[0].ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddEffect(h.ID, "Сбит с ног", 1, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.RemoveFromEncounter(ms[0].ID); err != nil {
		t.Fatal(err)
	}
	if a.Encounter() != nil {
		t.Fatal("остался один – бой окончен")
	}
	if c := a.find(h.ID); len(c.Effects) != 0 {
		t.Errorf("эффекты боя должны сняться: %+v", c.Effects)
	}
}
