package editor

import (
	"fmt"
	"path"

	g "github.com/AllenDang/giu"
	"github.com/chimera-rpg/go-editor/data"
	sdata "github.com/chimera-rpg/go-server/data"
)

type Archset struct {
	filename             string
	archs                []sdata.Archetype
	shouldClose          bool
	newDataName, newName string
}

func NewArchset(name string, archs map[string]*sdata.Archetype) *Archset {
	a := &Archset{
		filename: name,
	}

	for _, v := range archs {
		a.archs = append(a.archs, *v)
	}

	a.setDefaults()

	return a
}

func (a *Archset) draw(d *data.Manager) {
	windowOpen := true

	var newArchPopup bool

	g.WindowV(fmt.Sprintf("Archset: %s", a.filename), &windowOpen, g.WindowFlagsMenuBar, 210, 430, 300, 400, g.Layout{
		g.MenuBar(g.Layout{
			g.Menu("Archset", g.Layout{
				g.MenuItem("New Arch...", func() {
					newArchPopup = true
				}),
				g.Separator(),
				g.MenuItem("Save All", func() {}),
				g.Separator(),
				g.MenuItem("Close", func() {
					a.close()
				}),
			}),
		}),
		g.Custom(func() {
			if newArchPopup {
				g.OpenPopup("New Arch")
			}
		}),
		g.PopupModalV("New Arch", nil, g.WindowFlagsHorizontalScrollbar, g.Layout{
			g.Label("Create a new arch"),
			g.InputText("Data Name", 0, &a.newDataName),
			g.InputText("Name", 0, &a.newName),
			g.Line(
				g.Button("Create", func() {
					// TODO: Check if arch with the same name already exists
					//a.archs = append(a.archs, NewUnReArch(sdata.Archetype{
					// Name: a.newName,
					// }, a.newDataName))
					g.CloseCurrentPopup()
					a.setDefaults()
				}),
				g.Button("Cancel", func() {
					g.CloseCurrentPopup()
					a.setDefaults()
				}),
			),
		}),
	})

	if !windowOpen {
		a.close()
	}
}

func (a *Archset) setDefaults() {
	a.newName = "My Archetype"
	a.newDataName = path.Join(path.Dir(a.filename), "myarch")
}

func (a *Archset) close() {
	a.shouldClose = true
}
