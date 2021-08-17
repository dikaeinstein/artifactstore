package fsys

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

// InMemFS wraps fstest.MapFS to allow creating Files.
type InMemFS struct {
	m    map[string]fs.File
	root string
}

func NewInMemFS(root string) *InMemFS {
	return &InMemFS{m: make(map[string]fs.File), root: root}
}

func (inmem *InMemFS) Open(name string) (fs.File, error) {
	return nil, nil
}

func (inmem *InMemFS) Create(name string) (fs.File, error) {
	f, err := ioutil.TempFile(inmem.root, name)
	if err != nil {
		return f, err
	}

	inmem.m[name] = f
	return f, nil
}

func Cleanup(pattern string) error {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}

	return nil
}
