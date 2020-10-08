package data

import (
	g "github.com/AllenDang/giu"
)

// ImageTexture is a container around giu Textures.
type ImageTexture struct {
	Texture       *g.Texture
	Width, Height float32
}
