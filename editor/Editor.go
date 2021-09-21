package editor

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"os"
	"path"

	g "github.com/AllenDang/giu"
	imgui "github.com/AllenDang/imgui-go"
	"github.com/chimera-rpg/go-editor/data"
	"github.com/chimera-rpg/go-editor/editor/mapview"
	"github.com/chimera-rpg/go-editor/widgets"
	sdata "github.com/chimera-rpg/go-server/data"
	log "github.com/sirupsen/logrus"
)

type Editor struct {
	masterWindow   *g.MasterWindow
	archetypesMode bool
	isLoaded       bool
	isRunning      bool
	showSplash     bool
	mapsets        []*mapview.Mapset
	archsets       []*Archset
	animsets       []*Animset
	context        Context
	//
	pendingImages map[string]image.Image
	//
	openMapCWD, openMapFilename string
}

func (e *Editor) Setup(dataManager *data.Manager) (err error) {
	e.context = Context{
		dataManager:   dataManager,
		imageTextures: make(map[string]*data.ImageTexture),
	}
	e.isLoaded = false
	e.isRunning = true
	e.archetypesMode = true
	e.showSplash = false
	e.openMapCWD = dataManager.MapsPath

	e.masterWindow = g.NewMasterWindow("Editor", 1280, 720, g.MasterWindowFlagsMaximized)
	g.Context.GetRenderer().SetTextureMagFilter(g.TextureFilterNearest)
	imgui.CurrentIO().SetIniFilename(e.context.dataManager.GetEtcPath("chimera-editor.ini"))

	for imagePath, img := range dataManager.GetImages() {
		e.context.imageTextures[imagePath] = &data.ImageTexture{
			Texture: nil,
			Width:   float32(img.Bounds().Max.X),
			Height:  float32(img.Bounds().Max.Y),
		}
	}

	e.pendingImages = dataManager.GetImages()

	return nil
}

func (e *Editor) Destroy() {
	e.isRunning = false
}

func (e *Editor) Start() {
	log.Println("Editor: Start")

	e.masterWindow.Run(func() { e.loop() })
}

func (e *Editor) loop() {
	if !e.isRunning {
		os.Exit(0)
	}

	var openMapPopup bool

	if !e.isLoaded {
		g.OpenPopup("Loading...")
		g.PopupModal("Loading...").Flags(g.WindowFlagsNoResize).Layout(
			g.Label("Now loading files..."),
		).Build()

		for imagePath, img := range e.pendingImages {
			go func(imagePath string, img image.Image) {
				rgba := image.NewRGBA(img.Bounds())
				draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)
				g.NewTextureFromRgba(rgba, func(tex *g.Texture) {
					if tex == nil {
						log.Fatalln("couldn't load image")
					}
					if it, ok := e.context.imageTextures[imagePath]; ok {
						it.Texture = tex
					}
				})
			}(imagePath, img)
		}
		e.isLoaded = true

		return
	}

	g.MainMenuBar().Layout(
		g.Menu("File").Layout(
			g.MenuItem("New Mapset").OnClick(func() {
				e.mapsets = append(e.mapsets, mapview.NewMapset(&e.context, "", nil))
			}),
			g.MenuItem("Open Mapset...").OnClick(func() {
				openMapPopup = true
			}),
			g.Separator(),
			g.MenuItem("Exit").OnClick(func() { e.isRunning = false }),
		),
		g.Menu("Misc").Layout(
			g.Button("Button"),
		),
		widgets.KeyBinds(0,
			widgets.KeyBind(widgets.KeyBindFlagPressed, widgets.Keys(widgets.KeyControl), widgets.Keys(widgets.KeyO), func() {
				openMapPopup = true
			}),
		),
	).Build()

	if openMapPopup {
		g.OpenPopup("Open Mapset...")
	}

	g.PopupModal("Open Mapset...").Flags(g.WindowFlagsNoResize).Layout(
		g.Label("Select a maps file"),
		widgets.FileBrowser(&e.openMapCWD, &e.openMapFilename, nil),
		g.Row(
			g.Button("Cancel").OnClick(func() { g.CloseCurrentPopup() }),
			g.Button("Open").OnClick(func() {
				fullPath := path.Join(e.openMapCWD, e.openMapFilename)
				dMapset, err := e.context.dataManager.LoadMap(fullPath)
				if err != nil {
					// TODO: Popup some sort of error!
					log.Errorln(err)
				} else {
					e.mapsets = append(e.mapsets, mapview.NewMapset(&e.context, fullPath, dMapset))
				}
				g.CloseCurrentPopup()
			}),
		),
	).Build()

	for i, m := range e.mapsets {
		m.Draw()
		if m.ShouldClose {
			e.mapsets = append(e.mapsets[:i], e.mapsets[i+1:]...)
		}
	}

	for i, a := range e.archsets {
		a.draw()
		if a.shouldClose {
			e.archsets = append(e.archsets[:i], e.archsets[i+1:]...)
		}
	}

	for _, a := range e.animsets {
		a.draw()
	}

	e.drawArchetypes()
	e.drawAnimations()
	e.drawSplash()

}

func (e *Editor) drawSplash() {
	if e.showSplash {
		g.Window("splash").IsOpen(&e.showSplash).Flags(g.WindowFlagsNoCollapse|g.WindowFlagsNoResize).Pos(10, 30).Size(120, 120).Layout(
			g.Label("Chimera Editor"),
			g.Label("0.0.0"),
			g.Row(
				g.Button("Close").OnClick(func() { e.showSplash = false }),
			),
		)
	}
}

func (e *Editor) drawArchetypeTreeNode(node data.ArchetypeTreeNode, parent string, isFirst bool) g.Layout {
	var items g.Layout

	var flags g.TreeNodeFlags
	flags = g.TreeNodeFlagsSpanFullWidth
	var archName string
	if isFirst {
		flags |= g.TreeNodeFlagsLeaf | g.TreeNodeFlagsNoTreePushOnOpen
	}
	if len(parent) == 0 {
		archName = node.Name
	} else {
		archName = parent + "/" + node.Name
	}

	if node.IsTree {
		var childItems g.Layout
		for _, child := range node.Children {
			childItems = append(childItems, e.drawArchetypeTreeNode(child, archName, false))
		}
		items = append(items, g.TreeNode(node.Name).Flags(flags).Layout(childItems))
	} else {
		flags |= g.TreeNodeFlagsLeaf
		if archName == e.context.selectedArch {
			flags |= g.TreeNodeFlagsSelected
		}
		items = append(items, g.TreeNode("").Flags(flags).Layout(
			g.Custom(func(name string) func() {
				return func() {
					if g.IsItemHovered() {
						if g.IsMouseDoubleClicked(g.MouseButtonLeft) {
							e.openArchsetFromArchetype(name)
						} else if g.IsMouseClicked(g.MouseButtonLeft) {
							e.context.selectedArch = name
							log.Println(name)
						}
					}
				}
			}(archName)),
			g.Custom(func(archName string) func() {
				return func() {
					arch := e.context.dataManager.GetArchetype(archName)
					if arch == nil {
						return
					}
					anim, face := e.context.dataManager.GetAnimAndFace(arch, "", "")
					imageName, err := e.context.dataManager.GetAnimFaceImage(anim, face)
					if err != nil {
						return
					}
					//e.context.imageTexturesLock.Lock()
					t, ok := e.context.imageTextures[imageName]
					//e.context.imageTexturesLock.Unlock()
					if ok {
						g.SameLine()
						fmt.Printf("ok %s %+v\n", imageName, t)
						if t.Texture != nil {
							g.Image(t.Texture).Size(t.Width, t.Height).Build()
							fmt.Println("Wow, got stuff")
						}
					}
				}
			}(archName)),
			g.Custom(func() { g.SameLine() }),
			g.Label(node.Name),
		))
	}
	return items
}

func (e *Editor) drawArchetypes() {
	var items g.Layout
	if e.archetypesMode {
		items = e.drawArchetypeTreeNode(e.context.dataManager.GetArchetypesAsTree(), "", true)
	} else {
		archs := e.context.dataManager.GetArchetypeFiles()
		for _, archFile := range archs {
			var archItems []g.Widget
			for _, archName := range e.context.dataManager.GetArchetypeFile(archFile) {
				archItems = append(archItems, g.Layout{
					g.Label(archName),
					g.ContextMenu().Layout(
						g.Selectable("Argh").OnClick(func() {}),
					),
				})
			}
			items = append(items, g.Row(g.TreeNode(archFile).Flags(g.TreeNodeFlagsCollapsingHeader|g.TreeNodeFlagsDefaultOpen).Layout(archItems...)))
		}
	}
	var b bool
	g.Window("Archetypes").IsOpen(&b).Flags(g.WindowFlagsMenuBar).Pos(10, 30).Size(200, 400).Layout(
		g.MenuBar().Layout(
			g.Menu("File").Layout(
				g.MenuItem("New...").OnClick(func() {}),
				g.Separator(),
			),
			g.Menu("Misc").Layout(
				g.Checkbox("Archetype Mode", &e.archetypesMode),
				g.Button("Reload Archetypes").OnClick(func() {
					e.context.dataManager.ReloadArchetypes()
				}),
				g.Button("Reload Animations").OnClick(func() {
					e.context.dataManager.ReloadAnimations()
				}),
			),
		),
		items,
	)

}

func (e *Editor) drawAnimations() {
	var rows []*g.RowWidget
	rows = append(rows, g.Row(g.Label("anim")))
	var b bool
	g.Window("Animations").IsOpen(&b).Flags(g.WindowFlagsMenuBar).Pos(10, 500).Size(200, 400).Layout(
		g.MenuBar().Layout(
			g.Menu("File").Layout(
				g.MenuItem("New...").OnClick(func() {}),
				g.Separator(),
			),
		),
		//g.FastTable("animations", true, rows),
	)
}

func (e *Editor) openArchsetFromArchetype(archName string) {
	archFilename := e.context.dataManager.LookupArchetypeFile(archName)
	if archFilename == "" {
		log.Errorln(errors.New("No archetype file for arch"))
		return
	}
	archFile := e.context.dataManager.GetArchetypeFile(archFilename)
	if archFile == nil {
		log.Errorln(errors.New("Missing archetype file"))
		return
	}
	for _, a := range e.archsets {
		if a.filename == archFilename {
			log.Printf("Archset %s already open.", archFilename)
			return
		}
	}
	archMap := make(map[string]*sdata.Archetype)
	for _, archName := range archFile {
		arch := e.context.dataManager.GetArchetype(archName)
		if arch != nil {
			archMap[archName] = arch
		}
	}
	e.archsets = append(e.archsets, NewArchset(&e.context, archFilename, archMap))
}
