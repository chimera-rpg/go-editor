package editor

import (
	"github.com/chimera-rpg/go-editor/data"
	"github.com/chimera-rpg/go-editor/editor/mapview"
	"github.com/chimera-rpg/go-editor/widgets"
)

// Context provides a variety of editor state that is used between components.
type Context struct {
	dataManager   *data.Manager
	selectedArch  string
	cursorArch    []string
	imageTextures map[string]*data.ImageTexture
	archEditor    *widgets.ArchEditorWidget
	focusedMapset *mapview.Mapset
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

func (c *Context) SetSelectedArch(a string) {
	c.selectedArch = a
}

func (c *Context) ArchEditor() *widgets.ArchEditorWidget {
	return c.archEditor
}

func (c *Context) FocusedMapset() *mapview.Mapset {
	return c.focusedMapset
}

func (c *Context) SetFocusedMapset(m *mapview.Mapset) {
	c.focusedMapset = m
}
