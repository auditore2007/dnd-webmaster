package atlas

import (
	"embed"
	"image"
	"image/png"
	"math"
	"path"
	"sort"
	"strings"
	"sync"

	xdraw "golang.org/x/image/draw"
)

// Рисованные значки: у каждой расы свой замок-столица (sprites/castles/<раса>.png), поселения, мельницы
// и дикие места – картинки построек (sprites/buildings). Картинка ставится низом по центру на точку места.

//go:embed sprites/castles/*.png sprites/buildings/*.png
var spriteFS embed.FS

var sprites struct {
	once sync.Once
	imgs map[string]image.Image // "castles/elf", "buildings/house"
}

func loadSprites() map[string]image.Image {
	sprites.once.Do(func() {
		sprites.imgs = map[string]image.Image{}
		for _, dir := range []string{"castles", "buildings"} {
			entries, err := spriteFS.ReadDir("sprites/" + dir)
			if err != nil {
				continue
			}
			for _, e := range entries {
				f, err := spriteFS.Open(path.Join("sprites", dir, e.Name()))
				if err != nil {
					continue
				}
				img, err := png.Decode(f)
				f.Close()
				if err == nil {
					sprites.imgs[dir+"/"+strings.TrimSuffix(e.Name(), ".png")] = img
				}
			}
		}
	})
	return sprites.imgs
}

// HasCastle – есть ли у расы своя картинка замка-столицы.
func HasCastle(race string) bool {
	_, ok := loadSprites()["castles/"+race]
	return ok
}

// Site – дикое место на карте (руины, пещера, логово…); рисуется картинкой без подписи.
type Site struct {
	X, Y float64
	Kind string
}

// siteSprites – картинки диких мест; вариант выбирается по месту, чтобы карта повторялась.
var siteSprites = map[string][]string{
	"ruins": {"ruins", "ruinedtower"}, "cave": {"cave"}, "mine": {"mine"}, "camp": {"camp", "warcamp", "tusktent"},
	"lair": {"dungeon"}, "dungeon": {"gate"}, "tower": {"magetower", "watchtower", "observatory"}, "shrine": {"stones"},
	"forest": {"treehouse"}, "temple": {"temple"}, "tavern": {"tavern"}, "shop": {"market"}, "smithy": {"smithy"},
}

// variant – детерминированный выбор из списка по координатам.
func variant(xs []string, x, y float64) string {
	h := uint32(int(x)*73856093) ^ uint32(int(y)*19349663)
	return xs[h%uint32(len(xs))]
}

// burgSprite – картинка поселения и её высота в пикселях.
func (p *painter) burgSprite(b Burg) (string, float64) {
	s := p.scale
	switch b.Kind {
	case "capital":
		if HasCastle(b.Race) {
			return "castles/" + b.Race, 46 * s
		}
		return "buildings/castle", 40 * s
	case "castle":
		return "buildings/castle", 30 * s
	case "city":
		return "buildings/" + variant([]string{"academy", "market", "temple"}, b.X, b.Y), 30 * s
	case "port":
		return "buildings/harbor", 27 * s
	case "town":
		return "buildings/tavern", 26 * s
	}
	cell := p.cell[p.at(int(b.X), int(b.Y))]
	switch {
	case p.w.Temp[cell] < 0:
		return "buildings/lodge", 22 * s
	case p.w.TerrainOf(int(cell)) == Grass:
		return "buildings/" + variant([]string{"farm", "house"}, b.X, b.Y), 21 * s
	}
	return "buildings/house", 21 * s
}

func (p *painter) siteSprite(st Site) (string, float64) {
	xs, ok := siteSprites[st.Kind]
	if !ok {
		xs = []string{"ruins"}
	}
	return "buildings/" + variant(xs, st.X, st.Y), 24 * p.scale
}

// spriteBox – прямоугольник картинки высотой h, стоящей низом по центру на (x, y).
func spriteBox(name string, x, y, h float64) image.Rectangle {
	img, ok := loadSprites()[name]
	if !ok {
		return image.Rect(int(x-h/2), int(y-h), int(x+h/2), int(y))
	}
	b := img.Bounds()
	w := h * float64(b.Dx()) / float64(b.Dy())
	return image.Rect(int(math.Round(x-w/2)), int(math.Round(y-h)), int(math.Round(x+w/2)), int(math.Round(y)))
}

func (p *painter) drawSprite(name string, x, y, h float64) {
	img, ok := loadSprites()[name]
	if !ok {
		return
	}
	r := spriteBox(name, x, y, h)
	// мягкая тень под постройкой
	p.c.fill([][]pt{ellipse(x+h*0.06, y-h*0.04, float64(r.Dx())*0.46, h*0.1)}, rgb(40, 30, 20), 0.28)
	xdraw.CatmullRom.Scale(p.c.img, r, img, img.Bounds(), xdraw.Over, nil)
}

// burgObstacles – места поселений и диких мест заняты: туда не ставятся значки рельефа и надписи.
func (p *painter) burgObstacles() {
	for _, b := range p.w.Burgs {
		name, h := p.burgSprite(b)
		p.obstacles = append(p.obstacles, spriteBox(name, b.X, b.Y, h).Inset(-2))
	}
	for _, st := range p.w.Sites {
		name, h := p.siteSprite(st)
		p.obstacles = append(p.obstacles, spriteBox(name, st.X, st.Y, h).Inset(-2))
	}
	p.mills()
	p.hamlets()
}

// mills – у части деревень мельница: на реке – водяная, в степи и полях – ветряная.
func (p *painter) mills() {
	p.extras = nil
	for _, b := range p.w.Burgs {
		if b.Kind != "village" && b.Kind != "town" || variant([]string{"", "", "mill"}, b.X, b.Y) == "" {
			continue
		}
		cell := int(p.cell[p.at(int(b.X), int(b.Y))])
		name := "buildings/windmill"
		if p.w.River[cell] != 0 {
			name = "buildings/watermill"
		}
		h := 20 * p.scale
		for _, dx := range []float64{1.6, -1.6} {
			x, y := b.X+dx*h, b.Y+h*0.3
			box := spriteBox(name, x, y, h)
			i := p.at(int(x), int(y))
			if p.kind[i] != pxLand || !p.free(box) {
				continue
			}
			p.obstacles = append(p.obstacles, box.Inset(-2))
			p.extras = append(p.extras, extra{name, x, y, h})
			break
		}
	}
}

type extra struct {
	name    string
	x, y, h float64
}

// burgs рисует поселения, мельницы и дикие места сверху вниз, чтобы нижние перекрывали верхние.
func (p *painter) burgs() {
	items := append([]extra(nil), p.extras...)
	for _, b := range p.w.Burgs {
		name, h := p.burgSprite(b)
		items = append(items, extra{name, b.X, b.Y, h})
	}
	for _, st := range p.w.Sites {
		name, h := p.siteSprite(st)
		items = append(items, extra{name, st.X, st.Y, h})
	}
	sort.SliceStable(items, func(a, b int) bool { return items[a].y < items[b].y })
	for _, it := range items {
		p.drawSprite(it.name, it.x, it.y, it.h)
	}
}

// burgBox – прямоугольник значка поселения (для подписи рядом).
func (p *painter) burgBox(b Burg) image.Rectangle {
	name, h := p.burgSprite(b)
	return spriteBox(name, b.X, b.Y, h)
}
