package widgets

import (
	"image/color"

	log "github.com/sirupsen/logrus"

	g "github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/chimera-rpg/go-editor/data"
	sdata "github.com/chimera-rpg/go-server/data"
)

type Context interface {
	DataManager() *data.Manager
}

type StringPair struct {
	initial  string
	pending  string
	previous string
	reset    bool
}

type ArchEditorWidget struct {
	arch               *sdata.Archetype
	descEditor         imgui.TextEditor
	context            Context
	name               StringPair
	description        StringPair
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
	a.arch = arch

	//
	a.Refresh()
}

func (a *ArchEditorWidget) Refresh() {
	a.name = a.getStringPair("Name")
	a.description = a.getStringPair("Description")
}

func (a *ArchEditorWidget) getStringPair(field string) StringPair {
	dm := a.context.DataManager()
	s := StringPair{}
	v1 := dm.GetArchAncestryField(a.arch, field)
	if v1.IsValid() && !v1.IsNil() {
		s.previous = v1.Elem().String()
	}
	v2 := dm.GetArchField(a.arch, field)
	if v2.IsValid() && !v2.IsNil() {
		s.pending = v2.Elem().String()
		s.initial = v2.Elem().String()
	}
	s.reset = false
	return s
}

func (a *ArchEditorWidget) checkStringPair(field string, s *StringPair) {
	if s.reset {
		a.context.DataManager().ClearArchField(a.arch, field)
		s.reset = false
	} else if s.initial != s.pending {
		if err := a.context.DataManager().SetArchField(a.arch, field, s.pending); err != nil {
			log.Println(err)
		}
		s.initial = s.pending
	}
}

func (a *ArchEditorWidget) Layout() (l g.Layout) {
	l = g.Layout{
		g.MenuBar(g.Layout{}),
		a.ArchetypeLayout(),
	}

	return l
}

func (a *ArchEditorWidget) StringLayout(field string, target *StringPair) g.Layout {
	dm := a.context.DataManager()
	isLocal, _ := dm.IsArchFieldLocal(a.arch, field)
	var resetButton g.Widget
	resetButton = g.Button("reset", func() {
		target.reset = true
		target.pending = target.previous
	})
	if !isLocal || target.reset {
		resetButton = g.Dummy(0, 0)
	}

	var inputField g.Widget
	if field == "Description" {
		inputField = g.InputTextMultiline(field, &target.pending, 0, 0, g.InputTextFlagsNone, nil, nil)
	} else {
		inputField = g.InputText(field, 0, &target.pending)
	}

	return g.Layout{
		g.Custom(func() {
			if isLocal {
				g.PushColorText(color.RGBA{
					R: 200,
					G: 128,
					B: 0,
					A: 255,
				})
			}
		}),
		g.Line(
			inputField,
			resetButton,
		),
		g.Custom(func() {
			if isLocal {
				g.PopStyleColor()
			}
		}),
	}
}

func (a *ArchEditorWidget) ArchetypeLayout() (l g.Layout) {
	l = g.Layout{
		a.StringLayout("Name", &a.name),
		a.StringLayout("Description", &a.description),
		g.Line(
			g.Button("apply", func() {
				a.preChangeCallback()

				a.checkStringPair("Name", &a.name)
				a.checkStringPair("Description", &a.description)

				a.postChangeCallback()
				a.Refresh()
			}),
		),
	}
	return
}

func (a *ArchEditorWidget) SetContext(c Context) {
	a.context = c
}
