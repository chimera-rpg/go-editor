package editor

import (
	"errors"
	"fmt"
	"image"

	g "github.com/AllenDang/giu"
	"github.com/chimera-rpg/go-editor/data"
	sdata "github.com/chimera-rpg/go-server/data"
	"github.com/fogleman/gg"
	log "github.com/sirupsen/logrus"
)

type Maps struct {
	filename                     string
	maps                         map[string]*sdata.Map
	currentMap                   string
	mapTextures                  map[string]MapTexture
	focusedY, focusedX, focusedZ int
}

type MapTexture struct {
	texture *g.Texture
	width   int
	height  int
}

func NewMaps(name string, maps map[string]*sdata.Map) *Maps {
	m := &Maps{
		filename:    name,
		maps:        maps,
		mapTextures: make(map[string]MapTexture),
	}

	for k := range maps {
		m.currentMap = k
		break
	}

	return m
}

func (m *Maps) draw(d *data.Manager) {
	var tabs []g.Widget
	childPos := image.Point{0, 0}
	for k, v := range m.maps {
		t, ok := m.mapTextures[k]
		if !ok || t.texture == nil {
			go func() {
				m.createMapTexture(k, v, d)
				g.Update()
			}()
		} else {
			tabs = append(tabs, g.TabItem(v.Name, g.Layout{
				g.Child(v.Name, false, 0, 0, g.WindowFlagsHorizontalScrollbar, g.Layout{
					g.Custom(func() {
						childPos = g.GetCursorScreenPos()
					}),
					g.Image(t.texture, float32(t.width), float32(t.height)),
					g.Custom(func() {
						if g.IsItemHovered() && g.IsMouseClicked(g.MouseButtonLeft) {
							mousePos := g.GetMousePos()
							mousePos.X -= childPos.X
							mousePos.Y -= childPos.Y
							m.handleMapMouse(mousePos, 0)
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

func (m *Maps) handleMapMouse(p image.Point, which int) {
	_, ok := m.maps[m.currentMap]
	if !ok {
		return
	}
	scale := 4.0
	tWidth := 8
	tHeight := 6

	hitX := int(float64(p.X) / scale)
	hitY := int(float64(p.Y) / scale)

	focusedY := 3

	xOffset := focusedY * 1
	yOffset := focusedY * 4

	nearestX := (hitX + xOffset) / tWidth
	nearestY := (hitY - yOffset) / tHeight
	log.Printf("%dx%d: %dx%d\n", hitX, hitY, nearestX, nearestY)
}

func (m *Maps) createMapTexture(name string, sm *sdata.Map, dm *data.Manager) {
	mT := MapTexture{}
	scale := 4.0
	tWidth := 8
	tHeight := 6
	padding := 4
	cWidth := sm.Width * tWidth
	cHeight := sm.Depth * tHeight

	mT.width = int(float64(cWidth+(sm.Height*1)+padding*2) * scale)
	mT.height = int(float64(cHeight+(sm.Height*4)+padding*2) * scale)

	dc := gg.NewContext(int(mT.width), int(mT.height))
	dc.SetRGB(0.1, 0.1, 0.1)
	dc.Clear()

	startX := padding
	startY := padding

	// Draw archetypes.
	for y := 0; y < sm.Height; y++ {
		xOffset := y * 1
		yOffset := y * 4
		for x := 0; x < sm.Width; x++ {
			for z := 0; z < sm.Depth; z++ {
				oX := float64(x*tWidth+xOffset+startX) * scale
				oY := float64(z*tHeight-yOffset+startY) * scale
				for t := 0; t < len(sm.Tiles[y][x][z]); t++ {
					img, _ := m.GetImage(&sm.Tiles[y][x][z][t], dm, scale)
					if img != nil {
						dc.DrawImage(img, int(oX), int(oY))
					}
				}
			}
		}
	}

	// Draw grid.
	dc.SetLineWidth(1)
	for y := 0; y < sm.Height; y++ {
		xOffset := y * 1
		yOffset := y * 4
		for x := 0; x < sm.Width; x++ {
			for z := 0; z < sm.Depth; z++ {
				oX := float64(x*tWidth+xOffset+startX) * scale
				oY := float64(z*tHeight-yOffset+startY) * scale
				oW := float64(tWidth) * scale
				oH := float64(tHeight) * scale
				dc.DrawRectangle(oX, oY, oW, oH)
				if y == m.focusedY {
					if x == m.focusedX && z == m.focusedZ {
						dc.SetRGBA(0.2, 0.3, 0.6, 0.25)
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

func (m *Maps) GetAnimAndFace(dm *data.Manager, a *sdata.Archetype, anim, face string) (string, string) {
	if anim == "" && a.Anim != "" {
		anim = a.Anim
	}
	if face == "" && a.Face != "" {
		face = a.Face
	}

	if anim == "" || face == "" {
		if a.Arch != "" {
			o := dm.GetArchetype(a.Arch)
			if o != nil {
				anim, face = m.GetAnimAndFace(dm, o, anim, face)
				if anim != "" && face != "" {
					return anim, face
				}
			}
		}
		for _, name := range a.Archs {
			o := dm.GetArchetype(name)
			if o != nil {
				anim, face = m.GetAnimAndFace(dm, o, anim, face)
				if anim != "" && face != "" {
					return anim, face
				}
			}
		}
	}

	return anim, face
}

func (m *Maps) GetImage(a *sdata.Archetype, dm *data.Manager, scale float64) (img image.Image, err error) {
	anim, face := m.GetAnimAndFace(dm, a, "", "")

	imgName, err := dm.GetAnimFaceImage(anim, face)
	if err != nil {
		return nil, err
	}

	img = dm.GetScaledImage(scale, imgName)
	if img == nil {
		return nil, errors.New("missing image")
	}

	// Didn't find anything, return missing image...
	return
}

func (m *Maps) saveAll() {
	log.Println("TODO: Save all maps in file")
}

func (m *Maps) close() {
	log.Println("TODO: Issue close of map")
}
