package editor

import (
	"fmt"
	"path"

	g "github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/chimera-rpg/go-editor/data"
	sdata "github.com/chimera-rpg/go-server/data"
)

type Archset struct {
	filename             string
	archs                []UnReArch
	archsSauce           []string
	shouldClose          bool
	newDataName, newName string
	currentArchIndex     int
}

func NewArchset(name string, archs map[string]*sdata.Archetype) *Archset {
	a := &Archset{
		filename: name,
	}

	for k, v := range archs {
		a.archs = append(a.archs, NewUnReArch(*v, k))
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
			if imgui.BeginTabBarV("Archset", int(g.TabBarFlagsFittingPolicyScroll|g.TabBarFlagsFittingPolicyResizeDown)) {
				for archIndex, arch := range a.archs {
					if imgui.BeginTabItemV(arch.DataName(), nil, 0) {
						a.currentArchIndex = archIndex
						arch.textEditor.Render("Source", imgui.Vec2{X: 0, Y: 0}, false)
						if arch.textEditor.IsTextChanged() {
							// TODO: Store that we've changed so we can set the tab item flags to show unsaved.
						}
						imgui.EndTabItem()
					}
				}
				imgui.EndTabBar()
			}
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
					a.archs = append(a.archs, NewUnReArch(sdata.Archetype{
						Name: sdata.NewStringExpression(a.newName),
					}, a.newDataName))
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
