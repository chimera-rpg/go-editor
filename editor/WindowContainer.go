package editor

import g "github.com/AllenDang/giu"

type WindowContainer struct {
	window *g.WindowWidget
	title  string
	layout g.Layout
}

func (w *WindowContainer) Window() *g.WindowWidget {
	return w.window
}

func (w *WindowContainer) Title() string {
	return w.title
}

func (w *WindowContainer) Layout() g.Layout {
	return w.layout
}
