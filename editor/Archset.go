package editor

import (
	"fmt"

	g "github.com/AllenDang/giu"
	"github.com/chimera-rpg/go-editor/data"
)

type Archset struct {
	filename string
}

func (a *Archset) draw(d *data.Manager) {
	var b bool

	g.WindowV(fmt.Sprintf("Archset: %s", a.filename), &b, g.WindowFlagsMenuBar, 210, 430, 300, 400, g.Layout{})
}
