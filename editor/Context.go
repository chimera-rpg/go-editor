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

func (c *Context) DataManager() *data.Manager {
	return c.dataManager
}

func (c *Context) ImageTextures() map[string]*data.ImageTexture {
	return c.imageTextures
}

func (c *Context) SelectedArch() string {
	return c.selectedArch
}
