// Package model – сущности, не зависящие от конкретной системы правил.
package model

import (
	"maps"
	"slices"
)

type Weapon struct {
	Name    string `json:"name"`
	Dice    string `json:"dice"`
	Ability string `json:"ability"`
	Type    string `json:"type"`
	Finesse bool   `json:"finesse"` // можно бить Ловкостью
	Ranged  bool   `json:"ranged"`
	Bonus   int    `json:"bonus"` // магический бонус к атаке и урону (+1, +2, +3)
}

// Unarmed – запасной вариант, когда оружие не выбрано.
var Unarmed = Weapon{Name: "Безоружный удар", Dice: "1d4", Ability: "str", Type: "bludgeoning"}

// Item – предмет. В библиотеке (ID задан) это шаблон, в инвентаре – копия со своим количеством.
type Item struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Qty          int     `json:"qty"`
	Weight       float64 `json:"weight"`
	Desc         string  `json:"desc"`
	ACBonus      int     `json:"acBonus"`
	Ability      string  `json:"ability"`
	AbilityBonus int     `json:"abilityBonus"`
	Weapon       *Weapon `json:"weapon"`
	Equipped     bool    `json:"equipped"`
	ArmorBase    int     `json:"armorBase"` // базовый КД доспеха, 0 – не доспех
	ArmorType    string  `json:"armorType"` // light | medium | heavy
	Cat          string  `json:"cat"`       // weapon | armor | gear | tool | magic | potion | mount …
	Price        string  `json:"price"`
	Rarity       string  `json:"rarity"`
}

type Spell struct {
	Name  string `json:"name"`
	Level int    `json:"level"` // D&D: уровень ячейки, 0 – заговор
	Cost  int    `json:"cost"`  // Алдоран: мана
	Note  string `json:"note"`
	Ref   string `json:"ref"` // id заклинания из каталога
	// Auto – автоматизация собственного заклинания (когда его нет в каталоге).
	Auto *SpellAuto `json:"auto,omitempty"`
}

// SpellAuto описывает, как считать собственное заклинание: mode = attack | save | auto | heal.
type SpellAuto struct {
	Mode  string `json:"mode"`
	Dmg   string `json:"dmg"`
	Type  string `json:"type"`
	Save  string `json:"save"`
	Half  bool   `json:"half"`
	Area  bool   `json:"area"`
	Up    string `json:"up"`    // доп. кубики за уровень ячейки выше базового
	Scale bool   `json:"scale"` // заговор растёт с 5/11/17 уровня
}

// SpellTemplate – собственное заклинание в библиотеке мастера.
type SpellTemplate struct {
	ID     string     `json:"id"`
	Name   string     `json:"name"`
	Level  int        `json:"level"`
	School string     `json:"school"`
	Desc   string     `json:"desc"`
	Auto   *SpellAuto `json:"auto,omitempty"`
}

// Effect – действующий эффект на существе: заклинание с длительностью, состояние от умения монстра.
// Тикает в конце хода носителя; концентрационные эффекты снимаются, когда заклинатель теряет концентрацию.
type Effect struct {
	Name   string `json:"name"`
	Rounds int    `json:"rounds"` // сколько ходов носителя осталось; 0 – пока не снимут вручную
	Cond   string `json:"cond"`   // состояние, которое накладывает эффект
	Src    string `json:"src"`    // id заклинателя
	Conc   bool   `json:"conc"`   // держится на концентрации Src
	Save   string `json:"save"`   // повторный спасбросок в конце хода (характеристика), "" – нет
	DC     int    `json:"dc"`
	AC     int    `json:"ac"`   // бонус к КД
	Atk    string `json:"atk"`  // бонус к броску атаки, напр. "1d4"; со знаком «-» – штраф
	Once   bool   `json:"once"` // бонус к атаке тратится на первом же броске
	ID     int64  `json:"id"`
}

// Special – особая способность существа, которую можно применить из боя (дыхание, яд, паралич, захват).
type Special struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Mode     string `json:"mode"`   // area – по всем врагам; single – по одной цели
	Attack   bool   `json:"attack"` // single: сначала бросок атаки по КД
	Dmg      string `json:"dmg"`    // area: урон со спасброском; single: урон при попадании
	Type     string `json:"type"`
	XDmg     string `json:"xDmg"` // дополнительный урон со спасброском (яд, второй тип урона)
	XType    string `json:"xType"`
	Save     string `json:"save"`  // характеристика спасброска цели
	RSave    string `json:"rSave"` // характеристика повторного спасброска (по умолчанию как Save)
	DC       int    `json:"dc"`
	Half     bool   `json:"half"`
	Cond     string `json:"cond"`     // состояние при провале спасброска
	Rounds   int    `json:"rounds"`   // сколько ходов длится состояние
	Repeat   bool   `json:"repeat"`   // цель повторяет спасбросок в конце хода
	Recharge bool   `json:"recharge"` // восстанавливается на 5–6 в начале хода
	Once     bool   `json:"once"`     // один раз за бой
	Desc     string `json:"desc"`
	OnlyKind string `json:"onlyKind"` // area: действует только на существ этого типа (изгнание нежити)
}

// MonsterStat – фиксированные значения существа из бестиария.
type MonsterStat struct {
	CR       string    `json:"cr"`
	Kind     string    `json:"kind"` // тип существа (для картинки и фильтров)
	AC       int       `json:"ac"`
	MaxHP    int       `json:"maxHp"`
	Attack   int       `json:"attack"`
	Speed    int       `json:"speed"`
	Traits   []string  `json:"traits"`
	ID       string    `json:"id"`
	Multi    []int     `json:"multi"` // индексы оружия в мультиатаке
	Resist   []string  `json:"resist"`
	Vuln     []string  `json:"vuln"`
	Immune   []string  `json:"immune"`
	Specials []Special `json:"specials"`
}

type Character struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Kind       string          `json:"kind"` // "" – герой, "monster" – существо
	Ruleset    string          `json:"ruleset"`
	Race       string          `json:"race"`
	Subrace    string          `json:"subrace"`
	Class      string          `json:"class"`
	Level      int             `json:"level"`
	XP         int             `json:"xp"`
	Points     int             `json:"points"` // нераспределённые очки характеристик
	Abilities  map[string]int  `json:"abilities"`
	HP         int             `json:"hp"`
	TempHP     int             `json:"tempHp"`
	MP         int             `json:"mp"`
	ACBonus    int             `json:"acBonus"`
	Conditions []string        `json:"conditions"`
	Weapons    []Weapon        `json:"weapons"`
	Inventory  []Item          `json:"inventory"`
	Spells     []Spell         `json:"spells"`
	SlotsUsed  [9]int          `json:"slotsUsed"`
	DeathOk    int             `json:"deathOk"`
	DeathFail  int             `json:"deathFail"`
	Stable     bool            `json:"stable"`
	Dead       bool            `json:"dead"`
	Active     map[string]bool `json:"active"` // включённые способности
	Used       map[string]bool `json:"used"`   // потраченные способности
	Stat       *MonsterStat    `json:"stat,omitempty"`
	Subclass   string          `json:"subclass"`
	Skills     []string        `json:"skills"`
	Expert     []string        `json:"expert"`
	Portrait   string          `json:"portrait"` // сжатая картинка data:image/…
	Counters   map[string]int  `json:"counters"`
	Notes      string          `json:"notes"`
	Effects    []Effect        `json:"effects"`
	Conc       string          `json:"conc"`    // заклинание, на котором держится концентрация
	ConcDmg    int             `json:"concDmg"` // урон, полученный с последней проверки концентрации
	Purse      map[string]int  `json:"purse"`   // монеты: pp gp ep sp cp
}

type Feature struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Kind string `json:"kind"` // toggle | once | special | passive
	On   bool   `json:"on"`   // включено или потрачено
	Desc string `json:"desc"`
	Mode string `json:"mode"` // для special: area | single
	// Для способностей с ресурсом и целью (Kind == "ability"):
	Target string `json:"target"` // "" – на себя; ally – союзник; foe – враг
	Pool   string `json:"pool"`   // название ресурса («Кости превосходства»)
	Cost   int    `json:"cost"`   // сколько единиц ресурса тратится
	Left   int    `json:"left"`   // осталось единиц ресурса
	Max    int    `json:"max"`
	Rest   string `json:"rest"` // short | long | battle – когда восстанавливается
}

// PoolView – ресурс класса для отображения (ки, кости превосходства, наложение рук).
type PoolView struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Left int    `json:"left"`
	Max  int    `json:"max"`
	Rest string `json:"rest"`
}

type SlotView struct {
	Level int `json:"level"`
	Max   int `json:"max"`
}

// Derived – вычисляемые значения; хранить их нельзя, иначе они разойдутся с исходными.
type Derived struct {
	MaxHP      int            `json:"maxHp"`
	MaxMP      int            `json:"maxMp"`
	AC         int            `json:"ac"`
	Prof       int            `json:"prof"`
	AtkBonus   int            `json:"atkBonus"` // для существ: фиксированный бонус атаки
	Initiative int            `json:"initiative"`
	Speed      int            `json:"speed"`
	Mods       map[string]int `json:"mods"`
	Saves      map[string]int `json:"saves"`
	Traits     []string       `json:"traits"`
	Effects    []string       `json:"effects"`
	Features   []Feature      `json:"features"`
	Pools      []PoolView     `json:"pools"`
	Weapons    []Weapon       `json:"weapons"`
	Slots      []SlotView     `json:"slots"`
	Load       float64        `json:"load"`
	Carry      float64        `json:"carry"`
	NextXP     int            `json:"nextXp"`
	Skills     map[string]int `json:"skills"`
	Attacks    int            `json:"attacks"`
	SpellDC    int            `json:"spellDc"`
	SpellAtk   int            `json:"spellAtk"`
}

type AttackResult struct {
	Hit        bool   `json:"hit"`
	Crit       bool   `json:"crit"`
	Rolls      []int  `json:"rolls"`
	Kept       int    `json:"kept"` // засчитанный d20 (преимущество или помеха могли прийти от состояний)
	Total      int    `json:"total"`
	Target     int    `json:"target"`
	Damage     int    `json:"damage"`
	DamageText string `json:"damageText"`
}

// Clone – глубокая копия: ответы интерфейсу сериализуются без мьютекса и не должны делить с состоянием срезы и карты.
func (c *Character) Clone() Character {
	out := *c
	out.Abilities = maps.Clone(c.Abilities)
	out.Active = maps.Clone(c.Active)
	out.Used = maps.Clone(c.Used)
	out.Counters = maps.Clone(c.Counters)
	out.Purse = maps.Clone(c.Purse)
	out.Conditions = slices.Clone(c.Conditions)
	out.Weapons = slices.Clone(c.Weapons)
	out.Skills = slices.Clone(c.Skills)
	out.Expert = slices.Clone(c.Expert)
	out.Effects = slices.Clone(c.Effects)
	out.Inventory = slices.Clone(c.Inventory)
	for i, it := range out.Inventory {
		if it.Weapon != nil {
			w := *it.Weapon
			out.Inventory[i].Weapon = &w
		}
	}
	out.Spells = slices.Clone(c.Spells)
	for i, s := range out.Spells {
		if s.Auto != nil {
			au := *s.Auto
			out.Spells[i].Auto = &au
		}
	}
	if c.Stat != nil {
		st := *c.Stat
		st.Traits, st.Multi, st.Specials = slices.Clone(st.Traits), slices.Clone(st.Multi), slices.Clone(st.Specials)
		st.Resist, st.Vuln, st.Immune = slices.Clone(st.Resist), slices.Clone(st.Vuln), slices.Clone(st.Immune)
		out.Stat = &st
	}
	return out
}

func (c *Character) Down() bool     { return c.HP <= 0 }
func (c *Character) Standing() bool { return c.HP > 0 && !c.Dead }

// Damage сначала снимает временные HP, затем обычные; ниже нуля не опускается.
// Смерть и спасброски от смерти – дело системы правил (Combatant.AfterDamage).
func (c *Character) Damage(n int) {
	if n <= 0 || c.Dead {
		return
	}
	absorbed := min(c.TempHP, n)
	c.TempHP -= absorbed
	c.HP = max(0, c.HP-(n-absorbed))
}

// Heal возвращает здоровье; погибшего вылечить нельзя, потерявшему сознание лечение сбрасывает спасброски.
func (c *Character) Heal(n, maxHP int) {
	if n <= 0 || c.Dead {
		return
	}
	if c.HP <= 0 {
		c.DeathOk, c.DeathFail, c.Stable = 0, 0, false
	}
	c.HP = min(maxHP, c.HP+n)
}

// Normalize приводит значения к допустимым границам.
func (c *Character) Normalize(maxHP, maxMP int) {
	c.HP = max(0, min(c.HP, maxHP))
	c.MP = max(0, min(c.MP, maxMP))
	c.TempHP = max(0, c.TempHP)
}

func (c *Character) Has(cond string) bool { return slices.Contains(c.Conditions, cond) }
func (c *Character) Remove(cond string) {
	c.Conditions = slices.DeleteFunc(c.Conditions, func(x string) bool { return x == cond })
}

func (c *Character) Mark(key string) {
	if c.Used == nil {
		c.Used = map[string]bool{}
	}
	c.Used[key] = true
}

func (c *Character) Flip(key string) {
	if c.Active == nil {
		c.Active = map[string]bool{}
	}
	c.Active[key] = !c.Active[key]
}

// Count изменяет счётчик расходуемого ресурса (например, число ярости за отдых).
func (c *Character) Count(key string, delta int) int {
	if c.Counters == nil {
		c.Counters = map[string]int{}
	}
	c.Counters[key] = max(0, c.Counters[key]+delta)
	return c.Counters[key]
}

// Coins – порядок и вес монет (50 монет весят фунт).
var CoinOrder = []string{"pp", "gp", "ep", "sp", "cp"}

// CoinValueCP – стоимость монеты в медных.
var CoinValueCP = map[string]int{"pp": 1000, "gp": 100, "ep": 50, "sp": 10, "cp": 1}

func (c *Character) PurseWeight() float64 {
	n := 0
	for _, v := range c.Purse {
		n += max(0, v)
	}
	return float64(n) / 50
}

// AddCoins изменяет кошелёк; ниже нуля монет не бывает.
func (c *Character) AddCoins(kind string, n int) {
	if c.Purse == nil {
		c.Purse = map[string]int{}
	}
	c.Purse[kind] = max(0, c.Purse[kind]+n)
}
