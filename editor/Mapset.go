package editor

import (
	"fmt"
	"image"
	"image/color"

	g "github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/chimera-rpg/go-editor/data"
	sdata "github.com/chimera-rpg/go-server/data"
	"github.com/fogleman/gg"
	log "github.com/sirupsen/logrus"
)

type Mapset struct {
	filename                                      string
	maps                                          []UnReMap
	currentMapIndex                               int
	mapTextures                                   map[int]MapTexture
	focusedY, focusedX, focusedZ                  int
	resizeL, resizeR                              int32
	resizeT, resizeB                              int32
	resizeU, resizeD                              int32
	newH, newW, newD                              int32
	newDataName, newName, newDescription, newLore string
	zoom                                          int32
	showGrid                                      bool
	shouldClose                                   bool
}

type MapTexture struct {
	texture *g.Texture
	width   int
	height  int
}

func NewMapset(name string, maps map[string]*sdata.Map) *Mapset {
	m := &Mapset{
		filename:    name,
		mapTextures: make(map[int]MapTexture),
		zoom:        3.0,
		showGrid:    true,
		newW:        1,
		newH:        1,
		newD:        1,
	}

	for k, v := range maps {
		m.maps = append(m.maps, NewUnReMap(v, k))
	}

	return m
}

func (m *Mapset) draw(d *data.Manager) {
	childPos := image.Point{0, 0}

	var b bool

	var mapExists bool
	var resizeMapPopup, newMapPopup, adjustMapPopup, deleteMapPopup bool

	if m.CurrentMap() != nil {
		mapExists = true
	}

	filename := m.filename
	if filename == "" {
		filename = "Untitled Map"
	}
	g.WindowV(fmt.Sprintf("Mapset: %s", filename), &b, g.WindowFlagsMenuBar, 210, 30, 300, 400, g.Layout{
		g.MenuBar(g.Layout{
			g.Menu("Mapset", g.Layout{
				g.MenuItem("New Map...", func() {
					newMapPopup = true
				}),
				g.Separator(),
				g.MenuItem("Save All", func() { m.saveAll() }),
				g.Separator(),
				g.MenuItem("Close", func() { m.close() }),
			}),
			g.Menu("Map", g.Layout{
				g.MenuItemV("Properties...", false, mapExists, func() {
					cm := m.CurrentMap()
					m.newName = cm.Get().Name
					m.newDataName = cm.DataName()
					m.newDescription = cm.Get().Description
					m.newLore = cm.Get().Lore
					adjustMapPopup = true
				}),
				g.MenuItemV("Resize...", false, mapExists, func() {
					resizeMapPopup = true
				}),
				g.Separator(),
				g.MenuItemV("Undo", false, mapExists, func() {
					cm := m.CurrentMap()
					cm.Undo()
				}),
				g.MenuItemV("Redo", false, mapExists, func() {
					cm := m.CurrentMap()
					cm.Redo()
				}),
				g.Separator(),
				g.MenuItemV("Delete...", false, mapExists, func() {
					deleteMapPopup = true
				}),
			}),
			g.Menu("View", g.Layout{
				g.Checkbox("Grid", &m.showGrid, nil),
				g.SliderInt("Zoom", &m.zoom, 1, 8, "%d"),
			}),
		}),
		g.Custom(func() {
			if imgui.BeginTabBarV("Mapset", int(g.TabBarFlagsFittingPolicyScroll|g.TabBarFlagsFittingPolicyResizeDown)) {
				for mapIndex, v := range m.maps {
					if imgui.BeginTabItemV(fmt.Sprintf("%s(%s)", v.DataName(), v.Get().Name), nil, 0) {
						m.currentMapIndex = mapIndex
						// Generate texture.
						t, ok := m.mapTextures[mapIndex]
						go func() {
							m.createMapTexture(mapIndex, v.Get(), d)
							g.Update()
						}()
						// Render content (if texture is ready)
						if ok && t.texture != nil {
							var availW, availH float32
							g.Layout{
								g.Custom(func() {
									availW, availH = g.GetAvaiableRegion()
								}),
								g.Child(v.Get().Name, false, availW, availH-20, g.WindowFlagsHorizontalScrollbar, g.Layout{
									g.Custom(func() {
										childPos = g.GetCursorScreenPos()
									}),
									g.ImageButtonV(t.texture, float32(t.width), float32(t.height), image.Point{X: 0, Y: 0}, image.Point{X: 1, Y: 1}, 0, color.RGBA{0, 0, 0, 0}, color.RGBA{255, 255, 255, 255}, nil),
									g.Custom(func() {
										if g.IsItemHovered() && g.IsMouseClicked(g.MouseButtonLeft) {
											mousePos := g.GetMousePos()
											mousePos.X -= childPos.X
											mousePos.Y -= childPos.Y
											m.handleMapMouse(mousePos, 0, d)
										}
									}),
								}),
								g.Label("info bar"),
							}.Build()
						}
						imgui.EndTabItem()
					}
				}
				imgui.EndTabBar()
			}
		}),
		g.Custom(func() {
			if resizeMapPopup {
				g.OpenPopup("Resize Map")
			} else if newMapPopup {
				g.OpenPopup("New Map")
			} else if adjustMapPopup {
				g.OpenPopup("Map Properties")
			} else if deleteMapPopup {
				g.OpenPopup("Delete Map")
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
		g.PopupModalV("New Map", nil, g.WindowFlagsHorizontalScrollbar, g.Layout{
			g.Label("Create a new map"),
			g.InputText("Data Name", 0, &m.newDataName),
			g.InputText("Name", 0, &m.newName),
			g.InputTextMultiline("Description", &m.newDescription, 0, 0, g.InputTextFlagsAllowTabInput, nil, nil),
			g.InputTextMultiline("Lore", &m.newLore, 0, 0, g.InputTextFlagsAllowTabInput, nil, nil),
			g.SliderInt("Height", &m.newH, 1, 200, "%d"),
			g.SliderInt("Width ", &m.newW, 1, 200, "%d"),
			g.SliderInt("Depth ", &m.newD, 1, 200, "%d"),
			g.Line(
				g.Button("Create", func() {
					g.CloseCurrentPopup()
					// TODO: Check if map with same name already exists!
					newMap := m.createMap(m.newName, m.newDescription, m.newLore, 0, 0, int(m.newH), int(m.newW), int(m.newD))
					m.maps = append(m.maps, NewUnReMap(newMap, m.newDataName))
					m.newName, m.newDataName, m.newDescription, m.newLore = "", "", "", ""
				}),
				g.Button("Cancel", func() {
					g.CloseCurrentPopup()
					m.newName, m.newDataName, m.newDescription, m.newLore = "", "", "", ""
				}),
			),
		}),
		g.PopupModalV("Map Properties", nil, g.WindowFlagsHorizontalScrollbar, g.Layout{
			g.InputText("Data Name", 0, &m.newDataName),
			g.InputText("Name", 0, &m.newName),
			g.InputTextMultiline("Description", &m.newDescription, 0, 0, g.InputTextFlagsAllowTabInput, nil, nil),
			g.InputTextMultiline("Lore", &m.newLore, 0, 0, g.InputTextFlagsAllowTabInput, nil, nil),
			g.Line(
				g.Button("Save", func() {
					g.CloseCurrentPopup()
					//
					cm := m.CurrentMap()

					clone := m.cloneMap(cm.Get())
					clone.Name = m.newName
					clone.Description = m.newDescription
					clone.Lore = m.newLore
					cm.dataName = m.newDataName

					cm.Set(clone)

					m.newName, m.newDataName, m.newDescription, m.newLore = "", "", "", ""
				}),
				g.Button("Cancel", func() {
					g.CloseCurrentPopup()
					m.newName, m.newDataName, m.newDescription, m.newLore = "", "", "", ""
				}),
			),
		}),
		g.PopupModalV("Delete Map", nil, 0, g.Layout{
			g.Label("Delete map?"),
			g.Label("This cannot be recovered."),
			g.Line(
				g.Button("Delete", func() {
					m.deleteMap(m.currentMapIndex)
					g.CloseCurrentPopup()
				}),
				g.Button("Cancel", func() {
					g.CloseCurrentPopup()
				}),
			),
		}),
	})
}

func (m *Mapset) handleMapMouse(p image.Point, which int, dm *data.Manager) {
	sm := m.CurrentMap()

	scale := float64(m.zoom)
	padding := 4
	tWidth := int(dm.AnimationsConfig.TileWidth)
	tHeight := int(dm.AnimationsConfig.TileHeight)

	hitX := int(float64(p.X) / scale)
	hitY := int(float64(p.Y) / scale)

	xOffset := m.focusedY*int(dm.AnimationsConfig.YStep.X) + padding
	yOffset := m.focusedY*int(dm.AnimationsConfig.YStep.Y) + padding

	nearestX := (hitX+xOffset)/tWidth - 1
	nearestY := (hitY - yOffset) / tHeight
	if nearestX >= 0 && nearestX < sm.Get().Width && nearestY >= 0 && nearestY < sm.Get().Depth {
		m.focusedX = nearestX
		m.focusedZ = nearestY
	}
}

func (m *Mapset) createMapTexture(index int, sm *sdata.Map, dm *data.Manager) {
	mT := MapTexture{}
	scale := float64(m.zoom)
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
	if m.showGrid {
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
						dc.SetRGB(0.9, 0.9, 0.9)
					} else {
						dc.SetRGBA(0.9, 0.9, 0.9, 0.1)
					}
					dc.Stroke()
				}
			}
		}
	}

	// Draw selected.
	{
		xOffset := m.focusedY * int(yStep.X)
		yOffset := m.focusedY * int(yStep.Y)
		oX := float64(m.focusedX*tWidth+xOffset+startX) * scale
		oY := float64(m.focusedZ*tHeight-yOffset+startY) * scale
		oW := float64(tWidth) * scale
		oH := float64(tHeight) * scale
		dc.DrawRectangle(oX, oY, oW, oH)
		dc.SetLineWidth(2)
		dc.SetRGBA(0, 0, 0, 0.85)
		dc.StrokePreserve()
		dc.SetLineWidth(1)
		dc.SetRGBA(1, 0, 0, 0.85)
		dc.Stroke()
	}

	var err error
	mT.texture, err = g.NewTextureFromRgba(dc.Image().(*image.RGBA))
	if err != nil {
		log.Fatalln(err)
	}
	m.mapTextures[index] = mT
}

func (m *Mapset) saveAll() {
	log.Println("TODO: Save all maps in file")
}

func (m *Mapset) close() {
	m.shouldClose = true
}

func (m *Mapset) resizeMap(u, d, l, r, t, b int) {
	cm := m.CurrentMap()
	nH := cm.Get().Height + u + d
	nW := cm.Get().Width + l + r
	nD := cm.Get().Depth + t + b
	offsetY := d
	offsetX := l
	offsetZ := t
	// Make a new map according to the given dimensions
	newMap := m.createMap(
		cm.Get().Name,
		cm.Get().Description,
		cm.Get().Lore,
		cm.Get().Darkness,
		cm.Get().ResetTime,
		nH,
		nW,
		nD,
	)
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
	for y := 0; y < cm.Get().Height; y++ {
		if y+offsetY < 0 || y+offsetY >= newMap.Height {
			continue
		}
		for x := 0; x < cm.Get().Width; x++ {
			if x+offsetX < 0 || x+offsetX >= newMap.Width {
				continue
			}
			for z := 0; z < cm.Get().Depth; z++ {
				if z+offsetZ < 0 || z+offsetZ >= newMap.Depth {
					continue
				}
				newMap.Tiles[y+offsetY][x+offsetX][z+offsetZ] = cm.Get().Tiles[y][x][z]
			}
		}
	}
	cm.Set(newMap)
}

func (m *Mapset) cloneMap(t *sdata.Map) *sdata.Map {
	clone := m.createMap(
		t.Name,
		t.Description,
		t.Lore,
		t.Darkness,
		t.ResetTime,
		t.Height,
		t.Width,
		t.Depth,
	)
	// Create the new map according to dimensions.
	for y := 0; y < t.Height; y++ {
		clone.Tiles = append(clone.Tiles, [][][]sdata.Archetype{})
		for x := 0; x < t.Width; x++ {
			clone.Tiles[y] = append(clone.Tiles[y], [][]sdata.Archetype{})
			for z := 0; z < t.Depth; z++ {
				clone.Tiles[y][x] = append(t.Tiles[y][x], []sdata.Archetype{})
				clone.Tiles[y][x][z] = t.Tiles[y][x][z]
			}
		}
	}
	return clone
}

func (m *Mapset) createMap(name, desc, lore string, darkness, resettime int, h, w, d int) *sdata.Map {
	// Make a new map according to the given dimensions
	newMap := &sdata.Map{
		Name:        name,
		Description: desc,
		Darkness:    darkness,
		ResetTime:   resettime,
		Lore:        lore,
		Height:      h,
		Width:       w,
		Depth:       d,
	}
	// Create the new map according to dimensions.
	for y := 0; y < h; y++ {
		newMap.Tiles = append(newMap.Tiles, [][][]sdata.Archetype{})
		for x := 0; x < w; x++ {
			newMap.Tiles[y] = append(newMap.Tiles[y], [][]sdata.Archetype{})
			for z := 0; z < d; z++ {
				newMap.Tiles[y][x] = append(newMap.Tiles[y][x], []sdata.Archetype{})
			}
		}
	}
	return newMap
}

func (m *Mapset) deleteMap(index int) {
	if index >= len(m.maps) || index < 0 {
		return
	}
	m.maps = append(m.maps[:index], m.maps[index+1:]...)
	if m.currentMapIndex > index {
		m.currentMapIndex--
	}
	if m.currentMapIndex < 0 {
		m.currentMapIndex = 0
	}
}

func (m *Mapset) CurrentMap() *UnReMap {
	if m.currentMapIndex < 0 || m.currentMapIndex >= len(m.maps) {
		return nil
	}
	return &m.maps[m.currentMapIndex]
}
