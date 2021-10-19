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
	files := []string{"dropper", "dropper-focus", "eraser", "eraser-focus", "fill", "fill-focus", "insert", "insert-focus", "select", "select-focus", "cselect", "cselect-focus", "lselect", "lselect-focus", "wand", "wand-focus", "loading", "missing", "tl", "tr", "bl", "br", "l", "t", "r", "b", "u", "d", "delete", "blank"}
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
