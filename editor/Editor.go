package editor

import (
	"os"

	"path"

	g "github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/chimera-rpg/go-editor/data"
	log "github.com/sirupsen/logrus"
)

type Editor struct {
	masterWindow      *g.MasterWindow
	dataManager       *data.Manager
	archetypesMode    bool
	selectedArchetype string
	isRunning         bool
	showSplash        bool
	mapTexture        *g.Texture
	mapTextureW       float32
	mapTextureH       float32
	mapsMap           map[string]*Maps
}

func (e *Editor) Setup(dataManager *data.Manager) (err error) {
	e.dataManager = dataManager
	e.isRunning = true
	e.archetypesMode = true
	e.mapsMap = make(map[string]*Maps)

	return
}

func (e *Editor) Destroy() {
	e.isRunning = false
}

func (e *Editor) Start() {
	log.Println("Editor: Start")
	e.masterWindow = g.NewMasterWindow("Editor", 800, 600, 0, nil)
	e.showSplash = true

	fullPath := path.Join(e.dataManager.MapsPath, "ChamberOfOrigins.map.yaml")
	dMaps, err := e.dataManager.LoadMap(fullPath)
	if err != nil {
		log.Errorln(err)
	} else {
		e.mapsMap["ChamberOfOrigins.map.yaml"] = NewMaps("ChamberOfOrigins.map.yaml", dMaps)
	}

	e.masterWindow.Main(func() { e.loop() })
}

func (e *Editor) loop() {
	if !e.isRunning {
		os.Exit(0)
	}

	g.MainMenuBar(g.Layout{
		g.Menu("File", g.Layout{
			g.MenuItem("Open Map", func() {
				imgui.OpenPopup("Open Map...")
			}),
			g.Separator(),
			g.MenuItem("Exit", func() { e.isRunning = false }),
		}),
		g.Menu("Misc", g.Layout{
			g.Button("Button", nil),
		}),
	}).Build()

	g.PopupModal("Open Map...", g.Layout{
		g.Label("Select a file"),
		g.Line(
			g.Button("Cancel", func() { imgui.CloseCurrentPopup() }),
			g.Button("Open", nil),
		),
	}).Build()

	for _, m := range e.mapsMap {
		m.draw(e.dataManager)
	}
	e.drawArchetypes()
	e.drawAnimations()
	e.drawSplash()

}

func (e *Editor) drawSplash() {
	if e.showSplash {
		g.WindowV("splash", &e.showSplash, g.WindowFlagsNoCollapse|g.WindowFlagsNoResize, 10, 30, 120, 120, g.Layout{
			g.Label("Chimera Editor"),
			g.Label("0.0.0"),
			g.Line(
				g.Button("Close", func() { e.showSplash = false }),
			),
		})
	}
}

func (e *Editor) drawArchetypes() {

	var items g.Layout
	if e.archetypesMode {
		archs := e.dataManager.GetArchetypes()
		for _, archName := range archs {
			var flags g.TreeNodeFlags
			flags = imgui.TreeNodeFlagsLeaf
			if archName == e.selectedArchetype {
				flags |= imgui.TreeNodeFlagsSelected
			}
			items = append(items, g.TreeNode(archName, flags, g.Layout{
				g.ContextMenu(g.Layout{
					g.Selectable("Edit", func(name string) func() { return func() { log.Printf("edit %s\n", name) } }(archName)),
				}),
				g.Custom(func(name string) func() {
					return func() {
						if g.IsItemHovered() && g.IsMouseClicked(g.MouseButtonLeft) {
							e.selectedArchetype = name
						}
					}
				}(archName)),
			}))
		}
	} else {
		archs := e.dataManager.GetArchetypeFiles()
		for _, archFile := range archs {
			var archItems []g.Widget
			for archName := range e.dataManager.GetArchetypeFile(archFile) {
				archItems = append(archItems, g.Layout{
					g.Label(archName),
					g.ContextMenu(g.Layout{
						g.Selectable("Argh", func() {}),
					}),
				})
			}
			items = append(items, g.Row(g.TreeNode(archFile, imgui.TreeNodeFlagsCollapsingHeader|imgui.TreeNodeFlagsDefaultOpen, archItems)))
		}
	}
	var b bool
	g.WindowV("Archetypes", &b, g.WindowFlagsMenuBar, 10, 30, 200, 400, g.Layout{
		g.MenuBar(g.Layout{
			g.Menu("File", g.Layout{
				g.MenuItem("New...", func() {}),
				g.Separator(),
			}),
			g.Menu("Misc", g.Layout{
				g.Checkbox("Archetype Mode", &e.archetypesMode, func() {}),
			}),
		}),
		items,
	})

}

func (e *Editor) drawAnimations() {
	var rows []*g.RowWidget
	rows = append(rows, g.Row(g.Label("anim")))
	var b bool
	g.WindowV("Animations", &b, g.WindowFlagsMenuBar, 10, 500, 200, 400, g.Layout{
		g.MenuBar(g.Layout{
			g.Menu("File", g.Layout{
				g.MenuItem("New...", func() {}),
				g.Separator(),
			}),
		}),
		g.FastTable("animations", true, rows),
	})
}
