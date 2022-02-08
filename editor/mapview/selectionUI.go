package mapview

import (
	"math"

	g "github.com/AllenDang/giu"
	"github.com/chimera-rpg/go-editor/editor/icons"
	sdata "github.com/chimera-rpg/go-server/data"
)

type SelectionWidget struct {
	grow                   int32
	growDiagonal           bool
	growY, growX, growZ    bool
	checkY, checkX, checkZ bool
	outer                  bool
	edges                  bool
	replaceFocusedIndex    bool
	replaceTopmost         bool
	replaceOverwrite       bool
	replaceFocused         bool
}

func (s *SelectionWidget) Reset() {
	s.ResetResize()
	s.ResetBorderify()
	s.ResetReplace()
}

func (s *SelectionWidget) ResetResize() {
	s.grow = 1
	s.growDiagonal = true
	s.growY = false
	s.growX = true
	s.growZ = true
}

func (s *SelectionWidget) ResetBorderify() {
	s.outer = false
	s.edges = true
	s.checkX = true
	s.checkZ = true
}

func (s *SelectionWidget) ResetReplace() {
	s.replaceFocused = true
	s.replaceTopmost = false
	s.replaceFocusedIndex = false
	s.replaceOverwrite = true
}

func (s *SelectionWidget) Draw(m *Mapset) (l g.Layout) {
	l = g.Layout{
		// Move
		g.Label("Move"),
		g.Child().Size(-1, 110).Border(false).Layout(
			g.Row(
				g.ImageButton(icons.Textures["tl"].Texture).Size(30, 30).FramePadding(0).OnClick(func() {
					m.selectedCoords.Shift(0, -1, -1)
				}),
				g.ImageButton(icons.Textures["t"].Texture).Size(30, 30).FramePadding(0).OnClick(func() {
					m.selectedCoords.Shift(0, 0, -1)
				}),
				g.ImageButton(icons.Textures["tr"].Texture).Size(30, 30).FramePadding(0).OnClick(func() {
					m.selectedCoords.Shift(0, 1, -1)
				}),
				g.ImageButton(icons.Textures["u"].Texture).Size(30, 30).FramePadding(0).OnClick(func() {
					m.selectedCoords.Shift(1, 0, 0)
				}),
			),
			g.Row(
				g.ImageButton(icons.Textures["l"].Texture).Size(30, 30).FramePadding(0).OnClick(func() {
					m.selectedCoords.Shift(0, -1, 0)
				}),
				g.ImageButton(icons.Textures["blank"].Texture).Size(30, 30).FramePadding(0),
				g.ImageButton(icons.Textures["r"].Texture).Size(30, 30).FramePadding(0).OnClick(func() {
					m.selectedCoords.Shift(0, 1, 0)
				}),
			),
			g.Row(
				g.ImageButton(icons.Textures["bl"].Texture).Size(30, 30).FramePadding(0).OnClick(func() {
					m.selectedCoords.Shift(0, -1, 1)
				}),
				g.ImageButton(icons.Textures["b"].Texture).Size(30, 30).FramePadding(0).OnClick(func() {
					m.selectedCoords.Shift(0, 0, 1)
				}),
				g.ImageButton(icons.Textures["br"].Texture).Size(30, 30).FramePadding(0).OnClick(func() {
					m.selectedCoords.Shift(0, 1, 1)
				}),
				g.ImageButton(icons.Textures["d"].Texture).Size(30, 30).FramePadding(0).OnClick(func() {
					m.selectedCoords.Shift(-1, 0, 0)
				}),
			),
		),
		// Grow and Shrink
		g.Label("Grow/Shrink"),
		g.Child().Size(-1, 110).Layout(
			g.InputInt(&s.grow).Size(50).Label("Size"),
			g.Tooltip("The growth or shrink size."),
			g.Checkbox("diagonal", &s.growDiagonal),
			g.Tooltip("Whether it grows diagonal."),
			g.Row(
				g.Checkbox("Y", &s.growY),
				g.Tooltip("Whether it grows height (Y)."),
				g.Checkbox("X", &s.growX),
				g.Tooltip("Whether it grows width (X)."),
				g.Checkbox("Z", &s.growZ),
				g.Tooltip("Whether it grows depth (Z)."),
			),
			g.Row(
				g.Button("Reset").OnClick(func() {
					s.ResetResize()
				}),
				g.Button("Apply").OnClick(func() {
					if s.grow < 0 {
						grow := int(math.Abs(float64(s.grow)))
						m.selectedCoords.Grow(grow, false, s.growDiagonal, s.growY, s.growX, s.growZ)
					} else if s.grow > 0 {
						m.selectedCoords.Grow(int(s.grow), true, s.growDiagonal, s.growY, s.growX, s.growZ)
					}
				}),
			),
		),
		// Borderify
		g.Label("Borderify"),
		g.Child().Size(-1, 110).Layout(
			g.Checkbox("Outer", &s.outer),
			g.Tooltip("Use the outside of the selection as the border."),
			g.Checkbox("Corners", &s.edges),
			g.Tooltip("Whether or not corners are considered for border selection."),
			g.Row(
				g.Checkbox("Y", &s.checkY),
				g.Tooltip("Whether Y coordinates are checked for bordeirng."),
				g.Checkbox("X", &s.checkX),
				g.Tooltip("Whether X coordinates are checked for bordeirng."),
				g.Checkbox("Z", &s.checkZ),
				g.Tooltip("Whether Z coordinates are checked for bordeirng."),
			),
			g.Row(
				g.Button("Reset").OnClick(func() {
					s.ResetBorderify()
				}),
				g.Button("Apply").OnClick(func() {
					m.selectedCoords.Border(s.outer, s.edges, s.checkY, s.checkX, s.checkZ)
					// TODO: Grow m.selectedCoords
				}),
			),
		),
		// Replace
		g.Label("Replace"),
		g.Child().Size(-1, 110).Layout(
			g.Checkbox("Match Focused", &s.replaceFocused),
			g.Tooltip("Replace archetypes matching the focused archetype."),
			g.Checkbox("Focused Index Only", &s.replaceFocusedIndex),
			g.Tooltip("Only replace archetypes at the focused stack index."),
			g.Checkbox("Topmost", &s.replaceTopmost),
			g.Tooltip("Only replace the topmost archetype in the tile."),
			g.Checkbox("Overwrite", &s.replaceOverwrite),
			g.Tooltip("Overwrite archetype completely."),
			g.Row(
				g.Button("Reset").OnClick(func() {
					s.ResetReplace()
				}),
				g.Button("Replace").OnClick(func() {
					pos := 0
					if s.replaceFocused {
						pos = -2
					} else if s.replaceFocusedIndex {
						pos = m.focusedI
					} else if s.replaceTopmost {
						pos = -1
					}
					var match *sdata.Archetype
					if s.replaceFocused {
						focusedTiles := m.getTiles(m.CurrentMap().Get(), m.focusedY, m.focusedX, m.focusedZ)
						if m.focusedI >= 0 && m.focusedI < len(*focusedTiles) {
							match = &(*focusedTiles)[m.focusedI]
						}
					}
					m.replace(m.CurrentMap(), match, pos, s.replaceOverwrite)
				}),
			),
		),
	}
	return
}
