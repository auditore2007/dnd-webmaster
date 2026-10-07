package rules

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"heroesbook/internal/dice"
	"heroesbook/internal/model"
)

// fx описывает, что заклинание оставляет после себя: длительность, концентрацию, состояние, бонус к КД или атаке.
// Продолжительность – в раундах (ходах носителя). Rounds 0 – пока не снимут.
type fx struct {
	Dur   int
	Conc  bool
	Cond  string // состояние, которое накладывается
	Save  string // спасбросок цели при наложении; пусто – действует сразу
	Again string // повторный спасбросок в конце хода (по умолчанию как Save); "-" – без повторов
	AC    int
	Atk   string
	Tgt   string // ally | foe | self – кого выбирать
	Max   int    // сколько целей (0 – одна)
}

var spellFx = map[string]fx{
	"bless":               {Dur: 10, Conc: true, Atk: "1d4", Tgt: "ally", Max: 3},
	"bane":                {Dur: 10, Conc: true, Save: "cha", Atk: "-1d4", Tgt: "foe", Max: 3},
	"shield":              {Dur: 1, AC: 5, Tgt: "self"},
	"shieldoffaith":       {Dur: 100, Conc: true, AC: 2, Tgt: "ally"},
	"magearmor":           {AC: 3, Tgt: "ally"},
	"haste":               {Dur: 10, Conc: true, AC: 2, Tgt: "ally"},
	"slow":                {Dur: 10, Conc: true, Save: "wis", AC: -2, Tgt: "foe", Max: 6},
	"holdperson":          {Dur: 10, Conc: true, Cond: "Парализован", Save: "wis", Tgt: "foe"},
	"holdmonster":         {Dur: 10, Conc: true, Cond: "Парализован", Save: "wis", Tgt: "foe"},
	"entangle":            {Dur: 10, Conc: true, Cond: "Опутан", Save: "str", Tgt: "foe", Max: 6},
	"web":                 {Dur: 100, Conc: true, Cond: "Опутан", Save: "dex", Again: "str", Tgt: "foe", Max: 6},
	"blindnessdeafness":   {Dur: 10, Cond: "Ослеплён", Save: "con", Tgt: "foe"},
	"fear":                {Dur: 10, Conc: true, Cond: "Напуган", Save: "wis", Tgt: "foe", Max: 6},
	"causefear":           {Dur: 10, Conc: true, Cond: "Напуган", Save: "wis", Tgt: "foe"},
	"hypnoticpattern":     {Dur: 10, Conc: true, Cond: "Недееспособен", Save: "wis", Again: "-", Tgt: "foe", Max: 6},
	"banishment":          {Dur: 10, Conc: true, Cond: "Недееспособен", Save: "cha", Again: "-", Tgt: "foe"},
	"fleshtostone":        {Dur: 10, Conc: true, Cond: "Опутан", Save: "con", Tgt: "foe"},
	"invisibility":        {Dur: 600, Conc: true, Cond: "Невидим", Tgt: "ally"},
	"greaterinvisibility": {Dur: 10, Conc: true, Cond: "Невидим", Tgt: "ally"},
	"faeriefire":          {Dur: 10, Conc: true, Cond: "Подсвечен", Save: "dex", Again: "-", Tgt: "foe", Max: 6},
	"grease":              {Dur: 1, Cond: "Сбит с ног", Save: "dex", Again: "-", Tgt: "foe", Max: 6},
	"huntersmark":         {Dur: 600, Conc: true, Tgt: "foe"},
	"hex":                 {Dur: 600, Conc: true, Tgt: "foe"},
	"heroism":             {Dur: 10, Conc: true, Tgt: "ally"},
	"fly":                 {Dur: 100, Conc: true, Tgt: "ally"},
	"levitate":            {Dur: 100, Conc: true, Tgt: "ally"},
	"blur":                {Dur: 10, Conc: true, Tgt: "self"},
	"mirrorimage":         {Dur: 100, Tgt: "self"},
	"moonbeam":            {Dur: 10, Conc: true, Tgt: "self"},
	"spiritguardians":     {Dur: 100, Conc: true, Tgt: "self"},
	"calllightning":       {Dur: 100, Conc: true, Tgt: "self"},
	"flamingsphere":       {Dur: 10, Conc: true, Tgt: "self"},
	"spikegrowth":         {Dur: 100, Conc: true, Tgt: "self"},
	"silence":             {Dur: 100, Conc: true, Tgt: "self"},
	"darkness":            {Dur: 100, Conc: true, Tgt: "self"},
	"fogcloud":            {Dur: 600, Conc: true, Tgt: "self"},
	"sleetstorm":          {Dur: 10, Conc: true, Tgt: "self"},
	"stormsphere":         {Dur: 10, Conc: true, Tgt: "self"},
	"wallofforce":         {Dur: 100, Conc: true, Tgt: "self"},
	"wallofstone":         {Dur: 100, Conc: true, Tgt: "self"},
	"wallofthorns":        {Dur: 100, Conc: true, Tgt: "self"},
	"walloffire":          {Dur: 10, Conc: true, Tgt: "self"},
	"wallofice":           {Dur: 10, Conc: true, Tgt: "self"},
	"cloudkill":           {Dur: 100, Conc: true, Tgt: "self"},
	"stinkingcloud":       {Dur: 10, Conc: true, Tgt: "self"},
	"confusion":           {Dur: 10, Conc: true, Cond: "Недееспособен", Save: "wis", Tgt: "foe", Max: 6},
	"polymorph":           {Dur: 600, Conc: true, Tgt: "ally"},
	"dominateperson":      {Dur: 100, Conc: true, Cond: "Очарован", Save: "wis", Tgt: "foe"},
	"dominatebeast":       {Dur: 100, Conc: true, Cond: "Очарован", Save: "wis", Tgt: "foe"},
	"charmperson":         {Dur: 600, Cond: "Очарован", Save: "wis", Again: "-", Tgt: "foe"},
	"crownofmadness":      {Dur: 10, Conc: true, Cond: "Очарован", Save: "wis", Tgt: "foe"},
	"slowmo":              {},
}

func init() {
	delete(spellFx, "slowmo")
	for i := range spellDefs {
		if f, ok := spellFx[spellDefs[i].ID]; ok {
			spellDefs[i].Conc, spellDefs[i].Tgt = f.Conc, f.Tgt
		}
	}
}

// SpellConc: держится ли заклинание на концентрации.
func SpellConc(ref string) bool { return spellFx[ref].Conc }

var effSeq atomic.Int64

func nextEffID() int64 {
	for {
		old := effSeq.Load()
		n := max(old+1, time.Now().UnixNano())
		if effSeq.CompareAndSwap(old, n) {
			return n
		}
	}
}

// AddEffect вешает эффект на существо и накладывает его состояние. Одинаковый эффект от того же источника обновляется.
func AddEffect(c *model.Character, e model.Effect) {
	if e.ID == 0 {
		e.ID = nextEffID()
	}
	for i := range c.Effects {
		if c.Effects[i].Name == e.Name && c.Effects[i].Src == e.Src {
			c.Effects[i] = e
			syncCond(c, e.Cond)
			return
		}
	}
	c.Effects = append(c.Effects, e)
	syncCond(c, e.Cond)
}

func syncCond(c *model.Character, cond string) {
	if cond != "" && !c.Has(cond) {
		c.Conditions = append(c.Conditions, cond)
	}
}

// DropEffect снимает эффект по номеру; состояние пропадает, если его больше ничто не держит.
func DropEffect(c *model.Character, i int) {
	if i < 0 || i >= len(c.Effects) {
		return
	}
	cond := c.Effects[i].Cond
	c.Effects = slices.Delete(c.Effects, i, i+1)
	if cond != "" && !slices.ContainsFunc(c.Effects, func(e model.Effect) bool { return e.Cond == cond }) {
		c.Remove(cond)
	}
}

// ClearEffects снимает все эффекты (конец боя).
func ClearEffects(c *model.Character) {
	for len(c.Effects) > 0 {
		DropEffect(c, len(c.Effects)-1)
	}
	c.Conc, c.ConcDmg = "", 0
}

func effectBonus(c *model.Character) (ac int, atk []string) {
	for _, e := range c.Effects {
		ac += e.AC
		if e.Atk != "" {
			atk = append(atk, e.Atk)
		}
	}
	return
}

// atkDice бросает бонусы к атаке от эффектов («Благословение» +1d4, «Порча» −1d4).
func atkDice(r dice.Roller, c *model.Character) (int, string) {
	total, note := 0, ""
	for i := len(c.Effects) - 1; i >= 0; i-- {
		a := c.Effects[i].Atk
		if a == "" {
			continue
		}
		sign := 1
		if strings.HasPrefix(a, "-") {
			sign, a = -1, a[1:]
		}
		a = strings.TrimPrefix(a, "+")
		v := 0
		if n, err := strconv.Atoi(a); err == nil {
			v = n * sign
		} else if ex, err := dice.Parse(a); err == nil {
			v = ex.Roll(r, false).Total * sign
		} else {
			continue
		}
		total += v
		note += fmt.Sprintf(" %+d", v)
		if c.Effects[i].Once {
			DropEffect(c, i)
		}
	}
	return total, note
}

var (
	autoFail  = []string{"Парализован", "Окаменел", "Без сознания", "Ошеломлён"}
	saveFatal = -1000
)

// saveRoll – спасбросок цели с учётом состояний: парализованный, окаменевший, лишённый сознания и ошеломлённый
// автоматически проваливают Силу и Ловкость; опутанный бросает Ловкость с помехой.
func (d DnD5e) saveRoll(r dice.Roller, t *model.Character, ab string) int {
	if (ab == "str" || ab == "dex") && hasAny(t, autoFail) {
		return saveFatal
	}
	mode := dice.Normal
	if ab == "dex" && t.Has("Опутан") {
		mode = dice.Disadvantage
	}
	k, _ := dice.D20(r, mode)
	return k + d.Derive(t).Saves[ab]
}

// saveVs – спасбросок против сложности. Босс с легендарным сопротивлением превращает провал в успех.
func (d DnD5e) saveVs(r dice.Roller, t *model.Character, ab string, dc int) (tot int, txt string, ok bool) {
	tot = d.saveRoll(r, t, ab)
	txt, ok = saveText(tot), tot >= dc
	if !ok && t.Stat != nil && t.Stat.LegRes > t.Counters[legResKey] {
		t.Count(legResKey, 1)
		ok = true
		txt += fmt.Sprintf(" → легендарное сопротивление (осталось %d)", t.Stat.LegRes-t.Counters[legResKey])
	}
	return tot, txt, ok
}

func saveText(total int) string {
	if total == saveFatal {
		return "автоматический провал"
	}
	return fmt.Sprint(total)
}

// tick в начале хода носителя: идёт время эффектов, у существ восстанавливаются способности.
func (d DnD5e) tick(r dice.Roller, c *model.Character) (lines []string) {
	for i := len(c.Effects) - 1; i >= 0; i-- {
		e := c.Effects[i]
		if e.Rounds <= 0 {
			continue
		}
		c.Effects[i].Rounds--
		if c.Effects[i].Rounds == 0 {
			lines = append(lines, fmt.Sprintf("%s: «%s» заканчивается", c.Name, e.Name))
			DropEffect(c, i)
		}
	}
	if c.Stat != nil {
		for _, s := range c.Stat.Specials {
			if s.Recharge && c.Used[s.Key] {
				if (dice.Expr{Count: 1, Sides: 6}).Roll(r, false).Total >= 5 {
					delete(c.Used, s.Key)
					lines = append(lines, fmt.Sprintf("%s: «%s» восстанавливается", c.Name, s.Name))
				}
			}
		}
	}
	return lines
}

// EndTurn в конце хода носителя: повторные спасброски от эффектов (паралич, страх, паутина).
func (d DnD5e) EndTurn(r dice.Roller, c *model.Character) (lines []string) {
	for i := len(c.Effects) - 1; i >= 0; i-- {
		e := c.Effects[i]
		if e.Save == "" || e.DC == 0 {
			continue
		}
		tot := d.saveRoll(r, c, e.Save)
		if tot >= e.DC {
			lines = append(lines, fmt.Sprintf("%s: повторный спасбросок %s – %s против %d, «%s» снят", c.Name, e.Save, saveText(tot), e.DC, e.Name))
			DropEffect(c, i)
		} else {
			lines = append(lines, fmt.Sprintf("%s: повторный спасбросок – %s против %d, «%s» держится", c.Name, saveText(tot), e.DC, e.Name))
		}
	}
	return lines
}

// ConcCheck: после урона заклинатель, держащий концентрацию, бросает Телосложение (СЛ – 10 или половина урона).
func (d DnD5e) ConcCheck(r dice.Roller, c *model.Character) (lines []string, lost bool) {
	if c.Conc == "" {
		c.ConcDmg = 0
		return nil, false
	}
	dmg := c.ConcDmg
	c.ConcDmg = 0
	switch {
	case c.Dead || c.Down() || hasAny(c, []string{"Недееспособен", "Парализован", "Окаменел", "Без сознания", "Ошеломлён"}):
		return []string{fmt.Sprintf("%s теряет концентрацию на «%s»", c.Name, c.Conc)}, true
	case dmg <= 0:
		return nil, false
	}
	dc := max(10, dmg/2)
	tot := d.saveRoll(r, c, "con")
	if tot >= dc {
		return []string{fmt.Sprintf("%s: концентрация на «%s» – спасбросок Телосложения %d против %d, держится", c.Name, c.Conc, tot, dc)}, false
	}
	return []string{fmt.Sprintf("%s: концентрация на «%s» – спасбросок Телосложения %s против %d, заклинание сорвано", c.Name, c.Conc, saveText(tot), dc)}, true
}

// applySpellFx накладывает эффекты заклинания после броска ячейки.
func (d DnD5e) applySpellFx(r dice.Roller, src *model.Character, ref, name string, f fx, targets []*model.Character) (lines []string) {
	ds := d.Derive(src)
	targets = slices.Clone(targets)
	if f.Tgt == "self" || (len(targets) == 0 && f.Tgt != "foe") {
		targets = []*model.Character{src}
	}
	if f.Tgt == "foe" && len(targets) == 0 {
		lines = append(lines, "цель не выбрана – эффект не наложен")
	}
	limit := max(1, f.Max)
	if len(targets) > limit {
		targets = targets[:limit]
	}
	if f.Conc {
		src.Conc, src.ConcDmg = name, 0
	}
	placed := false
	for _, t := range targets {
		e := model.Effect{Name: name, Rounds: f.Dur, Cond: f.Cond, Src: src.ID, Conc: f.Conc, AC: f.AC, Atk: f.Atk}
		if f.Save != "" {
			_, txt, ok := d.saveVs(r, t, f.Save, ds.SpellDC)
			if ok {
				lines = append(lines, fmt.Sprintf("→ %s: спасбросок %s против СЛ %d – заклинание не действует", t.Name, txt, ds.SpellDC))
				continue
			}
			lines = append(lines, fmt.Sprintf("→ %s: спасбросок %s против СЛ %d – провал", t.Name, txt, ds.SpellDC))
			if f.Again != "-" {
				e.Save, e.DC = cmp2(f.Again, f.Save), ds.SpellDC
			}
		}
		AddEffect(t, e)
		placed = true
		if f.Cond != "" {
			lines = append(lines, fmt.Sprintf("→ %s: %s (%s)", t.Name, f.Cond, name))
		} else {
			lines = append(lines, fmt.Sprintf("→ %s: действует «%s»", t.Name, name))
		}
	}
	if f.Conc && !placed {
		// ни одна цель не поддалась, но концентрация уже потрачена – метка на заклинателе
		AddEffect(src, model.Effect{Name: name, Rounds: f.Dur, Src: src.ID, Conc: true})
	}
	return lines
}

func cmp2(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
