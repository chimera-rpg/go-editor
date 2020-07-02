package editor

import (
	"fmt"
	g "github.com/AllenDang/giu"
	sdata "github.com/chimera-rpg/go-server/data"
	"github.com/fogleman/gg"
	log "github.com/sirupsen/logrus"
	"image"
)

type Maps struct {
	filename    string
	maps        map[string]*sdata.Map
	mapTextures map[string]MapTexture
}

type MapTexture struct {
	texture *g.Texture
	width   float32
	height  float32
}

func NewMaps(name string, maps map[string]*sdata.Map) *Maps {
	return &Maps{
		filename:    name,
		maps:        maps,
		mapTextures: make(map[string]MapTexture),
	}
}

func (m *Maps) draw() {
	var tabs []g.Widget
	for k, v := range m.maps {
		t, ok := m.mapTextures[k]
		if !ok || t.texture == nil {
			go func() {
				m.createMapTexture(k, v)
				g.Update()
			}()
		} else {
			tabs = append(tabs, g.TabItem(v.Name, g.Layout{
				g.Image(t.texture, t.width, t.height),
			}))
		}
	}
	var b bool

	g.WindowV(fmt.Sprintf("MapSet: %s", m.filename), &b, g.WindowFlagsMenuBar, 210, 30, 300, 400, g.Layout{
		g.MenuBar(g.Layout{
			g.Menu("File", g.Layout{
				g.MenuItem("Save", nil),
				g.Separator(),
			}),
		}),
		g.TabBarV("Tabs", g.TabBarFlagsFittingPolicyScroll|g.TabBarFlagsFittingPolicyResizeDown, tabs),
	})
}

func (m *Maps) createMapTexture(name string, sm *sdata.Map) {
	mT := MapTexture{}
	mHeight := sm.Height
	mWidth := sm.Width
	mDepth := sm.Depth
	tWidth := 32
	tHeight := 24
	cWidth := mWidth * tWidth
	cHeight := mDepth * tHeight

	mT.width = float32(cWidth)
	mT.height = float32(cHeight)

	dc := gg.NewContext(cWidth, cHeight)
	dc.SetRGB(1, 1, 1)
	for y := 0; y < mHeight; y++ {
		for x := 0; x < mWidth; x++ {
			for z := 0; z < mDepth; z++ {
				dc.DrawRectangle(float64(x*tWidth), float64(z*tHeight), float64(tWidth), float64(tHeight))
			}
		}
	}
	dc.Stroke()

	var err error
	mT.texture, err = g.NewTextureFromRgba(dc.Image().(*image.RGBA))
	if err != nil {
		log.Fatalln(err)
	}
	m.mapTextures[name] = mT
}
