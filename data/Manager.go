package data

import (
	"errors"
	"image"
	"log"
	"reflect"
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
	EditorConfig        EditorConfig
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

	// Read editor config
	editorConfigPath := filepath.Join(m.EtcPath, "editor-config.yaml")
	r, err := ioutil.ReadFile(editorConfigPath)
	if err != nil {
		m.EditorConfig = EditorConfig{
			filePath: editorConfigPath,
		}
		if err := m.EditorConfig.Save(); err != nil {
			return err
		}
	} else {
		if err = yaml.Unmarshal(r, &m.EditorConfig); err != nil {
			return err
		}
	}
	m.EditorConfig.filePath = editorConfigPath

	// Read animations config
	animationsConfigPath := filepath.Join(m.ArchetypesPath, "config.yaml")
	r, err = ioutil.ReadFile(animationsConfigPath)
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

func (m *Manager) ReloadArchetypes() {
	m.archetypes = make(map[string]*sdata.Archetype)
	m.archetypeFiles = make(map[string][]string)
	m.archetypesOrder = make([]string, 0)
	if err := m.LoadArchetypes(); err != nil {
		return
	}
	log.Printf("Loaded %d archetypes\n", len(m.archetypes))
}

func (m *Manager) ReloadAnimations() error {
	m.animationFiles = make(map[string]map[string]struct{})
	m.animations = make(map[string]sdata.AnimationPre)

	animationsConfigPath := filepath.Join(m.ArchetypesPath, "config.yaml")
	r, err := ioutil.ReadFile(animationsConfigPath)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(r, &m.AnimationsConfig); err != nil {
		return err
	}

	if err = m.LoadAnimations(); err != nil {
		return err
	}

	if err = m.LoadImages(); err != nil {
		return err
	}
	log.Printf("Cached %d images\n", len(m.images))
	return nil
}

func (m *Manager) ReloadImages() error {
	m.images = make(map[string]image.Image)
	m.scaledImages = make(map[float64]map[string]image.Image)

	if err := m.LoadImages(); err != nil {
		return err
	}
	log.Printf("Cached %d images\n", len(m.images))
	return nil
}

// GetDataPath gets a path relative to the data path directory.
func (m *Manager) GetDataPath(parts ...string) string {
	return path.Join(m.DataPath, filepath.Clean(fmt.Sprintf("%c", filepath.Separator)+filepath.Join(parts...)))
}

func (m *Manager) GetRelativeMapPath(parts ...string) (string, error) {
	p := path.Join(parts...)
	return filepath.Rel(m.MapsPath, p)
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

func (m *Manager) GetAnimFaceFrames(anim, face string) (f []sdata.AnimationFramePre, e error) {
	a, ok := m.animations[anim]
	if !ok {
		return f, errors.New("missing animation")
	}
	f, ok = a.Faces[face]
	if !ok {
		return f, errors.New("missing face")
	}
	return f, nil
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

// GetMissingArchAncestors returns a string slice of missing archetype names.
func (m *Manager) GetMissingArchAncestors(a *sdata.Archetype) []string {
	var missing []string
	for _, archName := range a.Archs {
		if m.GetArchetype(archName) == nil {
			missing = append(missing, archName)
		}
	}
	return missing
}

// GetArchField does a proper hierarchical lookup for a given field.
func (m *Manager) GetArchField(a *sdata.Archetype, field string) reflect.Value {
	s := reflect.ValueOf(a).Elem()
	f := s.FieldByName(field)
	f2 := f
	if f.IsValid() && f.Kind() == reflect.Ptr && !f.IsNil() {
		return f
	} else if f.IsValid() && f.Kind() != reflect.Ptr {
		return f
	}
	f = m.GetArchAncestryField(a, field)
	if !f.IsValid() {
		return f2
	}
	return f
}

// GetArchAncestryField gets the arch's ancestry value for a given field.
func (m *Manager) GetArchAncestryField(a *sdata.Archetype, field string) reflect.Value {
	var f reflect.Value
	if a.Arch != "" {
		a2 := m.GetArchetype(a.Arch)
		if a2 != nil {
			f = m.GetArchField(a2, field)
			if f.IsValid() && ((f.Kind() == reflect.Ptr && !f.IsNil()) || f.Kind() != reflect.Ptr) {
				return f
			}
		}
	}
	for _, archName := range a.Archs {
		a2 := m.GetArchetype(archName)
		if a2 != nil {
			f = m.GetArchField(a2, field)
			if f.IsValid() && ((f.Kind() == reflect.Ptr && !f.IsNil()) || f.Kind() != reflect.Ptr) {
				return f
			}
		}
	}
	return f
}

// IsArchFieldLocal returns whether or not a given field is defined on the archetype or if it is from its ancestry.
func (m *Manager) IsArchFieldLocal(a *sdata.Archetype, field string) (bool, error) {
	s := reflect.ValueOf(a).Elem()
	f := s.FieldByName(field)
	if !f.IsValid() {
		return false, fmt.Errorf("field \"%s\" does not exist", field)
	}
	if !f.CanSet() {
		return false, fmt.Errorf("field \"%s\" cannot be set", field)
	}
	if f.Kind() == reflect.Ptr {
		if f.IsNil() {
			return false, nil
		}
		return true, nil
		// TODO: At the moment only strings exist as points, we need MatterType and others to also be pointers.
	}
	// Look up ancestry
	ancestorSame := true
	if a.Arch != "" {
		a2 := m.GetArchetype(a.Arch)
		if a2 != nil {
			f2 := m.GetArchField(a2, field)
			if !m.IsEqual(f, f2) {
				ancestorSame = false
			}
		}
	}
	for _, archName := range a.Archs {
		a2 := m.GetArchetype(archName)
		if a2 != nil {
			f2 := m.GetArchField(a2, field)
			if !m.IsEqual(f, f2) {
				ancestorSame = false
			}
		}
	}

	return !ancestorSame, nil
}

func (m *Manager) IsEqual(v1, v2 reflect.Value) bool {
	if v1.Kind() != v2.Kind() {
		return false
	}
	if v1.Interface() != v2.Interface() {
		return false
	}
	return true
}

func (m *Manager) SetArchField(a *sdata.Archetype, field string, v interface{}) error {
	s := reflect.ValueOf(a).Elem()
	f := s.FieldByName(field)
	if !f.IsValid() {
		return fmt.Errorf("field \"%s\" does not exist", field)
	}
	if !f.CanSet() {
		return fmt.Errorf("field \"%s\" cannot be set", field)
	}
	fmt.Println("set", field, v)
	if f.Kind() == reflect.Ptr {
		switch t := v.(type) {
		case string:
			s := new(string)
			*s = t
			f.Set(reflect.ValueOf(s))
		case *string:
			s := new(string)
			*s = *t
			f.Set(reflect.ValueOf(t))
		default:
			return fmt.Errorf("passed interface of %v is not a string or string pointer", v)
		}
	} else if f.Kind() == reflect.Uint8 {
		switch t := v.(type) {
		case uint8:
			f.Set(reflect.ValueOf(t))
		case uint16:
			f.Set(reflect.ValueOf(t))
		case uint32:
			f.Set(reflect.ValueOf(t))
		default:
			return fmt.Errorf("passed interface of %v is not a uint8", v)
		}
	} else if f.Kind() == reflect.Uint32 {
		switch t := v.(type) {
		case uint32:
			f.Set(reflect.ValueOf(t))
		case uint16:
			f.Set(reflect.ValueOf(t))
		case uint8:
			f.Set(reflect.ValueOf(t))
		default:
			return fmt.Errorf("passed interface of %v is not a uint32", v)
		}
	} else if f.Kind() == reflect.Int8 {
		switch t := v.(type) {
		case int8:
			f.Set(reflect.ValueOf(t))
		case int16:
			f.Set(reflect.ValueOf(t))
		case int32:
			f.Set(reflect.ValueOf(t))
		default:
			return fmt.Errorf("passed interface of %v is not a uint32", v)
		}
	} else if f.Kind() == reflect.Float32 {
		switch t := v.(type) {
		case float32:
			f.Set(reflect.ValueOf(t))
		default:
			return fmt.Errorf("passed interface of %v is not a float32", v)
		}
	} else {
		// FIXME: This is a bad catchall.
		f.Set(reflect.ValueOf(v))
	}
	return nil
}

func (m *Manager) ClearArchField(a *sdata.Archetype, field string) bool {
	f := m.GetArchField(a, field)
	if !f.IsValid() {
		return false
	}
	if f.Kind() == reflect.Ptr {
		f.Set(reflect.Zero(f.Type()))
		return true
	} else {
		f.Set(reflect.Zero(f.Type()))
		return true
	}
	return false
}

func (m *Manager) GetArchName(a *sdata.Archetype, name string) string {
	if name != "" {
		return name
	}

	v := m.GetArchField(a, "Name")
	if v.IsValid() && !v.IsNil() {
		return v.Elem().String()
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
