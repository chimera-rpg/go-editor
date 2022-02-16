package main

import (
	"os"
	"os/signal"
	"reflect"
	"runtime/debug"
	"syscall"

	"github.com/cosmos72/gomacro/imports"
	log "github.com/sirupsen/logrus"

	"github.com/chimera-rpg/go-editor/data"
	"github.com/chimera-rpg/go-editor/editor"
	sdata "github.com/chimera-rpg/go-server/data"
	sworld "github.com/chimera-rpg/go-server/world"
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{ForceColors: true}) // It would be ideal to only force colors on Windows 10+ -- checking this is possible with x/sys/windows/registry, though we'd need OS-specific source files for log initialization.
	var err error
	var dataManager data.Manager
	var editorInstance editor.Editor

	defer func() {
		if r := recover(); r != nil {
			log.Fatalln(r.(error))
			debug.PrintStack()
		}
	}()
	log.Print("Starting Chimera editor (golang)")

	// TODO: Properly setup Interpreter with a shared function with the server.
	{
		var o sworld.ObjectI
		var e sworld.EventI
		imports.Packages["chimera"] = imports.Package{
			Binds:    map[string]reflect.Value{},
			Types:    map[string]reflect.Type{},
			Proxies:  map[string]reflect.Type{},
			Untypeds: map[string]string{},
			Wrappers: map[string][]string{},
		}

		imports.Packages["chimera"].Binds["self"] = reflect.ValueOf(&o).Elem()
		imports.Packages["chimera"].Binds["event"] = reflect.ValueOf(&e).Elem()

		sdata.Interpreter.ImportPackage("lname", "chimera")
		sdata.Interpreter.ChangePackage("lname", "chimera")

		sworld.SetupInterpreterTypes(sdata.Interpreter)
	}

	if err = dataManager.Setup(); err != nil {
		log.Fatalln(err)
	}

	// Setup our UI
	/*if err = uiInstance.Setup(&dataManager); err != nil {
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
	uiInstance.Loop()*/

	if err = editorInstance.Setup(&dataManager); err != nil {
		log.Fatalln(err)
		return
	}
	defer editorInstance.Destroy()

	// Add cleanup handling on kill.
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		if err := dataManager.EditorConfig.Save(); err != nil {
			log.Errorln(err)
		}
		os.Exit(1)
	}()

	// Start the clientInstance's channel listening loop as a coroutine
	editorInstance.Start()

	// Save config.
	if err := dataManager.EditorConfig.Save(); err != nil {
		log.Errorln(err)
	}

	log.Print("Sayonara!")
}
