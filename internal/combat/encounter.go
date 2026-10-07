// Package combat – бой: инициатива, очередь ходов, атаки. Правила берутся у Ruleset героя.
package combat

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	"heroesbook/internal/dice"
	"heroesbook/internal/model"
	"heroesbook/internal/rules"
)

type Lookup func(id string) *model.Character

type Encounter struct {
	Order []string        `json:"order"`
	Init  map[string]int  `json:"init"`
	Turn  int             `json:"turn"`
	Round int             `json:"round"`
	Acted bool            `json:"acted"`
	Left  int             `json:"left"` // сколько атак осталось в этот ход (дополнительная атака, мультиатака)
	Log   []string        `json:"log"`
	Seq   int             `json:"seq"`   // сколько записей добавлено за всё время; нужно для журнала сессии
	Won   string          `json:"won"`   // кто победил: heroes | monsters | ""
	React map[string]bool `json:"react"` // у кого потрачена реакция в этом раунде
	// Настройки боя: существа ходят сами, проверяют мораль, бой идёт в логове босса.
	Auto      bool     `json:"auto"`
	Morale    bool     `json:"morale"`
	Lair      bool     `json:"lair"`
	LairRound int      `json:"lairRound"` // в каком раунде логово уже действовало
	Gone      []string `json:"gone"`      // существа, которые сбежали или сдались
}

func Start(r dice.Roller, ids []string, get Lookup) (*Encounter, error) {
	e := &Encounter{Init: map[string]int{}, Round: 1, React: map[string]bool{}, Auto: true, Morale: true}
	ruleset := ""
	for _, id := range ids {
		if _, dup := e.Init[id]; dup {
			continue
		}
		c := get(id)
		if c == nil {
			return nil, fmt.Errorf("герой %q не найден", id)
		}
		if ruleset == "" {
			ruleset = c.Ruleset
		} else if c.Ruleset != ruleset {
			return nil, errors.New("в одном бою все герои должны быть из одной системы правил")
		}
		rs, err := rules.Get(c.Ruleset)
		if err != nil {
			return nil, err
		}
		e.Init[id] = rs.Initiative(r, c)
		e.Order = append(e.Order, id)
	}
	if len(e.Order) < 2 {
		return nil, errors.New("для боя нужно минимум два участника")
	}
	sort.SliceStable(e.Order, func(i, j int) bool { return e.Init[e.Order[i]] > e.Init[e.Order[j]] })
	names := make([]string, len(e.Order))
	for i, id := range e.Order {
		names[i] = fmt.Sprintf("%s %d", get(id).Name, e.Init[id])
	}
	e.logf("Бой начался. Инициатива: %s", strings.Join(names, ", "))
	e.seek(r, get)
	return e, nil
}

// Clone – независимая копия боя для ответа интерфейсу (nil остаётся nil).
func (e *Encounter) Clone() *Encounter {
	if e == nil {
		return nil
	}
	out := *e
	out.Order = slices.Clone(e.Order)
	out.Init = maps.Clone(e.Init)
	out.Log = slices.Clone(e.Log)
	out.React = maps.Clone(e.React)
	out.Gone = slices.Clone(e.Gone)
	return &out
}

func (e *Encounter) logf(f string, a ...any) {
	e.Seq++
	e.Log = append([]string{fmt.Sprintf(f, a...)}, e.Log...)
	if len(e.Log) > 100 {
		e.Log = e.Log[:100]
	}
}

// Say добавляет строку в журнал боя.
func (e *Encounter) Say(s string) { e.logf("%s", s) }

// AddAttacks добавляет атаки в текущий ход; если герой уже отатаковал, ход даёт ещё n атак.
func (e *Encounter) AddAttacks(n int) {
	if n <= 0 {
		return
	}
	if e.Acted {
		e.Acted, e.Left = false, n
		return
	}
	e.Left += n
}

func (e *Encounter) step() {
	e.Turn++
	if e.Turn >= len(e.Order) {
		e.Turn, e.Round = 0, e.Round+1
	}
}

// seek переходит к ближайшему участнику, который может действовать. В начале каждого хода
// система правил применяет состояния и спасброски от смерти; упавшие, оглушённые и удалённые пропускаются.
func (e *Encounter) seek(r dice.Roller, get Lookup) {
	for i := 0; i < len(e.Order); i++ {
		e.lairTurn(r, get)
		if e.flees(r, get) {
			i-- // сбежавший убран из очереди: на его месте уже следующий, его ход тоже нужно начать
			continue
		}
		delete(e.React, e.Order[e.Turn])
		if c := get(e.Order[e.Turn]); c != nil && !c.Dead {
			if rs, err := rules.Get(c.Ruleset); err == nil {
				lines, skip := rs.TurnStart(r, c)
				for _, l := range lines {
					e.logf("%s", l)
				}
				e.settle(r, get)
				if !skip && !c.Down() {
					e.Acted = false
					e.Left = max(1, rs.Derive(c).Attacks)
					return
				}
			}
		}
		e.step()
	}
}

func (e *Encounter) Next(r dice.Roller, get Lookup) {
	if c := get(e.Current()); c != nil && !c.Dead && !c.Down() {
		if rs, err := rules.Get(c.Ruleset); err == nil {
			for _, l := range rs.EndTurn(r, c) {
				e.logf("%s", l)
			}
		}
	}
	if e.Auto {
		e.autoLegendary(r, get)
	}
	e.step()
	e.seek(r, get)
}

// Reaction отмечает реакцию участника потраченной (или возвращает её).
func (e *Encounter) Reaction(id string) error {
	if _, ok := e.Init[id]; !ok {
		return errors.New("участник не в бою")
	}
	if e.React == nil {
		e.React = map[string]bool{}
	}
	if e.React[id] {
		delete(e.React, id)
	} else {
		e.React[id] = true
	}
	return nil
}

// settle после любого действия: проверки концентрации у получивших урон, снятие «осиротевших» эффектов,
// вторая фаза боссов.
func (e *Encounter) settle(r dice.Roller, get Lookup) {
	e.phases(get)
	for _, id := range e.Order {
		c := get(id)
		if c == nil || c.Conc == "" {
			continue
		}
		rs, err := rules.Get(c.Ruleset)
		if err != nil {
			continue
		}
		lines, lost := rs.ConcCheck(r, c)
		for _, l := range lines {
			e.logf("%s", l)
		}
		if lost {
			e.DropConc(c, get)
		}
	}
	// концентрация закончилась сама (эффект истёк)
	for _, id := range e.Order {
		c := get(id)
		if c == nil || c.Conc == "" {
			continue
		}
		held := false
		for _, oid := range e.Order {
			if o := get(oid); o != nil {
				for _, ef := range o.Effects {
					if ef.Conc && ef.Src == id {
						held = true
					}
				}
			}
		}
		if !held {
			c.Conc = ""
		}
	}
}

// DropConc обрывает концентрацию заклинателя и снимает все её эффекты.
func (e *Encounter) DropConc(c *model.Character, get Lookup) {
	name := c.Conc
	c.Conc, c.ConcDmg = "", 0
	for _, id := range e.Order {
		o := get(id)
		if o == nil {
			continue
		}
		for i := len(o.Effects) - 1; i >= 0; i-- {
			if o.Effects[i].Conc && o.Effects[i].Src == c.ID {
				rules.DropEffect(o, i)
			}
		}
	}
	if name != "" {
		e.logf("Заклинание «%s» прекращает действовать", name)
	}
}

// concIDs – номера эффектов концентрации заклинателя (чтобы снять старые после нового заклинания).
func (e *Encounter) concIDs(src string, get Lookup) map[int64]bool {
	out := map[int64]bool{}
	for _, id := range e.Order {
		if o := get(id); o != nil {
			for _, ef := range o.Effects {
				if ef.Conc && ef.Src == src {
					out[ef.ID] = true
				}
			}
		}
	}
	return out
}

// Cleanup в конце боя: эффекты и концентрация снимаются, способности существ восстанавливаются.
func (e *Encounter) Cleanup(get Lookup) {
	for _, id := range append(slices.Clone(e.Order), e.Gone...) {
		if c := get(id); c != nil {
			rules.ClearEffects(c)
			if c.Stat != nil {
				rules.ResetBoss(c) // фаза и легендарные ресурсы – только на один бой
				c.Used = nil
			}
		}
	}
}

func (e *Encounter) Current() string {
	if e.Turn < 0 || e.Turn >= len(e.Order) {
		return ""
	}
	return e.Order[e.Turn]
}

// mixed – в бою есть и герои, и существа: тогда «враги» – это противоположная сторона. Если все участники одного вида
// (дуэль героев), врагом считается любой другой участник.
func (e *Encounter) mixed(get Lookup) bool {
	var mon, hero bool
	for _, id := range e.Order {
		if c := get(id); c != nil {
			if c.Kind == "monster" {
				mon = true
			} else {
				hero = true
			}
		}
	}
	return mon && hero
}

// Foes – стоящие на ногах противники участника id.
func (e *Encounter) Foes(get Lookup, id string) (out []*model.Character) {
	me := get(id)
	if me == nil {
		return nil
	}
	mixed := e.mixed(get)
	for _, oid := range e.Order {
		c := get(oid)
		if c == nil || oid == id || !c.Standing() {
			continue
		}
		if mixed && (c.Kind == "monster") == (me.Kind == "monster") {
			continue
		}
		out = append(out, c)
	}
	return out
}

func (e *Encounter) foes(get Lookup) []*model.Character { return e.Foes(get, e.Current()) }

// Add вводит новых участников в идущий бой: бросается инициатива, очередь сохраняется,
// ход остаётся у того, кто действовал.
func (e *Encounter) Add(r dice.Roller, ids []string, get Lookup) error {
	cur := e.Current()
	var add []string
	for _, id := range ids {
		if _, dup := e.Init[id]; dup {
			continue
		}
		c := get(id)
		if c == nil {
			return fmt.Errorf("участник %q не найден", id)
		}
		if c.Ruleset != get(cur).Ruleset {
			return errors.New("в одном бою все участники должны быть из одной системы правил")
		}
		rs, err := rules.Get(c.Ruleset)
		if err != nil {
			return err
		}
		e.Init[id] = rs.Initiative(r, c)
		add = append(add, id)
	}
	for _, id := range add {
		e.Order = append(e.Order, id)
		e.logf("В бой вступает %s (инициатива %d)", get(id).Name, e.Init[id])
	}
	sort.SliceStable(e.Order, func(i, j int) bool { return e.Init[e.Order[i]] > e.Init[e.Order[j]] })
	for i, id := range e.Order {
		if id == cur {
			e.Turn = i
		}
	}
	return nil
}

// Remove выводит участника из боя (убежал, убран мастером). Если это был текущий ход – ход переходит дальше.
func (e *Encounter) Remove(r dice.Roller, id string, get Lookup) {
	i := -1
	for k, x := range e.Order {
		if x == id {
			i = k
		}
	}
	if i < 0 {
		return
	}
	if c := get(id); c != nil && c.Conc != "" {
		e.DropConc(c, get)
	}
	delete(e.React, id)
	wasCurrent := i == e.Turn
	e.Order = append(e.Order[:i:i], e.Order[i+1:]...)
	delete(e.Init, id)
	switch {
	case len(e.Order) == 0:
		e.Turn = 0
	case i < e.Turn:
		e.Turn--
	case wasCurrent:
		if e.Turn >= len(e.Order) {
			e.Turn, e.Round = 0, e.Round+1
		}
		e.seek(r, get)
	}
}

// Winner: "heroes" или "monsters", если на ногах остались только герои либо только существа; "" – бой продолжается.
func (e *Encounter) Winner(get Lookup) string {
	var heroes, mons int
	for _, id := range e.Order {
		if c := get(id); c != nil && c.Standing() {
			if c.Kind == "monster" {
				mons++
			} else {
				heroes++
			}
		}
	}
	mixed := e.mixed(get) || (len(e.Gone) > 0 && heroes > 0)
	switch {
	case mons == 0 && heroes > 0 && mixed:
		return "heroes"
	case heroes == 0 && mons > 0 && mixed:
		return "monsters"
	}
	return ""
}

func (e *Encounter) Attack(r dice.Roller, get Lookup, targetID string, w model.Weapon, m dice.Mode) (model.AttackResult, error) {
	att, def := get(e.Current()), get(targetID)
	fail := func(s string) (model.AttackResult, error) { return model.AttackResult{}, errors.New(s) }
	switch {
	case e.Acted:
		return fail("в этот ход герой уже действовал")
	case att == nil || def == nil:
		return fail("атакующий или цель не найдены")
	case att.ID == def.ID:
		return fail("нельзя атаковать самого себя")
	case att.Down():
		return fail("атакующий без сознания")
	case def.Dead:
		return fail("цель уже погибла")
	case att.Ruleset != def.Ruleset:
		return fail("герои из разных систем правил")
	}
	if _, ok := e.Init[targetID]; !ok {
		return fail("цель не участвует в бою")
	}
	rs, err := rules.Get(att.Ruleset)
	if err != nil {
		return model.AttackResult{}, err
	}
	if def.Down() && !rs.Catalog().DeathSaves {
		return fail("цель уже без сознания")
	}
	res := rs.Attack(r, att, def, w, m)
	defer e.settle(r, get)
	e.Left--
	e.Acted = e.Left <= 0
	e.resolve(r, rs, att, def, w, res, "")
	return res, nil
}

// resolve записывает атаку в журнал и наносит урон; цель с правом ответного удара бьёт в ответ.
func (e *Encounter) resolve(r dice.Roller, rs rules.Ruleset, att, def *model.Character, w model.Weapon, res model.AttackResult, pre string) {
	e.report(rs, att, def, w, res, pre)
	if !res.Hit {
		return
	}
	lines, counter := rules.Strike(rs, def, res.Damage, res.Crit)
	for _, l := range lines {
		e.logf("%s", l)
	}
	if counter && def.Standing() && att.Standing() {
		cw := model.Unarmed
		if ws := rs.Derive(def).Weapons; len(ws) > 0 {
			cw = ws[0]
		}
		cr := rs.Attack(r, def, att, cw, dice.Normal)
		e.report(rs, def, att, cw, cr, "↩️ ")
		if cr.Hit {
			cl, _ := rules.Strike(rs, att, cr.Damage, cr.Crit)
			for _, l := range cl {
				e.logf("%s", l)
			}
		}
	}
}

func (e *Encounter) report(rs rules.Ruleset, a, d *model.Character, w model.Weapon, res model.AttackResult, pre string) {
	out := "промах"
	if res.Hit {
		out = fmt.Sprintf("попадание, урон %d", res.Damage)
		if res.Crit {
			out = "критическое " + out
		}
	}
	e.logf("%s%s → %s (%s): %d против %d – %s", pre, a.Name, d.Name, w.Name, res.Total, res.Target, out)
}

// Special применяет особую способность текущего участника (например, дыхание дракона) по всем врагам.
func (e *Encounter) Special(r dice.Roller, get Lookup, key string) error {
	return e.SpecialAt(r, get, key, "")
}

// SpecialAt: то же, но способность «по одной цели» бьёт указанного врага.
func (e *Encounter) SpecialAt(r dice.Roller, get Lookup, key, targetID string) error {
	att := get(e.Current())
	switch {
	case e.Acted:
		return errors.New("в этот ход герой уже действовал")
	case att == nil || att.Down():
		return errors.New("участник не может действовать")
	}
	rs, err := rules.Get(att.Ruleset)
	if err != nil {
		return err
	}
	foes := e.foes(get)
	if targetID != "" {
		t := get(targetID)
		if t == nil || !slices.ContainsFunc(foes, func(c *model.Character) bool { return c.ID == targetID }) {
			return errors.New("цель недоступна")
		}
		foes = []*model.Character{t}
	}
	lines, err := rs.Special(r, key, att, foes)
	if err != nil {
		return err
	}
	defer e.settle(r, get)
	e.Acted, e.Left = true, 0
	for _, l := range lines {
		e.logf("%s", l)
	}
	return nil
}

// Cast: заклинание текущего участника. Тратит действие целиком; ячейка списывается только при успехе.
func (e *Encounter) Cast(r dice.Roller, get Lookup, idx, slot int, targetIDs []string, m dice.Mode) error {
	att := get(e.Current())
	switch {
	case e.Acted:
		return errors.New("в этот ход герой уже действовал")
	case att == nil || att.Down():
		return errors.New("участник не может действовать")
	case idx < 0 || idx >= len(att.Spells):
		return errors.New("нет такого заклинания")
	}
	rs, err := rules.Get(att.Ruleset)
	if err != nil {
		return err
	}
	var targets []*model.Character
	for _, id := range targetIDs {
		t := get(id)
		if t == nil {
			return fmt.Errorf("цель %q не найдена", id)
		}
		if _, ok := e.Init[id]; !ok {
			return fmt.Errorf("%s не участвует в бою", t.Name)
		}
		targets = append(targets, t)
	}
	old := e.concIDs(att.ID, get)
	oldName := att.Conc
	lines, err := rs.Spell(r, att, att.Spells[idx], slot, targets, m)
	if err != nil {
		return err
	}
	defer e.settle(r, get)
	if len(old) > 0 && rules.SpellConc(att.Spells[idx].Ref) {
		for _, id := range e.Order {
			if o := get(id); o != nil {
				for i := len(o.Effects) - 1; i >= 0; i-- {
					if old[o.Effects[i].ID] {
						rules.DropEffect(o, i)
					}
				}
			}
		}
		lines = append([]string{fmt.Sprintf("%s прерывает концентрацию на «%s»", att.Name, oldName)}, lines...)
	}
	e.Acted, e.Left = true, 0
	for _, l := range lines {
		e.logf("%s", l)
	}
	return nil
}
