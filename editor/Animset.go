package editor

import (
	"fmt"

	g "github.com/AllenDang/giu"
)

type Animset struct {
	context  *Context
	filename string
}

func (a *Animset) draw() {
	var b bool

	g.Window(fmt.Sprintf("Animset: %s", a.filename)).IsOpen(&b).Flags(g.WindowFlagsMenuBar).Pos(210, 440).Size(300, 400)
}
