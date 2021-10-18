package mapview

import (
	"fmt"
	"math"

	"github.com/chimera-rpg/go-server/data"
)

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

// Empty checks if coordinates are selected.
func (s *SelectedCoords) Empty() bool {
	return len(s.selected) == 0
}

// Count returns the count of selected coordinates.
func (s *SelectedCoords) Count() int {
	return len(s.selected)
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

func (s *SelectedCoords) Clone() SelectedCoords {
	s2 := SelectedCoords{
		selected: make(map[[3]int]struct{}),
	}
	for k := range s.selected {
		s2.selected[k] = struct{}{}
	}
	return s2
}

func (s *SelectedCoords) Line(doSelect bool, y1, x1, z1, y2, x2, z2 int) {
	dx := int(math.Abs(float64(x2 - x1)))
	dz := int(math.Abs(float64(z2 - z1)))
	sx := -1
	if x1 < x2 {
		sx = 1
	}
	sz := -1
	if z1 < z2 {
		sz = 1
	}
	v := dx - dz

	for {
		if doSelect {
			s.Select(y1, x1, z1)
		} else {
			s.Unselect(y1, x1, z1)
		}

		if (x1 == x2) && (z1 == z2) {
			break
		}
		v2 := v * 2
		if v2 > -dz {
			v -= dz
			x1 += sx
		}
		if v2 < dx {
			v += dx
			z1 += sz
		}
	}
}

// Range selects or unselects between 2 coordinates.
// TODO: Add a doOutline bool for only doing outer edge for range.
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

func (s *SelectedCoords) RangeCircle(doSelect bool, y1, x1, z1, y2, x2, z2 int) {
	/*ymin := int(math.Min(float64(y1), float64(y2)))
	ymax := int(math.Max(float64(y1), float64(y2)))
	xmin := int(math.Min(float64(x1), float64(x2)))
	xmax := int(math.Max(float64(x1), float64(x2)))
	zmin := int(math.Min(float64(z1), float64(z2)))
	zmax := int(math.Max(float64(z1), float64(z2)))

	yRadius := ymax - ymin
	if yRadius == 0 {
		yRadius = 1
	}
	xRadius := xmax - xmin
	zRadius := zmax - zmin

	rx := (x2 - x1) / 2
	ry := (y2 - y1) / 2
	rz := (z2 - z1) / 2
	cx := (x2 + x1) / 2
	cy := (y2 + y1) / 2
	cz := (z2 + z1) / 2

	for a := 0; a < math.Pi; a++ {
		x := cx + math.Cos(a)*rx
		z := cz + math.Sin(a)*rz
		//y := cy + sin(a) * ry
		if a > 0 {
			// TODO: Draw line.
		}
	}*/
}

func (s *SelectedCoords) FloodSelect(doSelect bool, y1, x1, z1 int, m *Mapset) {
	sm := m.CurrentMap().Get()
	traveledTiles := make(map[Coords]struct{})
	t := m.getTiles(sm, y1, x1, z1)
	if t == nil {
		return
	}
	var target *data.Archetype
	if len(*t) > 0 {
		target = &((*t)[len(*t)-1])
	}
	var walk func(y, x, z int)
	walk = func(y, x, z int) {
		if _, ok := traveledTiles[Coords{y, x, z}]; ok {
			return
		}
		traveledTiles[Coords{y, x, z}] = struct{}{}

		tiles := m.getTiles(sm, y, x, z)
		if tiles == nil {
			return
		}

		var walkTarget *data.Archetype
		if len(*tiles) > 0 {
			walkTarget = &((*tiles)[len(*tiles)-1])
		}
		if walkTarget == nil && target == nil {
			if doSelect {
				s.Select(y, x, z)
			} else {
				s.Unselect(y, x, z)
			}
			// Now also iterate to others.
			walk(y, x+1, z)
			walk(y, x-1, z)
			walk(y, x, z+1)
			walk(y, x, z-1)
		} else if walkTarget != nil && target != nil && len(walkTarget.Archs) == len(target.Archs) {
			match := true
			for i, _ := range walkTarget.Archs {
				if walkTarget.Archs[i] != target.Archs[i] {
					match = false
				}
			}
			if walkTarget.Arch != target.Arch {
				match = false
			}
			if match {
				if doSelect {
					s.Select(y, x, z)
				} else {
					s.Unselect(y, x, z)
				}
				// Now also iterate to others.
				walk(y, x+1, z)
				walk(y, x-1, z)
				walk(y, x, z+1)
				walk(y, x, z-1)
			}
		}
	}
	fmt.Println("Okay, do a flood selection using", target)
	walk(y1, x1, z1)
}

func (s *SelectedCoords) getEmptyAdjacents(y, x, z int, diagonal, growY, growX, growZ bool) (c []Coords) {
	if growX {
		if _, ok := s.selected[Coords{y, x + 1, z}]; !ok {
			c = append(c, Coords{y, x + 1, z})
		}
		if _, ok := s.selected[Coords{y, x - 1, z}]; !ok {
			c = append(c, Coords{y, x - 1, z})
		}
	}
	if growZ {
		if _, ok := s.selected[Coords{y, x, z + 1}]; !ok {
			c = append(c, Coords{y, x, z + 1})
		}
		if _, ok := s.selected[Coords{y, x, z - 1}]; !ok {
			c = append(c, Coords{y, x, z - 1})
		}
	}
	if diagonal && growX && growZ {
		if _, ok := s.selected[Coords{y, x + 1, z + 1}]; !ok {
			c = append(c, Coords{y, x + 1, z + 1})
		}
		if _, ok := s.selected[Coords{y, x - 1, z - 1}]; !ok {
			c = append(c, Coords{y, x - 1, z - 1})
		}
		if _, ok := s.selected[Coords{y, x + 1, z - 1}]; !ok {
			c = append(c, Coords{y, x + 1, z - 1})
		}
		if _, ok := s.selected[Coords{y, x - 1, z + 1}]; !ok {
			c = append(c, Coords{y, x - 1, z + 1})
		}
	}
	if growY {
		if _, ok := s.selected[Coords{y + 1, x, z}]; !ok {
			c = append(c, Coords{y + 1, x, z})
		}
		if _, ok := s.selected[Coords{y - 1, x, z}]; !ok {
			c = append(c, Coords{y - 1, x, z})
		}
		if diagonal {
			if growX {
				if _, ok := s.selected[Coords{y + 1, x + 1, z}]; !ok {
					c = append(c, Coords{y + 1, x + 1, z})
				}
				if _, ok := s.selected[Coords{y + 1, x - 1, z}]; !ok {
					c = append(c, Coords{y + 1, x - 1, z})
				}
				if _, ok := s.selected[Coords{y - 1, x + 1, z}]; !ok {
					c = append(c, Coords{y - 1, x + 1, z})
				}
				if _, ok := s.selected[Coords{y - 1, x - 1, z}]; !ok {
					c = append(c, Coords{y - 1, x - 1, z})
				}
			}
			if growZ {
				if _, ok := s.selected[Coords{y + 1, x, z + 1}]; !ok {
					c = append(c, Coords{y + 1, x, z + 1})
				}
				if _, ok := s.selected[Coords{y + 1, x, z - 1}]; !ok {
					c = append(c, Coords{y + 1, x, z - 1})
				}
				if _, ok := s.selected[Coords{y - 1, x, z + 1}]; !ok {
					c = append(c, Coords{y - 1, x, z + 1})
				}
				if _, ok := s.selected[Coords{y - 1, x, z - 1}]; !ok {
					c = append(c, Coords{y - 1, x, z - 1})
				}
			}
			if growX && growZ {
				if _, ok := s.selected[Coords{y + 1, x + 1, z + 1}]; !ok {
					c = append(c, Coords{y + 1, x + 1, z + 1})
				}
				if _, ok := s.selected[Coords{y + 1, x - 1, z - 1}]; !ok {
					c = append(c, Coords{y + 1, x - 1, z - 1})
				}
				if _, ok := s.selected[Coords{y + 1, x + 1, z - 1}]; !ok {
					c = append(c, Coords{y + 1, x + 1, z - 1})
				}
				if _, ok := s.selected[Coords{y + 1, x - 1, z + 1}]; !ok {
					c = append(c, Coords{y + 1, x - 1, z + 1})
				}
				//
				if _, ok := s.selected[Coords{y - 1, x + 1, z + 1}]; !ok {
					c = append(c, Coords{y - 1, x + 1, z + 1})
				}
				if _, ok := s.selected[Coords{y - 1, x - 1, z - 1}]; !ok {
					c = append(c, Coords{y - 1, x - 1, z - 1})
				}
				if _, ok := s.selected[Coords{y - 1, x + 1, z - 1}]; !ok {
					c = append(c, Coords{y - 1, x + 1, z - 1})
				}
				if _, ok := s.selected[Coords{y - 1, x - 1, z + 1}]; !ok {
					c = append(c, Coords{y - 1, x - 1, z + 1})
				}
			}
		}
	}
	return
}

func (s *SelectedCoords) Grow(size int, doSelect bool, diagonal, growY, growX, growZ bool) {
	if size <= 0 {
		return
	}

	n := s.Clone()
	var targets []Coords
	for k := range n.selected {
		y := k[0]
		x := k[1]
		z := k[2]
		a := s.getEmptyAdjacents(y, x, z, diagonal, growY, growX, growZ)
		if doSelect {
			for _, c := range a {
				targets = append(targets, Coords{c[0], c[1], c[2]})
			}
		} else {
			if len(a) > 0 {
				targets = append(targets, Coords{y, x, z})
			}
		}
	}
	for _, k := range targets {
		if doSelect {
			s.Select(k[0], k[1], k[2])
		} else {
			s.Unselect(k[0], k[1], k[2])
		}
	}
	s.Grow(size-1, doSelect, diagonal, growY, growX, growZ)
}

func (s *SelectedCoords) Border(outer, edges bool, checkY, checkX, checkZ bool) {
	if outer {
		s.Grow(1, true, edges, true, true, true)
	}
	n := s.Clone()
	for k := range n.selected {
		a := n.getEmptyAdjacents(k[0], k[1], k[2], edges, checkY, checkX, checkZ)
		if len(a) == 0 {
			s.Unselect(k[0], k[1], k[2])
		}
	}
}

func (s *SelectedCoords) Shift(y, x, z int) {
	newSelection := make(map[Coords]struct{})
	for k := range s.selected {
		newSelection[Coords{k[0] + y, k[1] + x, k[2] + z}] = struct{}{}
	}
	s.selected = newSelection
}

func (s *SelectedCoords) ReplicateYSlice(doSelect bool, slice map[[3]int]struct{}, targetY int) {
	for yxz := range slice {
		if doSelect {
			s.Select(targetY, yxz[1], yxz[2])
		} else {
			s.Unselect(targetY, yxz[1], yxz[2])
		}
	}
}

func (s *SelectedCoords) GetYSlice(y int) map[[3]int]struct{} {
	results := make(map[Coords]struct{})
	for yxz := range s.Get() {
		if yxz[0] == y {
			results[yxz] = struct{}{}
		}
	}
	return results
}
