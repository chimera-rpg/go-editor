package mapview

import (
	"errors"
	"fmt"

	g "github.com/AllenDang/giu"
	"github.com/chimera-rpg/go-editor/data"
	"github.com/chimera-rpg/go-editor/widgets"
	sdata "github.com/chimera-rpg/go-server/data"
	log "github.com/sirupsen/logrus"
)

type ButtonState = uint8

const (
	Up      ButtonState = 0 // Released
	Down                = 1 // Pressed
	Held                = 2 // Held
	Trigger             = 3 // Triggered by direct call
)

const (
	noTool = iota
	selectTool
	cselectTool
	lselectTool
	wandTool
	insertTool
	pickTool
	eraseTool
	fillTool
)

func (m *Mapset) bindMouseToTool(btn g.MouseButton, toolIndex int) {
	// Remove old btn bind.
	delete(m.toolBinds, btn)
	// Find and remove binding for this tool.
	for k, v := range m.toolBinds {
		if v == toolIndex {
			delete(m.toolBinds, k)
		}
	}
	// Set btn bind to this one.
	m.toolBinds[btn] = toolIndex
}

func (m *Mapset) getMouseTool(btn g.MouseButton) int {
	if toolIndex, ok := m.toolBinds[btn]; ok {
		return toolIndex
	}
	return noTool
}

func (m *Mapset) isToolBound(toolIndex int) bool {
	for _, v := range m.toolBinds {
		if v == toolIndex {
			return true
		}
	}
	return false
}

func (m *Mapset) getToolButtonString(toolIndex int) string {
	for k, v := range m.toolBinds {
		if v == toolIndex {
			if k == g.MouseButtonLeft {
				return "L"
			} else if k == g.MouseButtonMiddle {
				return "M"
			} else if k == g.MouseButtonRight {
				return "R"
			}
		}
	}
	return "_"
}

func (m *Mapset) handleMouseTool(btn g.MouseButton, state ButtonState, y, x, z int) error {
	if toolIndex, ok := m.toolBinds[btn]; ok {
		if m.currentMapIndex < 0 || m.currentMapIndex >= len(m.maps) {
			return errors.New("no current map")
		}
		cm := m.maps[m.currentMapIndex]

		if m.uniqueTileVisits {
			if state == Down || state == Held {
				if m.visitedCoords.Selected(y, x, z) {
					return nil
				} else {
					m.visitedCoords.Select(y, x, z)
				}
			} else if state == Up {
				m.visitedCoords.Clear()
			}
		}

		if toolIndex == insertTool {
			return m.toolInsert(state, cm, y, x, z)
		} else if toolIndex == selectTool {
			return m.toolSelect(state, selectTool, cm, y, x, z)
		} else if toolIndex == cselectTool {
			return m.toolSelect(state, cselectTool, cm, y, x, z)
		} else if toolIndex == lselectTool {
			return m.toolSelect(state, lselectTool, cm, y, x, z)
		} else if toolIndex == wandTool {
			return m.toolSelect(state, wandTool, cm, y, x, z)
		} else if toolIndex == eraseTool {
			return m.toolErase(state, cm, y, x, z)
		} else if toolIndex == fillTool {
			return m.toolFill(state, cm, y, x, z)
		} else if toolIndex == pickTool {
			return m.toolPick(state, cm, y, x, z)
		}
	}
	return nil
}

func (m *Mapset) toolSelect(state ButtonState, subTool int, v *data.UnReMap, y, x, z int) (err error) {
	insertMode := 0 // replace
	widgets.KeyBinds(0,
		widgets.KeyBind(widgets.KeyBindFlagPressed, widgets.Keys(widgets.KeyShift), nil, func() {
			insertMode = 1 // append
		}),
		widgets.KeyBind(widgets.KeyBindFlagPressed, widgets.Keys(widgets.KeyControl), nil, func() {
			insertMode = 2 // remove
		}),
	).Build()

	if state == Down {
		m.selectingYStart, m.selectingYEnd = y, y
		m.selectingXStart, m.selectingXEnd = x, x
		m.selectingZStart, m.selectingZEnd = z, z
		m.selectingCoords.Clear()
		if subTool == cselectTool {
			m.selectingCoords.RangeCircle(true, m.selectingYStart, m.selectingXStart, m.selectingZStart, m.selectingYEnd, m.selectingXEnd, m.selectingZEnd)
		} else if subTool == lselectTool {
			m.selectingCoords.Line(true, m.selectingYStart, m.selectingXStart, m.selectingZStart, m.selectingYEnd, m.selectingXEnd, m.selectingZEnd)
		} else if subTool == wandTool {
			m.selectingCoords.FloodSelect(true, m.selectingYStart, m.selectingXStart, m.selectingZStart, m)
		} else {
			m.selectingCoords.Range(true, m.selectingYStart, m.selectingXStart, m.selectingZStart, m.selectingYEnd, m.selectingXEnd, m.selectingZEnd)
		}
	} else if state == Held {
		m.selectingYEnd = y
		m.selectingXEnd = x
		m.selectingZEnd = z
		m.selectingCoords.Clear()
		if subTool == cselectTool {
			m.selectingCoords.RangeCircle(true, m.selectingYStart, m.selectingXStart, m.selectingZStart, m.selectingYEnd, m.selectingXEnd, m.selectingZEnd)
		} else if subTool == lselectTool {
			m.selectingCoords.Line(true, m.selectingYStart, m.selectingXStart, m.selectingZStart, m.selectingYEnd, m.selectingXEnd, m.selectingZEnd)
		} else if subTool == wandTool {
			// TODO
		} else {
			m.selectingCoords.Range(true, m.selectingYStart, m.selectingXStart, m.selectingZStart, m.selectingYEnd, m.selectingXEnd, m.selectingZEnd)
		}
	} else if state == Up {
		m.selectingYEnd = y
		m.selectingXEnd = x
		m.selectingZEnd = z
		if subTool == cselectTool {
			m.selectingCoords.Clear()
			m.selectingCoords.RangeCircle(true, m.selectingYStart, m.selectingXStart, m.selectingZStart, m.selectingYEnd, m.selectingXEnd, m.selectingZEnd)
		} else if subTool == lselectTool {
			m.selectingCoords.Clear()
			m.selectingCoords.Line(true, m.selectingYStart, m.selectingXStart, m.selectingZStart, m.selectingYEnd, m.selectingXEnd, m.selectingZEnd)
		} else if subTool == wandTool {
		} else {
			m.selectingCoords.Clear()
			m.selectingCoords.Range(true, m.selectingYStart, m.selectingXStart, m.selectingZStart, m.selectingYEnd, m.selectingXEnd, m.selectingZEnd)
		}

		if insertMode == 0 { // replace
			m.selectedCoords.Set(m.selectingCoords)
		} else if insertMode == 1 { // append
			m.selectedCoords.Add(m.selectingCoords)
		} else if insertMode == 2 { // remove
			m.selectedCoords.Remove(m.selectingCoords)
		}
		m.selectingCoords.Clear()
		m.selectingYStart, m.selectingYEnd, m.selectingXStart, m.selectingXEnd, m.selectingZEnd, m.selectingZStart = -1, -1, -1, -1, -1, -1
		// And set focused to last coords.
		m.moveCursor(y, x, z, m.focusedI)
	} else {
		// TODO: Check if Shift or Ctrl is held!
		m.selectedCoords.Clear()
		m.selectedCoords.Select(y, x, z)
		m.moveCursor(y, x, z, m.focusedI)
	}
	return
}

func (m *Mapset) toolInsert(state ButtonState, v *data.UnReMap, y, x, z int) (err error) {
	fmt.Printf("insert %s\n", m.context.SelectedArch())
	// Bail if no archetype is selected.
	if m.context.SelectedArch() == "" {
		return
	}
	// Check if we should not insert if top tile is the same.
	if m.keepSameTile {
		tiles := m.getTiles(v.Get(), y, x, z)
		if tiles != nil && len(*tiles) > 0 {
			if (*tiles)[len(*tiles)-1].Arch == m.context.SelectedArch() {
				return
			}
			for _, a := range (*tiles)[len(*tiles)-1].Archs {
				if a == m.context.SelectedArch() {
					return
				}
			}
		}
	}
	// Otherwise attempt to insert.
	clone := v.Clone()
	if err := m.insertArchetype(clone, m.context.SelectedArch(), y, x, z, -1); err != nil {
		return err
	}
	v.Set(clone)
	return
}

func (m *Mapset) toolErase(state ButtonState, v *data.UnReMap, y, x, z int) (err error) {
	if state == Down {
		clone := v.Clone()
		if err := m.removeArchetype(clone, y, x, z, -1); err != nil {
			return err
		}
		v.Set(clone)
	} else if state == Trigger {
		clone := v.Clone()
		changed := false
		for coord := range m.selectedCoords.Get() {
			y, x, z := coord[0], coord[1], coord[2]
			if err := m.removeArchetype(clone, y, x, z, -1); err != nil {
				log.Println(err)
				continue
			} else {
				changed = true
			}
		}
		if changed {
			v.Set(clone)
		}
	}
	return
}

func (m *Mapset) remove(v *data.UnReMap, y, x, z, i int) (err error) {
	clone := v.Clone()
	changed := false
	if err := m.removeArchetype(clone, y, x, z, i); err != nil {
		log.Println(err)
		return err
	} else {
		changed = true
	}
	if changed {
		v.Set(clone)
	}
	return nil
}

func (m *Mapset) move(v *data.UnReMap, y1, x1, z1, p1, y2, x2, z2, p2 int) (err error) {
	clone := v.Clone()
	tiles1 := m.getTiles(clone, y1, x1, z1)
	if tiles1 == nil {
		return errors.New("tile OOB")
	}
	if p1 == -1 {
		p1 = len(*tiles1) - 1
	}
	if p1 == -1 {
		p1 = 0
	}
	if p1 >= len(*tiles1) {
		return errors.New("pos OOB")
	}

	tiles2 := m.getTiles(clone, y2, x2, z2)
	if tiles2 == nil {
		return errors.New("tile OOB")
	}
	if p2 == -1 {
		p2 = len(*tiles2)
	}
	if p2 == -1 {
		p2 = 0
	}
	if p2 > len(*tiles2) {
		return errors.New("pos OOB")
	}

	a := (*tiles1)[p1]
	m.removeArchetype(clone, y1, x1, z1, p1)
	if len(*tiles2) == p2 {
		*tiles2 = append(*tiles2, a)
	} else {
		*tiles2 = append((*tiles2)[:p2], append([]sdata.Archetype{a}, (*tiles2)[p2:]...)...)
	}

	v.Set(clone)

	return nil
}

func (m *Mapset) toolFill(state ButtonState, v *data.UnReMap, y, x, z int) (err error) {
	// Bail if no archetype is selected.
	if m.context.SelectedArch() == "" {
		return
	}
	if state == Trigger || state == Up {
		clone := v.Clone()
		changed := false
		for coord := range m.selectedCoords.Get() {
			y, x, z := coord[0], coord[1], coord[2]

			place := true
			// Check if we should not insert if top tile is the same.
			if m.keepSameTile {
				tiles := m.getTiles(v.Get(), y, x, z)
				if tiles != nil && len(*tiles) > 0 {
					if (*tiles)[len(*tiles)-1].Arch == m.context.SelectedArch() {
						place = false
					}
					if place {
						for _, a := range (*tiles)[len(*tiles)-1].Archs {
							if a == m.context.SelectedArch() {
								place = false
								break
							}
						}
					}
				}
			}
			if place {
				if err := m.insertArchetype(clone, m.context.SelectedArch(), y, x, z, -1); err != nil {
					log.Println(err)
					continue
				} else {
					changed = true
				}
			}
		}
		if changed {
			v.Set(clone)
		}
	}
	return
}

func (m *Mapset) toolPick(state ButtonState, v *data.UnReMap, y, x, z int) (err error) {
	if state == Trigger || state == Up {
		tiles := m.getTiles(v.Get(), y, x, z)
		if len(*tiles) > 0 {
			a := (*tiles)[len(*tiles)-1]
			arch := a.Arch
			if arch == "" {
				if len(a.Archs) > 0 {
					arch = a.Archs[0]
				}
			}
			if arch != "" {
				m.context.SetSelectedArch(arch)
			}
		}
	}
	return
}
