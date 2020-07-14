package editor

import (
	"fmt"
	"path"

	g "github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	sdata "github.com/chimera-rpg/go-server/data"
)

type Archset struct {
	context              *Context
	filename             string
	archs                []*UnReArch
	archsSauce           []string
	shouldClose          bool
	newDataName, newName string
	currentArchIndex     int
}

func NewArchset(context *Context, name string, archs map[string]*sdata.Archetype) *Archset {
	a := &Archset{
		filename: name,
		context:  context,
	}

	for k, v := range archs {
		a.archs = append(a.archs, NewUnReArch(*v, k))
	}

	a.setDefaults()

	return a
}

func (a *Archset) draw() {
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
					var flags g.TabItemFlags
					if arch.unsaved {
						flags |= g.TabItemFlagsUnsavedDocument
					}
					if imgui.BeginTabItemV(arch.DataName(), nil, int(flags)) {
						_, availH := g.GetAvaiableRegion()
						a.currentArchIndex = archIndex
						arch.textEditor.Render("Source", imgui.Vec2{X: 0, Y: availH - 20}, false)
						if arch.textEditor.IsTextChanged() {
							arch.SetUnsaved(true)
						}
						g.Line(
							g.Button("Reset", func() {
								arch.Reset()
							}),
							g.Button("Save", func() {
								arch.Save()
								// TODO: Resave current Archset with most recent saved versions of Archs.
							}),
						).Build()
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
