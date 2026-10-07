package dice

import "testing"

// seq отдаёт заранее заданные значения (Intn возвращает v % n).
type seq struct {
	v []int
	i int
}

func (s *seq) Intn(n int) int { x := s.v[s.i%len(s.v)]; s.i++; return x % n }

func TestParse(t *testing.T) {
	ok := map[string]Expr{"d20": {1, 20, 0}, "2d6+3": {2, 6, 3}, " 1D8 - 2 ": {1, 8, -2}}
	for in, want := range ok {
		if got, err := Parse(in); err != nil || got != want {
			t.Errorf("Parse(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	for _, in := range []string{"", "d", "2d", "0d6", "101d6", "1d1", "abc", "2d6+", "d20+99999", "99999999999999999999d6"} {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) должен вернуть ошибку", in)
		}
	}
}

func TestD20Modes(t *testing.T) {
	if k, r := D20(&seq{v: []int{4, 14}}, Advantage); k != 15 || len(r) != 2 {
		t.Errorf("преимущество: kept=%d rolls=%v", k, r)
	}
	if k, _ := D20(&seq{v: []int{4, 14}}, Disadvantage); k != 5 {
		t.Errorf("помеха: kept=%d, ожидалось 5", k)
	}
	if k, r := D20(&seq{v: []int{4, 14}}, Normal); k != 5 || len(r) != 1 {
		t.Errorf("обычный бросок: kept=%d rolls=%v", k, r)
	}
}

func TestCritDoublesDiceNotBonus(t *testing.T) {
	e := Expr{Count: 1, Sides: 6, Mod: 2}
	if r := e.Roll(&seq{v: []int{2, 5}}, true); r.Total != 3+6+2 || len(r.Rolls) != 2 {
		t.Errorf("крит: %+v", r)
	}
	if r := e.Roll(&seq{v: []int{2, 5}}, false); r.Total != 3+2 {
		t.Errorf("обычный: %+v", r)
	}
}
