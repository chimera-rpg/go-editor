package data

import (
	"errors"
	"image"
	"log"
	"strings"

	"fmt"
	_ "image/png"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/nfnt/resize"

	"github.com/imdario/mergo"

	cdata "github.com/chimera-rpg/go-common/data"
	sdata "github.com/chimera-rpg/go-server/data"
	"gopkg.in/yaml.v2"
)

// Manager handles access to files on the system.
type Manager struct {
	DataPath            string // Path for client data (fonts, etc.)
	MapsPath            string // Path for maps
	EtcPath             string // Path for configuration
	ArchetypesPath      string // Path for archetypes.
	images              map[string]image.Image
	scaledImages        map[float64]map[string]image.Image
	animations          map[string]sdata.AnimationPre
	archetypes          map[string]*sdata.Archetype // Parsed archetypes
	archetypesOrder     []string
	archetypeFiles      map[string][]string
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
	//
	if err = m.acquireEtcPath(); err != nil {
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
	m.archetypeFiles = make(map[string][]string)
	m.animationFiles = make(map[string]map[string]struct{})
	m.animations = make(map[string]sdata.AnimationPre)

	// Read animations config
	animationsConfigPath := filepath.Join(m.ArchetypesPath, "config.yaml")
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
	log.Printf("Loaded %d archetypes\n", len(m.archetypes))

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
	return path.Join(m.DataPath, filepath.Clean(fmt.Sprintf("%c", filepath.Separator)+filepath.Join(parts...)))
}

// GetEtcPath gets a path relative to the etc path directory.
func (m *Manager) GetEtcPath(parts ...string) string {
	return path.Join(m.EtcPath, filepath.Clean(fmt.Sprintf("%c", filepath.Separator)+filepath.Join(parts...)))
}

func (m *Manager) acquireDataPath() (err error) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	m.DataPath = filepath.Join(cwd, "share", "chimera", "editor")
	return
}

func (m *Manager) acquireEtcPath() (err error) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	m.EtcPath = filepath.Join(cwd, "etc", "chimera")
	return
}

func (m *Manager) acquireMapPath() (err error) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	m.MapsPath = filepath.Join(cwd, "share", "chimera", "maps")
	return
}

func (m *Manager) acquireArchetypesPath() (err error) {
	var dir string
	cwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	dir = filepath.Join(cwd, "share", "chimera", "archetypes")

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

func (m *Manager) SaveMap(filepath string, maps map[string]*sdata.Map) (err error) {
	out, err := yaml.Marshal(maps)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath, out, 0644)
	if err != nil {
		return err
	}
	return nil
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

func (m *Manager) LoadArchetypeFile(fpath string) error {
	r, err := ioutil.ReadFile(fpath)
	if err != nil {
		return err
	}

	archetypesMap := make(map[string]*sdata.Archetype)

	if err = yaml.Unmarshal(r, &archetypesMap); err != nil {
		return err
	}

	shortpath := filepath.ToSlash(fpath[len(m.ArchetypesPath)+1:])
	shortpath = shortpath[0 : len(shortpath)-len(".arch.yaml")]

	m.archetypeFilesOrder = append(m.archetypeFilesOrder, shortpath)

	for k, a := range archetypesMap {
		m.archetypes[k] = a
		m.archetypesOrder = append(m.archetypesOrder, k)
		m.archetypeFiles[shortpath] = append(m.archetypeFiles[shortpath], k)
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

func (m *Manager) GetArchetypeFile(f string) []string {
	return m.archetypeFiles[f]
}

func (m *Manager) LookupArchetypeFile(a string) string {
	for _, v := range m.archetypeFiles {
		for _, ak := range v {
			if ak == a {
				return ak
			}
		}
	}
	return ""
}

func (m *Manager) GetArchetypes() []string {
	return m.archetypesOrder
}

func (m *Manager) GetArchetype(f string) *sdata.Archetype {
	return m.archetypes[f]
}

// GetArchetypesAsTree returns the current archetypes as a tree.
func (m *Manager) GetArchetypesAsTree() ArchetypeTreeNode {
	return ParseArchetypesIntoTree(m.archetypesOrder)
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

func (m *Manager) GetImages() map[string]image.Image {
	return m.images
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

func (m *Manager) GetAnimAndFace(a *sdata.Archetype, anim, face string) (string, string) {
	if anim == "" && a.Anim != "" {
		anim = a.Anim
	}
	if face == "" && a.Face != "" {
		face = a.Face
	}

	if anim == "" || face == "" {
		if a.Arch != "" {
			o := m.GetArchetype(a.Arch)
			if o != nil {
				anim, face = m.GetAnimAndFace(o, anim, face)
				if anim != "" && face != "" {
					return anim, face
				}
			}
		}
		for _, name := range a.Archs {
			o := m.GetArchetype(name)
			if o != nil {
				anim, face = m.GetAnimAndFace(o, anim, face)
				if anim != "" && face != "" {
					return anim, face
				}
			}
		}
	}

	return anim, face
}

func (m *Manager) GetArchType(a *sdata.Archetype, atype cdata.ArchetypeType) cdata.ArchetypeType {
	if atype == 0 && a.Type != 0 {
		atype = a.Type
	}

	if atype == 0 {
		if a.Arch != "" {
			o := m.GetArchetype(a.Arch)
			if o != nil {
				atype = m.GetArchType(o, atype)
				if atype != 0 {
					return atype
				}
			}
		}
		for _, name := range a.Archs {
			o := m.GetArchetype(name)
			if o != nil {
				atype = m.GetArchType(o, atype)
				if atype != 0 {
					return atype
				}
			}
		}
	}

	return atype
}

func (m *Manager) GetArchName(a *sdata.Archetype, name string) string {
	aString, _ := a.Name.GetString()
	if name == "" && aString != "" {
		name = aString
	}

	if name == "" {
		if a.Arch != "" {
			o := m.GetArchetype(a.Arch)
			if o != nil {
				name = m.GetArchName(o, name)
				if name != "" {
					return name
				}
			}
		}
		for _, archName := range a.Archs {
			o := m.GetArchetype(archName)
			if o != nil {
				name = m.GetArchName(o, name)
				if name != "" {
					return name
				}
			}
		}
	}

	return name
}

func (m *Manager) GetArchDimensions(a *sdata.Archetype) (uint8, uint8, uint8) {
	var h, w, d uint8
	h, w, d = 0, 0, 0

	m.rGetArchDimensions(a, &h, &w, &d)

	return h, w, d
}
func (m *Manager) rGetArchDimensions(a *sdata.Archetype, h, w, d *uint8) {
	if *h == 0 && a.Height != 0 {
		*h = a.Height
	}
	if *w == 0 && a.Width != 0 {
		*w = a.Width
	}
	if *d == 0 && a.Depth != 0 {
		*d = a.Depth
	}
	if *h != 0 && *w != 0 && *d != 0 {
		return
	}

	archs := append(a.Archs, a.Arch)
	for _, name := range archs {
		if o := m.GetArchetype(name); o != nil {
			m.rGetArchDimensions(o, h, w, d)
		}
	}
}

func (m *Manager) GetArchImage(a *sdata.Archetype, scale float64) (img image.Image, err error) {
	anim, face := m.GetAnimAndFace(a, "", "")

	imgName, err := m.GetAnimFaceImage(anim, face)
	if err != nil {
		return nil, err
	}

	img = m.GetScaledImage(scale, imgName)
	if img == nil {
		return nil, errors.New("missing image")
	}

	// Didn't find anything, return missing image...
	return
}

func (m *Manager) cloneArchetype(archetype *sdata.Archetype) (nArchetype *sdata.Archetype) {
	nArchetype = &sdata.Archetype{}
	if err := mergo.Merge(nArchetype, archetype); err != nil {
		log.Printf("%+v", err)
		return nil
	}
	return nArchetype
}
