package mapview

import (
	"errors"
	"image"

	g "github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/chimera-rpg/go-editor/data"
	sdata "github.com/chimera-rpg/go-server/data"
	log "github.com/sirupsen/logrus"
)

type Mapset struct {
	context                        Context
	filename                       string
	maps                           []*data.UnReMap
	currentMapIndex                int
	focusedY, focusedX, focusedZ   int
	focusedI                       int
	selectedCoords                 SelectedCoords
	selectingYStart, selectingYEnd int
	selectingXStart, selectingXEnd int
	selectingZStart, selectingZEnd int
	selectingCoords                SelectedCoords
	resizeL, resizeR               int32
	resizeT, resizeB               int32
	resizeU, resizeD               int32
	newH, newW, newD               int32
	newDataName, newName           string
	loreEditor, descEditor         imgui.TextEditor
	zoom                           int32
	showGrid                       bool
	showYGrids                     bool
	onionskin                      bool
	keepSameTile                   bool
	uniqueTileVisits               bool
	ShouldClose                    bool
	visitedCoords                  SelectedCoords // Coordinates visited during mouse drag.
	mouseHeld                      map[g.MouseButton]bool
	toolBinds                      map[g.MouseButton]int
	blockScroll                    bool // Block map scrolling (true if ctrl or alt is held)
	unsaved                        bool
}

func NewMapset(context Context, name string, maps map[string]*sdata.Map) *Mapset {
	m := &Mapset{
		filename:         name,
		zoom:             3.0,
		showGrid:         true,
		showYGrids:       false,
		onionskin:        true,
		keepSameTile:     true,
		uniqueTileVisits: true,
		newW:             1,
		newH:             1,
		newD:             1,
		context:          context,
		loreEditor:       imgui.NewTextEditor(),
		descEditor:       imgui.NewTextEditor(),
		mouseHeld:        make(map[g.MouseButton]bool),
		toolBinds:        make(map[g.MouseButton]int),
	}
	m.loreEditor.SetShowWhitespaces(false)
	m.descEditor.SetShowWhitespaces(false)

	m.bindMouseToTool(g.MouseButtonLeft, selectTool)
	m.bindMouseToTool(g.MouseButtonMiddle, eraseTool)
	m.bindMouseToTool(g.MouseButtonRight, insertTool)

	for k, v := range maps {
		m.maps = append(m.maps, data.NewUnReMap(v, k))
	}

	m.selectedCoords.Clear()
	m.selectingCoords.Clear()
	m.visitedCoords.Clear()

	return m
}

func (m *Mapset) getMapPointFromMouse(p image.Point) (h image.Point, err error) {
	dm := m.context.DataManager()
	sm := m.CurrentMap()

	scale := float64(m.zoom)
	padding := 4
	tWidth := int(dm.AnimationsConfig.TileWidth)
	tHeight := int(dm.AnimationsConfig.TileHeight)

	hitX := int(float64(p.X) / scale)
	hitY := int(float64(p.Y) / scale)

	xOffset := m.focusedY*int(dm.AnimationsConfig.YStep.X) + padding
	yOffset := m.focusedY*int(dm.AnimationsConfig.YStep.Y) + padding + (sm.Get().Height * int(-dm.AnimationsConfig.YStep.Y))

	nearestX := (hitX+xOffset)/tWidth - 1
	nearestY := (hitY - yOffset) / tHeight
	if nearestX >= 0 && nearestX < sm.Get().Width && nearestY >= 0 && nearestY < sm.Get().Depth {
		h.X = nearestX
		h.Y = nearestY
	} else {
		err = errors.New("Point OOB")
	}
	return
}

func (m *Mapset) saveAll() {
	maps := make(map[string]*sdata.Map)
	for _, v := range m.maps {
		if v.Unsaved() {
			v.Save()
		}
		maps[v.DataName()] = v.SavedMap()
	}
	err := m.context.DataManager().SaveMap(m.filename, maps)
	if err != nil {
		m.unsaved = true
		log.Println(err)
		// TODO: Report error to the user.
		return
	}
	m.unsaved = false
	// TODO: Some sort of UI notification.
}

func (m *Mapset) close() {
	m.ShouldClose = true
}

func (m *Mapset) resizeMap(u, d, l, r, t, b int) {
	cm := m.CurrentMap()
	nH := cm.Get().Height + u + d
	nW := cm.Get().Width + l + r
	nD := cm.Get().Depth + t + b
	offsetY := d
	offsetX := l
	offsetZ := t
	// Make a new map according to the given dimensions
	newMap := m.createMap(
		cm.Get().Name,
		cm.Get().Description,
		cm.Get().Lore,
		cm.Get().Darkness,
		cm.Get().ResetTime,
		nH,
		nW,
		nD,
	)
	// Create the new map according to dimensions.
	for y := 0; y < nH; y++ {
		newMap.Tiles = append(newMap.Tiles, [][][]sdata.Archetype{})
		for x := 0; x < nW; x++ {
			newMap.Tiles[y] = append(newMap.Tiles[y], [][]sdata.Archetype{})
			for z := 0; z < nD; z++ {
				newMap.Tiles[y][x] = append(newMap.Tiles[y][x], []sdata.Archetype{})
			}
		}
	}
	// Iterate through old map tiles and copy what is in range.
	for y := 0; y < cm.Get().Height; y++ {
		if y+offsetY < 0 || y+offsetY >= newMap.Height {
			continue
		}
		for x := 0; x < cm.Get().Width; x++ {
			if x+offsetX < 0 || x+offsetX >= newMap.Width {
				continue
			}
			for z := 0; z < cm.Get().Depth; z++ {
				if z+offsetZ < 0 || z+offsetZ >= newMap.Depth {
					continue
				}
				newMap.Tiles[y+offsetY][x+offsetX][z+offsetZ] = cm.Get().Tiles[y][x][z]
			}
		}
	}
	cm.Set(newMap)
}

func (m *Mapset) insertArchetype(t *sdata.Map, arch string, y, x, z, pos int) error {
	tiles := m.getTiles(t, y, x, z)
	if tiles == nil {
		return errors.New("tile OOB")
	}

	if pos == -1 {
		pos = len(*tiles)
	}
	if pos == -1 {
		pos = 0
	}

	archetype := sdata.Archetype{
		Archs: []string{arch},
	}
	*tiles = append((*tiles)[:pos], append([]sdata.Archetype{archetype}, (*tiles)[pos:]...)...)

	return nil
}

func (m *Mapset) removeArchetype(t *sdata.Map, y, x, z, pos int) error {
	tiles := m.getTiles(t, y, x, z)
	if tiles == nil {
		return errors.New("tile OOB")
	}
	if pos == -1 {
		pos = len(*tiles) - 1
	}
	if pos == -1 {
		pos = 0
	}

	if pos >= len(*tiles) {
		return errors.New("pos OOB")
	}

	*tiles = append((*tiles)[:pos], (*tiles)[pos+1:]...)

	return nil
}

func (m *Mapset) getTiles(t *sdata.Map, y, x, z int) *[]sdata.Archetype {
	if len(t.Tiles) > y && y >= 0 {
		if len(t.Tiles[y]) > x && x >= 0 {
			if len(t.Tiles[y][x]) > z && z >= 0 {
				return &t.Tiles[y][x][z]
			}
		}
	}
	return nil
}

func (m *Mapset) createMap(name, desc, lore string, darkness, resettime int, h, w, d int) *sdata.Map {
	// Make a new map according to the given dimensions
	newMap := &sdata.Map{
		Name:        name,
		Description: desc,
		Darkness:    darkness,
		ResetTime:   resettime,
		Lore:        lore,
		Height:      h,
		Width:       w,
		Depth:       d,
	}
	// Create the new map according to dimensions.
	for y := 0; y < h; y++ {
		newMap.Tiles = append(newMap.Tiles, [][][]sdata.Archetype{})
		for x := 0; x < w; x++ {
			newMap.Tiles[y] = append(newMap.Tiles[y], [][]sdata.Archetype{})
			for z := 0; z < d; z++ {
				newMap.Tiles[y][x] = append(newMap.Tiles[y][x], []sdata.Archetype{})
			}
		}
	}
	return newMap
}

func (m *Mapset) deleteMap(index int) {
	if index >= len(m.maps) || index < 0 {
		return
	}
	m.maps = append(m.maps[:index], m.maps[index+1:]...)
	if m.currentMapIndex > index {
		m.currentMapIndex--
	}
	if m.currentMapIndex < 0 {
		m.currentMapIndex = 0
	}
}

func (m *Mapset) CurrentMap() *data.UnReMap {
	if m.currentMapIndex < 0 || m.currentMapIndex >= len(m.maps) {
		return nil
	}
	return m.maps[m.currentMapIndex]
}

func (m *Mapset) Unsaved() bool {
	if m.unsaved {
		return true
	}
	for _, v := range m.maps {
		if v.Unsaved() {
			return true
		}
	}
	return false
}
