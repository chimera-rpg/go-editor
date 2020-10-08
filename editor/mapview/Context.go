package mapview

import (
	"github.com/chimera-rpg/go-editor/data"
)

type Context interface {
	DataManager() *data.Manager
	ImageTextures() map[string]*data.ImageTexture
	SelectedArch() string
}
