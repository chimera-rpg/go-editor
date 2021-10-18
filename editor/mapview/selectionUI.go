package mapview

import (
	"math"

	g "github.com/AllenDang/giu"
	"github.com/chimera-rpg/go-editor/editor/icons"
)

type SelectionWidget struct {
	grow                int32
	growDiagonal        bool
	growY, growX, growZ bool
	border              int32
	inner               bool
	edges               bool
}

func (s *SelectionWidget) Reset() {
	s.ResetResize()
	s.ResetBorderify()
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
