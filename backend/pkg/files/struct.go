package files

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Structure struct {
	Name     string      `json:"name"`
	IsDir    bool        `json:"isDir"`
	Children []Structure `json:"children"`
}

func (st *Structure) Empty() bool {
	return len(st.Children) == 0
}

func (st *Structure) OnlyFilesByExt(ext string) (ok bool) {
	if st.IsDir {
		for _, child := range st.Children {
			if ok = child.OnlyFilesByExt(ext); !ok {
				return
			}
		}
		return
	}
	ok = strings.ToLower(ext) == strings.ToLower(filepath.Ext(st.Name))
	return
}

func (st *Structure) Get(name string) (target Structure, has bool) {
	dir, file := filepath.Split(name)
	if dir == "" {
		for _, child := range st.Children {
			if child.Name == file {
				target = child
				has = true
				return
			}
		}
		return
	}
	dirs := splitDirs(dir)
	for _, dir0 := range dirs {
		for _, child := range st.Children {
			if child.Name == dir0 {
				target, has = child.Get(filepath.Join(filepath.Join(dirs[1:]...), file))
				return
			}
		}
	}
	return
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
