package editor

import (
	"fmt"

	g "github.com/AllenDang/giu"
	"github.com/chimera-rpg/go-editor/data"
	sdata "github.com/chimera-rpg/go-server/data"
)

type Archset struct {
	filename string
	archs    []sdata.Archetype
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

	g.WindowV(fmt.Sprintf("Archset: %s", a.filename), &b, g.WindowFlagsMenuBar, 210, 430, 300, 400, g.Layout{})
}
