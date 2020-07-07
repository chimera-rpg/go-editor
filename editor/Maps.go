package editor

import (
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
	resizeL, resizeR             int32
	resizeT, resizeB             int32
	resizeU, resizeD             int32
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
	//sm := m.maps[m.currentMap]
	var tabs []g.Widget
	childPos := image.Point{0, 0}
	for k, v := range m.maps {
		t, ok := m.mapTextures[k]
		go func() {
			m.createMapTexture(k, v, d)
			g.Update()
		}()
		if ok && t.texture != nil {
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
							m.handleMapMouse(mousePos, 0, d)
						}
					}),
				}),
			}))
		}
	}
	var b bool

	var resizeMapPopup bool

	g.WindowV(fmt.Sprintf("Maps: %s", m.filename), &b, g.WindowFlagsMenuBar, 210, 30, 300, 400, g.Layout{
		g.MenuBar(g.Layout{
			g.Menu("Maps", g.Layout{
				g.MenuItem("Save", func() { m.saveAll() }),
				g.Separator(),
				g.MenuItem("Close", func() { m.close() }),
			}),
			g.Menu("Map", g.Layout{
				g.MenuItem("Resize", func() {
					resizeMapPopup = true
				}),
			}),
		}),
		g.TabBarV("Tabs", g.TabBarFlagsFittingPolicyScroll|g.TabBarFlagsFittingPolicyResizeDown, tabs),
		g.Custom(func() {
			if resizeMapPopup {
				g.OpenPopup("Resize Map")
			}
		}),
		g.PopupModalV("Resize Map", nil, 0, g.Layout{
			g.Label("Grow or Shrink the current map"),
			g.Line(
				g.InputInt("Up    ", 50, &m.resizeU),
				g.InputInt("Down  ", 50, &m.resizeD),
			),
			g.Line(
				g.InputInt("Left  ", 50, &m.resizeL),
				g.InputInt("Right ", 50, &m.resizeR),
			),
			g.Line(
				g.InputInt("Top   ", 50, &m.resizeT),
				g.InputInt("Bottom", 50, &m.resizeB),
			),
			g.Line(
				g.Button("Resize", func() {
					m.resizeMap(int(m.resizeU), int(m.resizeD), int(m.resizeL), int(m.resizeR), int(m.resizeT), int(m.resizeB))
					m.resizeU, m.resizeD, m.resizeL, m.resizeR, m.resizeT, m.resizeB = 0, 0, 0, 0, 0, 0
					g.CloseCurrentPopup()
				}),
				g.Button("Cancel", func() {
					m.resizeU, m.resizeD, m.resizeL, m.resizeR, m.resizeT, m.resizeB = 0, 0, 0, 0, 0, 0
					g.CloseCurrentPopup()
				}),
			),
		}),
	})
}

func (m *Maps) handleMapMouse(p image.Point, which int, dm *data.Manager) {
	sm, ok := m.maps[m.currentMap]
	if !ok {
		return
	}
	scale := 4.0
	padding := 4
	tWidth := int(dm.AnimationsConfig.TileWidth)
	tHeight := int(dm.AnimationsConfig.TileHeight)

	hitX := int(float64(p.X) / scale)
	hitY := int(float64(p.Y) / scale)

	xOffset := m.focusedY*int(dm.AnimationsConfig.YStep.X) + padding
	yOffset := m.focusedY*int(dm.AnimationsConfig.YStep.Y) + padding

	nearestX := (hitX+xOffset)/tWidth - 1
	nearestY := (hitY - yOffset) / tHeight
	if nearestX >= 0 && nearestX < sm.Width && nearestY >= 0 && nearestY < sm.Depth {
		m.focusedX = nearestX
		m.focusedZ = nearestY
	}
}

func (m *Maps) createMapTexture(name string, sm *sdata.Map, dm *data.Manager) {
	mT := MapTexture{}
	scale := 4.0
	tWidth := int(dm.AnimationsConfig.TileWidth)
	tHeight := int(dm.AnimationsConfig.TileHeight)
	yStep := dm.AnimationsConfig.YStep
	padding := 4
	cWidth := sm.Width * tWidth
	cHeight := sm.Depth * tHeight

	mT.width = int(float64(cWidth+(sm.Height*int(yStep.X))+padding*2) * scale)
	mT.height = int(float64(cHeight+(sm.Height*int(yStep.Y))+padding*4) * scale)

	dc := gg.NewContext(int(mT.width), int(mT.height))
	dc.SetRGB(0.1, 0.1, 0.1)
	dc.Clear()

	startX := padding
	startY := padding

	// Draw archetypes.
	for y := 0; y < sm.Height; y++ {
		xOffset := y * int(yStep.X)
		yOffset := y * int(yStep.Y)
		for x := sm.Width - 1; x >= 0; x-- {
			for z := 0; z < sm.Depth; z++ {
				oX := float64(x*tWidth+xOffset+startX) * scale
				oY := float64(z*tHeight-yOffset+startY) * scale
				for t := 0; t < len(sm.Tiles[y][x][z]); t++ {
					if adjustment, ok := dm.AnimationsConfig.Adjustments[dm.GetArchType(&sm.Tiles[y][x][z][t], 0)]; ok {
						oX += float64(adjustment.X) * scale
						oY += float64(adjustment.Y) * scale
					}
					img, _ := dm.GetArchImage(&sm.Tiles[y][x][z][t], scale)
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
		xOffset := y * int(yStep.X)
		yOffset := y * int(yStep.Y)
		for x := 0; x < sm.Width; x++ {
			for z := 0; z < sm.Depth; z++ {
				oX := float64(x*tWidth+xOffset+startX) * scale
				oY := float64(z*tHeight-yOffset+startY) * scale
				oW := float64(tWidth) * scale
				oH := float64(tHeight) * scale
				dc.DrawRectangle(oX, oY, oW, oH)
				if y == m.focusedY {
					if x == m.focusedX && z == m.focusedZ {
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

func (m *Maps) resizeMap(u, d, l, r, t, b int) {
	cm, ok := m.maps[m.currentMap]
	if !ok {
		return
	}
	nH := cm.Height + u + d
	nW := cm.Width + l + r
	nD := cm.Depth + t + b
	offsetY := d
	offsetX := l
	offsetZ := t
	// Make a new map according to the given dimensions
	newMap := &sdata.Map{
		Name:        cm.Name,
		Description: cm.Description,
		Darkness:    cm.Darkness,
		Lore:        cm.Lore,
		ResetTime:   cm.ResetTime,
		Height:      nH,
		Width:       nW,
		Depth:       nD,
	}
	// Create the new map according to dimensions.
	for y := 0; y < nH; y++ {
		newMap.Tiles = append(newMap.Tiles, [][][]sdata.Archetype{})
		for x := 0; x < nW; x++ {
			newMap.Tiles[y] = append(newMap.Tiles[y], [][]sdata.Archetype{})
			for z := 0; z < nD; z++ {
				newMap.Tiles[y][x] = append(newMap.Tiles[y][x], []sdata.Archetype{})
			}
		}
	}
	// Iterate through old map tiles and copy what is in range.
	for y := 0; y < cm.Height; y++ {
		if y+offsetY < 0 || y+offsetY >= newMap.Height {
			continue
		}
		for x := 0; x < cm.Width; x++ {
			if x+offsetX < 0 || x+offsetX >= newMap.Width {
				continue
			}
			for z := 0; z < cm.Depth; z++ {
				if z+offsetZ < 0 || z+offsetZ >= newMap.Depth {
					continue
				}
				newMap.Tiles[y+offsetY][x+offsetX][z+offsetZ] = cm.Tiles[y][x][z]
			}
		}
	}
	m.maps[m.currentMap] = newMap
	log.Printf("%+v\n", newMap)
}
