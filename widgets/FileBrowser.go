package widgets

import (
	"fmt"
	"github.com/AllenDang/giu/imgui"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type FileBrowserWidget struct {
	cwd       *string
	favorites []string
	filename  *string
}

func FileBrowser(cwd *string, filename *string, favorites []string) *FileBrowserWidget {
	return &FileBrowserWidget{
		cwd:       cwd,
		favorites: favorites,
		filename:  filename,
	}
}

func (f *FileBrowserWidget) Build() {
	// Render Favorites
	imgui.BeginChildV("left", imgui.Vec2{X: 200, Y: 0}, true, 0)
	{
	}
	imgui.EndChild()
	imgui.SameLine()
	// Render Path
	imgui.BeginChildV("right", imgui.Vec2{X: 0, Y: 0}, true, 0)
	{
		// Begin Address
		imgui.BeginChildV("address", imgui.Vec2{X: 0, Y: 50}, true, 0)
		{
			pathParts := strings.Split(*f.cwd, fmt.Sprintf("%c", filepath.Separator))
			for i, p := range pathParts {
				if imgui.ButtonV(p, imgui.Vec2{X: 0, Y: 0}) {
					*f.cwd = strings.Join(pathParts[:i+1], fmt.Sprintf("%c", filepath.Separator))
					if *f.cwd == "" {
						*f.cwd = fmt.Sprintf("%c", filepath.Separator)
					}
				}
				imgui.SameLine()
			}
		}
		imgui.EndChild()
		imgui.BeginChildV("browser", imgui.Vec2{X: 0, Y: 0}, true, 0)
		{
			files, err := ioutil.ReadDir(*f.cwd)
			if err == nil {
				for _, file := range files {
					if imgui.ButtonV(file.Name(), imgui.Vec2{X: 0, Y: 0}) {
						*f.cwd = fmt.Sprintf("%s%c%s", *f.cwd, filepath.Separator, file.Name())
						// TODO: move/open file
					}
				}
			}
		}
		imgui.EndChild()
		// Final name path
		imgui.BeginChildV("name", imgui.Vec2{X: 0, Y: 0}, true, 0)
		{
			if imgui.InputTextV("filename", f.filename, 0, nil) {
				// TODO: ... something?
			}
		}
		imgui.EndChild()
		// End Right
	}
	imgui.EndChild()
}
