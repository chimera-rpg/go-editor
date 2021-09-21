package widgets

import (
	g "github.com/AllenDang/giu"
)

type TaskbarWidget struct {
	ParentWidth, ParentHeight int
}

func NewTaskbar() *TaskbarWidget {
	return &TaskbarWidget{}
}

func (t *TaskbarWidget) Draw(windows []WindowContainerI) {
	w := g.Window("Taskbar")

	var items []g.Widget
	for _, win := range windows {
		items = append(items, func(win WindowContainerI) g.Widget {
			return g.Button(win.Title()).OnClick(func() {
				win.Window().BringToFront()
			})
		}(win))
	}

	w.Flags(g.WindowFlagsNoTitleBar|g.WindowFlagsNoResize|g.WindowFlagsNoMove|g.WindowFlagsNoSavedSettings|g.WindowFlagsAlwaysHorizontalScrollbar|g.WindowFlagsNoFocusOnAppearing|g.WindowFlagsNoScrollbar).Size(float32(t.ParentWidth), 48).Pos(0, float32(t.ParentHeight)-48).Layout(g.Row(items...))
	for _, win := range windows {
		win.Window().Layout(win.Layout())
	}
}
