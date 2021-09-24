package mapview

import (
	"errors"

	g "github.com/AllenDang/giu"
	"github.com/chimera-rpg/go-editor/data"
	"github.com/chimera-rpg/go-editor/widgets"
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
		} else if toolIndex == eraseTool {
			return m.toolErase(state, cm, y, x, z)
		} else if toolIndex == fillTool {
			return m.toolFill(state, cm, y, x, z)
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
		} else {
			m.selectingCoords.Range(true, m.selectingYStart, m.selectingXStart, m.selectingZStart, m.selectingYEnd, m.selectingXEnd, m.selectingZEnd)
		}
	} else if state == Up {
		m.selectingYEnd = y
		m.selectingXEnd = x
		m.selectingZEnd = z
		m.selectingCoords.Clear()
		if subTool == cselectTool {
			m.selectingCoords.RangeCircle(true, m.selectingYStart, m.selectingXStart, m.selectingZStart, m.selectingYEnd, m.selectingXEnd, m.selectingZEnd)
		} else if subTool == lselectTool {
			m.selectingCoords.Line(true, m.selectingYStart, m.selectingXStart, m.selectingZStart, m.selectingYEnd, m.selectingXEnd, m.selectingZEnd)
		} else {
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
		m.focusedY = y
		m.focusedX = x
		m.focusedZ = z
	} else {
		// TODO: Check if Shift or Ctrl is held!
		m.selectedCoords.Clear()
		m.selectedCoords.Select(y, x, z)
		m.focusedY = y
		m.focusedX = x
		m.focusedZ = z
	}
	return
}

func (m *Mapset) toolInsert(state ButtonState, v *data.UnReMap, y, x, z int) (err error) {
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
