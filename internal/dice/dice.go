// Package dice — броски кубиков. Roller вынесен в интерфейс, чтобы в тестах
// подставлять предсказуемый генератор.
package dice

import (
	"fmt"
	"math/rand/v2"
	"regexp"
	"strconv"
	"strings"
)

type Roller interface{ Intn(n int) int }

// RNG — боевой генератор.
type RNG struct{}

func (RNG) Intn(n int) int { return rand.IntN(n) }

type Mode int

const (
	Normal Mode = iota
	Advantage
	Disadvantage
)

func ParseMode(s string) Mode {
	switch s {
	case "adv":
		return Advantage
	case "dis":
		return Disadvantage
	}
	return Normal
}

// D20 бросает d20; при преимуществе/помехе бросает дважды и оставляет лучший/худший.
func D20(r Roller, m Mode) (kept int, rolls []int) {
	a := r.Intn(20) + 1
	if m == Normal {
		return a, []int{a}
	}
	b := r.Intn(20) + 1
	kept = a
	if (m == Advantage && b > a) || (m == Disadvantage && b < a) {
		kept = b
	}
	return kept, []int{a, b}
}

// Expr — запись вида 2d6+3.
type Expr struct{ Count, Sides, Mod int }

var exprRe = regexp.MustCompile(`^(\d*)d(\d+)\s*(?:([+-])\s*(\d+))?$`)

func Parse(s string) (Expr, error) {
	m := exprRe.FindStringSubmatch(strings.ToLower(strings.TrimSpace(s)))
	if m == nil {
		return Expr{}, fmt.Errorf("неверная запись кубиков: %q", s)
	}
	e := Expr{Count: 1}
	if m[1] != "" {
		e.Count, _ = strconv.Atoi(m[1])
	}
	e.Sides, _ = strconv.Atoi(m[2])
	if m[4] != "" {
		e.Mod, _ = strconv.Atoi(m[4])
		if m[3] == "-" {
			e.Mod = -e.Mod
		}
	}
	if e.Count < 1 || e.Count > 100 || e.Sides < 2 || e.Sides > 1000 || e.Mod > 10000 || e.Mod < -10000 {
		return Expr{}, fmt.Errorf("недопустимые значения: %q", s)
	}
	return e, nil
}

type Result struct {
	Rolls []int `json:"rolls"`
	Mod   int   `json:"mod"`
	Total int   `json:"total"`
}

// Roll бросает кубики. Критический удар удваивает кубики, но не бонус.
func (e Expr) Roll(r Roller, crit bool) Result {
	n := e.Count
	if crit {
		n *= 2
	}
	res := Result{Mod: e.Mod}
	for i := 0; i < n; i++ {
		v := r.Intn(e.Sides) + 1
		res.Rolls = append(res.Rolls, v)
		res.Total += v
	}
	res.Total += e.Mod
	return res
}
