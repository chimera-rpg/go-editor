package data

import (
	"github.com/chimera-rpg/go-editor/internal/unredo"
	sdata "github.com/chimera-rpg/go-server/data"
)

type UnReMap struct {
	undoer   unredo.Unredoabler
	dataName string
	savedMap *sdata.Map
	unsaved  bool
}

func NewUnReMap(m *sdata.Map, d string) *UnReMap {
	undoer := unredo.NewUnredoabler(m)
	u := &UnReMap{
		undoer:   undoer,
		dataName: d,
	}
	u.Save()

	return u
}

func (u *UnReMap) Replace(m *sdata.Map) {
	u.undoer.Replace(m)
}

func (u *UnReMap) Set(m *sdata.Map) {
	u.unsaved = true
	u.undoer.Push(m)
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
	u.undoer.Replace(u.savedMap)
	u.Save()
	u.unsaved = false
}

func (u *UnReMap) DataName() string {
	return u.dataName
}

func (u *UnReMap) SetDataName(s string) {
	u.dataName = s
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
		Script:      t.Script,
		Y:           t.Y,
		X:           t.X,
		Z:           t.Z,
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

func (u *UnReMap) GetArchs(y, x, z int) []sdata.Archetype {
	t := u.Get()
	if y >= 0 && y < t.Height {
		if x >= 0 && x < t.Width {
			if z >= 0 && z < t.Depth {
				return t.Tiles[y][x][z]
			}
		}
	}
	return nil
}
