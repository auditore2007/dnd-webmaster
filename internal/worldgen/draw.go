package worldgen

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"math/rand/v2"

	"heroesbook/internal/dice"
)

// Кодирование готовых картинок карт в data:URL и воспроизводимый генератор случайностей.

type canvas struct{ img *image.RGBA }

// dataURL кодирует картинку в JPEG (карты большие – PNG вышел бы в разы тяжелее).
func (c canvas) dataURL(quality int) (string, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, c.img, &jpeg.Options{Quality: quality}); err != nil {
		return "", err
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// seeded – воспроизводимый генератор: один seed – одна и та же карта.
type seeded struct{ r *rand.Rand }

func (s seeded) Intn(n int) int { return s.r.IntN(n) }

// Seeded возвращает генератор случайностей с заданным зерном.
func Seeded(seed int64) dice.Roller {
	return seeded{rand.New(rand.NewPCG(uint64(seed), uint64(seed)^0x9e3779b97f4a7c15))}
}
