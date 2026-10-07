package rules

import (
	"testing"

	"heroesbook/internal/dice"
	"heroesbook/internal/model"
)

func TestShieldGivesACUntilNextTurn(t *testing.T) {
	d := DnD5e{}
	w := caster(t, "wizard", 5, "int", 17)
	base := d.Derive(w).AC
	if _, err := d.Spell(&seq{v: []int{1}}, w, model.Spell{Name: "Щит", Level: 1, Ref: "shield"}, 1, nil, dice.Normal); err != nil {
		t.Fatal(err)
	}
	if got := d.Derive(w).AC; got != base+5 {
		t.Errorf("КД со щитом %d, ожидалось %d", got, base+5)
	}
	d.TurnStart(&seq{v: []int{1}}, w) // начало следующего хода – щит исчезает
	if got := d.Derive(w).AC; got != base || len(w.Effects) != 0 {
		t.Errorf("щит должен исчезнуть: КД %d, эффектов %d", got, len(w.Effects))
	}
}

func TestHoldPersonParalyzesAndRepeatSaveFrees(t *testing.T) {
	d := DnD5e{}
	w := caster(t, "wizard", 5, "int", 17) // СЛ 14
	foe := dummy(10, 50)
	hold := model.Spell{Name: "Удержание личности", Level: 2, Ref: "holdperson"}
	// спасбросок мишени: 1 + 0 = 1 < 14 – провал
	if _, err := d.Spell(&seq{v: []int{0}}, w, hold, 2, []*model.Character{foe}, dice.Normal); err != nil {
		t.Fatal(err)
	}
	if !foe.Has("Парализован") || w.Conc != "Удержание личности" || len(foe.Effects) != 1 {
		t.Fatalf("паралич не наложен: %v %q %d", foe.Conditions, w.Conc, len(foe.Effects))
	}
	// конец хода: бросок 19 → 20 ≥ 14, цель освобождается
	lines := d.EndTurn(&seq{v: []int{19}}, foe)
	if foe.Has("Парализован") || len(foe.Effects) != 0 || len(lines) == 0 {
		t.Errorf("повторный спасбросок должен освободить: %v %v", foe.Conditions, lines)
	}
	// удачный спасбросок сразу – эффекта нет
	w2 := caster(t, "wizard", 5, "int", 17)
	foe2 := dummy(10, 50)
	d.Spell(&seq{v: []int{19}}, w2, model.Spell{Name: "Удержание личности", Level: 2, Ref: "holdperson"}, 2, []*model.Character{foe2}, dice.Normal)
	if foe2.Has("Парализован") || len(foe2.Effects) != 0 {
		t.Errorf("цель устояла, паралича быть не должно: %v", foe2.Conditions)
	}
}

func TestEffectExpiresAfterRounds(t *testing.T) {
	d := DnD5e{}
	c := dummy(10, 20)
	AddEffect(c, model.Effect{Name: "Страх", Rounds: 2, Cond: "Напуган", Src: "x"})
	d.TurnStart(&seq{v: []int{1}}, c)
	if !c.Has("Напуган") {
		t.Fatal("после первого хода эффект ещё действует")
	}
	d.TurnStart(&seq{v: []int{1}}, c)
	if c.Has("Напуган") || len(c.Effects) != 0 {
		t.Errorf("после второго хода эффект должен кончиться: %v", c.Conditions)
	}
}

func TestConditionStaysWhileAnotherEffectHoldsIt(t *testing.T) {
	c := dummy(10, 20)
	AddEffect(c, model.Effect{Name: "А", Cond: "Опутан", Src: "1"})
	AddEffect(c, model.Effect{Name: "Б", Cond: "Опутан", Src: "2"})
	DropEffect(c, 0)
	if !c.Has("Опутан") {
		t.Error("состояние держит второй эффект")
	}
	DropEffect(c, 0)
	if c.Has("Опутан") {
		t.Error("состояние должно сняться")
	}
}

func TestConcentrationCheck(t *testing.T) {
	d := DnD5e{}
	w := caster(t, "wizard", 5, "int", 17)
	w.Conc = "Ускорение"
	// урон 30 → СЛ 15; бросок 2 (+0 Телосложение 10) → провал
	w.ConcDmg = 30
	lines, lost := d.ConcCheck(&seq{v: []int{1}}, w)
	if !lost || len(lines) != 1 {
		t.Errorf("концентрация должна сорваться: %v %v", lines, lost)
	}
	// малый урон → СЛ 10, бросок 19 → держится
	w.ConcDmg = 4
	if _, lost := d.ConcCheck(&seq{v: []int{18}}, w); lost {
		t.Error("концентрация должна удержаться")
	}
	// без сознания – теряется сразу
	w.ConcDmg = -1
	w.HP = 0
	if _, lost := d.ConcCheck(&seq{v: []int{18}}, w); !lost {
		t.Error("без сознания концентрация теряется")
	}
}

func TestStrikeRecordsConcentrationDamage(t *testing.T) {
	d := DnD5e{}
	w := caster(t, "wizard", 5, "int", 17)
	w.Conc = "Благословение"
	w.HP = 100
	Strike(d, w, 8, false)
	Strike(d, w, 3, false)
	if w.ConcDmg != 8 {
		t.Errorf("должен запоминаться наибольший удар: %d", w.ConcDmg)
	}
	Strike(d, w, 999, false)
	if w.ConcDmg != -1 {
		t.Errorf("упавший теряет концентрацию: %d", w.ConcDmg)
	}
}

func TestBlessAddsAttackBonus(t *testing.T) {
	d := DnD5e{}
	cl := caster(t, "cleric", 3, "wis", 16)
	ally := build(t, d, "human", "", "fighter")
	d.Spell(&seq{v: []int{1}}, cl, model.Spell{Name: "Благословение", Level: 1, Ref: "bless"}, 1, []*model.Character{ally}, dice.Normal)
	if len(ally.Effects) != 1 || ally.Effects[0].Atk != "1d4" {
		t.Fatalf("благословение не наложено: %+v", ally.Effects)
	}
	foe := dummy(15, 100)
	w := model.Weapon{Name: "Меч", Dice: "1d8", Ability: "str", Type: "slashing"}
	// d20=10, бонус d4=3 (значение 2 → 1+2), без эффекта было бы 10+0+2
	with := d.Attack(&seq{v: []int{9, 2, 1}}, ally, foe, w, dice.Normal)
	ally.Effects = nil
	without := d.Attack(&seq{v: []int{9, 2, 1}}, ally, foe, w, dice.Normal)
	if with.Total <= without.Total {
		t.Errorf("с благословением итог должен быть выше: %d и %d", with.Total, without.Total)
	}
}

func TestSavesAutoFailWhenParalyzed(t *testing.T) {
	d := DnD5e{}
	c := dummy(10, 10)
	c.Conditions = []string{"Парализован"}
	if got := d.saveRoll(&seq{v: []int{19}}, c, "dex"); got != saveFatal {
		t.Errorf("парализованный проваливает Ловкость: %d", got)
	}
	if got := d.saveRoll(&seq{v: []int{19}}, c, "wis"); got < 10 {
		t.Errorf("Мудрость бросается как обычно: %d", got)
	}
}

func monster(spec model.Special) *model.Character {
	c := dummy(12, 40)
	c.Name = "Тварь"
	c.ID = "mon"
	c.Stat.Attack = 5
	c.Stat.Specials = []model.Special{spec}
	return c
}

func TestMonsterBreathHitsAllFoesAndRecharges(t *testing.T) {
	d := DnD5e{}
	dragon := monster(model.Special{Key: "breath", Name: "Дыхание", Mode: "area", Dmg: "4d6", Type: "fire", Save: "dex", DC: 15, Half: true, Recharge: true})
	a, b := dummy(10, 100), dummy(10, 100)
	a.ID, b.ID = "a", "b"
	lines, err := d.Special(&seq{v: []int{1}}, "breath", dragon, []*model.Character{a, b})
	if err != nil || len(lines) == 0 {
		t.Fatal(err)
	}
	if a.HP >= 100 || b.HP >= 100 {
		t.Errorf("дыхание должно ранить всех: %d %d", a.HP, b.HP)
	}
	if _, err := d.Special(&seq{v: []int{1}}, "breath", dragon, []*model.Character{a}); err == nil {
		t.Error("дыхание не перезарядилось")
	}
	// перезарядка на 5–6: d6 = 6
	d.tick(&seq{v: []int{5}}, dragon)
	if dragon.Used["breath"] {
		t.Error("на 6 дыхание должно перезарядиться")
	}
}

func TestMonsterParalyzingClawsAppliesConditionOnFailedSave(t *testing.T) {
	d := DnD5e{}
	ghoul := monster(model.Special{Key: "claws", Name: "Когти", Mode: "single", Attack: true, Dmg: "2d4", Type: "slashing", Save: "con", DC: 10, Cond: "Парализован", Rounds: 10, Repeat: true})
	hero := build(t, d, "human", "", "fighter")
	hero.HP = 50
	// атака 19 (+5) попадает, затем спасбросок Телосложения 1+1 против 10 – провал
	_, err := d.Special(&seq{v: []int{18, 1, 1, 0}}, "claws", ghoul, []*model.Character{hero})
	if err != nil {
		t.Fatal(err)
	}
	if !hero.Has("Парализован") || len(hero.Effects) != 1 || hero.Effects[0].Save != "con" || hero.Effects[0].DC != 10 {
		t.Errorf("паралич должен наложиться с повторным спасброском: %v %+v", hero.Conditions, hero.Effects)
	}
}

func TestMonsterSpecialMissAndPoisonHalf(t *testing.T) {
	d := DnD5e{}
	snake := monster(model.Special{Key: "bite", Name: "Укус", Mode: "single", Attack: true, Dmg: "1d4", Type: "piercing", Save: "con", DC: 10, XDmg: "2d4", XType: "poison", Half: true})
	tgt := dummy(30, 40)
	lines, _ := d.Special(&seq{v: []int{0}}, "bite", snake, []*model.Character{tgt}) // 1+5=6 < 30
	if tgt.HP != 40 || len(lines) < 2 {
		t.Errorf("промах не должен ранить: %d %v", tgt.HP, lines)
	}
	if _, err := d.Special(&seq{v: []int{1}}, "nope", snake, []*model.Character{tgt}); err == nil {
		t.Error("неизвестная способность")
	}
}
