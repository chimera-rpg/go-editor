package editor

import (
	"os"

	"path"

	g "github.com/AllenDang/giu"
	"github.com/chimera-rpg/go-editor/data"
	log "github.com/sirupsen/logrus"
)

type Editor struct {
	masterWindow *g.MasterWindow
	dataManager  *data.Manager
	isRunning    bool
	showSplash   bool
	mapTexture   *g.Texture
	mapTextureW  float32
	mapTextureH  float32
	mapsMap      map[string]*Maps
}

func (e *Editor) Setup(dataManager *data.Manager) (err error) {
	e.dataManager = dataManager
	e.isRunning = true
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
			g.MenuItem("Open", nil),
			g.Separator(),
			g.MenuItem("Exit", func() { e.isRunning = false }),
		}),
		g.Menu("Misc", g.Layout{
			g.Button("Button", nil),
		}),
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
	var rows []*g.RowWidget
	archs := e.dataManager.GetArchetypeFiles()
	for _, archFile := range archs {
		for archName := range e.dataManager.GetArchetypeFile(archFile) {
			rows = append(rows, g.Row(g.Label(archName)))
		}
	}
	var b bool
	g.WindowV("Archetypes", &b, g.WindowFlagsMenuBar, 10, 30, 200, 400, g.Layout{
		g.MenuBar(g.Layout{
			g.Menu("File", g.Layout{
				g.MenuItem("New...", func() {}),
				g.Separator(),
			}),
		}),
		g.FastTable("archetypes", true, rows),
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
