package rules

import (
	"errors"
	"fmt"
	"slices"

	"heroesbook/internal/dice"
	"heroesbook/internal/model"
)

// Ячейки колдуна: все одного уровня, восстанавливаются на коротком отдыхе.
func pactSlots(level int) []model.SlotView {
	n, lv := 1, 1
	switch {
	case level >= 17:
		n, lv = 4, 5
	case level >= 11:
		n, lv = 3, 5
	case level >= 9:
		n, lv = 2, 5
	case level >= 7:
		n, lv = 2, 4
	case level >= 5:
		n, lv = 2, 3
	case level >= 3:
		n, lv = 2, 2
	case level >= 2:
		n, lv = 2, 1
	}
	return []model.SlotView{{Level: lv, Max: n}}
}

func rageBonus(lvl int) int {
	switch {
	case lvl >= 16:
		return 4
	case lvl >= 9:
		return 3
	}
	return 2
}

func setActive(c *model.Character, key string, v bool) {
	if c.Active == nil {
		c.Active = map[string]bool{}
	}
	c.Active[key] = v
}

func clampLevel(l int) int { return max(1, min(20, l)) }

// armorClass: лучший из доступных способов – надетый доспех, защита без доспехов класса или 10 + Ловкость.
func armorClass(c *model.Character, mods map[string]int, itemAC int) int {
	dex := mods["dex"]
	best, worn := 10+dex, false
	for _, it := range c.Inventory {
		if !it.Equipped || it.ArmorBase <= 0 {
			continue
		}
		v := it.ArmorBase + dex
		switch it.ArmorType {
		case "heavy":
			v = it.ArmorBase
		case "medium":
			v = it.ArmorBase + min(dex, 2)
		}
		if !worn || v > best {
			best, worn = v, true
		}
	}
	if !worn {
		switch {
		case c.Race == "tortle":
			best = 17
		case c.Race == "lizardfolk":
			best = max(best, 13+dex)
		case c.Class == "barbarian":
			best = 10 + dex + mods["con"]
		case c.Class == "monk":
			best = 10 + dex + mods["wis"]
		case c.Class == "sorcerer" && c.Subclass == "draconic":
			best = 13 + dex
		}
	}
	return best + c.ACBonus + itemAC + b2i(c.Race == "warforged")
}

func heavyWorn(c *model.Character) bool {
	return slices.ContainsFunc(c.Inventory, func(it model.Item) bool { return it.Equipped && it.ArmorType == "heavy" })
}

func (d DnD5e) features(c *model.Character, lvl int) (fs []model.Feature) {
	switch c.Class {
	case "barbarian":
		fs = append(fs, model.Feature{Key: "rage", Name: "Ярость", Kind: "toggle", On: c.Active["rage"],
			Desc: fmt.Sprintf("+%d к урону Силой, сопротивление физическому урону. Осталось %d из %d.", rageBonus(lvl), max(0, rages(lvl)-c.Counters["rage"]), rages(lvl))})
	case "fighter":
		fs = append(fs, model.Feature{Key: "secondwind", Name: "Второе дыхание", Kind: "once", On: c.Used["secondwind"],
			Desc: fmt.Sprintf("Лечение 1d10 + %d. Раз за короткий отдых.", lvl)})
	case "rogue":
		fs = append(fs, model.Feature{Key: "sneak", Name: "Скрытая атака", Kind: "passive", On: c.Active["sneaked"],
			Desc: fmt.Sprintf("+%dd6 при попадании с преимуществом оружием с ловкостью или дальним, раз за ход.", (lvl+1)/2)})
	}
	switch c.Race {
	case "halforc":
		fs = append(fs, model.Feature{Key: "relentless", Name: "Непоколебимая стойкость", Kind: "passive", On: c.Used["relentless"], Desc: "Раз до долгого отдыха остаётся с 1 HP вместо 0."})
	case "dragonborn":
		fs = append(fs, model.Feature{Key: "breath", Name: "Дыхание дракона", Kind: "special", Mode: "area", On: c.Used["breath"], Desc: "Конус по всем врагам, спасбросок Ловкости – половина урона. Применяется в бою."})
	}
	return fs
}

func (d DnD5e) Derive(c *model.Character) model.Derived {
	if c.Stat != nil {
		return monsterDerive(c)
	}
	race, _ := merge(dndRaces, c.Race, c.Subrace)
	cls, _ := findClass(c.Class)
	if cls.HitDie == 0 {
		cls.HitDie = 8
	}
	lvl := clampLevel(c.Level)
	fx, itemAC := itemEffects(c)
	eff := applyEffects(c.Abilities, fx)
	out := model.Derived{Mods: map[string]int{}, Saves: map[string]int{}, Skills: map[string]int{}, Prof: 2 + (lvl-1)/4, Speed: race.Speed, Traits: race.Traits,
		Weapons: append(allWeapons(c), classWeapons(c, lvl)...), Slots: slotsOf(c), Load: totalWeight(c), Carry: float64(15 * eff["str"]), NextXP: d.XPFor(lvl + 1),
		Attacks: attacksFor(c.Class, lvl)}
	for _, e := range fx {
		out.Effects = append(out.Effects, e.text(dndAbilities))
	}
	for _, a := range dndAbilities {
		out.Mods[a.ID] = Mod(eff[a.ID])
	}
	for _, a := range dndAbilities {
		out.Saves[a.ID] = out.Mods[a.ID]
		if slices.Contains(cls.Saves, a.ID) {
			out.Saves[a.ID] += out.Prof
		}
	}
	for _, s := range dndSkills {
		v := out.Mods[s.Ability]
		switch {
		case slices.Contains(c.Expert, s.ID):
			v += 2 * out.Prof
		case slices.Contains(c.Skills, s.ID):
			v += out.Prof
		}
		out.Skills[s.ID] = v
	}
	perLevel := max(1, cls.HitDie/2+1+out.Mods["con"]+race.HPLevel)
	out.MaxHP = max(1, cls.HitDie+out.Mods["con"]+race.HPLevel) + (lvl-1)*perLevel
	if c.Class == "sorcerer" && c.Subclass == "draconic" {
		out.MaxHP += lvl
	}
	acFx, _ := effectBonus(c)
	out.AC = armorClass(c, out.Mods, itemAC) + acFx
	out.Effects = append(out.Effects, effectNames(c)...)
	out.Initiative = out.Mods["dex"]
	if thirdCaster(c) {
		out.SpellAtk = out.Mods["int"] + out.Prof
		out.SpellDC = 8 + out.Mods["int"] + out.Prof
	} else if ab := spellAbility[c.Class]; ab != "" {
		out.SpellAtk = out.Mods[ab] + out.Prof
		out.SpellDC = 8 + out.Mods[ab] + out.Prof
	}
	if c.Class == "paladin" && lvl >= 6 {
		for _, a := range dndAbilities {
			out.Saves[a.ID] += max(1, out.Mods["cha"])
		}
	}
	switch {
	case c.Class == "barbarian" && c.Subclass == "berserker" && lvl >= 3 && c.Active["rage"]:
		out.Attacks++
	case c.Class == "bard" && c.Subclass == "valor" && lvl >= 6:
		out.Attacks = max(out.Attacks, 2)
	}
	out.Features = append(d.features(c, lvl), d.abilityFeatures(c)...)
	out.Pools = d.poolViews(c)
	return out
}

// classWeapons – оружие, которое даёт сам класс: удар монаха и психический клинок.
func classWeapons(c *model.Character, lvl int) (out []model.Weapon) {
	if c.Class == "monk" {
		out = append(out, model.Weapon{Name: "Удар монаха", Dice: "1" + martialDie(ctx{Lvl: lvl}), Ability: "dex", Type: "bludgeoning", Finesse: true})
	}
	if c.Class == "rogue" && c.Subclass == "soulknife" && lvl >= 3 {
		out = append(out, model.Weapon{Name: "Психический клинок", Dice: "1d6", Ability: "dex", Type: "psychic", Finesse: true})
	}
	return out
}

func monsterDerive(c *model.Character) model.Derived {
	s := c.Stat
	acFx, _ := effectBonus(c)
	out := model.Derived{Mods: map[string]int{}, Saves: map[string]int{}, MaxHP: s.MaxHP, AC: s.AC + c.ACBonus + acFx, Speed: s.Speed, Effects: effectNames(c),
		AtkBonus: s.Attack, Traits: s.Traits, Weapons: allWeapons(c), Load: totalWeight(c), Attacks: max(1, len(s.Multi))}
	for k, v := range c.Abilities {
		out.Mods[k], out.Saves[k] = Mod(v), Mod(v)
	}
	out.Initiative = out.Mods["dex"]
	for _, sp := range s.Specials {
		out.Features = append(out.Features, model.Feature{Key: sp.Key, Name: sp.Name, Kind: "special", Mode: sp.Mode, On: c.Used[sp.Key], Desc: specialText(sp)})
	}
	return out
}

// critMin: с какого значения кубика попадание критическое (улучшенный критический удар чемпиона).
func critMin(c *model.Character) int {
	if c.Stat == nil && c.Class == "fighter" && c.Subclass == "champion" {
		switch {
		case c.Level >= 15:
			return 18
		case c.Level >= 3:
			return 19
		}
	}
	return 20
}

// Attack: d20 + модификатор + мастерство против КД. Крит – 20 (или порог чемпиона), 1 – промах.
// Существа бьют с фиксированным бонусом. По парализованному или бессознательному цели попадание – критическое.
func (d DnD5e) Attack(r dice.Roller, a, t *model.Character, w model.Weapon, m dice.Mode) model.AttackResult {
	da, dt := d.Derive(a), d.Derive(t)
	ab := abilityOr(w.Ability, "str")
	if w.Finesse && da.Mods["dex"] > da.Mods[ab] {
		ab = "dex"
	}
	mod := da.Mods[ab]
	bonus := mod + da.Prof + w.Bonus
	if a.Stat != nil {
		bonus = da.AtkBonus
	}
	mode := condModeR(a, t, m, w.Ranged)
	kept, rolls := dice.D20(r, mode)
	fxBonus, _ := atkDice(r, a)
	total := kept + bonus + fxBonus
	res := model.AttackResult{Rolls: rolls, Kept: kept, Total: total, Target: dt.AC}
	cm := critMin(a)
	res.Hit = kept >= cm || (kept != 1 && total >= dt.AC)
	res.Crit = res.Hit && (kept >= cm || hasAny(t, dndAutoCrit) || t.Down() || (a.Stat == nil && riderCrit(a)))
	if !res.Hit {
		return res
	}
	e, err := dice.Parse(w.Dice)
	if err != nil {
		e = dice.Expr{Count: 1, Sides: 4}
	}
	dmg := e.Roll(r, res.Crit)
	if a.Stat != nil {
		res.DamageText = fmt.Sprint(dmg.Rolls, " ", fmt.Sprintf("%+d", e.Mod))
	} else {
		extra := mod + w.Bonus
		if a.Active["rage"] && ab == "str" && !w.Ranged {
			extra += rageBonus(clampLevel(a.Level))
		}
		dmg.Total += extra
		res.DamageText = fmt.Sprintf("%v %+d", dmg.Rolls, extra)
		if a.Class == "rogue" && (w.Finesse || w.Ranged) && mode == dice.Advantage && !a.Active["sneaked"] {
			n := (clampLevel(a.Level) + 1) / 2
			sn := dice.Expr{Count: n, Sides: 6}.Roll(r, res.Crit)
			dmg.Total += sn.Total
			res.DamageText += fmt.Sprintf(" + скрытая атака %v", sn.Rolls)
			setActive(a, "sneaked", true)
		}
	}
	res.Damage = max(0, int(float64(dmg.Total)*d.Resist(t, w.Type)))
	if a.Stat == nil {
		n, note := d.applyRiders(r, a, t, w, res.Crit)
		res.Damage += n
		res.DamageText += note
		n, note = d.autoBonus(r, a, t, w, res.Crit)
		res.Damage += n
		res.DamageText += note
		res.Damage = max(0, res.Damage)
	}
	return res
}

var physicalTypes = []string{"slashing", "piercing", "bludgeoning"}

// Resist: множитель урона. У существ – иммунитет (0), уязвимость (×2), сопротивление (×½); у героев – расовое и ярость варвара.
func (DnD5e) Resist(def *model.Character, ty string) float64 {
	if s := def.Stat; s != nil {
		switch {
		case slices.Contains(s.Immune, ty):
			return 0
		case slices.Contains(s.Vuln, ty):
			return 2
		case slices.Contains(s.Resist, ty):
			return .5
		}
		return 1
	}
	if race, err := merge(dndRaces, def.Race, def.Subrace); err == nil && slices.Contains(race.Resist, ty) {
		return .5
	}
	if def.Active["rage"] && (slices.Contains(physicalTypes, ty) || def.Subclass == "totem" && clampLevel(def.Level) >= 3 && ty != "psychic") {
		return .5
	}
	return 1
}

// Rest: battle – конец боя (сбрасывает ярость и скрытую атаку), short/long – отдыхи.
func (d DnD5e) Rest(c *model.Character, kind string) {
	switch kind {
	case "battle":
		delete(c.Active, "rage")
		delete(c.Active, "sneaked")
		delete(c.Active, "tb")
		restPools(c, kind)
	case "short":
		delete(c.Used, "breath")
		delete(c.Used, "secondwind")
		delete(c.Active, "sneaked")
		restPools(c, kind)
		if c.Class == "warlock" {
			c.SlotsUsed = [9]int{}
		}
	case "long":
		c.Used, c.Active, c.Counters, c.SlotsUsed = nil, nil, nil, [9]int{}
		c.TempHP, c.DeathOk, c.DeathFail, c.Stable = 0, 0, 0, false
		if !c.Dead {
			c.HP = d.Derive(c).MaxHP
		}
	}
}

func (d DnD5e) Toggle(c *model.Character, key string) (string, error) {
	if key != "rage" || c.Class != "barbarian" {
		return "", errors.New("у героя нет такой способности")
	}
	if c.Active["rage"] {
		setActive(c, "rage", false)
		return c.Name + " успокаивается", nil
	}
	lvl := clampLevel(c.Level)
	switch {
	case c.Down():
		return "", errors.New("герой без сознания")
	case heavyWorn(c):
		return "", errors.New("в тяжёлых доспехах впасть в ярость нельзя")
	case c.Counters["rage"] >= rages(lvl):
		return "", errors.New("ярость закончилась – нужен долгий отдых")
	}
	c.Count("rage", 1)
	setActive(c, "rage", true)
	return fmt.Sprintf("%s впадает в ярость (+%d к урону)", c.Name, rageBonus(lvl)), nil
}

func (d DnD5e) Act(r dice.Roller, c *model.Character, key string) (string, error) {
	if key != "secondwind" || c.Class != "fighter" {
		return "", errors.New("у героя нет такой способности")
	}
	switch {
	case c.Dead || c.Down():
		return "", errors.New("герой без сознания")
	case c.Used["secondwind"]:
		return "", errors.New("второе дыхание уже использовано – нужен отдых")
	}
	n := dice.Expr{Count: 1, Sides: 10, Mod: clampLevel(c.Level)}.Roll(r, false).Total
	before := c.HP
	c.Heal(n, d.Derive(c).MaxHP)
	c.Mark("secondwind")
	return fmt.Sprintf("%s: второе дыхание, +%d HP (%d → %d)", c.Name, c.HP-before, before, c.HP), nil
}

func tier(lvl int) int { return 1 + b2i(lvl >= 5) + b2i(lvl >= 11) + b2i(lvl >= 17) }

// Spell творит заклинание. Заклинания без автоматизации просто тратят ячейку. Все проверки идут до траты ячейки.
func (d DnD5e) Spell(r dice.Roller, src *model.Character, s model.Spell, slot int, targets []*model.Character, m dice.Mode) ([]string, error) {
	lines, err := d.spellCore(r, src, s, slot, targets, m)
	if err != nil {
		return nil, err
	}
	if f, ok := spellFx[s.Ref]; ok {
		lines = append(lines, d.applySpellFx(r, src, s.Ref, s.Name, f, targets)...)
	}
	return lines, nil
}

func (d DnD5e) spellCore(r dice.Roller, src *model.Character, s model.Spell, slot int, targets []*model.Character, m dice.Mode) ([]string, error) {
	def, ok := spellDef(s.Ref)
	if !ok && s.Auto != nil {
		a := s.Auto
		def, ok = SpellDef{ID: "custom", Name: s.Name, Level: s.Level, Mode: a.Mode, Dmg: a.Dmg, Type: a.Type, Save: a.Save, Half: a.Half, Area: a.Area, Up: a.Up, Scale: a.Scale}, true
	}
	if !ok || def.Mode == "" {
		msg, err := d.Cast(src, s, slot)
		if err != nil {
			return nil, err
		}
		return []string{msg}, nil
	}
	if def.Mode == "heal" && len(targets) == 0 {
		targets = []*model.Character{src}
	}
	if len(targets) == 0 {
		return nil, errors.New("выберите цель заклинания")
	}
	if src.Down() {
		return nil, errors.New("герой без сознания")
	}
	ds := d.Derive(src)
	msg, err := d.Cast(src, s, slot)
	if err != nil {
		return nil, err
	}
	lines := []string{msg}
	lvl := clampLevel(src.Level)
	expr, perr := dice.Parse(def.Dmg)
	if perr != nil {
		return lines, nil
	}
	bump := max(0, slot-def.Level)
	if def.Scale {
		expr.Count *= tier(lvl)
	}
	if def.Up != "" && def.Up != "ray" && def.Level > 0 {
		if up, err := dice.Parse(def.Up); err == nil {
			expr.Count += up.Count * bump
		}
	}
	hit := func(t *model.Character, dmg int, crit bool) {
		dmg = int(float64(dmg) * d.Resist(t, def.Type))
		lines = append(lines, fmt.Sprintf("→ %s: урон %d", t.Name, dmg))
		sl, _ := Strike(d, t, dmg, crit)
		lines = append(lines, sl...)
	}
	switch def.Mode {
	case "heal":
		if !def.Area {
			targets = targets[:1]
		}
		for _, t := range targets {
			n := expr.Roll(r, false).Total + ds.Mods[spellAbility[src.Class]]
			before := t.HP
			t.Heal(max(1, n), d.Derive(t).MaxHP)
			lines = append(lines, fmt.Sprintf("→ %s: +%d HP (%d → %d)", t.Name, t.HP-before, before, t.HP))
		}
	case "auto":
		for i := 0; i < 3+bump; i++ {
			hit(targets[i%len(targets)], expr.Roll(r, false).Total, false)
		}
	case "save":
		if !def.Area {
			targets = targets[:1]
		}
		dmg := expr.Roll(r, false).Total
		for _, t := range targets {
			tot := d.saveRoll(r, t, def.Save)
			lines = append(lines, fmt.Sprintf("%s: спасбросок %s против СЛ %d", t.Name, saveText(tot), ds.SpellDC))
			dm := dmg
			if tot >= ds.SpellDC {
				if !def.Half {
					lines = append(lines, "→ "+t.Name+": заклинание не действует")
					continue
				}
				dm /= 2
			}
			hit(t, dm, false)
		}
	case "attack":
		n := 1
		switch {
		case def.Rays > 0:
			n = def.Rays + b2i(def.Up == "ray")*bump
		case def.Beams:
			n = tier(lvl)
		}
		for i := 0; i < n; i++ {
			t := targets[i%len(targets)]
			dt := d.Derive(t)
			kept, _ := dice.D20(r, condMode(src, t, m))
			fb, _ := atkDice(r, src)
			tot := kept + ds.SpellAtk + fb
			ok := kept == 20 || (kept != 1 && tot >= dt.AC)
			crit := ok && (kept == 20 || hasAny(t, dndAutoCrit) || t.Down())
			lines = append(lines, fmt.Sprintf("%s: атака заклинанием %d против КД %d", t.Name, tot, dt.AC))
			if !ok {
				lines = append(lines, "→ промах")
				continue
			}
			hit(t, expr.Roll(r, crit).Total, crit)
		}
	}
	return lines, nil
}
