package editor

import (
	"image"
	"image/draw"
	"os"

	"path"

	g "github.com/AllenDang/giu"
	"github.com/chimera-rpg/go-editor/data"
	"github.com/chimera-rpg/go-editor/widgets"
	log "github.com/sirupsen/logrus"
)

type Editor struct {
	masterWindow      *g.MasterWindow
	dataManager       *data.Manager
	archetypesMode    bool
	selectedArchetype string
	isRunning         bool
	showSplash        bool
	mapsets           []*Mapset
	archsets          []*Archset
	animsets          []*Animset
	imageTextures     map[string]ImageTexture
	//
	openMapCWD, openMapFilename string
}

type ImageTexture struct {
	texture       *g.Texture
	width, height float32
}

func (e *Editor) Setup(dataManager *data.Manager) (err error) {
	e.dataManager = dataManager
	e.isRunning = true
	e.archetypesMode = true
	e.imageTextures = make(map[string]ImageTexture)
	e.openMapCWD = dataManager.MapsPath

	return
}

func (e *Editor) Destroy() {
	e.isRunning = false
}

func (e *Editor) Start() {
	log.Println("Editor: Start")
	e.masterWindow = g.NewMasterWindow("Editor", 800, 600, 0, nil)
	e.showSplash = true

	e.masterWindow.Main(func() { e.loop() })
}

func (e *Editor) loop() {
	if !e.isRunning {
		os.Exit(0)
	}

	var openMapPopup bool

	g.MainMenuBar(g.Layout{
		g.Menu("File", g.Layout{
			g.MenuItem("New Mapset", func() {
				e.mapsets = append(e.mapsets, NewMapset("", nil))
			}),
			g.MenuItem("Open Mapset...", func() {
				openMapPopup = true
			}),
			g.Separator(),
			g.MenuItem("Exit", func() { e.isRunning = false }),
		}),
		g.Menu("Misc", g.Layout{
			g.Button("Button", nil),
		}),
	}).Build()

	if openMapPopup {
		g.OpenPopup("Open Mapset...")
	}

	g.PopupModalV("Open Mapset...", nil, g.WindowFlagsNoResize, g.Layout{
		g.Label("Select a maps file"),
		widgets.FileBrowser(&e.openMapCWD, &e.openMapFilename, nil),
		g.Line(
			g.Button("Cancel", func() { g.CloseCurrentPopup() }),
			g.Button("Open", func() {
				fullPath := path.Join(e.openMapCWD, e.openMapFilename)
				dMapset, err := e.dataManager.LoadMap(fullPath)
				if err != nil {
					// TODO: Popup some sort of error!
					log.Errorln(err)
				} else {
					e.mapsets = append(e.mapsets, NewMapset(fullPath, dMapset))
				}
				g.CloseCurrentPopup()
			}),
		),
	}).Build()

	for _, m := range e.mapsets {
		m.draw(e.dataManager)
	}

	for _, a := range e.archsets {
		a.draw(e.dataManager)
	}

	for _, a := range e.animsets {
		a.draw(e.dataManager)
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
			flags = g.TreeNodeFlagsLeaf | g.TreeNodeFlagsSpanFullWidth
			if archName == e.selectedArchetype {
				flags |= g.TreeNodeFlagsSelected
			}
			items = append(items, g.TreeNode("", flags, g.Layout{
				g.Custom(func(name string) func() {
					return func() {
						if g.IsItemHovered() {
							if g.IsMouseDoubleClicked(g.MouseButtonLeft) {
								e.openArchsetFromArchetype(name)
							} else if g.IsMouseClicked(g.MouseButtonLeft) {
								e.selectedArchetype = name
							}
						}
					}
				}(archName)),
				g.Custom(func(archName string) func() {
					return func() {
						arch := e.dataManager.GetArchetype(archName)
						if arch == nil {
							return
						}
						anim, face := e.dataManager.GetAnimAndFace(arch, "", "")
						imageName, err := e.dataManager.GetAnimFaceImage(anim, face)
						if err != nil {
							return
						}
						img := e.dataManager.GetImage(imageName)
						if t, ok := e.imageTextures[imageName]; !ok || t.texture == nil {
							go func() {
								rgba := image.NewRGBA(img.Bounds())
								draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)
								tex, err := g.NewTextureFromRgba(rgba)
								if err != nil {
									log.Fatalln(err)
								}
								e.imageTextures[imageName] = ImageTexture{
									texture: tex,
									width:   float32(img.Bounds().Max.X),
									height:  float32(img.Bounds().Max.Y),
								}
								g.Update()
							}()
						} else {
							g.SameLine()
							g.Image(t.texture, t.width, t.height).Build()
						}
					}
				}(archName)),
				g.Custom(func() { g.SameLine() }),
				g.Label(archName),
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
			items = append(items, g.Row(g.TreeNode(archFile, g.TreeNodeFlagsCollapsingHeader|g.TreeNodeFlagsDefaultOpen, archItems)))
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

func (e *Editor) openArchsetFromArchetype(archName string) {
	log.Printf("Open archset editor for file that corresponds to %s\n", archName)
}
