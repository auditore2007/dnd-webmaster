package bestiary

import "heroesbook/internal/model"

// Боссы: легендарные действия, легендарное сопротивление, логово и вторая фаза (по мотивам Monster Manual, упрощённо).

type boss struct {
	Legendary, LegRes int
	LegActs           []model.LegAct
	Lair              []model.Special
	Phase             *model.Phase
}

func weaponAct(key, name string, cost, weapon int) model.LegAct {
	return model.LegAct{Key: key, Name: name, Cost: cost, Weapon: weapon}
}

func specialAct(cost int, s model.Special) model.LegAct {
	return model.LegAct{Key: s.Key, Name: s.Name, Cost: cost, Weapon: -1, Special: &s}
}

// wing – взмах крыльев дракона: все враги рядом, спасбросок Ловкости, при провале урон и сбит с ног.
func wing(dmg string, dc int) model.LegAct {
	return specialAct(2, sp{Key: "wing", Name: "Взмах крыльев", Mode: "area", Dmg: dmg, Type: "bludgeoning", Save: "dex", DC: dc, Cond: "Сбит с ног", Rounds: 1})
}

func dragon(wingDmg string, wingDC int, lair sp) boss {
	return boss{Legendary: 3, LegRes: 3,
		LegActs: []model.LegAct{weaponAct("tail", "Удар хвостом", 1, 2), wing(wingDmg, wingDC)},
		Lair:    []model.Special{lair},
		Phase:   &model.Phase{Pct: 50, Name: "Ярость дракона", Atk: 2, Recharge: true}}
}

var (
	lairWhite = sp{Key: "lair-ice", Name: "Ледяные осколки", Mode: "area", Dmg: "3d6", Type: "cold", Save: "dex", DC: 15, Half: true}
	lairBlack = sp{Key: "lair-mire", Name: "Хватающая трясина", Mode: "area", Save: "str", DC: 15, Cond: "Опутан", Rounds: 1}
	lairGreen = sp{Key: "lair-vines", Name: "Ядовитые лианы", Mode: "area", Dmg: "2d6", Type: "poison", Save: "con", DC: 15, Half: true}
	lairRed   = sp{Key: "lair-magma", Name: "Извержение магмы", Mode: "area", Dmg: "6d6", Type: "fire", Save: "dex", DC: 15, Half: true}
)

var bosses = map[string]boss{
	"adult-white-dragon":   dragon("2d6+6", 19, lairWhite),
	"adult-black-dragon":   dragon("2d6+6", 19, lairBlack),
	"adult-green-dragon":   dragon("2d6+6", 19, lairGreen),
	"adult-red-dragon":     dragon("2d6+8", 22, lairRed),
	"ancient-white-dragon": dragon("2d6+8", 22, lairWhite),
	"ancient-black-dragon": dragon("2d6+8", 23, lairBlack),
	"ancient-red-dragon":   dragon("2d6+10", 25, lairRed),
	"beholder": {Legendary: 3,
		LegActs: []model.LegAct{
			specialAct(1, sp{Key: "ray-paralyze", Name: "Луч паралича", Mode: "single", Save: "con", DC: 16, Cond: "Парализован", Rounds: 10, Repeat: true}),
			specialAct(1, sp{Key: "ray-fear", Name: "Луч страха", Mode: "single", Save: "wis", DC: 16, Cond: "Напуган", Rounds: 10, Repeat: true}),
			specialAct(1, sp{Key: "ray-enervate", Name: "Луч истощения", Mode: "single", Dmg: "8d8", Type: "necrotic", Save: "con", DC: 16, Half: true}),
		},
		Lair:  []model.Special{{Key: "lair-slime", Name: "Скользкая слизь", Mode: "area", Save: "dex", DC: 15, Cond: "Сбит с ног", Rounds: 1}},
		Phase: &model.Phase{Pct: 50, Name: "Безумие глаз", Atk: 2, Recharge: true}},
	"vampire": {Legendary: 3, LegRes: 3,
		LegActs: []model.LegAct{weaponAct("claw", "Удар когтями", 1, 0), weaponAct("bite", "Укус", 2, 1)},
		Phase:   &model.Phase{Pct: 50, Name: "Кровавая жажда", Atk: 2, Temp: 25}},
	"mummy-lord": {Legendary: 3, LegRes: 0,
		LegActs: []model.LegAct{weaponAct("fist", "Гнилой кулак", 1, 0),
			specialAct(2, sp{Key: "dust", Name: "Ослепляющая пыль", Mode: "area", Save: "con", DC: 16, Cond: "Ослеплён", Rounds: 1}),
			specialAct(2, sp{Key: "word", Name: "Богохульное слово", Mode: "area", Save: "con", DC: 16, Cond: "Ошеломлён", Rounds: 1})},
		Lair:  []model.Special{{Key: "lair-curse", Name: "Проклятие гробницы", Mode: "area", Dmg: "3d6", Type: "necrotic", Save: "wis", DC: 16, Half: true}},
		Phase: &model.Phase{Pct: 50, Name: "Гнев фараона", Atk: 2, AC: 2}},
	"lich": {Legendary: 3, LegRes: 3,
		LegActs: []model.LegAct{weaponAct("touch", "Парализующее касание", 2, 0),
			specialAct(2, sp{Key: "gaze", Name: "Устрашающий взгляд", Mode: "single", Save: "wis", DC: 18, Cond: "Напуган", Rounds: 1}),
			specialAct(3, sp{Key: "disrupt", Name: "Нарушение жизни", Mode: "area", Dmg: "6d6", Type: "necrotic", Save: "con", DC: 18, Half: true})},
		Lair:  []model.Special{{Key: "lair-tether", Name: "Тёмная связь", Mode: "area", Dmg: "4d6", Type: "necrotic", Save: "con", DC: 18, Half: true}},
		Phase: &model.Phase{Pct: 50, Name: "Последний ритуал", Atk: 2, Temp: 40, Recharge: true}},
	"solar": {Legendary: 3,
		LegActs: []model.LegAct{
			specialAct(2, sp{Key: "burst", Name: "Испепеляющий взрыв", Mode: "area", Dmg: "4d6", Type: "radiant", Save: "dex", DC: 23, Half: true}),
			specialAct(3, sp{Key: "blind", Name: "Ослепляющий взгляд", Mode: "single", Save: "con", DC: 15, Cond: "Ослеплён", Rounds: 10, Repeat: true})},
		Phase: &model.Phase{Pct: 50, Name: "Небесный гнев", Atk: 2}},
	"aboleth": {Legendary: 3,
		LegActs: []model.LegAct{weaponAct("tail", "Взмах хвостом", 1, 0),
			specialAct(2, sp{Key: "drain", Name: "Психическая выжимка", Mode: "area", Dmg: "3d6", Type: "psychic", Save: "wis", DC: 14, Half: true})},
		Lair: []model.Special{{Key: "lair-water", Name: "Водяные струи", Mode: "area", Save: "str", DC: 14, Cond: "Сбит с ног", Rounds: 1}}},
	"tarrasque": {Legendary: 3, LegRes: 3,
		LegActs: []model.LegAct{weaponAct("claw", "Удар когтем", 1, 1), weaponAct("tail", "Удар хвостом", 1, 3), weaponAct("bite", "Укус", 2, 0)},
		Phase:   &model.Phase{Pct: 50, Name: "Неудержимый", Atk: 2, Extra: 1}},
}
