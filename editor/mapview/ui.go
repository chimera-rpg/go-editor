package mapview

import (
	"fmt"
	"image"
	"image/color"
	"math"

	g "github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/chimera-rpg/go-editor/data"
	"github.com/chimera-rpg/go-editor/widgets"
	sdata "github.com/chimera-rpg/go-server/data"
	log "github.com/sirupsen/logrus"
)

var focusedBorderColor = color.RGBA{255, 0, 0, 128}
var focusedBackgroundColor = color.RGBA{255, 0, 0, 100}
var selectedBackgroundColor = color.RGBA{255, 255, 32, 100}
var selectingBackgroundColor = color.RGBA{255, 255, 32, 50}

func (m *Mapset) Draw() {
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
				g.Checkbox("Only Visit Unique Tiles", &m.uniqueTileVisits, nil),
			}),
			g.Menu("View", g.Layout{
				g.Checkbox("Z Onionskinning", &m.onionskinZ, nil),
				g.Checkbox("Y Onionskinning", &m.onionskinY, nil),
				g.Checkbox("X Onionskinning", &m.onionskinX, nil),
				g.SliderInt("Onionskin > Opacity", &m.onionSkinGtIntensity, 0, 255, "%d"),
				g.SliderInt("Onionskin < Opacity", &m.onionSkinLtIntensity, 0, 255, "%d"),
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
			imgui.SelectableV(fmt.Sprintf("fill (%s)", m.getToolButtonString(fillTool)), m.isToolBound(fillTool), 0, imgui.Vec2{X: toolWidth, Y: 0})
			if g.IsItemHovered() {
				if g.IsMouseClicked(g.MouseButtonLeft) {
					m.bindMouseToTool(g.MouseButtonLeft, fillTool)
				} else if g.IsMouseClicked(g.MouseButtonMiddle) {
					m.bindMouseToTool(g.MouseButtonMiddle, fillTool)
				} else if g.IsMouseClicked(g.MouseButtonRight) {
					m.bindMouseToTool(g.MouseButtonRight, fillTool)
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
					m.maps = append(m.maps, data.NewUnReMap(newMap, m.newDataName))
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

					cm.SetDataName(m.newDataName)

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
					m.toolInsert(3, cm, m.focusedY, m.focusedX, m.focusedZ)
				}
			}),
			widgets.KeyBind(widgets.KeyBindFlagPressed, widgets.Keys(widgets.KeyControl), widgets.Keys(widgets.KeyF), func() {
				if cm := m.CurrentMap(); cm != nil {
					m.toolFill(3, cm, m.focusedY, m.focusedX, m.focusedZ)
				}
			}),
			widgets.KeyBind(widgets.KeyBindFlagPressed, widgets.Keys(), widgets.Keys(widgets.KeyD), func() {
				if cm := m.CurrentMap(); cm != nil {
					m.toolErase(3, cm, m.focusedY, m.focusedX, m.focusedZ)
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
		func(mapIndex int, v *data.UnReMap) {
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
		}(mapIndex, v)
	}
	return g.Layout{g.TabBarV("Mapset", g.TabBarFlagsFittingPolicyScroll|g.TabBarFlagsFittingPolicyResizeDown, tabs)}
}

func (m *Mapset) layoutMapView(v *data.UnReMap) g.Layout {
	var availW, availH float32
	childPos := image.Point{0, 0}
	childFlags := g.WindowFlagsHorizontalScrollbar | imgui.WindowFlagsNoMove
	if m.blockScroll {
		childFlags |= imgui.WindowFlagsNoScrollWithMouse
	}
	hovered := false

	return g.Layout{
		g.Custom(func() {
			availW, availH = g.GetAvaiableRegion()
		}),
		g.Child(v.Get().Name, false, availW, availH, childFlags, g.Layout{
			g.Custom(func() {
				childPos = g.GetCursorScreenPos()
				m.drawMap(v)
			}),
			g.Custom(func() {
				if g.IsItemHovered() {
					hovered = true
					mousePos := g.GetMousePos()
					mousePos.X -= childPos.X
					mousePos.Y -= childPos.Y

					p, err := m.getMapPointFromMouse(mousePos)
					if err != nil {
						//log.Errorln(err)
						return
					}

					var state ButtonState
					// RMB
					if g.IsMouseDown(g.MouseButtonRight) {
						state = 2
						if _, ok := m.mouseHeld[g.MouseButtonRight]; !ok {
							m.mouseHeld[g.MouseButtonRight] = true
							state = 1
						}
						err := m.handleMouseTool(g.MouseButtonRight, state, m.focusedY, p.X, p.Y)
						if err != nil {
							log.Errorln(err)
						}
					} else if g.IsMouseReleased(g.MouseButtonRight) {
						state = 0
						err := m.handleMouseTool(g.MouseButtonRight, state, m.focusedY, p.X, p.Y)
						if err != nil {
							log.Errorln(err)
						}

						delete(m.mouseHeld, g.MouseButtonRight)
					}
					// MMB
					if g.IsMouseDown(g.MouseButtonMiddle) {
						state = 2
						if _, ok := m.mouseHeld[g.MouseButtonMiddle]; !ok {
							m.mouseHeld[g.MouseButtonMiddle] = true
							state = 1
						}
						err := m.handleMouseTool(g.MouseButtonMiddle, state, m.focusedY, p.X, p.Y)
						if err != nil {
							log.Errorln(err)
						}
					} else if g.IsMouseReleased(g.MouseButtonMiddle) {
						state = 0
						err := m.handleMouseTool(g.MouseButtonMiddle, state, m.focusedY, p.X, p.Y)
						if err != nil {
							log.Errorln(err)
						}
						delete(m.mouseHeld, g.MouseButtonMiddle)
					}
					// LMB
					if g.IsMouseDown(g.MouseButtonLeft) {
						state = 2
						if _, ok := m.mouseHeld[g.MouseButtonLeft]; !ok {
							m.mouseHeld[g.MouseButtonLeft] = true
							state = 1
						}
						err := m.handleMouseTool(g.MouseButtonLeft, state, m.focusedY, p.X, p.Y)
						if err != nil {
							log.Errorln(err)
						}
					} else if g.IsMouseReleased(g.MouseButtonLeft) {
						state = 0
						err := m.handleMouseTool(g.MouseButtonLeft, state, m.focusedY, p.X, p.Y)
						if err != nil {
							log.Errorln(err)
						}
						delete(m.mouseHeld, g.MouseButtonLeft)
					}
				}
			}),
			g.Custom(func() {
				if hovered {
					widgets.KeyBinds(0,
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
					).Build()
				}
			}),
		}),
	}
}

func (m *Mapset) layoutArchsList(v *data.UnReMap) g.Layout {
	dm := m.context.DataManager()
	sm := m.CurrentMap()

	var yItems g.Layout
	// Collect the entire Y "stack" as separate lists.
	for y := sm.Get().Height - 1; y >= 0; y-- {
		var items g.Layout
		func(y int) {
			archs := sm.GetArchs(y, m.focusedX, m.focusedZ)
			if len(archs) > 0 {
				for i := range archs {
					index := len(archs) - 1 - i
					arch := archs[index]
					func(index int, arch sdata.Archetype) {
						archName := dm.GetArchName(&arch, "")
						var flags g.TreeNodeFlags
						flags = g.TreeNodeFlagsLeaf | g.TreeNodeFlagsSpanFullWidth
						if index == m.focusedI && m.focusedY == y {
							flags |= g.TreeNodeFlagsSelected
						}
						items = append(items, g.TreeNode("", flags, g.Layout{
							g.Custom(func() {
								if g.IsItemHovered() {
									if g.IsMouseDoubleClicked(g.MouseButtonLeft) {
										//e.openArchsetFromArchetype(name)
									} else if g.IsMouseClicked(g.MouseButtonLeft) {
										m.focusedI = index
										m.focusedY = y
									}
								}
							}),
							g.Custom(func() {
								anim, face := dm.GetAnimAndFace(&arch, "", "")
								imageName, err := dm.GetAnimFaceImage(anim, face)
								if err == nil {
									if tex, ok := m.context.ImageTextures()[imageName]; ok {
										g.SameLine()
										if tex.Texture != nil {
											g.Image(tex.Texture, tex.Width, tex.Height).Build()
										}
										return
									}
								}
								g.SameLine()
								g.Dummy(float32(dm.AnimationsConfig.TileWidth), float32(dm.AnimationsConfig.TileHeight))
							}),
							g.Custom(func() {
								//imgui.PushStyleColor(imgui.StyleColorText, g.ToVec4Color(color.RGBA{255, 0, 0, 255}))
								g.SameLine()
							}),
							g.Label(archName),
							g.Custom(func() {
								//imgui.PopStyleColorV(1)
							}),
						}))
					}(index, arch)
				}
			} else {
				var flags g.TreeNodeFlags
				flags = g.TreeNodeFlagsLeaf | g.TreeNodeFlagsSpanFullWidth
				items = append(items, g.TreeNode("", flags, g.Layout{
					g.Custom(func() {
						g.SameLine()
						g.Dummy(float32(dm.AnimationsConfig.TileWidth), float32(dm.AnimationsConfig.TileHeight))
					}),
					g.Custom(func() { g.SameLine() }),
					g.Label("-"),
				}))
			}
		}(y)

		flags := g.TreeNodeFlagsDefaultOpen | g.TreeNodeFlagsSpanFullWidth | g.TreeNodeFlagsOpenOnArrow | g.TreeNodeFlagsOpenOnDoubleClick
		if y == m.focusedY {
			yItems = append(yItems, g.Custom(func() {
				//imgui.PushStyleColor(imgui.StyleColorText, g.ToVec4Color(color.RGBA{32, 128, 255, 255}))
			}))
			flags |= g.TreeNodeFlagsSelected
		}
		yItems = append(yItems, g.TreeNode(fmt.Sprintf("%d", y), flags, items))
		if y == m.focusedY {
			yItems = append(yItems, g.Custom(func() {
				//imgui.PopStyleColor()
			}))
		}
	}
	return yItems
}

func (m *Mapset) layoutSelectedArch(v *data.UnReMap) g.Layout {
	sm := m.CurrentMap()
	dm := m.context.DataManager()

	archs := sm.GetArchs(m.focusedY, m.focusedX, m.focusedZ)
	if m.focusedI >= 0 && m.focusedI < len(archs) {
		arch := archs[m.focusedI]
		archName := dm.GetArchName(&arch, "")
		return g.Layout{
			g.Label(archName),
		}
	}
	return g.Layout{
		g.Label("no archetype selected"),
	}
}
