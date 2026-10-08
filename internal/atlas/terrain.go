package atlas

import (
	"image/color"
	"math"
)

// Суша и вода «как на рисованной карте»: цвет земли смешивается из биома и климата (температура × влажность),
// поверх – крупные пятна и мелкая фактура, полог леса, камень и снег по высоте и крутизне, пляжи у воды;
// рельеф освещён с северо-запада по нормалям (тёплый свет, холодные тени). Вода темнеет с глубиной,
// у берега – бирюзовое мелководье, пена и пара береговых линий.

var (
	sandColor   = rgb(224, 206, 158)
	rockLight   = rgb(166, 150, 128)
	rockDark    = rgb(110, 98, 86)
	coastWater  = rgb(150, 198, 188)
	shallowSea  = rgb(92, 160, 168)
	midSea      = rgb(46, 100, 138)
	deepSea     = rgb(24, 58, 96)
	foamColor   = rgb(226, 240, 234)
	lightTint   = rgb(255, 240, 204)
	shadowTint  = rgb(58, 66, 104)
	forestTones = map[uint8]color.RGBA{
		bTropicalSeasonalForest: rgb(92, 122, 52), bTemperateDeciduousForest: rgb(62, 102, 50),
		bTropicalRainforest: rgb(38, 90, 44), bTemperateRainforest: rgb(46, 92, 56), bTaiga: rgb(48, 78, 58),
		bWetland: rgb(70, 96, 66),
	}
)

// climatePalette – цвет земли по влажности (сухо, средне, влажно) для холода и жары.
var climatePalette = [2][3]color.RGBA{
	{rgb(190, 186, 164), rgb(132, 148, 104), rgb(80, 104, 80)},
	{rgb(214, 196, 148), rgb(146, 162, 86), rgb(64, 112, 58)},
}

func climateColor(temp, moist float64) color.RGBA {
	t := clamp((temp+8)/36, 0, 1)
	m := clamp((moist+2)/28, 0, 1) * 2
	k := min(1, int(m))
	f := m - float64(k)
	cold := mix(climatePalette[0][k], climatePalette[0][k+1], f)
	hot := mix(climatePalette[1][k], climatePalette[1][k+1], f)
	return mix(cold, hot, t)
}

// base рисует суше и воду попиксельно, параллельно по полосам строк.
func (p *painter) base() {
	img := p.c.img
	parallelRows(p.H, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := range p.W {
				i := y*p.W + x
				var col color.RGBA
				switch p.kind[i] {
				case pxLand:
					col = p.landColor(x, y, i)
				case pxOcean:
					col = p.seaColor(x, y, i)
				default:
					col = p.lakeColor(x, y, i)
				}
				o := img.PixOffset(x, y)
				img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = col.R, col.G, col.B, 255
			}
		}
	})
}

// cellMix – взвешенное по билинейной схеме среднее цвета соседних клеток суши.
func (p *painter) cellMix(u, v float64, colorOf func(c int) (color.RGBA, bool)) (color.RGBA, float64) {
	g := p.w.g
	i0, j0 := int(math.Floor(u)), int(math.Floor(v))
	tx, ty := u-float64(i0), v-float64(j0)
	var r, gg, b, sum float64
	for _, k := range [4][3]float64{{0, 0, (1 - tx) * (1 - ty)}, {1, 0, tx * (1 - ty)}, {0, 1, (1 - tx) * ty}, {1, 1, tx * ty}} {
		ci := min(g.nx-1, max(0, i0+int(k[0])))
		cj := min(g.ny-1, max(0, j0+int(k[1])))
		col, ok := colorOf(cj*g.nx + ci)
		if !ok || k[2] <= 0 {
			continue
		}
		r, gg, b, sum = r+float64(col.R)*k[2], gg+float64(col.G)*k[2], b+float64(col.B)*k[2], sum+k[2]
	}
	if sum == 0 {
		return color.RGBA{}, 0
	}
	return color.RGBA{uint8(r / sum), uint8(gg / sum), uint8(b / sum), 255}, sum
}

func (p *painter) landColor(x, y, i int) color.RGBA {
	w := p.w
	fx, fy := float64(x), float64(y)
	u, v := p.lattice(fx, fy)
	col, ok := p.cellMix(u, v, func(c int) (color.RGBA, bool) {
		if !w.land(c) {
			return color.RGBA{}, false
		}
		moist := float64(w.Prec[c])
		if w.River[c] != 0 {
			moist += 6
		}
		return mix(biomePaint[w.Biome[c]], climateColor(float64(w.Temp[c]), moist), 0.55), true
	})
	if ok == 0 {
		col = biomePaint[bGrassland]
	}
	// лес: тёмный полог с мелкой «кронистой» фактурой
	forest, fw := p.cellMix(u, v, func(c int) (color.RGBA, bool) {
		t, ok := forestTones[w.Biome[c]]
		return t, ok && w.land(c)
	})
	if fw > 0 {
		canopy := p.detail.fbm(fx/(2.4*p.scale), fy/(2.4*p.scale), 2)
		col = mix(col, shade(forest, 0.85+0.35*canopy), fw*0.55)
	}
	// крупные пятна (поля, луга, пустоши) и мелкая фактура травы
	patch := p.warpX.fbm(fx/(70*p.scale)+13, fy/(70*p.scale)+29, 3)
	col = mix(col, rgb(178, 172, 110), clamp(patch-0.58, 0, 0.2))
	col = mix(col, rgb(84, 118, 64), clamp(0.45-patch, 0, 0.25))
	col = shade(col, 0.94+0.12*p.warpY.fbm(fx/(5*p.scale), fy/(5*p.scale), 2))

	h := float64(p.hp[i])
	gx := float64(p.hp[p.at(x+1, y)]-p.hp[p.at(x-1, y)]) / 2
	gy := float64(p.hp[p.at(x, y+1)]-p.hp[p.at(x, y-1)]) / 2
	slope := math.Hypot(gx, gy)
	// камень на крутых и высоких склонах, со слоями породы
	rock := clamp(smoothstep(46, 66, h)+smoothstep(0.9, 2.6, slope)*smoothstep(36, 50, h), 0, 1)
	if rock > 0 {
		strata := p.detail.fbm(fx/(5*p.scale), fy/(16*p.scale)+40, 3)
		col = mix(col, mix(rockDark, rockLight, strata), rock*0.9)
	}
	// снег: в холодных краях ниже, на крутизне не держится
	line := 84.0
	if t := p.bilinear(u, v, func(c int) float64 { return float64(w.Temp[c]) }); t < 4 {
		line = 64 + t*3
	}
	patchy := (p.warpX.fbm(fx/(9*p.scale)+70, fy/(9*p.scale), 3) - 0.5) * 16 // снег лежит пятнами, по ложбинам ниже
	if snow := smoothstep(line-2, line+12, h+patchy) * (1 - smoothstep(2, 4.5, slope)*0.7); snow > 0 {
		col = mix(col, rgb(226, 232, 240), snow)
	}
	// пляж у моря
	if d := float64(p.dWater[i]); d < 3*p.scale && h < 30 && w.Temp[p.cell[i]] > -2 {
		col = mix(col, sandColor, 0.65*(1-d/(3*p.scale)))
	}
	relief := 0.25 + 0.75*smoothstep(32, 62, h) // равнины почти плоские, горы – в полную силу
	return p.hillshade(col, gx*relief, gy*relief, h)
}

// hillshade – освещение по нормали поверхности: свет с северо-запада, на солнечных склонах тёплый, в тени холодный.
func (p *painter) hillshade(col color.RGBA, gx, gy, h float64) color.RGBA {
	z := 1.1 / p.scale
	nx, ny, nz := -gx*z, -gy*z, 1.0
	l := math.Sqrt(nx*nx + ny*ny + nz*nz)
	const lx, ly, lz = -0.55, -0.55, 0.63
	lambert := (nx*lx + ny*ly + nz*lz) / l
	k := clamp(lambert/lz, 0.35, 1.45)
	out := shade(col, 0.2+0.8*k)
	if k < 1 {
		out = mix(out, shadowTint, (1-k)*0.25)
	} else {
		out = mix(out, lightTint, (k-1)*0.35)
	}
	return out
}

func (p *painter) seaColor(x, y, i int) color.RGBA {
	fx, fy := float64(x), float64(y)
	h := float64(p.hp[i])
	d := float64(p.dLand[i])
	t := clamp(0.5*(seaLevel-h)/18+0.5*d/(130*p.scale), 0, 1)
	var col color.RGBA
	switch {
	case t < 0.12:
		col = mix(coastWater, shallowSea, t/0.12)
	case t < 0.45:
		col = mix(shallowSea, midSea, (t-0.12)/0.33)
	default:
		col = mix(midSea, deepSea, (t-0.45)/0.55)
	}
	col = shade(col, 0.94+0.12*p.warpY.fbm(fx/(90*p.scale)+7, fy/(90*p.scale), 3))
	col = shade(col, 0.985+0.03*p.detail.fbm(fx/(3*p.scale), fy/(1.4*p.scale), 2))
	// береговые линии и пена у берега
	for k, ring := range []float64{6, 13} {
		on := clamp(1-math.Abs(d-ring*p.scale)/(0.7*p.scale), 0, 1)
		col = mix(col, rgb(36, 74, 104), on*(0.22-float64(k)*0.08))
	}
	col = mix(col, foamColor, 0.4*clamp(1-d/(4*p.scale), 0, 1))
	return p.shore(col, h, d)
}

func (p *painter) lakeColor(x, y, i int) color.RGBA {
	d := float64(p.dLand[i])
	col := mix(rgb(120, 176, 182), rgb(56, 112, 140), clamp(d/(16*p.scale), 0, 1))
	col = shade(col, 0.96+0.08*p.warpY.fbm(float64(x)/40, float64(y)/40, 2))
	col = mix(col, foamColor, 0.3*clamp(1-d/(3*p.scale), 0, 1))
	return p.shore(col, float64(p.hp[i]), d)
}

// shore – у самой кромки вода переходит в песок (сглаживание берега), по кромке – тонкая тёмная линия.
func (p *painter) shore(col color.RGBA, h, d float64) color.RGBA {
	if d > 2 {
		return col
	}
	col = mix(col, sandColor, clamp(1-(seaLevel-h)/1.5, 0, 1)*0.7)
	return mix(col, rgb(70, 62, 46), clamp(1.6-d, 0, 1)*0.5)
}
