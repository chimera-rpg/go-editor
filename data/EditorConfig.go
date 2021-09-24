package data

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type EditorConfig struct {
	filePath string
	OpenMaps []string
}

// Save saves the configuration to disk.
func (e *EditorConfig) Save() (err error) {
	// Ensure path to cfg exists.
	if _, err = os.Stat(filepath.Dir(e.filePath)); os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(e.filePath), os.ModePerm); err != nil {
			return
		}
	}
	// Write out default config.
	log.Printf("Saving config \"%s\"\n", e.filePath)
	bytes, _ := yaml.Marshal(e)
	err = ioutil.WriteFile(e.filePath, bytes, 0644)
	if err != nil {
		return
	}
	return
}

func (e *EditorConfig) AddMap(fullPath string) error {
	for _, oM := range e.OpenMaps {
		if oM == fullPath {
			return fmt.Errorf("map already exists")
		}
	}
	e.OpenMaps = append(e.OpenMaps, fullPath)
	return nil
}
func (e *EditorConfig) RemoveMap(fullPath string) error {
	for oI, oM := range e.OpenMaps {
		if fullPath == oM {
			e.OpenMaps = append(e.OpenMaps[:oI], e.OpenMaps[oI+1:]...)
			return nil
		}
	}
	return fmt.Errorf("map does not exist")
}
