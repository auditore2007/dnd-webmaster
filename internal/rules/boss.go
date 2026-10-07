package rules

import (
	"fmt"
	"slices"
	"strings"

	"heroesbook/internal/dice"
	"heroesbook/internal/model"
)

// Боссы и мораль: легендарные действия, вторая фаза, бегство и сдача врагов.

// Boss – необязательные возможности системы правил, которыми пользуется бой (легендарные действия, фазы, мораль).
type Boss interface {
	// RunSpecial исполняет описание способности (легендарное действие, логово) от имени существа.
	RunSpecial(r dice.Roller, s model.Special, src *model.Character, foes []*model.Character) ([]string, error)
	// PhaseCheck включает вторую фазу босса, когда его здоровье упало до порога.
	PhaseCheck(c *model.Character) []string
	// Morale – проверка морали существа: gone – оно бежит или сдаётся.
	Morale(r dice.Roller, c *model.Character, leaderDown bool) (lines []string, gone bool)
}

const (
	legKey        = "leg"    // счётчик потраченных за раунд очков легендарных действий
	moraleKey     = "morale" // мораль уже проверялась
	phaseKey      = "phase"  // вторая фаза уже началась
	legResKey     = "legres"
	phaseAtkKey   = "phase-atk"
	phaseACKey    = "phase-ac"
	phaseExtraKey = "phase-extra"
	moraleDC      = 10
	moralePct     = 25 // проверка морали, когда здоровья осталось не больше этой доли
	surrenderIQ   = 8  // разумные гуманоиды не бегут, а сдаются
)

// fearless – кто не знает страха: нежить, конструкты, слизи, растения и боссы.
var fearless = []string{"undead", "construct", "ooze", "plant"}

func (d DnD5e) RunSpecial(r dice.Roller, s model.Special, src *model.Character, foes []*model.Character) ([]string, error) {
	s.Recharge, s.Once = false, false // перезарядкой легендарных действий управляет счётчик очков
	return d.runSpecial(r, s, src, foes)
}

// LegendaryActions – легендарные действия существа: заданные в бестиарии или (если их нет) атаки каждым оружием.
func LegendaryActions(c *model.Character) []model.LegAct {
	s := c.Stat
	if s == nil || s.Legendary <= 0 {
		return nil
	}
	if len(s.LegActs) > 0 {
		return s.LegActs
	}
	out := make([]model.LegAct, 0, len(c.Weapons))
	for i, w := range c.Weapons {
		out = append(out, model.LegAct{Key: fmt.Sprintf("w%d", i), Name: "Атака: " + w.Name, Cost: 1, Weapon: i})
	}
	return out
}

// ResetBoss в конце боя снимает прибавки второй фазы и возвращает легендарные действия и сопротивления.
func ResetBoss(c *model.Character) {
	s := c.Stat
	if s == nil {
		return
	}
	s.Attack -= c.Counters[phaseAtkKey]
	c.ACBonus -= c.Counters[phaseACKey]
	if n := c.Counters[phaseExtraKey]; n > 0 && n <= len(s.Multi) {
		s.Multi = s.Multi[:len(s.Multi)-n]
	}
	for _, k := range []string{phaseAtkKey, phaseACKey, phaseExtraKey, legKey, legResKey} {
		delete(c.Counters, k)
	}
}

// LegendaryLeft – сколько очков легендарных действий осталось в этом раунде.
func LegendaryLeft(c *model.Character) int {
	if c.Stat == nil {
		return 0
	}
	return max(0, c.Stat.Legendary-c.Counters[legKey])
}

// SpendLegendary тратит очки легендарных действий.
func SpendLegendary(c *model.Character, n int) { c.Count(legKey, n) }

// Incapacitated – существо не может действовать (паралич, оглушение, окаменение…).
func Incapacitated(c *model.Character) bool { return hasAny(c, dndIncapacitate) }

func (d DnD5e) PhaseCheck(c *model.Character) []string {
	s := c.Stat
	if s == nil || s.Phase == nil || c.Used[phaseKey] || !c.Standing() || c.HP*100 > s.Phase.Pct*s.MaxHP {
		return nil
	}
	p := s.Phase
	c.Mark(phaseKey)
	s.Attack += p.Atk
	c.ACBonus += p.AC
	for i := 0; i < p.Extra; i++ {
		s.Multi = append(s.Multi, firstOr(s.Multi, 0))
	}
	// запоминаем прибавки, чтобы ResetBoss снял их после боя
	c.Count(phaseAtkKey, p.Atk)
	c.Count(phaseACKey, p.AC)
	c.Count(phaseExtraKey, p.Extra)
	c.TempHP = max(c.TempHP, p.Temp)
	if p.Recharge {
		for _, sp := range s.Specials {
			delete(c.Used, sp.Key)
		}
	}
	var gains []string
	for _, g := range []struct {
		on   bool
		text string
	}{{p.Atk > 0, fmt.Sprintf("атака +%d", p.Atk)}, {p.AC > 0, fmt.Sprintf("КД +%d", p.AC)}, {p.Extra > 0, fmt.Sprintf("+%d атак за ход", p.Extra)},
		{p.Temp > 0, fmt.Sprintf("%d временных HP", p.Temp)}, {p.Recharge, "способности восстановлены"}} {
		if g.on {
			gains = append(gains, g.text)
		}
	}
	line := fmt.Sprintf("🔥 %s: вторая фаза – «%s»", c.Name, p.Name)
	if len(gains) > 0 {
		line += " (" + strings.Join(gains, ", ") + ")"
	}
	return []string{line}
}

func (d DnD5e) Morale(r dice.Roller, c *model.Character, leaderDown bool) ([]string, bool) {
	s := c.Stat
	switch {
	case s == nil || !c.Standing() || c.Used[moraleKey] || s.Legendary > 0 || slices.Contains(fearless, s.Kind):
		return nil, false
	case c.HP*100 > moralePct*s.MaxHP && !leaderDown:
		return nil, false
	}
	c.Mark(moraleKey)
	why := "тяжело ранен"
	if leaderDown {
		why = "вожак пал"
	}
	tot := d.saveRoll(r, c, "wis")
	if tot >= moraleDC {
		return []string{fmt.Sprintf("%s: %s, проверка морали %s против %d – держится", c.Name, why, saveText(tot), moraleDC)}, false
	}
	verb := "в ужасе бежит с поля боя"
	if s.Kind == "humanoid" && c.Abilities["int"] >= surrenderIQ {
		verb = "бросает оружие и сдаётся"
		c.Mark("surrendered")
	} else {
		c.Mark("fled")
	}
	return []string{fmt.Sprintf("🏳 %s: %s, проверка морали %s против %d – %s", c.Name, why, saveText(tot), moraleDC, verb)}, true
}

func firstOr(xs []int, def int) int {
	if len(xs) == 0 {
		return def
	}
	return xs[0]
}
