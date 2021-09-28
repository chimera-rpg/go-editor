package mapview

import (
	"github.com/chimera-rpg/go-editor/data"
	"github.com/chimera-rpg/go-editor/widgets"
)

type Context interface {
	DataManager() *data.Manager
	ImageTextures() map[string]*data.ImageTexture
	SelectedArch() string
	SetSelectedArch(string)
	ArchEditor() *widgets.ArchEditorWidget
	FocusedMapset() *Mapset
	SetFocusedMapset(*Mapset)
}
