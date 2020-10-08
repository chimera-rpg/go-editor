package editor

import (
	"github.com/chimera-rpg/go-editor/data"
)

// Context provides a variety of editor state that is used between components.
type Context struct {
	dataManager   *data.Manager
	selectedArch  string
	cursorArch    []string
	imageTextures map[string]*data.ImageTexture
}
