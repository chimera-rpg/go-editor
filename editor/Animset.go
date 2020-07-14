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

	g.WindowV(fmt.Sprintf("Animset: %s", a.filename), &b, g.WindowFlagsMenuBar, 210, 440, 300, 400, g.Layout{})
}
