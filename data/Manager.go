package data

import (
	"image"
	"log"
	"strings"

	sdata "github.com/chimera-rpg/go-server/data"
	"gopkg.in/yaml.v2"
	_ "image/png"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

// Manager handles access to files on the system.
type Manager struct {
	DataPath            string // Path for client data (fonts, etc.)
	MapsPath            string // Path for maps
	ArchetypesPath      string // Path for archetypes.
	images              map[uint32]image.Image
	archetypeFiles      map[string]map[string]*sdata.Archetype
	archetypeFilesOrder []string
	animationFiles      map[string]map[string]struct{}
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
	// Acquire our various paths.
	if err = m.acquireMapPath(); err != nil {
		return
	}
	if err = m.acquireArchetypesPath(); err != nil {
		return
	}

	m.images = make(map[uint32]image.Image)
	m.archetypeFiles = make(map[string]map[string]*sdata.Archetype)
	m.animationFiles = make(map[string]map[string]struct{})

	if err = m.LoadArchetypes(); err != nil {
		return
	}

	if err = m.LoadAnimations(); err != nil {
		return
	}

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

func (m *Manager) acquireMapPath() (err error) {
	var dir string
	// Set our path which should be <parent of cmd>/share/chimera/client.
	if dir, err = filepath.Abs(os.Args[0]); err != nil {
		return
	}
	dir = path.Join(filepath.Dir(filepath.Dir(dir)), "share", "chimera", "maps")

	m.MapsPath = dir
	return
}

func (m *Manager) acquireArchetypesPath() (err error) {
	var dir string
	if dir, err = filepath.Abs(os.Args[0]); err != nil {
		return
	}
	dir = path.Join(filepath.Dir(filepath.Dir(dir)), "share", "chimera", "archetypes")

	m.ArchetypesPath = dir
	return
}

func (m *Manager) LoadMap(filepath string) (maps map[string]*sdata.Map, err error) {
	r, err := ioutil.ReadFile(filepath)
	if err != nil {
		return
	}

	if err = yaml.Unmarshal(r, &maps); err != nil {
		return
	}
	return
}

func (m *Manager) LoadArchetypes() error {
	err := filepath.Walk(m.ArchetypesPath, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if strings.HasSuffix(file, ".arch.yaml") {
				err = m.LoadArchetypeFile(file)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) LoadArchetypeFile(filepath string) error {
	r, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	archetypesMap := make(map[string]*sdata.Archetype)

	if err = yaml.Unmarshal(r, &archetypesMap); err != nil {
		return err
	}

	m.archetypeFiles[filepath] = archetypesMap
	m.archetypeFilesOrder = append(m.archetypeFilesOrder, filepath)

	return nil
}

func (m *Manager) LoadAnimations() error {
	err := filepath.Walk(m.ArchetypesPath, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if strings.HasSuffix(file, ".anim.yaml") {
				err = m.LoadAnimationFile(file)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) LoadAnimationFile(filepath string) error {
	log.Printf("Load anim %s\n", filepath)
	//
	return nil
}

func (m *Manager) GetArchetypeFiles() []string {
	return m.archetypeFilesOrder
}

func (m *Manager) GetArchetypeFile(f string) map[string]*sdata.Archetype {
	return m.archetypeFiles[f]
}
