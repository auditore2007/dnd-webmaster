package rules

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"heroesbook/internal/dice"
	"heroesbook/internal/model"
)

func effectNames(c *model.Character) (out []string) {
	for _, e := range c.Effects {
		n := e.Name
		if e.Rounds > 0 {
			n += fmt.Sprintf(" · %d р.", e.Rounds)
		}
		if e.Conc {
			n += " ⌛"
		}
		out = append(out, n)
	}
	return out
}

func specialText(s model.Special) string {
	var p []string
	if s.Mode == "area" {
		p = append(p, "по всем врагам")
	} else {
		p = append(p, "по одной цели")
	}
	if s.Attack {
		p = append(p, "бросок атаки")
	}
	if s.Dmg != "" {
		p = append(p, s.Dmg+" "+s.Type)
	}
	if s.XDmg != "" {
		p = append(p, "+ "+s.XDmg+" "+s.XType)
	}
	if s.Save != "" {
		p = append(p, fmt.Sprintf("спасбросок %s, СЛ %d", s.Save, s.DC))
	}
	if s.Cond != "" {
		p = append(p, s.Cond)
	}
	switch {
	case s.Recharge:
		p = append(p, "перезарядка 5–6")
	case s.Once:
		p = append(p, "раз за бой")
	}
	t := strings.Join(p, ", ")
	if s.Desc != "" {
		t += ". " + s.Desc
	}
	return t
}

// monsterSpecial применяет особую способность существа: по всем врагам (area) или по одной цели (single).
func (d DnD5e) monsterSpecial(r dice.Roller, key string, src *model.Character, foes []*model.Character) ([]string, error) {
	i := slices.IndexFunc(src.Stat.Specials, func(s model.Special) bool { return s.Key == key })
	if i < 0 {
		return nil, errors.New("у существа нет такой способности")
	}
	return d.runSpecial(r, src.Stat.Specials[i], src, foes)
}

// runSpecial исполняет описание способности (у существа или героя).
func (d DnD5e) runSpecial(r dice.Roller, s model.Special, src *model.Character, foes []*model.Character) ([]string, error) {
	key := s.Key
	if s.OnlyKind != "" {
		foes = slices.DeleteFunc(slices.Clone(foes), func(t *model.Character) bool { return t.Stat == nil || t.Stat.Kind != s.OnlyKind })
		if len(foes) == 0 {
			return nil, errors.New("нет подходящих целей")
		}
	}
	if (s.Recharge || s.Once) && src.Used[key] {
		return nil, errors.New("способность ещё не восстановилась")
	}
	if len(foes) == 0 {
		return nil, errors.New("нет целей")
	}
	if s.Mode != "area" {
		foes = foes[:1]
	}
	if s.Recharge || s.Once {
		src.Mark(key)
	}
	lines := []string{fmt.Sprintf("%s: %s", src.Name, s.Name)}
	roll := func(expr string, crit bool) int {
		if expr == "" {
			return 0
		}
		ex, err := dice.Parse(expr)
		if err != nil {
			return 0
		}
		return ex.Roll(r, crit).Total
	}
	hit := func(t *model.Character, n int, ty string, crit bool) {
		n = int(float64(n) * d.Resist(t, ty))
		if n <= 0 {
			return
		}
		lines = append(lines, fmt.Sprintf("→ %s: урон %d (%s)", t.Name, n, ty))
		sl, _ := Strike(d, t, n, crit)
		lines = append(lines, sl...)
	}
	cond := func(t *model.Character) {
		if s.Cond == "" {
			return
		}
		e := model.Effect{Name: s.Name, Rounds: s.Rounds, Cond: s.Cond, Src: src.ID}
		if s.Repeat {
			e.Save, e.DC = cmp2(s.RSave, s.Save), s.DC
		}
		AddEffect(t, e)
		lines = append(lines, fmt.Sprintf("→ %s: %s", t.Name, s.Cond))
	}
	if s.Mode == "area" {
		dmg, x := roll(s.Dmg, false), roll(s.XDmg, false)
		for _, t := range foes {
			ok := false
			if s.Save != "" { // без спасброска область задевает всех
				var txt string
				_, txt, ok = d.saveVs(r, t, s.Save, s.DC)
				lines = append(lines, fmt.Sprintf("%s: спасбросок %s против %d – %s", t.Name, txt, s.DC, map[bool]string{true: "успех", false: "провал"}[ok]))
			}
			a, b := dmg, x
			if ok {
				if !s.Half {
					continue
				}
				a, b = a/2, b/2
			}
			hit(t, a, s.Type, false)
			hit(t, b, s.XType, false)
			if !ok {
				cond(t)
			}
		}
		return lines, nil
	}
	t := foes[0]
	crit := false
	if s.Attack {
		kept, _ := dice.D20(r, condModeR(src, t, dice.Normal, false))
		fb, _ := atkDice(r, src)
		ds, dt := d.Derive(src), d.Derive(t)
		tot := kept + ds.AtkBonus + fb
		ok := kept == 20 || (kept != 1 && tot >= dt.AC)
		lines = append(lines, fmt.Sprintf("→ %s: атака %d против КД %d", t.Name, tot, dt.AC))
		if !ok {
			return append(lines, "→ промах"), nil
		}
		crit = kept == 20 || hasAny(t, autoCritConds) || t.Down()
		hit(t, roll(s.Dmg, crit), s.Type, crit)
		if s.Save == "" {
			cond(t)
			return lines, nil
		}
	} else if s.Dmg != "" {
		hit(t, roll(s.Dmg, false), s.Type, false)
	}
	if s.Save != "" {
		_, txt, ok := d.saveVs(r, t, s.Save, s.DC)
		lines = append(lines, fmt.Sprintf("%s: спасбросок %s против %d – %s", t.Name, txt, s.DC, map[bool]string{true: "успех", false: "провал"}[ok]))
		x := roll(s.XDmg, crit)
		switch {
		case !ok:
			hit(t, x, s.XType, false)
			cond(t)
		case s.Half:
			hit(t, x/2, s.XType, false)
		}
	}
	return lines, nil
}

var autoCritConds = []string{"Парализован", "Без сознания"}
