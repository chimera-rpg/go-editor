package mapview

import "math"

// Coords is our type alias to [y, x, z]
type Coords = [3]int

// SelectedCoords is a simple structure that provides functionality to select/unselect arbitrary coordinates.
type SelectedCoords struct {
	selected map[Coords]struct{}
}

// Get returns the underlying map.
func (s *SelectedCoords) Get() map[[3]int]struct{} {
	return s.selected
}

// Select selects the given coordinates.
func (s *SelectedCoords) Select(y, x, z int) {
	s.selected[Coords{y, x, z}] = struct{}{}
}

// Unselect unselects the given coordinates.
func (s *SelectedCoords) Unselect(y, x, z int) {
	delete(s.selected, Coords{y, x, z})
}

// Clear clears all selected coordinates.
func (s *SelectedCoords) Clear() {
	s.selected = make(map[Coords]struct{})
}

// Selected checks if the coordinate is selected.
func (s *SelectedCoords) Selected(y, x, z int) bool {
	if _, ok := s.selected[Coords{y, x, z}]; ok {
		return true
	}
	return false
}

// Unselected checks if the coordinate is unselected.
func (s *SelectedCoords) Unselected(y, x, z int) bool {
	return !s.Selected(y, x, z)
}

// Set sets the coords to match the selection of another.
func (s *SelectedCoords) Set(o SelectedCoords) {
	s.Clear()
	for k := range o.selected {
		s.selected[k] = struct{}{}
	}
}

// Add adds the coords of another to itself.
func (s *SelectedCoords) Add(o SelectedCoords) {
	for k := range o.selected {
		s.selected[k] = struct{}{}
	}
}

// Remove adds the coords of another to itself.
func (s *SelectedCoords) Remove(o SelectedCoords) {
	for k := range o.selected {
		delete(s.selected, k)
	}
}

// Range selects or unselects between 2 coordinates.
func (s *SelectedCoords) Range(doSelect bool, y1, x1, z1, y2, x2, z2 int) {
	ymin := int(math.Min(float64(y1), float64(y2)))
	ymax := int(math.Max(float64(y1), float64(y2)))
	xmin := int(math.Min(float64(x1), float64(x2)))
	xmax := int(math.Max(float64(x1), float64(x2)))
	zmin := int(math.Min(float64(z1), float64(z2)))
	zmax := int(math.Max(float64(z1), float64(z2)))
	for y := ymin; y <= ymax; y++ {
		for x := xmin; x <= xmax; x++ {
			for z := zmin; z <= zmax; z++ {
				if doSelect {
					s.Select(y, x, z)
				} else {
					s.Unselect(y, x, z)
				}
			}
		}
	}
}
