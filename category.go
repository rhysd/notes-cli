package notes

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Category represents a category directory which contains some notes
type Category struct {
	// Path is a path to the category directory
	Path string
	// Name is a name of category
	Name string
	// NotePaths are paths to notes of the category
	NotePaths []string
}

// Notes returns all Note instances which belong to the category
func (cat *Category) Notes(c *Config) ([]*Note, error) {
	notes := make([]*Note, 0, len(cat.NotePaths))
	for _, p := range cat.NotePaths {
		n, err := LoadNote(p, c)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}

// Categories is a map from category name to Category instance
type Categories map[string]*Category

// Names returns all category names as slice
func (cats Categories) Names() []string {
	ss := make([]string, 0, len(cats))
	for n := range cats {
		ss = append(ss, n)
	}
	return ss
}

// Notes returns all Note instances which belong to the categories
func (cats Categories) Notes(cfg *Config) ([]*Note, error) {
	numNotes := 0
	for _, c := range cats {
		numNotes += len(c.NotePaths)
	}

	notes := make([]*Note, 0, numNotes)
	for _, c := range cats {
		for _, p := range c.NotePaths {
			n, err := LoadNote(p, cfg)
			if err != nil {
				return nil, err
			}
			notes = append(notes, n)
		}
	}
	return notes, nil
}

// CollectCategories collects all categories under home
func CollectCategories(cfg *Config) (Categories, error) {
	cats := Categories(map[string]*Category{})

	fs, err := ioutil.ReadDir(cfg.HomePath)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot read home")
	}

	for _, f := range fs {
		name := f.Name()
		if !f.IsDir() || strings.HasPrefix(name, ".") {
			continue
		}

		root := filepath.Join(cfg.HomePath, name)
		if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			path = normPathNFD(path)
			name := info.Name()

			if info.IsDir() {
				if strings.HasPrefix(name, ".") {
					return filepath.SkipDir
				}
				rel := strings.TrimPrefix(path, cfg.HomePath)
				n := strings.TrimPrefix(filepath.ToSlash(rel), "/")
				cats[n] = &Category{Name: n, Path: path}
				return nil
			}

			if strings.HasPrefix(name, ".") || !strings.HasSuffix(name, ".md") {
				return nil
			}

			rel := strings.TrimPrefix(filepath.Dir(path), cfg.HomePath)
			cat := cats[strings.TrimPrefix(filepath.ToSlash(rel), "/")]
			cat.NotePaths = append(cat.NotePaths, path)
			return nil
		}); err != nil {
			return nil, errors.Wrapf(err, "Cannot walk on directory for category %q", name)
		}
	}

	// Remove category which has no note
	for n, c := range cats {
		if len(c.NotePaths) == 0 {
			delete(cats, n)
		}
	}

	return cats, nil
}
