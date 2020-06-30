package main

import (
	"runtime/debug"

	log "github.com/sirupsen/logrus"

	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-editor/data"
	"github.com/chimera-rpg/go-editor/editor"
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{ForceColors: true}) // It would be ideal to only force colors on Windows 10+ -- checking this is possible with x/sys/windows/registry, though we'd need OS-specific source files for log initialization.
	var err error
	var dataManager data.Manager
	var editorInstance editor.Editor
	var uiInstance ui.Instance

	defer func() {
		if r := recover(); r != nil {
			ui.ShowError("%v", r.(error).Error())
			debug.PrintStack()
		}
	}()
	log.Print("Starting Chimera editor (golang)")

	if err = dataManager.Setup(); err != nil {
		ui.ShowError("%s", err)
	}

	// Setup our UI
	if err = uiInstance.Setup(&dataManager); err != nil {
		ui.ShowError("%s", err)
		return
	}
	defer uiInstance.Cleanup()

	ui.GlobalInstance = &uiInstance

	// Setup our Editor
	if err = editorInstance.Setup(&dataManager, &uiInstance); err != nil {
		ui.ShowError("%s", err)
		return
	}
	defer editorInstance.Destroy()
	// Start the clientInstance's channel listening loop as a coroutine
	go editorInstance.Loop()

	// Start our UI Loop.
	uiInstance.Loop()

	log.Print("Sayonara!")
}
