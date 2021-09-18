package mapview

import (
	"image"
	"image/color"
	"math"
	"sort"

	g "github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/chimera-rpg/go-editor/data"
)

type archDrawable struct {
	z    int
	x, y int
	w, h int
	t    *data.ImageTexture
	c    color.RGBA
}

func (m *Mapset) drawMap(v *data.UnReMap) {
	sm := v.Get()

	pos := g.GetCursorScreenPos()
	canvas := g.GetCanvas()
	dm := m.context.DataManager()
	scale := int(m.zoom)
	tWidth := int(dm.AnimationsConfig.TileWidth)
	tHeight := int(dm.AnimationsConfig.TileHeight)
	yStep := dm.AnimationsConfig.YStep
	padding := 4
	cWidth := sm.Width * tWidth
	cHeight := sm.Depth * tHeight

	canvasWidth := int((cWidth + (sm.Height * int(yStep.X)) + padding*2) * scale)
	canvasHeight := int((cHeight + (sm.Height * int(-yStep.Y)) + padding*2) * scale)

	imgui.BeginChildV("map", imgui.Vec2{X: float32(canvasWidth), Y: float32(canvasHeight)}, false, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoMouseInputs)

	startX := padding
	startY := padding + (sm.Height * int(-yStep.Y))

	col := color.RGBA{0, 0, 0, 255}
	canvas.AddRectFilled(pos, pos.Add(image.Pt(canvasWidth, canvasHeight)), col, 0, 0)

	col = color.RGBA{255, 255, 255, 255}
	var drawables []archDrawable
	// Draw archetypes.
	var alphaY, alphaX, alphaZ float64
	// TODO: Adjust onion skins based upon distance from cursor.
	for y := 0; y < sm.Height; y++ {
		alphaY = 255
		if m.onionskinY {
			if y > m.focusedY {
				alphaY = 50
			}
		}
		xOffset := y * int(yStep.X)
		yOffset := y * int(-yStep.Y)
		for x := sm.Width - 1; x >= 0; x-- {
			alphaX = 255
			if m.onionskinX {
				if x < m.focusedX {
					alphaX = 50
				}
			}
			for z := 0; z < sm.Depth; z++ {
				alphaZ = 255
				if m.onionskinZ {
					if z > m.focusedZ {
						alphaZ = 50
					}
				}
				col.A = uint8(math.Min(math.Min(alphaX, alphaY), alphaZ))
				for t := 0; t < len(sm.Tiles[y][x][z]); t++ {
					oX := pos.X + (x*tWidth+xOffset+startX)*scale
					oY := pos.Y + (z*tHeight-yOffset+startY)*scale
					oH, _, oD := dm.GetArchDimensions(&sm.Tiles[y][x][z][t])
					if adjustment, ok := dm.AnimationsConfig.Adjustments[dm.GetArchType(&sm.Tiles[y][x][z][t], 0)]; ok {
						oX += int(adjustment.X) * scale
						oY += int(adjustment.Y) * scale
					}

					// calc render z
					indexZ := z
					indexX := x
					indexY := y
					zIndex := (indexZ * sm.Height * sm.Width) + (sm.Depth * indexY) - (indexX) + t

					anim, face := dm.GetAnimAndFace(&sm.Tiles[y][x][z][t], "", "")
					imageName, err := dm.GetAnimFaceImage(anim, face)
					if err != nil {
						continue
					}

					if tex, ok := m.context.ImageTextures()[imageName]; ok {
						if (oH > 1 || oD > 1) && int(tex.Height*float32(scale)) > tHeight*scale {
							oY -= int(tex.Height*float32(scale)) - (tHeight * scale)
						}
						drawables = append(drawables, archDrawable{
							z: zIndex,
							x: oX,
							y: oY,
							w: oX + int(tex.Width)*scale,
							h: oY + int(tex.Height)*scale,
							c: col,
							t: tex,
						})
					} else {
						//log.Println(err)
					}
				}
			}
		}
	}

	// Sort our drawables.
	sort.Slice(drawables, func(i, j int) bool {
		return drawables[i].z < drawables[j].z
	})
	// Render them.
	for _, d := range drawables {
		canvas.AddImageV(d.t.Texture, image.Pt(d.x, d.y), image.Pt(d.w, d.h), image.Pt(0, 0), image.Pt(1, 1), d.c)
	}

	// Draw grid.
	if m.showGrid {
		for y := 0; y < sm.Height; y++ {
			xOffset := y * int(yStep.X)
			yOffset := y * int(-yStep.Y)
			col.A = 0
			if m.showYGrids {
				// TODO: fade out based upon distance from focusedY
				col.A = 15
			}
			if m.focusedY == y {
				col.A = 50
			}
			for x := 0; x < sm.Width; x++ {
				for z := 0; z < sm.Depth; z++ {
					oX := pos.X + (x*tWidth+xOffset+startX)*scale
					oY := pos.Y + (z*tHeight-yOffset+startY)*scale
					oW := (tWidth) * scale
					oH := (tHeight) * scale
					canvas.AddRect(image.Pt(oX, oY), image.Pt(oX+oW, oY+oH), col, 0, 0, 0.5)
				}
			}
		}
	}

	focusedDrawn := false
	// Draw focused arch location, if possible
	{
		archs := v.GetArchs(m.focusedY, m.focusedX, m.focusedZ)
		if m.focusedI >= 0 && m.focusedI < len(archs) {
			xOffset := m.focusedY * int(yStep.X)
			yOffset := m.focusedY * int(-yStep.Y)
			oX := pos.X + (m.focusedX*tWidth+xOffset+startX)*scale
			oY := pos.Y + (m.focusedZ*tHeight-yOffset+startY)*scale
			oW := (tWidth) * scale
			oH := (tHeight) * scale

			canvas.AddRectFilled(image.Pt(oX, oY), image.Pt(oX+oW, oY+oH), focusedBackgroundColor, 0, 0)
			focusedDrawn = true
		}
	}

	// Draw selected.
	{
		for yxz := range m.selectedCoords.Get() {
			y, x, z := yxz[0], yxz[1], yxz[2]
			if focusedDrawn && m.focusedY == y && m.focusedX == x && m.focusedZ == z {
				continue
			}

			xOffset := y * int(yStep.X)
			yOffset := y * int(-yStep.Y)
			oX := pos.X + (x*tWidth+xOffset+startX)*scale
			oY := pos.Y + (z*tHeight-yOffset+startY)*scale
			oW := (tWidth) * scale
			oH := (tHeight) * scale

			canvas.AddRectFilled(image.Pt(oX, oY), image.Pt(oX+oW, oY+oH), selectedBackgroundColor, 0, 0)
		}
	}

	// Draw selecting.
	{
		for yxz := range m.selectingCoords.Get() {
			y, x, z := yxz[0], yxz[1], yxz[2]
			if focusedDrawn && m.focusedY == y && m.focusedX == x && m.focusedZ == z {
				continue
			}

			xOffset := y * int(yStep.X)
			yOffset := y * int(-yStep.Y)
			oX := pos.X + (x*tWidth+xOffset+startX)*scale
			oY := pos.Y + (z*tHeight-yOffset+startY)*scale
			oW := (tWidth) * scale
			oH := (tHeight) * scale

			canvas.AddRectFilled(image.Pt(oX, oY), image.Pt(oX+oW, oY+oH), selectingBackgroundColor, 0, 0)
		}
	}

	// Draw focused.
	{
		xOffset := m.focusedY * int(yStep.X)
		yOffset := m.focusedY * int(-yStep.Y)
		oX := pos.X + (m.focusedX*tWidth+xOffset+startX)*scale
		oY := pos.Y + (m.focusedZ*tHeight-yOffset+startY)*scale
		oW := (tWidth) * scale
		oH := (tHeight) * scale

		canvas.AddRect(image.Pt(oX, oY), image.Pt(oX+oW, oY+oH), focusedBorderColor, 0, 0, 1)
	}

	imgui.EndChild()
}
