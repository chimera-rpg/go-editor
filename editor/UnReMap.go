package editor

import (
	sdata "github.com/chimera-rpg/go-server/data"
	undo "github.com/iomodo/a-simple-undo-redo"
)

type UnReMap struct {
	undoer undo.Undoer
}

func NewUnReMap(m *sdata.Map) UnReMap {
	undoer := undo.NewUndoer(0)
	undoer.Save(m)
	return UnReMap{
		undoer: undoer,
	}
}

func (u *UnReMap) Set(m *sdata.Map) {
	u.undoer.Save(m)
}

func (u *UnReMap) Get() *sdata.Map {
	return u.undoer.State().(*sdata.Map)
}

func (u *UnReMap) Undo() {
	u.undoer.Undo()
}

func (u *UnReMap) Redo() {
	u.undoer.Redo()
}
