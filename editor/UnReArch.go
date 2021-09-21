package editor

import (
	imgui "github.com/AllenDang/imgui-go"
	"github.com/chimera-rpg/go-editor/internal/unredo"
	sdata "github.com/chimera-rpg/go-server/data"
	"gopkg.in/yaml.v2"
)

type UnReArch struct {
	undoer     unredo.Unredoabler
	source     string
	dataName   string
	textEditor imgui.TextEditor
	savedArch  sdata.Archetype
	unsaved    bool
}

func NewUnReArch(a sdata.Archetype, d string) *UnReArch {
	undoer := unredo.NewUnredoabler(a)
	u := &UnReArch{
		undoer:     undoer,
		dataName:   d,
		textEditor: imgui.NewTextEditor(),
		savedArch:  a,
	}
	u.textEditor.SetShowWhitespaces(false)

	return u
}

func (u *UnReArch) Set(a sdata.Archetype) error {
	u.undoer.Push(a)

	bytes, err := yaml.Marshal(a)
	if err != nil {
		return err
	}
	u.source = string(bytes)
	u.textEditor.SetText(u.source)
	return nil
}

func (u *UnReArch) SetSource(s string) error {
	var a sdata.Archetype

	err := yaml.Unmarshal([]byte(s), &a)
	if err != nil {
		return err
	}
	u.Set(a)
	return nil
}

func (u *UnReArch) SyncSourceToSave() error {
	bytes, err := yaml.Marshal(u.savedArch)
	if err != nil {
		return err
	}
	u.source = string(bytes)
	u.textEditor.SetText(u.source)
	return nil
}

func (u *UnReArch) Get() sdata.Archetype {
	return u.undoer.State().(sdata.Archetype)
}

func (u *UnReArch) GetSource() string {
	return u.source
}

func (u *UnReArch) Undo() {
	u.undoer.Undo()
}

func (u *UnReArch) Redo() {
	u.undoer.Redo()
}

func (u *UnReArch) SavedArch() sdata.Archetype {
	return u.savedArch
}

func (u *UnReArch) DataName() string {
	return u.dataName
}

func (u *UnReArch) SetUnsaved(b bool) {
	u.unsaved = b
}

func (u *UnReArch) Save() {
	u.SetSource(u.textEditor.GetText())
	u.savedArch = u.Get()
	u.SyncSourceToSave()
	u.unsaved = false
}

func (u *UnReArch) Reset() {
	u.Set(u.savedArch)
	u.SyncSourceToSave()
	u.unsaved = false
}
