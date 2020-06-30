package data

import (
	"bytes"
	"image"
	"strconv"
	"strings"

	_ "image/png"
	"os"
	"path"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/kettek/apng"
)

// Manager handles access to files on the system.
type Manager struct {
	DataPath       string // Path for client data (fonts, etc.)
	MapPath        string // Path for maps
	ArchetypesPath string // Path for archetypes.
	images         map[uint32]image.Image
}

// Setup gets the required data paths and creates them if needed.
func (m *Manager) Setup() (err error) {
	// Acquire our various paths.
	if err = m.acquireDataPath(); err != nil {
		return
	}
	// Ensure each exists.
	if _, err = os.Stat(m.DataPath); err != nil {
		// DataPath does not exist!
		return
	}

	m.images = make(map[uint32]image.Image)

	return
}

// GetDataPath gets a path relative to the data path directory.
func (m *Manager) GetDataPath(parts ...string) string {
	return path.Join(m.DataPath, path.Clean("/"+path.Join(parts...)))
}

func (m *Manager) acquireDataPath() (err error) {
	var dir string
	// Set our path which should be <parent of cmd>/share/chimera/client.
	if dir, err = filepath.Abs(os.Args[0]); err != nil {
		return
	}
	dir = path.Join(filepath.Dir(filepath.Dir(dir)), "share", "chimera", "editor")

	m.DataPath = dir
	return
}
