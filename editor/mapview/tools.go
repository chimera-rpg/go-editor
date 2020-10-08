package mapview

import (
	"errors"

	g "github.com/AllenDang/giu"
	"github.com/chimera-rpg/go-editor/data"
)

type ButtonState = uint8

const (
	Up      ButtonState = 0 // Released
	Down                = 1 // Pressed
	Held                = 2 // Held
	Trigger             = 3 // Triggered by direct call
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

		if toolIndex == insertTool {
			return m.toolInsert(state, cm, y, x, z)
		} else if toolIndex == selectTool {
			return m.toolSelect(state, cm, y, x, z)
		} else if toolIndex == eraseTool {
			return m.toolErase(state, cm, y, x, z)
		}
	}
	return nil
}

func (m *Mapset) toolSelect(state ButtonState, v *data.UnReMap, y, x, z int) (err error) {
	// TODO: Check if Shift or Ctrl is held!
	m.selectedCoords.Clear()
	m.selectedCoords.Select(y, x, z)
	m.focusedY = y
	m.focusedX = x
	m.focusedZ = z
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
	clone := v.Clone()
	if err := m.removeArchetype(clone, y, x, z, -1); err != nil {
		return err
	}
	v.Set(clone)
	return
}
