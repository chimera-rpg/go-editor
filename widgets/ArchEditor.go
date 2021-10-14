package widgets

import (
	"image/color"
	"log"
	"reflect"

	g "github.com/AllenDang/giu"
	imgui "github.com/AllenDang/imgui-go"
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-editor/data"
	sdata "github.com/chimera-rpg/go-server/data"
)

type Context interface {
	DataManager() *data.Manager
}

type StringPair struct {
	initial       reflect.Value
	pending       reflect.Value
	previous      reflect.Value
	pendingStr    string
	initialStr    string
	previousStr   string
	pendingFloat  float32
	initialFloat  float32
	previousFloat float32
	pendingInt    int32
	initialInt    int32
	previousInt   int32
	reset         bool
}

type ArchEditorWidget struct {
	arch               *sdata.Archetype
	descEditor         imgui.TextEditor
	context            Context
	pairs              map[string]*StringPair
	preChangeCallback  func() bool
	postChangeCallback func() bool
	requestSave        func() bool
	requestUndo        func() bool
	requestRedo        func() bool
}

func NewArchEditor() *ArchEditorWidget {
	a := &ArchEditorWidget{
		descEditor: imgui.NewTextEditor(),
		pairs:      make(map[string]*StringPair),
	}
	a.descEditor.SetShowWhitespaces(false)
	return a
}

func (a *ArchEditorWidget) Draw() (title string, w *g.WindowWidget, layout g.Layout) {
	layout = a.Layout()
	title = "Active Archetype"
	w = g.Window(title)
	return
}

func (a *ArchEditorWidget) ResetCallbacks() {
	a.preChangeCallback = nil
	a.postChangeCallback = nil
	a.requestRedo = nil
	a.requestUndo = nil
	a.requestSave = nil
	a.arch = nil
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
	if a.arch == nil {
		return
	}
	a.pairs["Name"] = a.getStringPair("Name")
	a.pairs["Description"] = a.getStringPair("Description")
	a.pairs["Value"] = a.getStringPair("Value")
	a.pairs["Count"] = a.getStringPair("Count")
	a.pairs["Audio"] = a.getStringPair("Audio")
	a.pairs["SoundSet"] = a.getStringPair("SoundSet")
	a.pairs["SoundIndex"] = a.getStringPair("SoundIndex")
}

func (a *ArchEditorWidget) getStringPair(field string) *StringPair {
	dm := a.context.DataManager()
	s := StringPair{}
	v1 := dm.GetArchAncestryField(a.arch, field)
	if v1.IsValid() {
		s.previous = v1
		if s.previous.Kind() == reflect.Ptr && !s.previous.IsNil() {
			if s.previous.Elem().Kind() == reflect.String {
				s.previousStr = s.previous.Elem().String()
			}
		} else {
			if s.previous.Kind() == reflect.String {
				s.previousStr = s.previous.String()
			} else if s.previous.Kind() == reflect.Float32 {
				s.previousFloat = float32(s.previous.Float())
			} else if s.previous.Kind() == reflect.Uint || s.previous.Kind() == reflect.Uint8 || s.previous.Kind() == reflect.Uint16 || s.previous.Kind() == reflect.Uint32 {
				s.previousInt = int32(s.previous.Uint())
			} else if s.previous.Kind() == reflect.Int || s.previous.Kind() == reflect.Int32 || s.previous.Kind() == reflect.Int8 {
				s.previousInt = int32(s.previous.Int())
			}
		}
	}
	v2 := dm.GetArchField(a.arch, field)
	if v2.IsValid() {
		s.pending = v2
		s.initial = v2
		if s.pending.Kind() == reflect.Ptr && !s.pending.IsNil() {
			if s.pending.Elem().Kind() == reflect.String {
				s.pendingStr = s.pending.Elem().String()
				s.initialStr = s.pendingStr
			}
		} else {
			if s.pending.Kind() == reflect.String {
				s.pendingStr = s.pending.String()
				s.initialStr = s.pendingStr
			} else if s.pending.Kind() == reflect.Float32 {
				s.pendingFloat = float32(s.pending.Float())
				s.initialFloat = s.pendingFloat
			} else if s.pending.Kind() == reflect.Uint || s.pending.Kind() == reflect.Uint8 || s.pending.Kind() == reflect.Uint16 || s.pending.Kind() == reflect.Uint32 {
				s.pendingInt = int32(s.pending.Uint())
				s.initialInt = s.pendingInt
			} else if s.pending.Kind() == reflect.Int || s.pending.Kind() == reflect.Int32 || s.pending.Kind() == reflect.Int8 {
				s.pendingInt = int32(s.pending.Int())
				s.initialInt = s.pendingInt
			}
		}
	}
	s.reset = false
	return &s
}

func (a *ArchEditorWidget) checkStringPair(field string, s *StringPair) {
	if s.reset {
		a.context.DataManager().ClearArchField(a.arch, field)
		s.reset = false
	} else if s.initialStr != s.pendingStr || s.initialInt != s.pendingInt || s.initialFloat != s.pendingFloat {
		if s.pending.Kind() == reflect.Ptr {
			if s.pending.Type().Elem().Kind() == reflect.String {
				if err := a.context.DataManager().SetArchField(a.arch, field, s.pendingStr); err != nil {
					log.Println(err)
				}
			}
		} else {
			if s.pending.Kind() == reflect.String {
				if err := a.context.DataManager().SetArchField(a.arch, field, s.pendingStr); err != nil {
					log.Println(err)
				}
			} else if s.pending.Kind() == reflect.Float32 {
				if err := a.context.DataManager().SetArchField(a.arch, field, s.pendingFloat); err != nil {
					log.Println(err)
				}
			} else if s.pending.Kind() == reflect.Uint {
				if err := a.context.DataManager().SetArchField(a.arch, field, uint32(s.pendingInt)); err != nil {
					log.Println(err)
				}
			} else if s.pending.Kind() == reflect.Int {
				if err := a.context.DataManager().SetArchField(a.arch, field, int32(s.pendingInt)); err != nil {
					log.Println(err)
				}
			} else if s.pending.Kind() == reflect.Int8 {
				if err := a.context.DataManager().SetArchField(a.arch, field, int8(s.pendingInt)); err != nil {
					log.Println(err)
				}
			}
		}
		s.initial = s.pending
		s.initialStr = s.pendingStr
		s.initialInt = s.pendingInt
		s.initialFloat = s.pendingFloat
	}
}

func (a *ArchEditorWidget) Layout() (l g.Layout) {
	l = g.Layout{
		g.MenuBar().Layout(),
		a.ArchetypeLayout(),
	}

	return l
}

func (a *ArchEditorWidget) StringLayout(title, field, tooltip string) g.Layout {
	dm := a.context.DataManager()
	isLocal, _ := dm.IsArchFieldLocal(a.arch, field)
	var resetButton g.Widget
	target := a.pairs[field]
	resetButton = g.Button("reset").OnClick(func() {
		target.reset = true
		target.pending = target.previous
	})
	if !isLocal || target.reset {
		resetButton = g.Dummy(0, 0)
	}

	var inputField g.Widget
	if field == "Description" {
		//inputField = g.InputTextMultiline(field, &target.pending, 0, 0, g.InputTextFlagsNone, nil, nil)
	}
	k := target.pending.Kind()
	if k == reflect.Ptr {
		inputField = g.InputText(&target.pendingStr).Label(title)
	} else {
		if k == reflect.String {
			inputField = g.InputText(&target.pendingStr).Label(title)
		} else if k == reflect.Uint || k == reflect.Uint8 {
			inputField = g.InputInt(&target.pendingInt).Label(title)
		} else if k == reflect.Int || k == reflect.Int8 {
			inputField = g.InputInt(&target.pendingInt).Label(title)
		} else if k == reflect.Float32 {
			inputField = g.InputFloat(&target.pendingFloat).Label(title)
		} else {
			inputField = g.Label("unhandled")
		}
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
		g.Row(
			inputField,
			g.Tooltip(tooltip),
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
	if a.arch == nil {
		return g.Layout{
			g.Label("no archetype selected"),
		}
	}
	l = g.Layout{
		a.StringLayout("Name", "Name", "The name of the object."),
		a.StringLayout("Description", "Description", "The description of the object."),
		a.musicLayout(),
		g.Row(
			g.Button("apply").OnClick(func() {
				a.preChangeCallback()

				for k, p := range a.pairs {
					a.checkStringPair(k, p)
				}

				a.postChangeCallback()
				a.Refresh()
			}),
		),
	}
	return
}

func (a *ArchEditorWidget) musicLayout() (l g.Layout) {
	dm := a.context.DataManager()
	v := dm.GetArchAncestryField(a.arch, "Type")
	if !v.IsValid() {
		return
	}
	t := cdata.ArchetypeType(v.Uint())
	if t != cdata.ArchetypeAudio {
		return
	}
	l = g.Layout{
		a.StringLayout("Audio", "Audio", "The audio definition to use."),
		a.StringLayout("SoundSet", "SoundSet", "The sound set to use."),
		a.StringLayout("SoundIndex", "SoundIndex", "The sound index to use for playback."),
		a.StringLayout("Volume", "Value", "Volume from 0.0 to 1.0"),
		a.StringLayout("Playback Count", "Count", "The playback count. -1 loops indefinitely."),
	}

	return
}

func (a *ArchEditorWidget) SetContext(c Context) {
	a.context = c
}
