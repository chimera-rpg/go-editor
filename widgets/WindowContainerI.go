package widgets

import g "github.com/AllenDang/giu"

type WindowContainerI interface {
	Window() *g.WindowWidget
	Title() string
	Layout() g.Layout
}
