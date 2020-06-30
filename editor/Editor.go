package editor

import (
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-editor/data"
	log "github.com/sirupsen/logrus"
)

type Editor struct {
	dataManager *data.Manager
	rootWindow  *ui.Window
	isRunning   bool

	UpdateChannel chan struct{}
}

func (e *Editor) Setup(dataManager *data.Manager, inst *ui.Instance) (err error) {
	e.UpdateChannel = make(chan struct{})
	e.rootWindow = &inst.RootWindow
	e.dataManager = dataManager
	e.isRunning = true

	return
}

func (e *Editor) Destroy() {
	e.isRunning = false
}

func (e *Editor) Loop() {
	log.Println("Editor: Loop")
	for e.isRunning {
		select {
		case <-e.UpdateChannel:
			if e.isRunning {
				// do something
			}
		}
	}
}
