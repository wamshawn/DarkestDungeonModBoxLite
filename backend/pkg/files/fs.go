package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type FileInfo struct {
	Name     string
	Filepath string
	Size     int64
	IsDir    bool
	Children []*FileInfo
}

func NewDirFS(filename string) (v *DirFS, err error) {
	stat, statErr := os.Stat(filename)
	if statErr != nil {
		err = statErr
		return
	}
	if !stat.IsDir() {
		err = errors.New(filename + " is not a directory")
		return
	}
	v = &DirFS{
		path: filename,
		dir:  os.DirFS(filename),
	}
	return
}

type DirFS struct {
	path string
	dir  fs.FS
}

func (df *DirFS) Path() string { return df.path }

func (df *DirFS) DirList() (v []string, err error) {
	entries, dirErr := fs.ReadDir(df.dir, ".")
	if dirErr != nil {
		err = dirErr
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			v = append(v, entry.Name())
		}
	}
	return
}

func (df *DirFS) Dir(name string) (v *DirFS) {
	path := filepath.Join(df.path, name)
	var err error
	v, err = NewDirFS(path)
	if err != nil {
		panic(err)
	}
	return
}

func (df *DirFS) DirInfo(name string) (v *FileInfo, err error) {
	stat, statErr := fs.Stat(df.dir, name)
	if statErr != nil {
		err = statErr
		return
	}
	if !stat.IsDir() {
		err = errors.New(name + " is not a directory")
		return
	}
	v = &FileInfo{
		Name:     name,
		Filepath: filepath.Join(df.path, name),
		Size:     stat.Size(),
		IsDir:    stat.IsDir(),
		Children: nil,
	}
	entries, dirErr := fs.ReadDir(df.dir, name)
	if dirErr != nil {
		err = dirErr
		return
	}
	if len(entries) == 0 {
		return
	}
	v.Children = make([]*FileInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			subFS, subFSErr := NewDirFS(filepath.Join(df.path, entry.Name()))
			if subFSErr != nil {
				err = subFSErr
				return
			}
			sub, subErr := subFS.DirInfo(".")
			if subErr != nil {
				err = subErr
				return
			}
			v.Children = append(v.Children, sub)
			continue
		}
		sub, subErr := df.FileInfo(entry.Name())
		if subErr != nil {
			err = subErr
			return
		}
		v.Children = append(v.Children, sub)
	}
	return
}

func (df *DirFS) FileInfo(name string) (v *FileInfo, err error) {
	stat, statErr := fs.Stat(df.dir, name)
	if statErr != nil {
		err = statErr
		return
	}
	if stat.IsDir() {
		err = errors.New(name + " is a directory")
		return
	}
	v = &FileInfo{
		Name:     name,
		Filepath: filepath.Join(df.path, name),
		Size:     stat.Size(),
		IsDir:    stat.IsDir(),
		Children: nil,
	}
	return
}

func (df *DirFS) ReadFile(name string) (data []byte, err error) {
	return fs.ReadFile(df.dir, name)
}

func (df *DirFS) WriteFile(name string, data []byte) (err error) {
	err = os.WriteFile(filepath.Join(df.path, name), data, 0644)
	return
}
