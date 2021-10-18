package mapview

import (
	"math"

	g "github.com/AllenDang/giu"
)

type SelectionWidget struct {
	grow                int32
	growDiagonal        bool
	growY, growX, growZ bool
	border              int32
	inner               bool
	edges               bool
}

func (s *SelectionWidget) ResetResize() {
	s.grow = 0
	s.growDiagonal = true
	s.growY = false
	s.growX = true
	s.growZ = true
}

func (s *SelectionWidget) ResetBorderify() {
	s.border = 1
	s.inner = false
	s.edges = true
}

func (s *SelectionWidget) Draw(m *Mapset) (l g.Layout) {
	l = g.Layout{
		g.Label("Grow/Shrink"),
		g.Child().Size(-1, 110).Layout(
			// Grow and Shrink
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
			g.InputInt(&s.border).Size(50).Label("Width"),
			g.Tooltip("The width of the border"),
			g.Checkbox("Inner", &s.inner),
			g.Tooltip("Use the inside of the selection as the border."),
			g.Checkbox("Corners", &s.edges),
			g.Tooltip("Whether or not corners are considered for border selection."),
			g.Row(
				g.Button("Reset").OnClick(func() {
					s.ResetBorderify()
				}),
				g.Button("Apply").OnClick(func() {
					m.selectedCoords.Border(int(s.border), s.inner, s.edges)
					// TODO: Grow m.selectedCoords
				}),
			),
		),
	}
	return
}
