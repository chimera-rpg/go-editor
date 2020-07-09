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
	imgui.BeginGroup()
	{
	}
	imgui.EndGroup()
	imgui.SameLine()
	// Render Path
	imgui.BeginGroup()
	{
		// Address Begin
		imgui.BeginGroup()
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
		imgui.EndGroup()
		// Address End
		// Browser Begin
		imgui.BeginGroup()
		imgui.BeginChildV("browser", imgui.Vec2{X: 0, Y: 300}, false, 0)
		{
			files, err := ioutil.ReadDir(*f.cwd)
			if err == nil {
				for _, file := range files {
					if imgui.ButtonV(file.Name(), imgui.Vec2{X: 0, Y: 0}) {
						if file.IsDir() {
							*f.cwd = fmt.Sprintf("%s%c%s", *f.cwd, filepath.Separator, file.Name())
						} else {
							*f.filename = file.Name()
						}
						// TODO: move/open file
					}
				}
			}
		}
		imgui.EndChild()
		imgui.EndGroup()
		// Browser End
		// FileName Begin
		imgui.BeginGroup()
		{
			if imgui.InputTextV("filename", f.filename, 0, nil) {
				// TODO: ... something?
			}
		}
		imgui.EndGroup()
		// FileName End
	}
	imgui.EndGroup()
}
