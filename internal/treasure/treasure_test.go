package treasure

import (
	"testing"

	"heroesbook/internal/dice"
)

type rr struct{ v, i int }

func (s *rr) Intn(n int) int { s.i++; return (s.v + s.i) % n }

func TestTierBoundaries(t *testing.T) {
	for cr, want := range map[string]int{"0": 0, "1/4": 0, "4": 0, "5": 1, "10": 1, "11": 2, "16": 2, "17": 3, "30": 3} {
		if got := Tier(cr); got != want {
			t.Errorf("CR %s: ярус %d, ожидался %d", cr, got, want)
		}
	}
}

func TestIndividualGivesCoinsOnly(t *testing.T) {
	for seed := 0; seed < 100; seed++ {
		tr := Individual(&rr{v: seed}, "3")
		n := 0
		for _, v := range tr.Coins {
			n += v
		}
		if n <= 0 || len(tr.Items) != 0 {
			t.Fatalf("seed %d: монеты %v, предметы %d", seed, tr.Coins, len(tr.Items))
		}
	}
}

func TestHoardAlwaysHasCoinsAndValidItems(t *testing.T) {
	magic := 0
	for seed := 0; seed < 400; seed++ {
		for _, cr := range []string{"1", "7", "13", "20"} {
			tr := Hoard(&rr{v: seed * 7}, cr)
			if tr.Coins["gp"] == 0 && tr.Coins["cp"] == 0 {
				t.Fatalf("сокровищница без монет: %v", tr.Coins)
			}
			for _, it := range tr.Items {
				if it.Name == "" || it.Qty < 1 {
					t.Fatalf("кривой предмет %+v", it)
				}
				if it.Cat != "gem" && it.Cat != "art" {
					magic++
					if it.Rarity == "" {
						t.Errorf("магический предмет без редкости: %s", it.Name)
					}
				}
			}
		}
	}
	if magic == 0 {
		t.Error("за 1600 бросков не выпало ни одного магического предмета")
	}
}

func TestForCRsAndValue(t *testing.T) {
	tr := ForCRs(dice.RNG{}, []string{"1/4", "1/4", "2"}, false)
	if GoldValue(tr) <= 0 {
		t.Errorf("добыча за троих должна стоить денег: %v", tr.Coins)
	}
	if h := ForCRs(dice.RNG{}, []string{"1", "12"}, true); !h.Hoard || h.CR != "12" {
		t.Errorf("сокровищница берётся по самому опасному: %+v", h)
	}
	if e := ForCRs(dice.RNG{}, nil, false); len(e.Coins) != 0 {
		t.Error("без существ добычи нет")
	}
}
