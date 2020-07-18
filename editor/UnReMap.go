package editor

import (
	sdata "github.com/chimera-rpg/go-server/data"
	undo "github.com/iomodo/a-simple-undo-redo"
)

type UnReMap struct {
	undoer   undo.Undoer
	dataName string
	savedMap *sdata.Map
	unsaved  bool
}

func NewUnReMap(m *sdata.Map, d string) *UnReMap {
	undoer := undo.NewUndoer(0)
	u := &UnReMap{
		undoer:   undoer,
		dataName: d,
	}
	u.Set(m)
	u.Save()

	return u
}

func (u *UnReMap) Set(m *sdata.Map) {
	u.unsaved = true
	u.undoer.Save(m)
}

func (u *UnReMap) Get() *sdata.Map {
	return u.undoer.State().(*sdata.Map)
}

func (u *UnReMap) Undo() {
	u.unsaved = true
	u.undoer.Undo()
}

func (u *UnReMap) Redo() {
	u.unsaved = true
	u.undoer.Redo()
}

func (u *UnReMap) SavedMap() *sdata.Map {
	return u.savedMap
}

func (u *UnReMap) Unsaved() bool {
	return u.unsaved
}

func (u *UnReMap) SetUnsaved(b bool) {
	u.unsaved = b
}

func (u *UnReMap) Save() {
	u.savedMap = u.Clone()
	u.unsaved = false
}

func (u *UnReMap) Reset() {
	u.Set(u.savedMap)
	u.Save()
	u.unsaved = false
}

func (u *UnReMap) DataName() string {
	return u.dataName
}

func (u *UnReMap) Clone() *sdata.Map {
	t := u.Get()
	// Make a new map according to the given dimensions
	clone := &sdata.Map{
		Name:        t.Name,
		Description: t.Description,
		Darkness:    t.Darkness,
		ResetTime:   t.ResetTime,
		Lore:        t.Lore,
		Height:      t.Height,
		Width:       t.Width,
		Depth:       t.Depth,
	}
	// Create the new map according to dimensions.
	for y := 0; y < t.Height; y++ {
		clone.Tiles = append(clone.Tiles, [][][]sdata.Archetype{})
		for x := 0; x < t.Width; x++ {
			clone.Tiles[y] = append(clone.Tiles[y], [][]sdata.Archetype{})
			for z := 0; z < t.Depth; z++ {
				clone.Tiles[y][x] = append(clone.Tiles[y][x], []sdata.Archetype{})
				for a := 0; a < len(t.Tiles[y][x][z]); a++ {
					clone.Tiles[y][x][z] = append(clone.Tiles[y][x][z], t.Tiles[y][x][z][a])
				}
			}
		}
	}
	return clone
}
