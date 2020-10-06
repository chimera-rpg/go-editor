package editor

import (
	"errors"
	"image"
	"image/draw"
	"os"
	"path"

	g "github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/chimera-rpg/go-editor/data"
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
	mapsets        []*Mapset
	archsets       []*Archset
	animsets       []*Animset
	context        Context
	//
	openMapCWD, openMapFilename string
}

type ImageTexture struct {
	texture       *g.Texture
	width, height float32
}

func (e *Editor) Setup(dataManager *data.Manager) (err error) {
	e.context = Context{
		dataManager:   dataManager,
		imageTextures: make(map[string]*ImageTexture),
	}
	e.isLoaded = false
	e.isRunning = true
	e.archetypesMode = true
	e.showSplash = true
	e.openMapCWD = dataManager.MapsPath

	e.masterWindow = g.NewMasterWindow("Editor", 800, 600, 0, nil)
	g.Context.GetRenderer().SetTextureMagFilter(g.TextureFilterNearest)
	imgui.CurrentIO().SetIniFilename(e.context.dataManager.GetEtcPath("chimera-editor.ini"))

	for imagePath, img := range dataManager.GetImages() {
		//e.context.imageTexturesLock.Lock()
		e.context.imageTextures[imagePath] = &ImageTexture{
			texture: nil,
			width:   float32(img.Bounds().Max.X),
			height:  float32(img.Bounds().Max.Y),
		}
		//e.context.imageTexturesLock.Unlock()
	}

	go func() {
		for imagePath, img := range dataManager.GetImages() {
			rgba := image.NewRGBA(img.Bounds())
			draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)
			tex, err := g.NewTextureFromRgba(rgba)
			if err != nil {
				log.Fatalln(err)
			}
			if it, ok := e.context.imageTextures[imagePath]; ok {
				it.texture = tex
			}
		}
		e.isLoaded = true
	}()

	return nil
}

func (e *Editor) Destroy() {
	e.isRunning = false
}

func (e *Editor) Start() {
	log.Println("Editor: Start")

	e.masterWindow.Main(func() { e.loop() })
}

func (e *Editor) loop() {
	if !e.isRunning {
		os.Exit(0)
	}

	var openMapPopup bool

	if !e.isLoaded {
		g.OpenPopup("Loading...")
		g.PopupModalV("Loading...", nil, g.WindowFlagsNoResize, g.Layout{
			g.Label("Now loading files..."),
		}).Build()
		return
	}

	g.MainMenuBar(g.Layout{
		g.Menu("File", g.Layout{
			g.MenuItem("New Mapset", func() {
				e.mapsets = append(e.mapsets, NewMapset(&e.context, "", nil))
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
				dMapset, err := e.context.dataManager.LoadMap(fullPath)
				if err != nil {
					// TODO: Popup some sort of error!
					log.Errorln(err)
				} else {
					e.mapsets = append(e.mapsets, NewMapset(&e.context, fullPath, dMapset))
				}
				g.CloseCurrentPopup()
			}),
		),
	}).Build()

	for i, m := range e.mapsets {
		m.draw()
		if m.shouldClose {
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
		archs := e.context.dataManager.GetArchetypes()
		for _, archName := range archs {
			var flags g.TreeNodeFlags
			flags = g.TreeNodeFlagsLeaf | g.TreeNodeFlagsSpanFullWidth
			if archName == e.context.selectedArch {
				flags |= g.TreeNodeFlagsSelected
			}
			items = append(items, g.TreeNode("", flags, g.Layout{
				g.Custom(func(name string) func() {
					return func() {
						if g.IsItemHovered() {
							if g.IsMouseDoubleClicked(g.MouseButtonLeft) {
								e.openArchsetFromArchetype(name)
							} else if g.IsMouseClicked(g.MouseButtonLeft) {
								e.context.selectedArch = name
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
							if t.texture != nil {
								g.Image(t.texture, t.width, t.height).Build()
							}
						}
					}
				}(archName)),
				g.Custom(func() { g.SameLine() }),
				g.Label(archName),
			}))
		}
	} else {
		archs := e.context.dataManager.GetArchetypeFiles()
		for _, archFile := range archs {
			var archItems []g.Widget
			for _, archName := range e.context.dataManager.GetArchetypeFile(archFile) {
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
