package editor

import (
	"github.com/AllenDang/giu/imgui"
	sdata "github.com/chimera-rpg/go-server/data"
	undo "github.com/iomodo/a-simple-undo-redo"
	"gopkg.in/yaml.v2"
)

type UnReArch struct {
	undoer     undo.Undoer
	source     string
	dataName   string
	textEditor imgui.TextEditor
}

func NewUnReArch(a sdata.Archetype, d string) UnReArch {
	undoer := undo.NewUndoer(0)
	u := UnReArch{
		undoer:     undoer,
		dataName:   d,
		textEditor: imgui.NewTextEditor(),
	}
	u.Set(a)

	return u
}

func (u *UnReArch) Set(a sdata.Archetype) error {
	u.undoer.Save(a)

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

func (u *UnReArch) DataName() string {
	return u.dataName
}