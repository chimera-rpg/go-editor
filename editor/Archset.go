package editor

import (
	"fmt"
	"path"

	g "github.com/AllenDang/giu"
	imgui "github.com/AllenDang/imgui-go"
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

	g.Window(fmt.Sprintf("Archset: %s", a.filename)).IsOpen(&windowOpen).Flags(g.WindowFlagsMenuBar).Pos(210, 430).Size(300, 400).Layout(
		g.MenuBar().Layout(
			g.Menu("Archset").Layout(
				g.MenuItem("New Arch...").OnClick(func() {
					newArchPopup = true
				}),
				g.Separator(),
				g.MenuItem("Save All").OnClick(func() {}),
				g.Separator(),
				g.MenuItem("Close").OnClick(func() {
					a.close()
				}),
			),
		),
		g.Custom(func() {
			if imgui.BeginTabBarV("Archset", int(g.TabBarFlagsFittingPolicyScroll|g.TabBarFlagsFittingPolicyResizeDown)) {
				for archIndex, arch := range a.archs {
					var flags g.TabItemFlags
					if arch.unsaved {
						flags |= g.TabItemFlagsUnsavedDocument
					}
					if imgui.BeginTabItemV(arch.DataName(), nil, int(flags)) {
						_, availH := g.GetAvailableRegion()
						a.currentArchIndex = archIndex
						arch.textEditor.Render("Source", imgui.Vec2{X: 0, Y: availH - 20}, false)
						if arch.textEditor.IsTextChanged() {
							arch.SetUnsaved(true)
						}
						g.Row(
							g.Button("Reset").OnClick(func() {
								arch.Reset()
							}),
							g.Button("Save").OnClick(func() {
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
		g.PopupModal("New Arch").IsOpen(&newArchPopup).Flags(g.WindowFlagsHorizontalScrollbar).Layout(
			g.Label("Create a new arch"),
			g.InputText(&a.newDataName).Label("Data Name"),
			g.InputText(&a.newName).Label("Name"),
			g.Row(
				g.Button("Create").OnClick(func() {
					// TODO: Check if arch with the same name already exists
					a.archs = append(a.archs, NewUnReArch(sdata.Archetype{
						Name: &*&a.newName, // TODO: Replace with a NewString(...) function call.
					}, a.newDataName))
					g.CloseCurrentPopup()
					a.setDefaults()
				}),
				g.Button("Cancel").OnClick(func() {
					g.CloseCurrentPopup()
					a.setDefaults()
				}),
			),
		),
	)

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
