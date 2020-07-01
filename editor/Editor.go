package editor

import (
	"fmt"
	"os"

	g "github.com/AllenDang/giu"
	"github.com/chimera-rpg/go-editor/data"
	log "github.com/sirupsen/logrus"
)

type Editor struct {
	masterWindow *g.MasterWindow
	dataManager  *data.Manager
	isRunning    bool
	showSplash   bool
}

func (e *Editor) Setup(dataManager *data.Manager) (err error) {
	e.dataManager = dataManager
	e.isRunning = true

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

	e.drawArchetypes()
	e.drawMap()
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
	rows = append(rows, g.Row(g.Label("test")))
	g.Window("Archetypes", 10, 30, 200, 400, g.Layout{
		g.FastTable("yeet", true, rows),
	})
}

func (e *Editor) drawMap( /* m Map* */ ) {
	var yChildren []g.Widget
	mHeight := 8
	mWidth := 8
	mDepth := 8
	tWidth := 32
	tHeight := 24
	for y := 0; y < mHeight; y++ {
		var rows []*g.RowWidget
		for z := 0; z < mDepth; z++ {
			var cells []g.Widget
			for x := 0; x < mWidth; x++ {
				cells = append(cells, g.SelectableV("", false, g.SelectableFlagsNone, float32(tWidth), float32(tHeight), func() {
					fmt.Printf("%dx%dx%d\n", y, x, z)
				}))
			}
			rows = append(rows, g.Row(cells...))
		}
		yChildren = append(yChildren, g.Child(fmt.Sprintf("layer %d", y), true, float32(mWidth*tWidth), float32(mDepth*tHeight), g.WindowFlagsNoBackground|g.WindowFlagsNoCollapse|g.WindowFlagsNoResize|g.WindowFlagsNoDecoration, g.Layout{
			g.Table("", false, rows),
		}))
	}
	g.Window("Map", 210, 30, float32(mWidth*tWidth), float32(mWidth*tWidth), yChildren)
}
