package widgets

import (
	log "github.com/sirupsen/logrus"

	g "github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/chimera-rpg/go-editor/data"
	sdata "github.com/chimera-rpg/go-server/data"
)

type Context interface {
	DataManager() *data.Manager
}

type ArchEditorWidget struct {
	arch               *sdata.Archetype
	descEditor         imgui.TextEditor
	context            Context
	name               string
	preChangeCallback  func() bool
	postChangeCallback func() bool
	requestSave        func() bool
	requestUndo        func() bool
	requestRedo        func() bool
}

func NewArchEditor() *ArchEditorWidget {
	a := &ArchEditorWidget{
		descEditor: imgui.NewTextEditor(),
	}
	a.descEditor.SetShowWhitespaces(false)
	return a
}

func (a *ArchEditorWidget) SetPreChangeCallback(f func() bool) {
	a.preChangeCallback = f
}
func (a *ArchEditorWidget) SetPostChangeCallback(f func() bool) {
	a.postChangeCallback = f
}

func (a *ArchEditorWidget) SetSaveCallback(f func() bool) {
	a.requestSave = f
}

func (a *ArchEditorWidget) SetUndoCallback(f func() bool) {
	a.requestUndo = f
}

func (a *ArchEditorWidget) SetRedoCallback(f func() bool) {
	a.requestRedo = f
}

func (a *ArchEditorWidget) SetArchetype(arch *sdata.Archetype) {
	if arch == a.arch {
		return
	}
	dm := a.context.DataManager()
	a.arch = arch
	a.name = dm.GetArchName(a.arch, "")
}

func (a *ArchEditorWidget) Layout() (l g.Layout) {
	l = g.Layout{
		g.MenuBar(g.Layout{}),
		a.ArchetypeLayout(),
	}

	return l
}

func (a *ArchEditorWidget) ArchetypeLayout() (l g.Layout) {
	l = g.Layout{
		g.InputText("Name", 0, &a.name),
		g.Line(
			g.Button("apply", func() {
				dm := a.context.DataManager()
				a.preChangeCallback()
				if a.name != dm.GetArchName(a.arch, "") {
					if err := dm.SetArchField(a.arch, "Name", a.name); err != nil {
						log.Println(err)
					}
				}
				a.postChangeCallback()
			}),
		),
	}
	return
}

func (a *ArchEditorWidget) SetContext(c Context) {
	a.context = c
}
