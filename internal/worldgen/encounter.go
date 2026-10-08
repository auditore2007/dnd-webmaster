package worldgen

import (
	"heroesbook/internal/model"
	"heroesbook/internal/travel"
)

// Случайные встречи в пути: кто попадается в местности и как это выглядит.

// encounterKind – из какого «дикого места» брать существ для местности.
var encounterKind = map[uint8]string{
	travel.Road: "camp", travel.Plain: "camp", travel.Forest: "forest", travel.Hills: "cave",
	travel.Mountain: "mountain", travel.Swamp: "swamp", travel.Desert: "ruins", travel.Snow: "mountain",
}

// EncounterChance – вероятность встречи за день пути по местности.
var EncounterChance = map[uint8]float64{
	travel.Road: 0.08, travel.Plain: 0.12, travel.Forest: 0.2, travel.Hills: 0.2,
	travel.Mountain: 0.25, travel.Swamp: 0.25, travel.Desert: 0.2, travel.Snow: 0.2,
}

var encounterTexts = map[uint8][]string{
	travel.Road:     {"На дороге – перевёрнутая телега и засада", "Навстречу едут вооружённые всадники", "У моста требуют плату за проезд"},
	travel.Plain:    {"Над травой поднимается пыль – кто-то идёт наперерез", "У заброшенного колодца разбит чужой лагерь", "Ночью у костра слышны шаги"},
	travel.Forest:   {"Из чащи выходят на тропу", "Под деревьями – следы и обглоданные кости", "На привале в лесу стало слишком тихо"},
	travel.Hills:    {"Из пещеры в склоне холма тянет зверем", "Камнепад, а за ним – засада", "На вершине холма кто-то разжёг сигнальный костёр"},
	travel.Mountain: {"На узком карнизе путь преграждают", "Сверху на тропу падает тень", "В ущелье эхом отдаются тяжёлые шаги"},
	travel.Swamp:    {"Из тумана над топью поднимаются фигуры", "Кочка под ногой оказывается чьей-то спиной", "Блуждающие огни уводят с тропы"},
	travel.Desert:   {"Из песка показывается край древней стены – и не только", "Мираж оказывается лагерем", "Песчаная буря, а в ней – чужие силуэты"},
	travel.Snow:     {"В метели проступают чужие следы", "Под снегом трещит лёд", "На перевале путь преграждают"},
}

// Encounter – случайная встреча в местности: короткое описание и существа под уровень группы.
func (g Gen) Encounter(terrain uint8) (string, []model.Foe) {
	kind, ok := encounterKind[terrain]
	if !ok {
		kind, terrain = "camp", travel.Plain
	}
	return pick(g.R, encounterTexts[terrain]), g.foes(kind)
}
