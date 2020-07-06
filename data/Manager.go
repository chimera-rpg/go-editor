package data

import (
	"errors"
	"image"
	"log"
	"strings"

	_ "image/png"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/nfnt/resize"

	cdata "github.com/chimera-rpg/go-common/data"
	sdata "github.com/chimera-rpg/go-server/data"
	"gopkg.in/yaml.v2"
)

// Manager handles access to files on the system.
type Manager struct {
	DataPath            string // Path for client data (fonts, etc.)
	MapsPath            string // Path for maps
	ArchetypesPath      string // Path for archetypes.
	images              map[string]image.Image
	scaledImages        map[float64]map[string]image.Image
	animations          map[string]sdata.AnimationPre
	archetypes          map[string]*sdata.Archetype
	archetypeFiles      map[string]map[string]*sdata.Archetype
	AnimationsConfig    cdata.AnimationsConfig
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

	m.images = make(map[string]image.Image)
	m.scaledImages = make(map[float64]map[string]image.Image)
	m.archetypes = make(map[string]*sdata.Archetype)
	m.archetypeFiles = make(map[string]map[string]*sdata.Archetype)
	m.animationFiles = make(map[string]map[string]struct{})
	m.animations = make(map[string]sdata.AnimationPre)

	// Read animations config
	animationsConfigPath := path.Join(m.ArchetypesPath, "config.yaml")
	r, err := ioutil.ReadFile(animationsConfigPath)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(r, &m.AnimationsConfig); err != nil {
		return err
	}

	if err = m.LoadArchetypes(); err != nil {
		return
	}

	if err = m.LoadAnimations(); err != nil {
		return
	}

	if err = m.LoadImages(); err != nil {
		return
	}
	log.Printf("Cached %d images\n", len(m.images))

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

	for k, a := range archetypesMap {
		m.archetypes[k] = a
	}

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
	r, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	animationsMap := make(map[string]sdata.AnimationPre)

	if err = yaml.Unmarshal(r, &animationsMap); err != nil {
		return err
	}

	for k, a := range animationsMap {
		m.animations[k] = a
	}

	return nil
}

func (m *Manager) LoadImages() error {
	err := filepath.Walk(m.ArchetypesPath, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if strings.HasSuffix(file, ".png") {
				err = m.LoadImage(file)
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

func (m *Manager) LoadImage(p string) error {
	shortpath := filepath.ToSlash(p[len(m.ArchetypesPath)+1:])
	if _, ok := m.images[shortpath]; ok {
		return nil
	}

	reader, err := os.Open(p)
	if err != nil {
		log.Fatal(err)
	}

	img, _, err := image.Decode(reader)
	if err != nil {
		return err
	}

	m.images[shortpath] = img
	return nil
}

func (m *Manager) GetArchetypeFiles() []string {
	return m.archetypeFilesOrder
}

func (m *Manager) GetArchetypeFile(f string) map[string]*sdata.Archetype {
	return m.archetypeFiles[f]
}

func (m *Manager) GetArchetype(f string) *sdata.Archetype {
	return m.archetypes[f]
}

func (m *Manager) GetAnimFaceImage(anim, face string) (string, error) {
	a, ok := m.animations[anim]
	if !ok {
		return "", errors.New("missing animation")
	}
	f, ok := a.Faces[face]
	if !ok {
		return "", errors.New("missing face")
	}
	if len(f) < 0 {
		return "", errors.New("missing frame")
	}
	return f[0].Image, nil
}

func (m *Manager) GetImage(i string) image.Image {
	return m.images[i]
}

func (m *Manager) GetScaledImage(scale float64, name string) image.Image {
	img := m.GetImage(name)
	if img == nil {
		return nil
	}
	if _, ok := m.scaledImages[scale]; !ok {
		m.scaledImages[scale] = make(map[string]image.Image)
	}
	scaledImage, ok := m.scaledImages[scale][name]
	if !ok {
		scaledImage = resize.Resize(uint(float64(img.Bounds().Max.X)*scale), uint(float64(img.Bounds().Max.Y)*scale), img, resize.NearestNeighbor)
		m.scaledImages[scale][name] = scaledImage
	}
	return scaledImage
}
