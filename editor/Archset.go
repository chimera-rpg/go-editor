package editor

import (
	"fmt"

	g "github.com/AllenDang/giu"
	"github.com/chimera-rpg/go-editor/data"
	sdata "github.com/chimera-rpg/go-server/data"
)

type Archset struct {
	filename    string
	archs       []sdata.Archetype
	shouldClose bool
}

func NewArchset(name string, archs map[string]*sdata.Archetype) *Archset {
	a := &Archset{
		filename: name,
	}

	for _, v := range archs {
		a.archs = append(a.archs, *v)
	}

	return a
}

func (a *Archset) draw(d *data.Manager) {
	var b bool

	var newArchPopup bool

	g.WindowV(fmt.Sprintf("Archset: %s", a.filename), &b, g.WindowFlagsMenuBar, 210, 430, 300, 400, g.Layout{
		g.MenuBar(g.Layout{
			g.Menu("Archset", g.Layout{
				g.MenuItem("New Arch...", func() {
					newArchPopup = true
				}),
				g.Separator(),
				g.MenuItem("Save All", func() {}),
				g.Separator(),
				g.MenuItem("Close", func() {
					a.shouldClose = true
				}),
			}),
		}),
	})
}
