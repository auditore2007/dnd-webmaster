package atlas

import "math"

// Климат по Azgaar (MIT, см. LICENSE-AZGAAR): температура от широты и высоты, осадки приносят ветра
// широтных поясов, горы выше 85 их задерживают; биом – по влажности и температуре.

const (
	tempEquator   = 27.0
	tempNorthPole = -30.0
	heightExp     = 1.8 // высота 100 ≈ 2,8 км: холод в горах
	maxPassable   = 85  // через горы выше ветер влагу не переносит
)

// Биомы Azgaar.
const (
	bMarine = iota
	bHotDesert
	bColdDesert
	bSavanna
	bGrassland
	bTropicalSeasonalForest
	bTemperateDeciduousForest
	bTropicalRainforest
	bTemperateRainforest
	bTaiga
	bTundra
	bGlacier
	bWetland
)

var habitability = [13]float64{0, 4, 10, 22, 30, 50, 100, 80, 90, 12, 4, 0, 12}

// biomeCost – трудность продвижения государства через биом.
var biomeCost = [13]float64{10, 200, 150, 60, 50, 70, 70, 80, 90, 200, 1000, 5000, 150}

// biomesMatrix: строки – влажность 0–4 (сухо → влажно), столбцы – 20 − температура (жарко → холодно).
var biomesMatrix = [5][26]uint8{
	{1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 10},
	{3, 3, 3, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 9, 9, 9, 9, 10, 10, 10},
	{5, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 9, 9, 9, 9, 9, 10, 10, 10},
	{5, 6, 6, 6, 6, 6, 6, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9, 9, 9, 9, 9, 9, 10, 10, 10},
	{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9, 9, 9, 9, 9, 9, 9, 10, 10},
}

// latitudeModifier – осадки по поясам в 5° широты (у экватора дожди, в 20–30° пустыни, в 50–60° снова сыро).
var latitudeModifier = []float64{4, 2, 2, 2, 1, 1, 2, 2, 2, 2, 3, 3, 2, 2, 1, 1, 1, 0.5}

// winds – направление ветра по поясам в 30° с севера на юг (градусы, как у Azgaar по умолчанию).
var winds = []float64{225, 45, 225, 315, 135, 315}

func (w *World) latitude(y float64) float64 { return w.latN - y/float64(w.g.h)*w.latT }

func temperatures(w *World) []int8 {
	g := w.g
	temp := make([]int8, g.size())
	tropicN := tempEquator - 16*0.15
	gradN := (tropicN - tempNorthPole) / (90 - 16)
	tropicS := tempEquator - 20*0.15
	gradS := (tropicS - (-15.0)) / (90 - 20)
	for row := range g.ny {
		lat := w.latitude(g.py[row*g.nx])
		var sea float64
		switch {
		case lat <= 16 && lat >= -20:
			sea = tempEquator - math.Abs(lat)*0.15
		case lat > 0:
			sea = tropicN - (lat-16)*gradN
		default:
			sea = tropicS + (lat+20)*gradS
		}
		for c := row * g.nx; c < (row+1)*g.nx; c++ {
			drop := 0.0
			if w.H[c] >= seaLevel {
				drop = math.Round(math.Pow(float64(w.H[c])-18, heightExp) / 1000 * 6.5)
			}
			temp[c] = int8(clamp(math.Round(sea-drop), -128, 127))
		}
	}
	return temp
}

func latMod(lat float64) float64 {
	i := int((math.Abs(lat) - 1) / 5)
	return latitudeModifier[min(len(latitudeModifier)-1, max(0, i))]
}

func precipitation(w *World, r rng) []uint8 {
	g, h, temp := w.g, w.H, w.Temp
	prec := make([]float64, g.size())
	modifier := math.Pow(float64(g.size())/10000, 0.25)
	height := func(i int) float64 {
		if i < 0 || i >= len(h) {
			return 0
		}
		return float64(h[i])
	}
	precipitationAt := func(humidity float64, i, next int) float64 {
		normal := math.Max(humidity/(10*modifier), 1)
		diff := math.Max(height(i+next)-height(i), 0)
		return clamp(normal+diff*sq(height(i+next)/70), 1, humidity)
	}
	// pass – ветер входит в клетку first и идёт шагами next, теряя влагу над сушей и набирая над морем.
	pass := func(first int, maxPrec float64, next, steps int) {
		humidity := maxPrec - float64(h[first])
		if humidity <= 0 {
			return
		}
		for s, cur := 0, first; s < steps && cur >= 0 && cur < len(h); s, cur = s+1, cur+next {
			if temp[cur] < -5 {
				continue
			}
			if h[cur] < seaLevel {
				switch {
				case height(cur+next) >= seaLevel && cur+next < len(prec):
					prec[cur+next] += math.Max(humidity/float64(r.intIn(10, 20)), 1) // дожди на побережье
				case height(cur+next) < seaLevel:
					humidity = math.Min(humidity+5*modifier, maxPrec)
					prec[cur] += 5 * modifier
				}
				continue
			}
			if height(cur+next) > maxPassable {
				prec[cur] += humidity // горы забирают всю влагу
				humidity = 0
				continue
			}
			p := precipitationAt(humidity, cur, next)
			prec[cur] += p
			evaporation := 0.0
			if p > 1.5 {
				evaporation = 1
			}
			humidity = clamp(humidity-p+evaporation, 0, maxPrec)
		}
	}
	var northerly, southerly float64
	for row := range g.ny {
		lat := w.latitude(g.py[row*g.nx])
		tier := min(len(winds)-1, int(math.Abs(lat-89)/30))
		angle := winds[tier]
		maxPrec := math.Min(120*modifier*latMod(lat), 255)
		if angle > 40 && angle < 140 && row > 0 {
			pass(row*g.nx, maxPrec, 1, g.nx)
		}
		if angle > 220 && angle < 320 {
			pass(row*g.nx+g.nx-1, maxPrec, -1, g.nx)
		}
		if angle > 100 && angle < 260 {
			northerly++
		}
		if angle > 280 || angle < 80 {
			southerly++
		}
	}
	if total := northerly + southerly; total > 0 {
		if northerly > 0 {
			maxPrec := northerly / total * 60 * modifier * latMod(w.latN)
			for x := range g.nx {
				pass(x, maxPrec, g.nx, g.ny)
			}
		}
		if southerly > 0 {
			maxPrec := southerly / total * 60 * modifier * latMod(w.latN-w.latT)
			for x := range g.nx {
				pass(g.size()-g.nx+x, maxPrec, -g.nx, g.ny)
			}
		}
	}
	out := make([]uint8, len(prec))
	for i, v := range prec {
		out[i] = uint8(clamp(v, 0, 255))
	}
	return out
}

// biomeOf – биом клетки по влажности, температуре и высоте (getId у Azgaar).
func biomeOf(moisture float64, temperature int8, height uint8, river bool) uint8 {
	t := float64(temperature)
	switch {
	case height < seaLevel:
		return bMarine
	case t < -5:
		return bGlacier
	case t >= 25 && !river && moisture < 8:
		return bHotDesert
	case t > -2 && (moisture > 40 && height < 25 || moisture > 24 && height > 24 && height < 60):
		return bWetland
	}
	band := min(int(moisture/5), 4)
	tb := int(clamp(20-t, 0, 25))
	return biomesMatrix[band][tb]
}

func biomes(w *World) []uint8 {
	out := make([]uint8, w.g.size())
	for i, h := range w.H {
		if h < seaLevel || w.Lake[i] {
			continue
		}
		own := float64(w.Prec[i])
		if w.River[i] != 0 {
			own += math.Max(w.Flux[i]/10, 2)
		}
		sum, n := own, 1.0
		for _, c := range w.g.nb[i] {
			if w.H[c] >= seaLevel {
				sum += float64(w.Prec[c])
				n++
			}
		}
		out[i] = biomeOf(math.Round(4+sum/n), w.Temp[i], h, w.River[i] != 0)
	}
	return out
}
