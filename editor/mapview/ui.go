package mapview

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"path"

	g "github.com/AllenDang/giu"
	imgui "github.com/AllenDang/imgui-go"
	"github.com/chimera-rpg/go-editor/data"
	"github.com/chimera-rpg/go-editor/editor/icons"
	"github.com/chimera-rpg/go-editor/widgets"
	sdata "github.com/chimera-rpg/go-server/data"
	log "github.com/sirupsen/logrus"
)

var focusedBorderColor = color.RGBA{255, 0, 0, 128}
var focusedBackgroundColor = color.RGBA{255, 0, 0, 100}
var selectedBackgroundColor = color.RGBA{128, 128, 128, 128}
var selectingBackgroundColor = color.RGBA{128, 128, 128, 64}
var hoveredBorderColor = color.RGBA{255, 255, 0, 128}
var hoveredBackgroundColor = color.RGBA{255, 255, 0, 0}

func (m *Mapset) Draw() (title string, w *g.WindowWidget, layout g.Layout) {
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
	//toolWidth, _ := g.CalcTextSize("_________")
	selectImage := "select"
	if m.isToolBound(selectTool) {
		selectImage += "-focus"
	}
	cselectImage := "cselect"
	if m.isToolBound(cselectTool) {
		cselectImage += "-focus"
	}
	lselectImage := "lselect"
	if m.isToolBound(lselectTool) {
		lselectImage += "-focus"
	}
	fillImage := "fill"
	if m.isToolBound(fillTool) {
		fillImage += "-focus"
	}
	pickImage := "dropper"
	if m.isToolBound(pickTool) {
		pickImage += "-focus"
	}
	eraseImage := "eraser"
	if m.isToolBound(eraseTool) {
		eraseImage += "-focus"
	}
	insertImage := "insert"
	if m.isToolBound(insertTool) {
		insertImage += "-focus"
	}
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
	title = fmt.Sprintf("Mapset: %s", filename)
	w = g.Window(title)
	w.IsOpen(&windowOpen).Flags(windowFlags).Pos(210, 30).Size(300, 400)
	layout = g.Layout{g.MenuBar().Layout(
		g.Menu("Mapset").Layout(
			g.MenuItem("New Map...").OnClick(func() {
				newMapPopup = true
				m.descEditor.SetText("")
				m.loreEditor.SetText("")
			}),
			g.Separator(),
			g.MenuItem("Save All").OnClick(func() { m.saveAll() }),
			g.Separator(),
			g.MenuItem("Close").OnClick(func() { m.close() }),
		),
		g.Menu("Map").Layout(
			g.MenuItem("Properties...").Enabled(mapExists).OnClick(func() {
				cm := m.CurrentMap()
				m.newName = cm.Get().Name
				m.newDataName = cm.DataName()
				m.descEditor.SetText(cm.Get().Description)
				m.loreEditor.SetText(cm.Get().Lore)
				adjustMapPopup = true
			}),
			g.MenuItem("Resize...").Enabled(mapExists).OnClick(func() {
				resizeMapPopup = true
			}),
			g.Separator(),
			g.MenuItem("Undo").Enabled(mapExists).OnClick(func() {
				cm := m.CurrentMap()
				cm.Undo()
			}),
			g.MenuItem("Redo").Enabled(mapExists).OnClick(func() {
				cm := m.CurrentMap()
				cm.Redo()
			}),
			g.Separator(),
			g.MenuItem("Delete...").Enabled(mapExists).OnClick(func() {
				deleteMapPopup = true
			}),
		),
		g.Menu("Settings").Layout(
			g.Checkbox("Keep Same Tile", &m.keepSameTile),
			g.Checkbox("Only Visit Unique Tiles", &m.uniqueTileVisits),
		),
		g.Menu("View").Layout(
			g.Checkbox("Z Onionskinning", &m.onionskinZ),
			g.Checkbox("Y Onionskinning", &m.onionskinY),
			g.Checkbox("X Onionskinning", &m.onionskinX),
			g.SliderInt(&m.onionSkinGtIntensity, 0, 255).Label("Onionskin > Opacity").Format("%d"),
			g.SliderInt(&m.onionSkinLtIntensity, 0, 255).Label("Onionskin < Opacity").Format("%d"),
			g.Checkbox("Grid", &m.showGrid),
			g.Checkbox("Y Grids", &m.showYGrids),
			g.SliderInt(&m.zoom, 1, 8).Label("Zoom").Format("%d"),
		),
	),
		g.Row(
			g.Column(
				g.Row(
					g.ImageButton(icons.Textures[selectImage].Texture).Size(30, 30).FramePadding(0).OnClick(func() {
						m.bindMouseToTool(g.MouseButtonLeft, selectTool)
					}),
					g.Tooltip("selection tool"),
					g.ImageButton(icons.Textures[cselectImage].Texture).Size(30, 30).FramePadding(0).OnClick(func() {
						m.bindMouseToTool(g.MouseButtonLeft, cselectTool)
					}),
					g.Tooltip("circular selection tool"),
					g.ImageButton(icons.Textures[lselectImage].Texture).Size(30, 30).FramePadding(0).OnClick(func() {
						m.bindMouseToTool(g.MouseButtonLeft, lselectTool)
					}),
					g.Tooltip("line selection tool"),
				),
				g.Row(
					g.ImageButton(icons.Textures[insertImage].Texture).Size(30, 30).FramePadding(0).OnClick(func() {
						m.bindMouseToTool(g.MouseButtonLeft, insertTool)
					}),
					g.Tooltip("insertion tool"),
					g.ImageButton(icons.Textures[eraseImage].Texture).Size(30, 30).FramePadding(0).OnClick(func() {
						m.bindMouseToTool(g.MouseButtonLeft, eraseTool)
					}),
					g.Tooltip("erase tool"),
					g.ImageButton(icons.Textures[fillImage].Texture).Size(30, 30).FramePadding(0).OnClick(func() {
						m.bindMouseToTool(g.MouseButtonLeft, fillTool)
					}),
					g.Tooltip("fill tool"),
				),
				g.Row(
					g.ImageButton(icons.Textures[pickImage].Texture).Size(30, 30).FramePadding(0).OnClick(func() {
						m.bindMouseToTool(g.MouseButtonLeft, pickTool)
					}),
					g.Tooltip("pick from map tool"),
				),
				g.Row(
					g.Child().Size(105, g.Auto).Layout(
						g.Custom(func() {
							g.Label("TODO").Build()
						}),
					),
				),
			),
			g.Child().Layout(
				m.layoutMapTabs(),
			),
		),
		g.Custom(func() {
			if m.showSave {
				g.OpenPopup("Save Map")
			} else if resizeMapPopup {
				g.OpenPopup("Resize Map")
			} else if newMapPopup {
				g.OpenPopup("New Map")
			} else if adjustMapPopup {
				g.OpenPopup("Map Properties")
			} else if deleteMapPopup {
				g.OpenPopup("Delete Map")
			}
		}),
		g.PopupModal("Save Map").Layout(
			g.Label("Save mapset to file"),
			widgets.FileBrowser(&m.saveMapCWD, &m.saveMapFilename, nil),
			// Line
			g.Row(
				g.Button("Cancel").OnClick(func() {
					m.saveMapCWD = m.context.DataManager().MapsPath
					m.saveMapFilename = ""
					m.showSave = false
					g.CloseCurrentPopup()
				}),
				g.Button("Save").OnClick(func() {
					fullPath := path.Join(m.saveMapCWD, m.saveMapFilename)
					m.saveMapCWD = m.context.DataManager().MapsPath
					m.saveMapFilename = ""
					m.pendingFilename = fullPath
					m.showSave = false
					m.saveAll()
					g.CloseCurrentPopup()
				}),
			),
		),
		g.PopupModal("Resize Map").Layout(
			g.Label("Grow or Shrink the current map"),
			g.Row(
				g.InputInt(&m.resizeU).Size(50).Label("Up    "),
				g.InputInt(&m.resizeD).Size(50).Label("Down  "),
			),
			g.Row(
				g.InputInt(&m.resizeL).Size(50).Label("Left  "),
				g.InputInt(&m.resizeR).Size(50).Label("Right "),
			),
			g.Row(
				g.InputInt(&m.resizeT).Size(50).Label("Top   "),
				g.InputInt(&m.resizeB).Size(50).Label("Bottom"),
			),
			g.Row(
				g.Button("Resize").OnClick(func() {
					m.resizeMap(int(m.resizeU), int(m.resizeD), int(m.resizeL), int(m.resizeR), int(m.resizeT), int(m.resizeB))
					m.resizeU, m.resizeD, m.resizeL, m.resizeR, m.resizeT, m.resizeB = 0, 0, 0, 0, 0, 0
					g.CloseCurrentPopup()
				}),
				g.Button("Cancel").OnClick(func() {
					m.resizeU, m.resizeD, m.resizeL, m.resizeR, m.resizeT, m.resizeB = 0, 0, 0, 0, 0, 0
					g.CloseCurrentPopup()
				}),
			),
		),
		g.PopupModal("New Map").Flags(g.WindowFlagsHorizontalScrollbar).Layout(
			g.Label("Create a new map"),
			g.InputText(&m.newDataName).Label("Data Name"),
			g.InputText(&m.newName).Label("Name"),
			g.Custom(func() {
				availW, _ := g.GetAvailableRegion()
				labelV := imgui.CalcTextSize("Description", false, 0)
				m.descEditor.Render("Description", imgui.Vec2{X: availW - labelV.X - 5, Y: 200}, false)
				imgui.SameLine()
				g.Label("Description").Build()
				m.loreEditor.Render("Lore", imgui.Vec2{X: availW - labelV.X - 5, Y: 200}, false)
				imgui.SameLine()
				g.Label("Lore").Build()
			}),
			g.SliderInt(&m.newH, 1, 200).Label("Height").Format("%d"),
			g.SliderInt(&m.newW, 1, 200).Label("Width ").Format("%d"),
			g.SliderInt(&m.newD, 1, 200).Label("Depth ").Format("%d"),
			g.Row(
				g.Button("Create").OnClick(func() {
					g.CloseCurrentPopup()
					lore := m.loreEditor.GetText()
					desc := m.descEditor.GetText()
					// TODO: Check if map with same name already exists!
					newMap := m.createMap(m.newName, desc, lore, 0, 0, int(m.newH), int(m.newW), int(m.newD))
					m.maps = append(m.maps, data.NewUnReMap(newMap, m.newDataName))
					m.newName, m.newDataName = "", ""
				}),
				g.Button("Cancel").OnClick(func() {
					g.CloseCurrentPopup()
					m.newName, m.newDataName = "", ""
				}),
			),
		),
		g.PopupModal("Map Properties").Flags(g.WindowFlagsHorizontalScrollbar).Layout(
			g.InputText(&m.newDataName).Label("Data Name"),
			g.InputText(&m.newName).Label("Name"),
			g.Custom(func() {
				availW, availH := g.GetAvailableRegion()
				labelV := imgui.CalcTextSize("Description", false, 0)
				m.descEditor.Render("Description", imgui.Vec2{X: availW - labelV.X - 5, Y: availH/2 - labelV.Y - 3}, false)
				imgui.SameLine()
				g.Label("Description").Build()
				m.loreEditor.Render("Lore", imgui.Vec2{X: availW - labelV.X - 5, Y: availH/2 - labelV.Y - 3}, false)
				imgui.SameLine()
				g.Label("Lore").Build()
			}),
			g.Row(
				g.Button("Save").OnClick(func() {
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
				g.Button("Cancel").OnClick(func() {
					g.CloseCurrentPopup()
					m.newName, m.newDataName = "", ""
				}),
			),
		),
		g.PopupModal("Delete Map").Layout(
			g.Label("Delete map?"),
			g.Label("This cannot be recovered."),
			g.Row(
				g.Button("Delete").OnClick(func() {
					m.deleteMap(m.currentMapIndex)
					g.CloseCurrentPopup()
				}),
				g.Button("Cancel").OnClick(func() {
					g.CloseCurrentPopup()
				}),
			),
		),
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
		g.Custom(func() {
			if !windowOpen {
				m.close()
			}
		}),
	}
	return title, w, layout
}

func (m *Mapset) layoutMapTabs() g.Layout {
	var tabs []*g.TabItemWidget
	for mapIndex, v := range m.maps {
		func(mapIndex int, v *data.UnReMap) {
			var flags g.TabItemFlags
			if v.Unsaved() {
				flags |= g.TabItemFlagsUnsavedDocument
			}
			tab := g.TabItem(fmt.Sprintf("%s(%s)", v.DataName(), v.Get().Name)).Flags(flags).Layout(
				g.Custom(func() {
					m.currentMapIndex = mapIndex
					availW, availH := g.GetAvailableRegion()
					defaultW := float32(math.Round(float64(availW - availW/4)))
					defaultH := float32(math.Round(float64(availH - availH/4)))
					g.SplitLayout(g.DirectionVertical, true, defaultH, g.Layout{
						g.SplitLayout(g.DirectionHorizontal, true, defaultW,
							m.layoutMapView(v),
							m.layoutArchsList(v),
						),
					}, m.layoutSelectedArch(v)).Build()
				}),
			)

			tabs = append(tabs, tab)
		}(mapIndex, v)
	}
	return g.Layout{g.TabBar().Flags(g.TabBarFlagsFittingPolicyScroll | g.TabBarFlagsFittingPolicyResizeDown).TabItems(tabs...)}
}

func (m *Mapset) layoutMapView(v *data.UnReMap) g.Layout {
	var availW, availH float32
	childPos := image.Point{0, 0}
	childFlags := g.WindowFlagsHorizontalScrollbar | imgui.WindowFlagsNoMove
	if m.blockScroll {
		childFlags |= imgui.WindowFlagsNoScrollWithMouse
	}
	hovered := false
	lineHeight := imgui.CalcTextSize("Toolbar", false, 0)
	var canvasWidth, canvasHeight float32

	return g.Layout{
		g.Custom(func() {
			availW, availH = g.GetAvailableRegion()
			g.Child().Border(false).Flags(childFlags).Size(availW, availH-lineHeight.Y*2).Layout(
				g.Custom(func() {
					childPos = g.GetCursorScreenPos()
					canvasWidth, canvasHeight = m.getMapSize(v)
					g.Child().Border(false).Flags(g.WindowFlagsNoMouseInputs|g.WindowFlagsNoMove).Size(canvasWidth, canvasHeight).Layout(
						g.Custom(func() {
							m.drawMap(v)
						}),
					).Build()
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

						m.hoveredY = m.focusedY
						m.hoveredX = p.X
						m.hoveredZ = p.Y

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
			).Build()
		}),
		m.layoutMapInfobar(v),
	}
}

func (m *Mapset) layoutMapInfobar(v *data.UnReMap) g.Layout {
	dm := m.context.DataManager()
	tiles := m.getTiles(v.Get(), m.hoveredY, m.hoveredX, m.hoveredZ)
	var hoveredArch sdata.Archetype
	hoveredHasMore := false
	if len(*tiles) > 0 {
		hoveredArch = (*tiles)[len(*tiles)-1]
		if len(*tiles) > 1 {
			hoveredHasMore = true
		}
	}
	hoveredArchName := dm.GetArchName(&hoveredArch, "")
	if hoveredHasMore {
		hoveredArchName += "*"
	}

	tiles = m.getTiles(v.Get(), m.focusedY, m.focusedX, m.focusedZ)
	var focusedArch sdata.Archetype
	focusedHasMore := false
	if len(*tiles) > 0 {
		focusedArch = (*tiles)[len(*tiles)-1]
		if len(*tiles) > 1 {
			focusedHasMore = true
		}
	}
	focusedArchName := dm.GetArchName(&focusedArch, "")
	if focusedHasMore {
		focusedArchName += "*"
	}

	return g.Layout{
		g.Row(
			g.Label(fmt.Sprintf("%dx%dx%d", m.focusedX, m.focusedZ, m.focusedY)),
			g.Custom(func() {
				anim, face := dm.GetAnimAndFace(&focusedArch, "", "")
				imageName, err := dm.GetAnimFaceImage(anim, face)
				if err == nil {
					if tex, ok := m.context.ImageTextures()[imageName]; ok {
						g.SameLine()
						if tex.Texture != nil {
							g.Image(tex.Texture).Size(tex.Width, tex.Height).Build()
						}
						return
					}
				}
				g.SameLine()
				g.Dummy(float32(dm.AnimationsConfig.TileWidth), float32(dm.AnimationsConfig.TileHeight))
			}),
			g.Label(focusedArchName),
			g.Label(fmt.Sprintf("(%dx%dx%d)", m.hoveredX, m.hoveredZ, m.hoveredY)),
			g.Custom(func() {
				anim, face := dm.GetAnimAndFace(&hoveredArch, "", "")
				imageName, err := dm.GetAnimFaceImage(anim, face)
				if err == nil {
					if tex, ok := m.context.ImageTextures()[imageName]; ok {
						g.SameLine()
						if tex.Texture != nil {
							g.Image(tex.Texture).Size(tex.Width, tex.Height).Build()
						}
						return
					}
				}
				g.SameLine()
				g.Dummy(float32(dm.AnimationsConfig.TileWidth), float32(dm.AnimationsConfig.TileHeight))
			}),
			g.Label(hoveredArchName),
		),
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
						items = append(items, g.TreeNode("").Flags(flags).Layout(
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
											g.Image(tex.Texture).Size(tex.Width, tex.Height).Build()
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
						))
					}(index, arch)
				}
			} else {
				var flags g.TreeNodeFlags
				flags = g.TreeNodeFlagsLeaf | g.TreeNodeFlagsSpanFullWidth
				items = append(items, g.TreeNode("").Flags(flags).Layout(
					g.Custom(func() {
						g.SameLine()
						g.Dummy(float32(dm.AnimationsConfig.TileWidth), float32(dm.AnimationsConfig.TileHeight))
					}),
					g.Custom(func() { g.SameLine() }),
					g.Label("-"),
				))
			}
		}(y)

		flags := g.TreeNodeFlagsDefaultOpen | g.TreeNodeFlagsSpanFullWidth | g.TreeNodeFlagsOpenOnArrow | g.TreeNodeFlagsOpenOnDoubleClick
		if y == m.focusedY {
			yItems = append(yItems, g.Custom(func() {
				//imgui.PushStyleColor(imgui.StyleColorText, g.ToVec4Color(color.RGBA{32, 128, 255, 255}))
			}))
			flags |= g.TreeNodeFlagsSelected
		}
		yItems = append(yItems, g.TreeNode(fmt.Sprintf("%d", y)).Flags(flags).Layout(items))
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

	archs := sm.GetArchs(m.focusedY, m.focusedX, m.focusedZ)
	if m.focusedI >= 0 && m.focusedI < len(archs) {
		m.archEditor.SetArchetype(&archs[m.focusedI])
		return m.archEditor.Layout()
	}
	return g.Layout{
		g.Label("no archetype selected"),
	}
}
