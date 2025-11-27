package files

import (
	"io/fs"
	"os"
	"path/filepath"
)

type Structure struct {
	Name     string      `json:"name"`
	IsDir    bool        `json:"isDir"`
	Children []Structure `json:"children"`
}

func FileStructure(filename string) (v *Structure, err error) {
	stat, statErr := os.Stat(filename)
	if statErr != nil {
		err = statErr
		return
	}
	v = &Structure{
		Name:     filepath.Base(filename),
		IsDir:    false,
		Children: nil,
	}
	if !stat.IsDir() {
		return
	}
	v.IsDir = true

	entries, dirErr := fs.ReadDir(os.DirFS(filename), ".")
	if dirErr != nil {
		err = dirErr
		return
	}
	for _, entry := range entries {
		child, childErr := FileStructure(filepath.Join(filename, entry.Name()))
		if childErr != nil {
			err = childErr
			return
		}
		v.Children = append(v.Children, *child)
	}
	return
}
