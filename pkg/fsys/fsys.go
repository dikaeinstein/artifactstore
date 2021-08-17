package fsys

import (
	"fmt"
	"io/fs"
)

// CreatFS is a filesystem that can create a new file.
type CreatFS interface {
	fs.FS
	Create(name string) (fs.File, error)
}

// Create a new file using the given filesystem.
func Create(fsys fs.FS, name string) (fs.File, error) {
	if fsys, ok := fsys.(CreatFS); ok {
		return fsys.Create(name)
	}

	return nil, fmt.Errorf("create %s: operation not supported", name)
}
