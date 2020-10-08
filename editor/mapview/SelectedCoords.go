package mapview

// SelectedCoords is a simple structure that provides functionality to select/unselect arbitrary coordinates.
type SelectedCoords struct {
	selected map[[3]int]struct{}
}

// Get returns the underlying map.
func (s *SelectedCoords) Get() map[[3]int]struct{} {
	return s.selected
}

// Select selects the given coordinates.
func (s *SelectedCoords) Select(y, x, z int) {
	s.selected[[3]int{y, x, z}] = struct{}{}
}

// Unselect unselects the given coordinates.
func (s *SelectedCoords) Unselect(y, x, z int) {
	delete(s.selected, [3]int{y, x, z})
}

// Clear clears all selected coordinates.
func (s *SelectedCoords) Clear() {
	s.selected = make(map[[3]int]struct{})
}
