package editor

import (
	"fmt"
	"image"

	g "github.com/AllenDang/giu"
	sdata "github.com/chimera-rpg/go-server/data"
	"github.com/fogleman/gg"
	log "github.com/sirupsen/logrus"
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
	childPos := image.Point{0, 0}
	for k, v := range m.maps {
		t, ok := m.mapTextures[k]
		if !ok || t.texture == nil {
			go func() {
				m.createMapTexture(k, v)
				g.Update()
			}()
		} else {
			tabs = append(tabs, g.TabItem(v.Name, g.Layout{
				g.Child(v.Name, false, 0, 0, g.WindowFlagsHorizontalScrollbar, g.Layout{
					g.Custom(func() {
						childPos = g.GetCursorScreenPos()
					}),
					g.Image(t.texture, t.width, t.height),
					g.Custom(func() {
						if g.IsItemHovered() && g.IsMouseClicked(g.MouseButtonLeft) {
							mousePos := g.GetMousePos()
							mousePos.X -= childPos.X
							mousePos.Y -= childPos.Y
							// TODO: Send mousePos as a click for the selected map.
							log.Println(mousePos)
						}
					}),
				}),
			}))
		}
	}
	var b bool

	g.WindowV(fmt.Sprintf("Maps: %s", m.filename), &b, g.WindowFlagsMenuBar, 210, 30, 300, 400, g.Layout{
		g.MenuBar(g.Layout{
			g.Menu("File", g.Layout{
				g.MenuItem("Save", func() { m.saveAll() }),
				g.Separator(),
				g.MenuItem("Close", func() { m.close() }),
			}),
		}),
		g.TabBarV("Tabs", g.TabBarFlagsFittingPolicyScroll|g.TabBarFlagsFittingPolicyResizeDown, tabs),
	})
}

func (m *Maps) createMapTexture(name string, sm *sdata.Map) {
	mT := MapTexture{}
	scale := 4.0
	//mHeight := sm.Height
	mHeight := 4
	mWidth := sm.Width
	mDepth := sm.Depth
	tWidth := 8
	tHeight := 6
	cWidth := mWidth * tWidth
	cHeight := mDepth * tHeight

	focusedY := 3
	focusedX := 2
	focusedZ := 2

	mT.width = float32(cWidth+(mHeight*1)) * float32(scale)
	mT.height = float32(cHeight+(mHeight*4)) * float32(scale)

	dc := gg.NewContext(int(mT.width), int(mT.height))
	dc.SetRGB(0.1, 0.1, 0.1)
	dc.Clear()
	dc.SetLineWidth(1)
	startY := mHeight * 4
	for y := 0; y < mHeight; y++ {
		xOffset := y * 1
		yOffset := y * 4
		for x := 0; x < mWidth; x++ {
			for z := 0; z < mDepth; z++ {
				oX := float64(x*tWidth+xOffset) * scale
				oY := float64(z*tHeight-yOffset+startY) * scale
				oW := float64(tWidth) * scale
				oH := float64(tHeight) * scale
				dc.DrawRectangle(oX, oY, oW, oH)
				if y == focusedY {
					if x == focusedX && z == focusedZ {
						dc.SetRGBA(0.2, 0.3, 0.6, 0.5)
						dc.FillPreserve()
					}
					dc.SetRGB(0.9, 0.9, 0.9)
				} else {
					dc.SetRGBA(0.9, 0.9, 0.9, 0.1)
				}
				dc.Stroke()
			}
		}
	}

	var err error
	mT.texture, err = g.NewTextureFromRgba(dc.Image().(*image.RGBA))
	if err != nil {
		log.Fatalln(err)
	}
	m.mapTextures[name] = mT
}

func (m *Maps) saveAll() {
	log.Println("TODO: Save all maps in file")
}

func (m *Maps) close() {
	log.Println("TODO: Issue close of map")
}
