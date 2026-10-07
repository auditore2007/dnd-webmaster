// Package srd – снаряжение D&D 5e: оружие, доспехи, снаряжение, инструменты, транспорт, магические предметы.
package srd

import (
	"fmt"
	"strconv"
	"strings"

	"heroesbook/internal/model"
)

var rarity = map[string]string{"c": "обычный", "u": "необычный", "r": "редкий", "v": "очень редкий", "l": "легендарный", "a": "артефакт"}

// parse читает таблицу. Категория берётся из заголовка "#cat"; предмет с ar= – доспех, с ac= – щит.
func parse(table string, out *[]model.Item) {
	cat := "gear"
	for _, ln := range strings.Split(table, "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		if strings.HasPrefix(ln, "#") {
			cat = ln[1:]
			continue
		}
		p := strings.Split(ln, "|")
		if len(p) < 4 {
			continue
		}
		it := model.Item{Name: p[0], Qty: 1, Cat: cat, Desc: p[3]}
		if p[1] != "–" {
			it.Price = p[1]
		}
		it.Weight, _ = strconv.ParseFloat(p[2], 64)
		if len(p) > 4 {
			var w model.Weapon
			isW := false
			for _, tok := range strings.Fields(p[4]) {
				k, v, _ := strings.Cut(tok, "=")
				switch k {
				case "d":
					w.Dice, isW = v, true
				case "t":
					w.Type = v
				case "f":
					w.Finesse = true
				case "r":
					w.Ranged = true
				case "b":
					w.Bonus, _ = strconv.Atoi(v)
				case "ac":
					it.ACBonus, _ = strconv.Atoi(v)
					if it.Cat == "armor" || it.Cat == "weapon" {
						it.Cat = "shield"
					}
				case "ar":
					b, kind, _ := strings.Cut(v, ":")
					it.ArmorBase, _ = strconv.Atoi(b)
					it.ArmorType = kind
				case "rar":
					it.Rarity = rarity[v]
				}
			}
			if isW {
				w.Name = it.Name
				w.Ability = "str"
				if w.Ranged {
					w.Ability = "dex"
				}
				it.Weapon = &w
			}
		}
		*out = append(*out, it)
	}
}

// Items возвращает весь каталог. Порядок: оружие, доспехи, снаряжение, магия.
func Items() []model.Item {
	var out []model.Item
	parse(mundaneTable, &out)
	parse(magicTable, &out)
	out = append(out, enchanted(out)...)
	return out
}

// enchanted порождает «+1/+2/+3» версии обычного оружия, доспехов и щитов.
func enchanted(base []model.Item) []model.Item {
	rar := []string{"необычный", "редкий", "очень редкий"}
	price := []string{"1000 зм", "4000 зм", "16000 зм"}
	weapons := map[string]bool{"Кинжал": true, "Короткий меч": true, "Скимитар": true, "Рапира": true, "Длинный меч": true, "Двуручный меч": true,
		"Боевой топор": true, "Большой топор": true, "Боевой молот": true, "Булава": true, "Копьё": true, "Короткий лук": true, "Длинный лук": true, "Лёгкий арбалет": true, "Тяжёлый арбалет": true}
	armors := map[string]bool{"Кожаный доспех": true, "Проклёпанная кожа": true, "Кольчужная рубаха": true, "Чешуйчатый доспех": true, "Полулаты": true, "Кольчуга": true, "Наборный доспех": true, "Латы": true}
	var out []model.Item
	for _, b := range base {
		for n := 1; n <= 3; n++ {
			it := b
			it.Name = fmt.Sprintf("%s +%d", b.Name, n)
			it.Rarity, it.Price = rar[n-1], price[n-1]
			it.Desc = fmt.Sprintf("Магический предмет: +%d. %s", n, b.Desc)
			switch {
			case b.Cat == "weapon" && weapons[b.Name] && b.Weapon != nil:
				w := *b.Weapon
				w.Name, w.Bonus = it.Name, n
				it.Weapon = &w
			case b.Cat == "armor" && armors[b.Name]:
				it.ArmorBase += n
			case b.Name == "Щит":
				it.ACBonus += n
				it.Cat = "shield"
			default:
				continue
			}
			out = append(out, it)
		}
	}
	return out
}
