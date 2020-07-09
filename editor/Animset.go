package editor

import (
	"fmt"

	g "github.com/AllenDang/giu"
	"github.com/chimera-rpg/go-editor/data"
)

type Animset struct {
	filename string
}

func (a *Animset) draw(d *data.Manager) {
	var b bool

	g.WindowV(fmt.Sprintf("Animset: %s", a.filename), &b, g.WindowFlagsMenuBar, 210, 440, 300, 400, g.Layout{})
}
