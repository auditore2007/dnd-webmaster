package rules

import (
	"fmt"
	"slices"
	"sync"

	"heroesbook/internal/dice"
	"heroesbook/internal/model"
)

// DnD5e – правила D&D 5e (SRD): шесть характеристик, КД, бонус мастерства.
type DnD5e struct{}

type bonus = map[string]int

var dndAbilities = []Ability{{"str", "Сила"}, {"dex", "Ловкость"}, {"con", "Телосложение"}, {"int", "Интеллект"}, {"wis", "Мудрость"}, {"cha", "Харизма"}}

var dndConditions = []string{"Ослеплён", "Очарован", "Оглохший", "Напуган", "Схвачен", "Недееспособен", "Невидим", "Парализован", "Окаменел", "Отравлен", "Сбит с ног", "Опутан", "Ошеломлён", "Без сознания", "Концентрация", "Уклонение", "Преимущество", "Безрассудство"}

var dndRests = []Ability{{"short", "Короткий отдых"}, {"long", "Долгий отдых"}}

var dndDamage = []Ability{{"slashing", "Рубящий"}, {"piercing", "Колющий"}, {"bludgeoning", "Дробящий"}, {"fire", "Огонь"}, {"cold", "Холод"}, {"lightning", "Молния"}, {"thunder", "Звук"}, {"poison", "Яд"}, {"acid", "Кислота"}, {"necrotic", "Некротический"}, {"radiant", "Излучение"}, {"psychic", "Психический"}, {"force", "Силовое поле"}}

func (DnD5e) ID() string { return "dnd5e" }

// Catalog собирается один раз: его вызывают почти в каждом действии, а данные неизменны. Вызывающие только читают его.
func (DnD5e) Catalog() Catalog { return dndCatalog() }

var dndCatalog = sync.OnceValue(func() Catalog {
	return Catalog{ID: "dnd5e", Name: "D&D 5e (SRD)", ACLabel: "КД", DeathSaves: true, Abilities: dndAbilities, Races: withText(dndRaces), Classes: withClassText(dndClasses), Conditions: dndConditions, DamageTypes: dndDamage, Rests: dndRests, Skills: dndSkills, Spells: spellDefs}
})

func findClass(id string) (Class, bool) {
	i := slices.IndexFunc(dndClasses, func(c Class) bool { return c.ID == id })
	if i < 0 {
		return Class{}, false
	}
	return dndClasses[i], true
}

// Mod – модификатор характеристики. Сдвиг округляет вниз и для отрицательных (9 → −1), деление в Go – нет.
func Mod(score int) int { return (score - 10) >> 1 }

func (d DnD5e) Build(c *model.Character) error {
	race, err := merge(dndRaces, c.Race, c.Subrace)
	if err != nil {
		return err
	}
	if _, ok := findClass(c.Class); !ok {
		return fmt.Errorf("неизвестный класс: %q", c.Class)
	}
	c.Level = 1
	c.Abilities = map[string]int{}
	for _, a := range dndAbilities {
		c.Abilities[a.ID] = 10 + race.Bonus[a.ID]
	}
	c.HP = d.Derive(c).MaxHP
	return nil
}

// Ячейки заклинаний полного заклинателя по уровням 1–20: цифры – число ячеек 1-го, 2-го… уровня.
var fullSlots = []string{"2", "3", "42", "43", "432", "433", "4331", "4332", "43331", "43332", "433321", "433321", "4333211", "4333211", "43332111", "43332111", "433321111", "433331111", "433332111", "433332211"}

// Ячейки «третьего заклинателя» (мистический рыцарь, mistic trickster) по уровням 3–20.
var thirdSlots = map[int]string{3: "2", 4: "3", 5: "3", 6: "3", 7: "42", 8: "42", 9: "42", 10: "43", 11: "43", 12: "43", 13: "432", 14: "432", 15: "432", 16: "433", 17: "433", 18: "433", 19: "4331", 20: "4331"}

// thirdCaster: воин-мистический рыцарь и плут-мистический ловкач колдуют как третьи заклинатели на Интеллекте.
func thirdCaster(c *model.Character) bool {
	return c.Stat == nil && c.Level >= 3 && (c.Class == "fighter" && c.Subclass == "eldritchknight" || c.Class == "rogue" && c.Subclass == "arcanetrickster")
}

func slotsOf(c *model.Character) []model.SlotView {
	if thirdCaster(c) {
		var out []model.SlotView
		for i, ch := range thirdSlots[clampLevel(c.Level)] {
			out = append(out, model.SlotView{Level: i + 1, Max: int(ch - '0')})
		}
		return out
	}
	return slotsFor(c.Class, clampLevel(c.Level))
}

func slotsFor(class string, level int) []model.SlotView {
	caster := level
	switch class {
	case "warlock":
		return pactSlots(level)
	case "bard", "cleric", "druid", "sorcerer", "wizard":
	case "artificer":
		caster = (level + 1) / 2
	case "paladin", "ranger":
		caster = (level + 1) / 2
		if level < 2 {
			caster = 0
		}
	default:
		return nil
	}
	if caster < 1 {
		return nil
	}
	var out []model.SlotView
	for i, ch := range fullSlots[min(caster, 20)-1] {
		out = append(out, model.SlotView{Level: i + 1, Max: int(ch - '0')})
	}
	return out
}

var dndXP = []int{0, 0, 300, 900, 2700, 6500, 14000, 23000, 34000, 48000, 64000, 85000, 100000, 120000, 140000, 165000, 195000, 225000, 265000, 305000, 355000}

func (DnD5e) XPFor(level int) int {
	if level < 1 || level >= len(dndXP) {
		return 1 << 30
	}
	return dndXP[level]
}

func (DnD5e) LevelPoints(level int) int {
	if slices.Contains([]int{4, 8, 12, 16, 19}, level) {
		return 2
	}
	return 0
}

func (d DnD5e) Initiative(r dice.Roller, c *model.Character) int {
	k, _ := dice.D20(r, dice.Normal)
	return k + d.Derive(c).Initiative
}

var (
	dndDisAttack    = []string{"Отравлен", "Напуган", "Опутан", "Ослеплён", "Сбит с ног"}
	dndAdvAgainst   = []string{"Парализован", "Окаменел", "Ошеломлён", "Без сознания", "Опутан", "Ослеплён", "Подсвечен"}
	dndAutoCrit     = []string{"Парализован", "Без сознания"}
	dndIncapacitate = []string{"Недееспособен", "Парализован", "Окаменел", "Ошеломлён", "Без сознания"}
)

func hasAny(c *model.Character, conds []string) bool {
	return slices.ContainsFunc(conds, c.Has)
}

// condMode: преимущество и помеха из состояний и выбранного режима; если есть и то и другое – они гасят друг друга.
func condMode(a, t *model.Character, m dice.Mode) dice.Mode { return condModeR(a, t, m, false) }

// condModeR: то же, но для дальних атак: по лежащему цели дальняя атака идёт с помехой, а не с преимуществом.
func condModeR(a, t *model.Character, m dice.Mode, ranged bool) dice.Mode {
	against := hasAny(t, dndAdvAgainst) || (!ranged && t.Has("Сбит с ног"))
	against = against || t.Has("Безрассудство")
	adv := m == dice.Advantage || against || a.Has("Невидим") || t.Down() || a.Has("Преимущество") || a.Has("Безрассудство")
	dis := m == dice.Disadvantage || hasAny(a, dndDisAttack) || t.Has("Невидим") || t.Has("Уклонение") || (ranged && t.Has("Сбит с ног"))
	switch {
	case adv && !dis:
		return dice.Advantage
	case dis && !adv:
		return dice.Disadvantage
	}
	return dice.Normal
}

// AfterDamage: падение на 0 HP запускает спасброски от смерти; урон по лежащему – провал (крит – два). Существа гибнут сразу.
func (DnD5e) AfterDamage(c *model.Character, wasDown, crit bool) {
	if c.HP > 0 || c.Dead {
		return
	}
	switch {
	case c.Stat != nil:
		c.Dead = true
	case !wasDown:
		c.DeathOk, c.DeathFail, c.Stable = 0, 0, false
	default:
		c.Stable = false
		c.DeathFail++
		if crit {
			c.DeathFail++
		}
		if c.DeathFail >= 3 {
			c.Dead = true
		}
	}
}

func (d DnD5e) DeathSave(r dice.Roller, c *model.Character) (string, error) {
	switch {
	case c.Stat != nil:
		return "", fmt.Errorf("существа не делают спасбросков от смерти")
	case c.Dead:
		return "", fmt.Errorf("%s уже погиб", c.Name)
	case c.HP > 0:
		return "", fmt.Errorf("%s в сознании", c.Name)
	case c.Stable:
		return "", fmt.Errorf("%s уже стабилизирован", c.Name)
	}
	k, _ := dice.D20(r, dice.Normal)
	out := fmt.Sprintf("спасбросок от смерти: %d", k)
	switch {
	case k == 20:
		c.HP, c.DeathOk, c.DeathFail = 1, 0, 0
		return out + " – приходит в себя с 1 HP", nil
	case k == 1:
		c.DeathFail += 2
	case k >= 10:
		c.DeathOk++
	default:
		c.DeathFail++
	}
	switch {
	case c.DeathFail >= 3:
		c.Dead = true
		out += " – погибает"
	case c.DeathOk >= 3:
		c.Stable, c.DeathOk, c.DeathFail = true, 0, 0
		out += " – стабилизируется"
	default:
		out += fmt.Sprintf(" (успехов %d, провалов %d)", c.DeathOk, c.DeathFail)
	}
	return out, nil
}

func (d DnD5e) TurnStart(r dice.Roller, c *model.Character) ([]string, bool) {
	if c.Dead {
		return nil, true
	}
	lines := d.tick(r, c)
	switch {
	case c.HP > 0:
		delete(c.Active, "sneaked")
		if hasAny(c, dndIncapacitate) {
			lines = append(lines, c.Name+" не может действовать")
			return append(lines, d.EndTurn(r, c)...), true
		}
		return lines, false
	}
	if msg, err := d.DeathSave(r, c); err == nil {
		return append(lines, c.Name+": "+msg), true
	}
	return lines, true
}

func (DnD5e) Survives(c *model.Character) (string, bool, bool) {
	if c.Race == "halforc" && c.Stat == nil && !c.Used["relentless"] {
		c.Mark("relentless")
		return c.Name + ": непоколебимая стойкость – остаётся с 1 HP", false, true
	}
	return "", false, false
}

// Special: дыхание дракорождённого – 2d6 (3d6/4d6/5d6 с 6/11/16 уровня), СЛ = 8 + Телосложение + мастерство.
func (d DnD5e) Special(r dice.Roller, key string, src *model.Character, foes []*model.Character) ([]string, error) {
	if src.Stat != nil {
		return d.monsterSpecial(r, key, src, foes)
	}
	if key != "breath" {
		return d.heroSpecial(r, key, src, foes)
	}
	if src.Race != "dragonborn" {
		return nil, fmt.Errorf("у героя нет такой способности")
	}
	if src.Used["breath"] {
		return nil, fmt.Errorf("дыхание уже использовано – нужен отдых")
	}
	ds := d.Derive(src)
	dc, n := 8+ds.Mods["con"]+ds.Prof, 2+b2i(src.Level >= 6)+b2i(src.Level >= 11)+b2i(src.Level >= 16)
	src.Mark("breath")
	lines := []string{fmt.Sprintf("%s: дыхание дракона (%dd6, СЛ %d)", src.Name, n, dc)}
	for _, f := range foes {
		tot := d.saveRoll(r, f, "dex")
		dmg := dice.Expr{Count: n, Sides: 6}.Roll(r, false).Total
		if tot >= dc {
			dmg /= 2
		}
		dmg = int(float64(dmg) * d.Resist(f, "fire"))
		lines = append(lines, fmt.Sprintf("→ %s: спасбросок %s против %d, урон %d", f.Name, saveText(tot), dc, dmg))
		sl, _ := Strike(d, f, dmg, false)
		lines = append(lines, sl...)
	}
	return lines, nil
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (d DnD5e) Cast(c *model.Character, s model.Spell, slot int) (string, error) {
	if s.Level == 0 {
		return c.Name + " читает заговор «" + s.Name + "»", nil
	}
	if slot < s.Level || slot > 9 {
		return "", fmt.Errorf("для «%s» нужна ячейка не ниже %d уровня", s.Name, s.Level)
	}
	for _, sv := range d.Derive(c).Slots {
		if sv.Level == slot {
			if c.SlotsUsed[slot-1] >= sv.Max {
				return "", fmt.Errorf("ячейки %d уровня закончились", slot)
			}
			c.SlotsUsed[slot-1]++
			return fmt.Sprintf("%s творит «%s» (ячейка %d уровня)", c.Name, s.Name, slot), nil
		}
	}
	return "", fmt.Errorf("нет ячеек %d уровня", slot)
}
