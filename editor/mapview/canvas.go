package mapview

import (
	"image"
	"image/color"
	"math"
	"sort"

	g "github.com/AllenDang/giu"
	"github.com/chimera-rpg/go-editor/data"
	"github.com/chimera-rpg/go-editor/editor/icons"
)

type archDrawable struct {
	z     int
	x, y  int
	cY    int
	w, h  int
	large bool
	t     *data.ImageTexture
	c     color.RGBA
}

func (m *Mapset) getMapSize(v *data.UnReMap) (float32, float32) {
	sm := v.Get()
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
	return float32(canvasWidth), float32(canvasHeight)
}

func (m *Mapset) drawMap(v *data.UnReMap) {
	sm := v.Get()
	dm := m.context.DataManager()

	canvas := g.GetCanvas()
	pos := g.GetCursorScreenPos()
	scale := int(m.zoom)
	tWidth := int(dm.AnimationsConfig.TileWidth)
	tHeight := int(dm.AnimationsConfig.TileHeight)
	yStep := dm.AnimationsConfig.YStep
	padding := 4
	cWidth := sm.Width * tWidth
	cHeight := sm.Depth * tHeight

	canvasWidth := int((cWidth + (sm.Height * int(yStep.X)) + padding*2) * scale)
	canvasHeight := int((cHeight + (sm.Height * int(-yStep.Y)) + padding*2) * scale)

	startX := padding
	startY := padding + (sm.Height * int(-yStep.Y))

	drawRect := func(y, x, z int, col color.RGBA) {
		xOffset := y * int(yStep.X)
		yOffset := y * int(-yStep.Y)
		oX := pos.X + (x*tWidth+xOffset+startX)*scale
		oY := pos.Y + (z*tHeight-yOffset+startY)*scale
		oW := (tWidth) * scale
		oH := (tHeight) * scale

		canvas.AddRectFilled(image.Pt(oX, oY), image.Pt(oX+oW, oY+oH), col, 0, 0)
	}

	drawBox := func(y, x, z int, col color.RGBA) {
		xOffset := y * int(yStep.X)
		yOffset := y * int(-yStep.Y)

		// Calc bottom
		o1X := pos.X + (x*tWidth+xOffset+startX)*scale
		o1Y := pos.Y + (z*tHeight-yOffset+startY)*scale

		// Calc top
		y += 1
		xOffset = y * int(yStep.X)
		yOffset = y * int(-yStep.Y)
		o2X := pos.X + (x*tWidth+xOffset+startX)*scale
		o2Y := pos.Y + (z*tHeight-yOffset+startY)*scale

		oW := (tWidth) * scale
		oH := (tHeight) * scale

		// Left
		canvas.AddQuad(image.Pt(o1X, o1Y), image.Pt(o2X, o2Y+2), image.Pt(o2X, o2Y+oH-1), image.Pt(o1X, o1Y+oH-3), col, 1)
		// Front
		canvas.AddQuad(image.Pt(o1X, o1Y+oH), image.Pt(o2X, o2Y+oH), image.Pt(o2X+oW, o2Y+oH), image.Pt(o1X+oW, o1Y+oH), col, 1)
		// Top
		canvas.AddQuad(image.Pt(o2X+1, o2Y), image.Pt(o2X+oW, o2Y), image.Pt(o2X+oW, o2Y+oH-1), image.Pt(o2X+1, o2Y+oH-1), col, 1)
	}

	drawHeightBox := func(y, x, z int, col color.RGBA) {
		// Get position of closest arch below the target coordinates.
		yPos := y
		for ; yPos >= 0; yPos-- {
			if len(sm.Tiles[yPos][x][z]) > 0 {
				maxH := 0
				for t := 0; t < len(sm.Tiles[yPos][x][z]); t++ {
					h, _, _ := dm.GetArchDimensions(&sm.Tiles[yPos][x][z][t])
					if int(h) > maxH {
						maxH = int(h)
					}
				}
				yPos += maxH
				break
			}
		}

		xOffset := y * int(yStep.X)
		yOffset := y * int(-yStep.Y)

		// Calc bottom
		o1X := pos.X + (x*tWidth+xOffset+startX)*scale
		o1Y := pos.Y + (z*tHeight-yOffset+startY)*scale

		// Calc top
		y = yPos
		xOffset = y * int(yStep.X)
		yOffset = y * int(-yStep.Y)
		o2X := pos.X + (x*tWidth+xOffset+startX)*scale
		o2Y := pos.Y + (z*tHeight-yOffset+startY)*scale

		oW := (tWidth) * scale
		oH := (tHeight) * scale

		canvas.AddLine(image.Pt(o1X, o1Y), image.Pt(o2X, o2Y), col, 1)
		//canvas.AddLine(image.Pt(o1X+oW, o1Y), image.Pt(o2X+oW, o2Y), col, 1)
		canvas.AddLine(image.Pt(o1X+oW, o1Y+oH), image.Pt(o2X+oW, o2Y+oH), col, 1)
		canvas.AddLine(image.Pt(o1X, o1Y+oH), image.Pt(o2X, o2Y+oH), col, 1)
	}

	col := color.RGBA{0, 0, 0, 255}
	canvas.AddRectFilled(pos, pos.Add(image.Pt(canvasWidth, canvasHeight)), col, 0, 0)

	col = color.RGBA{255, 255, 255, 255}
	var drawables []archDrawable
	// Draw archetypes.
	var alphaY, alphaX, alphaZ int32
	// TODO: Adjust onion skins based upon distance from cursor.
	for y := 0; y < sm.Height; y++ {
		alphaY = 255
		if m.onionskinY {
			if y < m.focusedY {
				alphaY = m.onionSkinGtIntensity
			} else if y > m.focusedY {
				alphaY = m.onionSkinLtIntensity
			}
		}
		xOffset := y * int(yStep.X)
		yOffset := y * int(-yStep.Y)
		for x := sm.Width - 1; x >= 0; x-- {
			alphaX = 255
			if m.onionskinX {
				if x < m.focusedX {
					alphaX = m.onionSkinGtIntensity
				} else if x > m.focusedX {
					alphaX = m.onionSkinLtIntensity
				}
			}
			for z := 0; z < sm.Depth; z++ {
				alphaZ = 255
				if m.onionskinZ {
					if z > m.focusedZ {
						alphaZ = m.onionSkinGtIntensity
					} else if z < m.focusedZ {
						alphaZ = m.onionSkinLtIntensity
					}
				}
				col.A = uint8(math.Min(math.Min(float64(alphaX), float64(alphaY)), float64(alphaZ)))
				for t := 0; t < len(sm.Tiles[y][x][z]); t++ {
					oX := pos.X + (x*tWidth+xOffset+startX)*scale
					oY := pos.Y + (z*tHeight-yOffset+startY)*scale
					oH, oW, oD := dm.GetArchDimensions(&sm.Tiles[y][x][z][t])
					large := false
					if oH > 1 || oW > 1 || oD > 1 {
						large = true
					}
					if adjustment, ok := dm.AnimationsConfig.Adjustments[dm.GetArchType(&sm.Tiles[y][x][z][t], 0)]; ok {
						oX += int(adjustment.X) * scale
						oY += int(adjustment.Y) * scale
					}

					// calc render z
					indexZ := z
					indexX := x
					indexY := y
					zIndex := (indexZ * sm.Height * sm.Width) + (sm.Depth * indexY) - (indexX) + t

					var tex *data.ImageTexture
					var ok bool
					anim, face := dm.GetAnimAndFace(&sm.Tiles[y][x][z][t], "", "")
					imageName, err := dm.GetAnimFaceImage(anim, face)
					if err != nil {
						tex, ok = icons.Textures["missing"]
					} else {
						tex, ok = m.context.ImageTextures()[imageName]
					}

					if ok {
						cY := oY
						if (oH > 1 || oD > 1) && int(tex.Height*float32(scale)) > tHeight*scale {
							oY -= int(tex.Height*float32(scale)) - (tHeight * scale)
						}
						drawables = append(drawables, archDrawable{
							z:     zIndex,
							x:     oX,
							y:     oY,
							cY:    cY,
							w:     oX + int(tex.Width)*scale,
							h:     oY + int(tex.Height)*scale,
							c:     col,
							t:     tex,
							large: large,
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
	for _, d := range drawables {
		if d.large {
			col := color.RGBA{
				R: 0,
				G: 128,
				B: 255,
				A: d.c.A,
			}
			canvas.AddRect(image.Pt(d.x, d.cY), image.Pt(d.x+tWidth*scale, d.cY+tHeight*scale), col, 0, 0, 0.5)
		}
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
			drawRect(m.focusedY, m.focusedX, m.focusedZ, focusedBackgroundColor)
			focusedDrawn = true
		}
	}

	// Draw selected.
	{
		coords := m.selectedCoords.Get()
		lowestY := math.MaxInt
		for yxz := range coords {
			if yxz[0] < lowestY {
				lowestY = yxz[0]
			}
		}
		for yxz := range m.selectedCoords.Get() {
			y, x, z := yxz[0], yxz[1], yxz[2]
			if focusedDrawn && m.focusedY == y && m.focusedX == x && m.focusedZ == z {
				continue
			}

			if y == lowestY {
				drawRect(y, x, z, selectedBackgroundColor)
			}

			drawBox(y, x, z, selectedBackgroundColor)
		}
	}

	// Draw selecting.
	{
		for yxz := range m.selectingCoords.Get() {
			y, x, z := yxz[0], yxz[1], yxz[2]
			if focusedDrawn && m.focusedY == y && m.focusedX == x && m.focusedZ == z {
				continue
			}
			drawRect(y, x, z, selectedBackgroundColor)
			drawBox(y, x, z, selectedBackgroundColor)
		}
	}

	// Draw focused.
	{
		/*xOffset := m.focusedY * int(yStep.X)
		yOffset := m.focusedY * int(-yStep.Y)
		oX := pos.X + (m.focusedX*tWidth+xOffset+startX)*scale
		oY := pos.Y + (m.focusedZ*tHeight-yOffset+startY)*scale
		oW := (tWidth) * scale
		oH := (tHeight) * scale

		canvas.AddRect(image.Pt(oX, oY), image.Pt(oX+oW, oY+oH), focusedBorderColor, 0, 0, 1)*/
		drawHeightBox(m.focusedY, m.focusedX, m.focusedZ, focusedHeightBoxColor)
		drawBox(m.focusedY, m.focusedX, m.focusedZ, focusedBorderColor)
	}

	// Draw hovered.
	{
		drawHeightBox(m.hoveredY, m.hoveredX, m.hoveredZ, hoveredHeightBoxColor)
		drawBox(m.hoveredY, m.hoveredX, m.hoveredZ, hoveredBorderColor)
		/*xOffset := m.hoveredY * int(yStep.X)
		yOffset := m.hoveredY * int(-yStep.Y)
		oX := pos.X + (m.hoveredX*tWidth+xOffset+startX)*scale
		oY := pos.Y + (m.hoveredZ*tHeight-yOffset+startY)*scale
		oW := (tWidth) * scale
		oH := (tHeight) * scale

		canvas.AddRect(image.Pt(oX, oY), image.Pt(oX+oW, oY+oH), hoveredBorderColor, 0, 0, 1)
		*/
	}
}
