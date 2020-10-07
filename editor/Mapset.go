package editor

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"

	g "github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/chimera-rpg/go-editor/widgets"
	sdata "github.com/chimera-rpg/go-server/data"
	log "github.com/sirupsen/logrus"
)

type Mapset struct {
	context                      *Context
	filename                     string
	maps                         []*UnReMap
	currentMapIndex              int
	focusedY, focusedX, focusedZ int
	selectedYStart, selectedYEnd int
	selectedXStart, selectedXEnd int
	selectedZStart, selectedZEnd int
	selectedCoords               SelectedCoords
	resizeL, resizeR             int32
	resizeT, resizeB             int32
	resizeU, resizeD             int32
	newH, newW, newD             int32
	newDataName, newName         string
	loreEditor, descEditor       imgui.TextEditor
	zoom                         int32
	showGrid                     bool
	showYGrids                   bool
	onionskin                    bool
	keepSameTile                 bool
	shouldClose                  bool
	visitedTiles                 map[image.Point]bool // Coordinates visited during mouse drag.
	mouseHeld                    map[g.MouseButton]bool
	toolBinds                    map[g.MouseButton]int
	blockScroll                  bool // Block map scrolling (true if ctrl or alt is held)
	unsaved                      bool
}

const (
	noTool = iota
	selectTool
	insertTool
	pickTool
	eraseTool
)

func NewMapset(context *Context, name string, maps map[string]*sdata.Map) *Mapset {
	m := &Mapset{
		filename:     name,
		zoom:         3.0,
		showGrid:     true,
		showYGrids:   false,
		onionskin:    true,
		keepSameTile: true,
		newW:         1,
		newH:         1,
		newD:         1,
		context:      context,
		loreEditor:   imgui.NewTextEditor(),
		descEditor:   imgui.NewTextEditor(),
		visitedTiles: make(map[image.Point]bool),
		mouseHeld:    make(map[g.MouseButton]bool),
		toolBinds:    make(map[g.MouseButton]int),
	}
	m.loreEditor.SetShowWhitespaces(false)
	m.descEditor.SetShowWhitespaces(false)

	m.bindMouseToTool(g.MouseButtonLeft, selectTool)
	m.bindMouseToTool(g.MouseButtonMiddle, eraseTool)
	m.bindMouseToTool(g.MouseButtonRight, insertTool)

	for k, v := range maps {
		m.maps = append(m.maps, NewUnReMap(v, k))
	}

	m.selectedCoords.Clear()

	return m
}

func (m *Mapset) draw() {
	windowOpen := true

	var mapExists bool
	var resizeMapPopup, newMapPopup, adjustMapPopup, deleteMapPopup bool

	if m.CurrentMap() != nil {
		mapExists = true
	}

	filename := m.filename
	if filename == "" {
		filename = "Untitled Map"
	}
	toolWidth, _ := g.CalcTextSize("_________")

	// Block mousewheel scrolling if alt or ctrl is held.
	m.blockScroll = false
	widgets.KeyBinds(0,
		widgets.KeyBind(widgets.KeyBindFlagPressed, widgets.Keys(widgets.KeyAlt), nil, func() {
			m.blockScroll = true
		}),
		widgets.KeyBind(widgets.KeyBindFlagPressed, widgets.Keys(widgets.KeyControl), nil, func() {
			m.blockScroll = true
		}),
	).Build()

	windowFlags := g.WindowFlagsMenuBar

	if m.Unsaved() {
		windowFlags |= g.WindowFlagsUnsavedDocument
	}
	g.WindowV(fmt.Sprintf("Mapset: %s", filename), &windowOpen, windowFlags, 210, 30, 300, 400, g.Layout{
		g.MenuBar(g.Layout{
			g.Menu("Mapset", g.Layout{
				g.MenuItem("New Map...", func() {
					newMapPopup = true
					m.descEditor.SetText("")
					m.loreEditor.SetText("")
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
					m.descEditor.SetText(cm.Get().Description)
					m.loreEditor.SetText(cm.Get().Lore)
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
			g.Menu("Settings", g.Layout{
				g.Checkbox("Keep Same Tile", &m.keepSameTile, nil),
			}),
			g.Menu("View", g.Layout{
				g.Checkbox("Onionskinning", &m.onionskin, nil),
				g.Checkbox("Grid", &m.showGrid, nil),
				g.Checkbox("Y Grids", &m.showYGrids, nil),
				g.SliderInt("Zoom", &m.zoom, 1, 8, "%d"),
			}),
		}),
		g.Custom(func() {
			imgui.SelectableV(fmt.Sprintf("select (%s)", m.getToolButtonString(selectTool)), m.isToolBound(selectTool), 0, imgui.Vec2{X: toolWidth, Y: 0})
			if g.IsItemHovered() {
				if g.IsMouseClicked(g.MouseButtonLeft) {
					m.bindMouseToTool(g.MouseButtonLeft, selectTool)
				} else if g.IsMouseClicked(g.MouseButtonMiddle) {
					m.bindMouseToTool(g.MouseButtonMiddle, selectTool)
				} else if g.IsMouseClicked(g.MouseButtonRight) {
					m.bindMouseToTool(g.MouseButtonRight, selectTool)
				}
			}
			imgui.SameLine()
			imgui.SelectableV(fmt.Sprintf("insert (%s)", m.getToolButtonString(insertTool)), m.isToolBound(insertTool), 0, imgui.Vec2{X: toolWidth, Y: 0})
			if g.IsItemHovered() {
				if g.IsMouseClicked(g.MouseButtonLeft) {
					m.bindMouseToTool(g.MouseButtonLeft, insertTool)
				} else if g.IsMouseClicked(g.MouseButtonMiddle) {
					m.bindMouseToTool(g.MouseButtonMiddle, insertTool)
				} else if g.IsMouseClicked(g.MouseButtonRight) {
					m.bindMouseToTool(g.MouseButtonRight, insertTool)
				}
			}
			imgui.SameLine()
			imgui.SelectableV(fmt.Sprintf("pick (%s)", m.getToolButtonString(pickTool)), m.isToolBound(pickTool), 0, imgui.Vec2{X: toolWidth, Y: 0})
			if g.IsItemHovered() {
				if g.IsMouseClicked(g.MouseButtonLeft) {
					m.bindMouseToTool(g.MouseButtonLeft, pickTool)
				} else if g.IsMouseClicked(g.MouseButtonMiddle) {
					m.bindMouseToTool(g.MouseButtonMiddle, pickTool)
				} else if g.IsMouseClicked(g.MouseButtonRight) {
					m.bindMouseToTool(g.MouseButtonRight, pickTool)
				}
			}
			imgui.SameLine()
			imgui.SelectableV(fmt.Sprintf("erase (%s)", m.getToolButtonString(eraseTool)), m.isToolBound(eraseTool), 0, imgui.Vec2{X: toolWidth, Y: 0})
			if g.IsItemHovered() {
				if g.IsMouseClicked(g.MouseButtonLeft) {
					m.bindMouseToTool(g.MouseButtonLeft, eraseTool)
				} else if g.IsMouseClicked(g.MouseButtonMiddle) {
					m.bindMouseToTool(g.MouseButtonMiddle, eraseTool)
				} else if g.IsMouseClicked(g.MouseButtonRight) {
					m.bindMouseToTool(g.MouseButtonRight, eraseTool)
				}
			}
		}),
		m.layoutMapTabs(),
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
			g.Custom(func() {
				availW, _ := g.GetAvaiableRegion()
				labelV := imgui.CalcTextSize("Description", false, 0)
				m.descEditor.Render("Description", imgui.Vec2{X: availW - labelV.X - 5, Y: 200}, false)
				imgui.SameLine()
				g.Label("Description").Build()
				m.loreEditor.Render("Lore", imgui.Vec2{X: availW - labelV.X - 5, Y: 200}, false)
				imgui.SameLine()
				g.Label("Lore").Build()
			}),
			g.SliderInt("Height", &m.newH, 1, 200, "%d"),
			g.SliderInt("Width ", &m.newW, 1, 200, "%d"),
			g.SliderInt("Depth ", &m.newD, 1, 200, "%d"),
			g.Line(
				g.Button("Create", func() {
					g.CloseCurrentPopup()
					lore := m.loreEditor.GetText()
					desc := m.descEditor.GetText()
					// TODO: Check if map with same name already exists!
					newMap := m.createMap(m.newName, desc, lore, 0, 0, int(m.newH), int(m.newW), int(m.newD))
					m.maps = append(m.maps, NewUnReMap(newMap, m.newDataName))
					m.newName, m.newDataName = "", ""
				}),
				g.Button("Cancel", func() {
					g.CloseCurrentPopup()
					m.newName, m.newDataName = "", ""
				}),
			),
		}),
		g.PopupModalV("Map Properties", nil, g.WindowFlagsHorizontalScrollbar, g.Layout{
			g.InputText("Data Name", 0, &m.newDataName),
			g.InputText("Name", 0, &m.newName),
			g.Custom(func() {
				availW, availH := g.GetAvaiableRegion()
				labelV := imgui.CalcTextSize("Description", false, 0)
				m.descEditor.Render("Description", imgui.Vec2{X: availW - labelV.X - 5, Y: availH/2 - labelV.Y - 3}, false)
				imgui.SameLine()
				g.Label("Description").Build()
				m.loreEditor.Render("Lore", imgui.Vec2{X: availW - labelV.X - 5, Y: availH/2 - labelV.Y - 3}, false)
				imgui.SameLine()
				g.Label("Lore").Build()
			}),
			g.Line(
				g.Button("Save", func() {
					g.CloseCurrentPopup()
					//
					cm := m.CurrentMap()

					clone := cm.Clone()
					clone.Name = m.newName
					clone.Description = m.descEditor.GetText()
					clone.Lore = m.loreEditor.GetText()

					cm.dataName = m.newDataName

					cm.Set(clone)

					m.newName, m.newDataName = "", ""
				}),
				g.Button("Cancel", func() {
					g.CloseCurrentPopup()
					m.newName, m.newDataName = "", ""
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
		widgets.KeyBinds(widgets.KeyBindsFlagWindowFocused,
			widgets.KeyBind(widgets.KeyBindFlagPressed, widgets.Keys(widgets.KeyShift, widgets.KeyControl), widgets.Keys(widgets.KeyZ), func() {
				if cm := m.CurrentMap(); cm != nil {
					cm.Redo()
				}
			}),
			widgets.KeyBind(widgets.KeyBindFlagPressed, widgets.Keys(widgets.KeyControl), widgets.Keys(widgets.KeyZ), func() {
				if cm := m.CurrentMap(); cm != nil {
					cm.Undo()
				}
			}),
			widgets.KeyBind(widgets.KeyBindFlagPressed, widgets.Keys(widgets.KeyControl), widgets.Keys(widgets.KeyY), func() {
				if cm := m.CurrentMap(); cm != nil {
					cm.Redo()
				}
			}),
			widgets.KeyBind(widgets.KeyBindFlagPressed, widgets.Keys(), widgets.Keys(widgets.KeyLeft), func() {
				if m.focusedX > 0 {
					m.focusedX -= 1
				}
			}),
			widgets.KeyBind(widgets.KeyBindFlagPressed, widgets.Keys(), widgets.Keys(widgets.KeyRight), func() {
				if cm := m.CurrentMap(); cm != nil {
					if m.focusedX < cm.Get().Width-1 {
						m.focusedX += 1
					}
				}
			}),
			widgets.KeyBind(widgets.KeyBindFlagPressed, widgets.Keys(), widgets.Keys(widgets.KeyUp), func() {
				if m.focusedZ > 0 {
					m.focusedZ -= 1
				}
			}),
			widgets.KeyBind(widgets.KeyBindFlagPressed, widgets.Keys(), widgets.Keys(widgets.KeyDown), func() {
				if cm := m.CurrentMap(); cm != nil {
					if m.focusedZ < cm.Get().Depth-1 {
						m.focusedZ += 1
					}
				}
			}),
			widgets.KeyBind(widgets.KeyBindFlagPressed, widgets.Keys(widgets.KeyAlt), widgets.Keys(widgets.KeyUp), func() {
				if cm := m.CurrentMap(); cm != nil {
					if m.focusedY < cm.Get().Height-1 {
						m.focusedY += 1
					}
				}
			}),
			widgets.KeyBind(widgets.KeyBindFlagPressed, widgets.Keys(widgets.KeyAlt), widgets.Keys(widgets.KeyDown), func() {
				if m.focusedY > 0 {
					m.focusedY -= 1
				}
			}),
			widgets.KeyBind(widgets.KeyBindFlagPressed, widgets.Keys(), widgets.Keys(widgets.KeyA), func() {
				if cm := m.CurrentMap(); cm != nil {
					m.toolInsert(cm, m.focusedY, m.focusedX, m.focusedZ)
				}
			}),
			widgets.KeyBind(widgets.KeyBindFlagPressed, widgets.Keys(), widgets.Keys(widgets.KeyD), func() {
				if cm := m.CurrentMap(); cm != nil {
					m.toolErase(cm, m.focusedY, m.focusedX, m.focusedZ)
				}
			}),
		),
	})

	if !windowOpen {
		m.close()
	}
}

func (m *Mapset) layoutMapTabs() g.Layout {
	var tabs g.Layout
	for mapIndex, v := range m.maps {
		var flags g.TabItemFlags
		if v.Unsaved() {
			flags |= g.TabItemFlagsUnsavedDocument
		}
		tab := g.TabItemV(fmt.Sprintf("%s(%s)", v.DataName(), v.Get().Name), nil, flags, g.Layout{
			g.Custom(func() {
				m.currentMapIndex = mapIndex
				availW, availH := g.GetAvaiableRegion()
				defaultW := float32(math.Round(float64(availW - availW/4)))
				defaultH := float32(math.Round(float64(availH - availH/4)))
				g.SplitLayout("vsplit", g.DirectionVertical, true, defaultH, g.Layout{
					g.SplitLayout("hsplit", g.DirectionHorizontal, true, defaultW,
						m.layoutMapView(v),
						m.layoutArchsList(v),
					),
				}, m.layoutSelectedArch(v)).Build()
			}),
		})

		tabs = append(tabs, tab)
	}
	return g.Layout{g.TabBarV("Mapset", g.TabBarFlagsFittingPolicyScroll|g.TabBarFlagsFittingPolicyResizeDown, tabs)}
}

func (m *Mapset) layoutMapView(v *UnReMap) g.Layout {
	var availW, availH float32
	childPos := image.Point{0, 0}
	childFlags := g.WindowFlagsHorizontalScrollbar | imgui.WindowFlagsNoMove
	if m.blockScroll {
		childFlags |= imgui.WindowFlagsNoScrollWithMouse
	}

	return g.Layout{
		g.Custom(func() {
			availW, availH = g.GetAvaiableRegion()
		}),
		g.Child(v.Get().Name, false, availW, availH-20, childFlags, g.Layout{
			g.Custom(func() {
				childPos = g.GetCursorScreenPos()
				m.drawMap(v)
			}),
			g.Custom(func() {
				if g.IsItemHovered() {
					mousePos := g.GetMousePos()
					mousePos.X -= childPos.X
					mousePos.Y -= childPos.Y

					// RMB
					if g.IsMouseDown(g.MouseButtonRight) {
						if _, ok := m.mouseHeld[g.MouseButtonRight]; !ok {
							m.mouseHeld[g.MouseButtonRight] = true
						}
						if p, err := m.getMapPointFromMouse(mousePos); err == nil {
							if _, ok := m.visitedTiles[p]; !ok {
								err := m.handleMouseTool(g.MouseButtonRight, m.focusedY, p.X, p.Y)
								if err != nil {
									log.Errorln(err)
								}
								m.visitedTiles[p] = true
							}
						}
					} else if g.IsMouseReleased(g.MouseButtonRight) {
						delete(m.mouseHeld, g.MouseButtonRight)
						m.visitedTiles = make(map[image.Point]bool)
					}
					// MMB
					if g.IsMouseDown(g.MouseButtonMiddle) {
						if _, ok := m.mouseHeld[g.MouseButtonMiddle]; !ok {
							m.mouseHeld[g.MouseButtonMiddle] = true
						}
						if p, err := m.getMapPointFromMouse(mousePos); err == nil {
							if _, ok := m.visitedTiles[p]; !ok {
								err := m.handleMouseTool(g.MouseButtonMiddle, m.focusedY, p.X, p.Y)
								if err != nil {
									log.Errorln(err)
								}
								m.visitedTiles[p] = true
							}
						}
					} else if g.IsMouseReleased(g.MouseButtonMiddle) {
						delete(m.mouseHeld, g.MouseButtonMiddle)
						m.visitedTiles = make(map[image.Point]bool)
					}
					// LMB
					if g.IsMouseDown(g.MouseButtonLeft) {
						if _, ok := m.mouseHeld[g.MouseButtonLeft]; !ok {
							m.mouseHeld[g.MouseButtonLeft] = true
						}
						if p, err := m.getMapPointFromMouse(mousePos); err == nil {
							if _, ok := m.visitedTiles[p]; !ok {
								err := m.handleMouseTool(g.MouseButtonLeft, m.focusedY, p.X, p.Y)
								if err != nil {
									log.Errorln(err)
								}
								m.visitedTiles[p] = true
							}
						}
					} else if g.IsMouseReleased(g.MouseButtonLeft) {
						delete(m.mouseHeld, g.MouseButtonLeft)
						m.visitedTiles = make(map[image.Point]bool)
					}
				}
			}),
		}),
		widgets.KeyBinds(widgets.KeyBindsFlagItemHovered,
			widgets.KeyBind(widgets.KeyBindFlagDown, widgets.Keys(), widgets.Keys(widgets.KeyAlt), func() {
				mouseWheelDelta, _ := g.Context.IO().GetMouseWheelDelta(), g.Context.IO().GetMouseWheelHDelta()
				if mouseWheelDelta != 0 {
					m.focusedY += int(mouseWheelDelta)
					if m.focusedY < 0 {
						m.focusedY = 0
					} else if m.focusedY >= v.Get().Height {
						m.focusedY = v.Get().Height - 1
					}
				}
			}),
			widgets.KeyBind(widgets.KeyBindFlagDown, widgets.Keys(), widgets.Keys(widgets.KeyControl), func() {
				mouseWheelDelta, _ := g.Context.IO().GetMouseWheelDelta(), g.Context.IO().GetMouseWheelHDelta()
				if mouseWheelDelta != 0 {
					m.zoom += int32(mouseWheelDelta)
					if m.zoom < 1 {
						m.zoom = 1
					} else if m.zoom > 8 {
						m.zoom = 8
					}
				}
			}),
		),
	}
}

func (m *Mapset) layoutArchsList(v *UnReMap) g.Layout {
	return g.Layout{
		g.Label("tile archetypes"),
	}
}

func (m *Mapset) layoutSelectedArch(v *UnReMap) g.Layout {
	return g.Layout{
		g.Label("current archetype"),
	}
}

func (m *Mapset) getMapPointFromMouse(p image.Point) (h image.Point, err error) {
	dm := m.context.dataManager
	sm := m.CurrentMap()

	scale := float64(m.zoom)
	padding := 4
	tWidth := int(dm.AnimationsConfig.TileWidth)
	tHeight := int(dm.AnimationsConfig.TileHeight)

	hitX := int(float64(p.X) / scale)
	hitY := int(float64(p.Y) / scale)

	xOffset := m.focusedY*int(dm.AnimationsConfig.YStep.X) + padding
	yOffset := m.focusedY*int(dm.AnimationsConfig.YStep.Y) + padding + (sm.Get().Height * int(-dm.AnimationsConfig.YStep.Y))

	nearestX := (hitX+xOffset)/tWidth - 1
	nearestY := (hitY - yOffset) / tHeight
	if nearestX >= 0 && nearestX < sm.Get().Width && nearestY >= 0 && nearestY < sm.Get().Depth {
		h.X = nearestX
		h.Y = nearestY
	} else {
		err = errors.New("Point OOB")
	}
	return
}

func (m *Mapset) bindMouseToTool(btn g.MouseButton, toolIndex int) {
	// Remove old btn bind.
	delete(m.toolBinds, btn)
	// Find and remove binding for this tool.
	for k, v := range m.toolBinds {
		if v == toolIndex {
			delete(m.toolBinds, k)
		}
	}
	// Set btn bind to this one.
	m.toolBinds[btn] = toolIndex
}

func (m *Mapset) getMouseTool(btn g.MouseButton) int {
	if toolIndex, ok := m.toolBinds[btn]; ok {
		return toolIndex
	}
	return noTool
}

func (m *Mapset) isToolBound(toolIndex int) bool {
	for _, v := range m.toolBinds {
		if v == toolIndex {
			return true
		}
	}
	return false
}

func (m *Mapset) getToolButtonString(toolIndex int) string {
	for k, v := range m.toolBinds {
		if v == toolIndex {
			if k == g.MouseButtonLeft {
				return "L"
			} else if k == g.MouseButtonMiddle {
				return "M"
			} else if k == g.MouseButtonRight {
				return "R"
			}
		}
	}
	return "_"
}

func (m *Mapset) handleMouseTool(btn g.MouseButton, y, x, z int) error {
	if toolIndex, ok := m.toolBinds[btn]; ok {
		if m.currentMapIndex < 0 || m.currentMapIndex >= len(m.maps) {
			return errors.New("no current map")
		}
		cm := m.maps[m.currentMapIndex]

		if toolIndex == insertTool {
			return m.toolInsert(cm, y, x, z)
		} else if toolIndex == selectTool {
			return m.toolSelect(cm, y, x, z)
		} else if toolIndex == eraseTool {
			return m.toolErase(cm, y, x, z)
		}
	}
	return nil
}

func (m *Mapset) toolSelect(v *UnReMap, y, x, z int) (err error) {
	// TODO: Check if Shift or Ctrl is held!
	m.selectedCoords.Clear()
	m.selectedCoords.Select(y, x, z)
	m.focusedY = y
	m.focusedX = x
	m.focusedZ = z
	return
}

func (m *Mapset) toolInsert(v *UnReMap, y, x, z int) (err error) {
	// Bail if no archetype is selected.
	if m.context.selectedArch == "" {
		return
	}
	// Check if we should not insert if top tile is the same.
	if m.keepSameTile {
		tiles := m.getTiles(v.Get(), y, x, z)
		if tiles != nil && len(*tiles) > 0 {
			if (*tiles)[len(*tiles)-1].Arch == m.context.selectedArch {
				return
			}
			for _, a := range (*tiles)[len(*tiles)-1].Archs {
				if a == m.context.selectedArch {
					return
				}
			}
		}
	}
	// Otherwise attempt to insert.
	clone := v.Clone()
	if err := m.insertArchetype(clone, m.context.selectedArch, y, x, z, -1); err != nil {
		return err
	}
	v.Set(clone)
	return
}

func (m *Mapset) toolErase(v *UnReMap, y, x, z int) (err error) {
	clone := v.Clone()
	if err := m.removeArchetype(clone, y, x, z, -1); err != nil {
		return err
	}
	v.Set(clone)
	return
}

type archDrawable struct {
	z    int
	x, y int
	w, h int
	t    *ImageTexture
	c    color.RGBA
}

func (m *Mapset) drawMap(v *UnReMap) {
	sm := v.Get()

	pos := g.GetCursorScreenPos()
	canvas := g.GetCanvas()
	dm := m.context.dataManager
	scale := int(m.zoom)
	tWidth := int(dm.AnimationsConfig.TileWidth)
	tHeight := int(dm.AnimationsConfig.TileHeight)
	yStep := dm.AnimationsConfig.YStep
	padding := 4
	cWidth := sm.Width * tWidth
	cHeight := sm.Depth * tHeight

	canvasWidth := int((cWidth + (sm.Height * int(yStep.X)) + padding*2) * scale)
	canvasHeight := int((cHeight + (sm.Height * int(-yStep.Y)) + padding*2) * scale)

	imgui.BeginChildV("map", imgui.Vec2{X: float32(canvasWidth), Y: float32(canvasHeight)}, false, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoMouseInputs)

	startX := padding
	startY := padding + (sm.Height * int(-yStep.Y))

	col := color.RGBA{0, 0, 0, 255}
	canvas.AddRectFilled(pos, pos.Add(image.Pt(canvasWidth, canvasHeight)), col, 0, 0)

	col = color.RGBA{255, 255, 255, 255}
	var drawables []archDrawable
	// Draw archetypes.
	for y := 0; y < sm.Height; y++ {
		if m.onionskin {
			// TODO: adjust alpha based upon distance of y from focusedY
			if y < m.focusedY {
				col.A = 200
			} else if y > m.focusedY {
				col.A = 50
			} else {
				col.A = 255
			}
		}
		xOffset := y * int(yStep.X)
		yOffset := y * int(-yStep.Y)
		for x := sm.Width - 1; x >= 0; x-- {
			for z := 0; z < sm.Depth; z++ {
				for t := 0; t < len(sm.Tiles[y][x][z]); t++ {
					oX := pos.X + (x*tWidth+xOffset+startX)*scale
					oY := pos.Y + (z*tHeight-yOffset+startY)*scale
					oH, _, oD := dm.GetArchDimensions(&sm.Tiles[y][x][z][t])
					if adjustment, ok := dm.AnimationsConfig.Adjustments[dm.GetArchType(&sm.Tiles[y][x][z][t], 0)]; ok {
						oX += int(adjustment.X) * scale
						oY += int(adjustment.Y) * scale
					}

					// calc render z
					indexZ := z
					indexX := x
					indexY := y
					zIndex := (indexZ * sm.Height * sm.Width) + (sm.Depth * indexY) - (indexX) + t

					anim, face := m.context.dataManager.GetAnimAndFace(&sm.Tiles[y][x][z][t], "", "")
					imageName, err := m.context.dataManager.GetAnimFaceImage(anim, face)
					if err != nil {
						continue
					}

					if tex, ok := m.context.imageTextures[imageName]; ok {
						if (oH > 1 || oD > 1) && int(tex.height*float32(scale)) > tHeight*scale {
							oY -= int(tex.height*float32(scale)) - (tHeight * scale)
						}
						drawables = append(drawables, archDrawable{
							z: zIndex,
							x: oX,
							y: oY,
							w: oX + int(tex.width)*scale,
							h: oY + int(tex.height)*scale,
							c: col,
							t: tex,
						})
					} else {
						//log.Println(err)
					}
				}
			}
		}
	}

	// Sort our drawables.
	sort.Slice(drawables, func(i, j int) bool {
		return drawables[i].z < drawables[j].z
	})
	// Render them.
	for _, d := range drawables {
		canvas.AddImageV(d.t.texture, image.Pt(d.x, d.y), image.Pt(d.w, d.h), image.Pt(0, 0), image.Pt(1, 1), d.c)
	}

	// Draw grid.
	if m.showGrid {
		for y := 0; y < sm.Height; y++ {
			xOffset := y * int(yStep.X)
			yOffset := y * int(-yStep.Y)
			col.A = 0
			if m.showYGrids {
				// TODO: fade out based upon distance from focusedY
				col.A = 15
			}
			if m.focusedY == y {
				col.A = 50
			}
			for x := 0; x < sm.Width; x++ {
				for z := 0; z < sm.Depth; z++ {
					oX := pos.X + (x*tWidth+xOffset+startX)*scale
					oY := pos.Y + (z*tHeight-yOffset+startY)*scale
					oW := (tWidth) * scale
					oH := (tHeight) * scale
					canvas.AddRect(image.Pt(oX, oY), image.Pt(oX+oW, oY+oH), col, 0, 0, 0.5)
				}
			}
		}
	}

	// Draw selected.
	{
		for yxz := range m.selectedCoords.Get() {
			y, x, z := yxz[0], yxz[1], yxz[2]
			xOffset := y * int(yStep.X)
			yOffset := y * int(-yStep.Y)
			oX := pos.X + (x*tWidth+xOffset+startX)*scale
			oY := pos.Y + (z*tHeight-yOffset+startY)*scale
			oW := (tWidth) * scale
			oH := (tHeight) * scale

			col = color.RGBA{255, 0, 255, 255}
			canvas.AddRect(image.Pt(oX, oY), image.Pt(oX+oW, oY+oH), col, 0, 0, 2)
		}
	}

	// Draw focused.
	{
		xOffset := m.focusedY * int(yStep.X)
		yOffset := m.focusedY * int(-yStep.Y)
		oX := pos.X + (m.focusedX*tWidth+xOffset+startX)*scale
		oY := pos.Y + (m.focusedZ*tHeight-yOffset+startY)*scale
		oW := (tWidth) * scale
		oH := (tHeight) * scale

		col = color.RGBA{255, 0, 0, 255}
		canvas.AddRect(image.Pt(oX, oY), image.Pt(oX+oW, oY+oH), col, 0, 0, 2)
	}

	imgui.EndChild()
}

func (m *Mapset) saveAll() {
	maps := make(map[string]*sdata.Map)
	for _, v := range m.maps {
		if v.Unsaved() {
			v.Save()
		}
		maps[v.DataName()] = v.SavedMap()
	}
	err := m.context.dataManager.SaveMap(m.filename, maps)
	if err != nil {
		m.unsaved = true
		log.Println(err)
		// TODO: Report error to the user.
		return
	}
	m.unsaved = false
	// TODO: Some sort of UI notification.
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

func (m *Mapset) insertArchetype(t *sdata.Map, arch string, y, x, z, pos int) error {
	tiles := m.getTiles(t, y, x, z)
	if tiles == nil {
		return errors.New("tile OOB")
	}

	if pos == -1 {
		pos = len(*tiles)
	}
	if pos == -1 {
		pos = 0
	}

	archetype := sdata.Archetype{
		Archs: []string{arch},
	}
	*tiles = append((*tiles)[:pos], append([]sdata.Archetype{archetype}, (*tiles)[pos:]...)...)

	return nil
}

func (m *Mapset) removeArchetype(t *sdata.Map, y, x, z, pos int) error {
	tiles := m.getTiles(t, y, x, z)
	if tiles == nil {
		return errors.New("tile OOB")
	}
	if pos == -1 {
		pos = len(*tiles) - 1
	}
	if pos == -1 {
		pos = 0
	}

	if pos >= len(*tiles) {
		return errors.New("pos OOB")
	}

	*tiles = append((*tiles)[:pos], (*tiles)[pos+1:]...)

	return nil
}

func (m *Mapset) getTiles(t *sdata.Map, y, x, z int) *[]sdata.Archetype {
	if len(t.Tiles) > y && y >= 0 {
		if len(t.Tiles[y]) > x && x >= 0 {
			if len(t.Tiles[y][x]) > z && z >= 0 {
				return &t.Tiles[y][x][z]
			}
		}
	}
	return nil
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
	return m.maps[m.currentMapIndex]
}

func (m *Mapset) Unsaved() bool {
	if m.unsaved {
		return true
	}
	for _, v := range m.maps {
		if v.Unsaved() {
			return true
		}
	}
	return false
}
