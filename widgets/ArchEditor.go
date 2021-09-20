package widgets

import (
	g "github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/chimera-rpg/go-editor/data"
	sdata "github.com/chimera-rpg/go-server/data"
)

type Context interface {
	DataManager() *data.Manager
}

type ArchEditorWidget struct {
	arch       *sdata.Archetype
	descEditor imgui.TextEditor
	context    Context
}

func NewArchEditor() *ArchEditorWidget {
	a := &ArchEditorWidget{
		descEditor: imgui.NewTextEditor(),
	}
	a.descEditor.SetShowWhitespaces(false)
	return a
}

func (a *ArchEditorWidget) SetArchetype(arch *sdata.Archetype) {
	a.arch = arch
}

func (a *ArchEditorWidget) Layout() (l g.Layout) {
	l = g.Layout{
		g.MenuBar(g.Layout{}),
		a.ArchetypeLayout(),
	}

	return l
}

func (a *ArchEditorWidget) ArchetypeLayout() (l g.Layout) {
	dm := a.context.DataManager()
	l = g.Layout{
		g.Label(dm.GetArchName(a.arch, "")),
	}
	return
}

func (a *ArchEditorWidget) SetContext(c Context) {
	a.context = c
}
