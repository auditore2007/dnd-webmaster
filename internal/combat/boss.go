package combat

import (
	"errors"
	"fmt"
	"slices"

	"heroesbook/internal/bestiary"
	"heroesbook/internal/dice"
	"heroesbook/internal/model"
	"heroesbook/internal/rules"
)

// Боссы, мораль и автоматический ход существ.

const smartIQ = 8 // существа с Интеллектом от 8 бьют самого раненого; звери – кого придётся

func bossOf(c *model.Character) (rules.Ruleset, rules.Boss, bool) {
	rs, err := rules.Get(c.Ruleset)
	if err != nil {
		return nil, nil, false
	}
	b, ok := rs.(rules.Boss)
	return rs, b, ok
}

// phases включает вторую фазу у боссов, чьё здоровье упало до порога.
func (e *Encounter) phases(get Lookup) {
	for _, id := range e.Order {
		c := get(id)
		if c == nil || c.Stat == nil || c.Stat.Phase == nil {
			continue
		}
		if _, b, ok := bossOf(c); ok {
			for _, l := range b.PhaseCheck(c) {
				e.logf("%s", l)
			}
		}
	}
}

// lairTurn: в начале каждого раунда (бой идёт в логове) каждый босс с логовом применяет одно действие логова по очереди.
func (e *Encounter) lairTurn(r dice.Roller, get Lookup) {
	if !e.Lair || e.Round <= e.LairRound {
		return
	}
	e.LairRound = e.Round
	for _, id := range slices.Clone(e.Order) {
		c := get(id)
		if c == nil || c.Stat == nil || len(c.Stat.Lair) == 0 || !c.Standing() {
			continue
		}
		_, b, ok := bossOf(c)
		foes := e.Foes(get, id)
		if !ok || len(foes) == 0 {
			continue
		}
		s := c.Stat.Lair[(e.Round-1)%len(c.Stat.Lair)]
		lines, err := b.RunSpecial(r, s, c, foes)
		if err != nil {
			continue
		}
		e.logf("🏰 Логово: %s", s.Name)
		for _, l := range lines[1:] {
			e.logf("%s", l)
		}
	}
	e.settle(r, get)
}

// leaderDown – пал вожак: все сильнейшие (по CR) существа боя повержены, а более слабые ещё стоят.
func (e *Encounter) leaderDown(get Lookup) bool {
	var mons []*model.Character
	top := -1.0
	for _, id := range e.Order {
		if c := get(id); c != nil && c.Stat != nil {
			mons = append(mons, c)
			top = max(top, bestiary.CRValue(c.Stat.CR))
		}
	}
	leadersDown, others := true, 0
	for _, c := range mons {
		switch {
		case bestiary.CRValue(c.Stat.CR) == top:
			leadersDown = leadersDown && !c.Standing()
		case c.Standing():
			others++
		}
	}
	return len(mons) > 0 && leadersDown && others > 0
}

// flees: в начале хода существо проверяет мораль; сбежавший или сдавшийся покидает бой. true – участник убран.
func (e *Encounter) flees(r dice.Roller, get Lookup) bool {
	if !e.Morale || len(e.Order) <= 1 { // последний участник не бежит: бой не остаётся без очереди
		return false
	}
	id := e.Order[e.Turn]
	c := get(id)
	if c == nil || c.Stat == nil || !c.Standing() {
		return false
	}
	_, b, ok := bossOf(c)
	if !ok {
		return false
	}
	lines, gone := b.Morale(r, c, e.leaderDown(get))
	for _, l := range lines {
		e.logf("%s", l)
	}
	if !gone {
		return false
	}
	if c.Conc != "" {
		e.DropConc(c, get)
	}
	e.Order = slices.Delete(e.Order, e.Turn, e.Turn+1)
	delete(e.Init, id)
	delete(e.React, id)
	delete(e.Team, id)
	rules.ClearEffects(c)
	e.Gone = append(e.Gone, id)
	if e.Turn >= len(e.Order) {
		e.Turn, e.Round = 0, e.Round+1
	}
	return true
}

// Legendary – легендарное действие существа actorID вне его хода. targetID нужен для атак и способностей по одной цели.
func (e *Encounter) Legendary(r dice.Roller, get Lookup, actorID, key, targetID string) (*model.AttackResult, error) {
	c := get(actorID)
	switch {
	case c == nil || c.Stat == nil:
		return nil, errors.New("существо не найдено")
	case !slices.Contains(e.Order, actorID):
		return nil, errors.New("существо не участвует в бою")
	case e.Current() == actorID:
		return nil, errors.New("легендарные действия – только в чужой ход")
	case !c.Standing() || rules.Incapacitated(c):
		return nil, errors.New("существо не может действовать")
	}
	acts := rules.LegendaryActions(c)
	i := slices.IndexFunc(acts, func(a model.LegAct) bool { return a.Key == key })
	if i < 0 {
		return nil, errors.New("нет такого легендарного действия")
	}
	act := acts[i]
	if rules.LegendaryLeft(c) < act.Cost {
		return nil, fmt.Errorf("не хватает очков легендарных действий (нужно %d, осталось %d)", act.Cost, rules.LegendaryLeft(c))
	}
	rs, b, ok := bossOf(c)
	if !ok {
		return nil, errors.New("система правил не поддерживает легендарные действия")
	}
	foes := e.Foes(get, actorID)
	if len(foes) == 0 {
		return nil, errors.New("нет целей")
	}
	var target *model.Character
	if targetID != "" {
		j := slices.IndexFunc(foes, func(f *model.Character) bool { return f.ID == targetID })
		if j < 0 {
			return nil, errors.New("цель недоступна")
		}
		target = foes[j]
	}
	defer e.settle(r, get)
	if act.Special == nil {
		if act.Weapon < 0 || act.Weapon >= len(c.Weapons) {
			return nil, errors.New("у существа нет такой атаки")
		}
		if target == nil {
			target = foes[0]
		}
		rules.SpendLegendary(c, act.Cost)
		w := c.Weapons[act.Weapon]
		res := rs.Attack(r, c, target, w, dice.Normal)
		e.resolve(r, rs, c, target, w, res, "⚡ ")
		return &res, nil
	}
	if act.Special.Mode != "area" && target != nil {
		foes = []*model.Character{target}
	}
	lines, err := b.RunSpecial(r, *act.Special, c, foes)
	if err != nil {
		return nil, err
	}
	rules.SpendLegendary(c, act.Cost)
	e.logf("⚡ Легендарное действие – %s", lines[0])
	for _, l := range lines[1:] {
		e.logf("%s", l)
	}
	return nil, nil
}

// autoLegendary: после чужого хода каждый босс тратит одно легендарное действие (в автоматическом режиме).
func (e *Encounter) autoLegendary(r dice.Roller, get Lookup) {
	cur := e.Current()
	for _, id := range slices.Clone(e.Order) {
		c := get(id)
		if id == cur || c == nil || c.Stat == nil || rules.LegendaryLeft(c) == 0 {
			continue
		}
		foes := e.Foes(get, id)
		if len(foes) == 0 {
			continue
		}
		act, ok := pickLegendary(c, len(foes))
		if !ok {
			continue
		}
		target := ""
		if act.Special == nil || act.Special.Mode != "area" {
			target = pickTarget(r, c, foes).ID
		}
		_, _ = e.Legendary(r, get, id, act.Key, target)
	}
}

// pickLegendary: по двум и более врагам – самое дорогое доступное действие по области, иначе самое дешёвое.
func pickLegendary(c *model.Character, foes int) (model.LegAct, bool) {
	left := rules.LegendaryLeft(c)
	var best model.LegAct
	found := false
	for _, a := range rules.LegendaryActions(c) {
		if a.Cost > left {
			continue
		}
		area := a.Special != nil && a.Special.Mode == "area"
		switch {
		case !found:
			best, found = a, true
		case foes >= 2 && area && (best.Special == nil || best.Special.Mode != "area" || a.Cost > best.Cost):
			best = a
		case !(foes >= 2 && best.Special != nil && best.Special.Mode == "area") && a.Cost < best.Cost:
			best = a
		}
	}
	return best, found
}

// pickTarget: разумные существа добивают самого раненого, звери и неразумные бьют кого придётся.
func pickTarget(r dice.Roller, c *model.Character, foes []*model.Character) *model.Character {
	if c.Abilities["int"] < smartIQ {
		return foes[r.Intn(len(foes))]
	}
	best := foes[0]
	for _, f := range foes[1:] {
		if f.HP < best.HP {
			best = f
		}
	}
	return best
}

// pickSpecial: дыхание, страх и другие перезаряжаемые способности – как только готовы (по области – если врагов двое
// и больше); обычные способности по цели – вместо атаки, если нет мультиатаки, иначе изредка.
func pickSpecial(r dice.Roller, c *model.Character, foes int) (model.Special, bool) {
	var fallback *model.Special
	for i := range c.Stat.Specials {
		s := c.Stat.Specials[i]
		limited := s.Recharge || s.Once
		if s.OnlyKind != "" || (limited && c.Used[s.Key]) {
			continue
		}
		if limited && (s.Mode != "area" || foes >= 2 || len(c.Stat.Specials) == 1) {
			return s, true
		}
		if !limited && s.Mode != "area" && fallback == nil {
			fallback = &c.Stat.Specials[i]
		}
	}
	if fallback != nil && (len(c.Stat.Multi) <= 1 || r.Intn(3) == 0) {
		return *fallback, true
	}
	return model.Special{}, false
}

// MonsterTurn – полный ход текущего существа: способность или серия атак, затем ход переходит дальше.
// Возвращает броски атак и имена целей (для анимации кубиков).
func (e *Encounter) MonsterTurn(r dice.Roller, get Lookup) ([]model.AttackResult, []string, error) {
	cur := get(e.Current())
	if cur == nil || cur.Stat == nil {
		return nil, nil, errors.New("сейчас ходит не существо")
	}
	var results []model.AttackResult
	var targets []string
	if !e.Acted && cur.Standing() {
		results, targets = e.monsterAct(r, get, cur)
	}
	e.Next(r, get)
	return results, targets, nil
}

func (e *Encounter) monsterAct(r dice.Roller, get Lookup, cur *model.Character) (results []model.AttackResult, targets []string) {
	foes := e.Foes(get, cur.ID)
	if len(foes) == 0 {
		return nil, nil
	}
	if s, ok := pickSpecial(r, cur, len(foes)); ok {
		target := ""
		if s.Mode != "area" {
			target = pickTarget(r, cur, foes).ID
		}
		if e.SpecialAt(r, get, s.Key, target) == nil {
			return nil, nil
		}
	}
	rs, err := rules.Get(cur.Ruleset)
	if err != nil {
		return nil, nil
	}
	for _, w := range attackPlan(rs, cur, e.Left) {
		if e.Acted {
			break
		}
		foes = e.Foes(get, cur.ID)
		if len(foes) == 0 {
			break
		}
		t := pickTarget(r, cur, foes)
		res, err := e.Attack(r, get, t.ID, w, dice.Normal)
		if err != nil {
			break
		}
		results, targets = append(results, res), append(targets, t.Name)
	}
	return results, targets
}

// attackPlan – оружие для каждой атаки хода: мультиатака существа или первое оружие столько раз, сколько атак.
func attackPlan(rs rules.Ruleset, c *model.Character, left int) []model.Weapon {
	ws := rs.Derive(c).Weapons
	if len(ws) == 0 {
		ws = []model.Weapon{model.Unarmed}
	}
	var plan []model.Weapon
	for _, i := range c.Stat.Multi {
		if i >= 0 && i < len(ws) {
			plan = append(plan, ws[i])
		}
	}
	if len(plan) == 0 {
		plan = []model.Weapon{ws[0]}
	}
	if left > 0 && len(plan) > left {
		plan = plan[len(plan)-left:]
	}
	return plan
}
