package icons

import (
	"embed"
	"image/png"

	g "github.com/AllenDang/giu"
	"github.com/chimera-rpg/go-editor/data"
)

//go:embed *.png
var f embed.FS

var Textures map[string]*data.ImageTexture

func Load() {
	Textures = make(map[string]*data.ImageTexture)
	files := []string{"dropper", "eraser", "fill", "insert", "select", "loading", "missing"}
	for _, name := range files {
		go func(name string) {
			filedata, _ := f.Open(name + ".png")
			img, _ := png.Decode(filedata)
			g.NewTextureFromRgba(img, func(t *g.Texture) {
				Textures[name] = &data.ImageTexture{
					Width:   float32(img.Bounds().Max.X),
					Height:  float32(img.Bounds().Max.Y),
					Texture: t,
				}
			})
		}(name)
	}
}
