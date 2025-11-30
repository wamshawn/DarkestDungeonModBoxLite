package files

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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
	path      string
	dir       fs.FS
	scratches []string
}

func (df *DirFS) Path() string { return df.path }

func (df *DirFS) Size() int64 {
	info, infoErr := os.Stat(df.path)
	if infoErr != nil {
		if os.IsNotExist(infoErr) {
			return 0
		}
		panic(infoErr)
	}
	return info.Size()
}

func (df *DirFS) Rollback() {
	for _, entry := range df.scratches {
		_ = os.RemoveAll(entry)
	}
	df.scratches = df.scratches[:0]
}

func (df *DirFS) ListDir() (v []string, err error) {
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
	dir, file := filepath.Split(name)
	if dir == "" {
		data, err = fs.ReadFile(df.dir, name)
		return
	}
	dirs := splitDirs(dir)
	sub, subErr := NewDirFS(filepath.Join(df.path, dirs[0]))
	if subErr != nil {
		err = subErr
		return
	}
	data, err = sub.ReadFile(filepath.Join(filepath.Join(dirs[1:]...), file))
	return
}

func (df *DirFS) OpenFile(name string) (v *os.File, err error) {
	dir, file := filepath.Split(name)
	if dir == "" {
		v, err = os.OpenFile(filepath.Join(df.path, file), os.O_RDONLY, 0644)
		return
	}
	dirs := splitDirs(dir)
	sub, subErr := NewDirFS(filepath.Join(df.path, dirs[0]))
	if subErr != nil {
		err = subErr
		return
	}
	v, err = sub.OpenFile(filepath.Join(filepath.Join(dirs[1:]...), file))
	return
}

func (df *DirFS) WriteFile(name string, data []byte) (err error) {
	name = strings.TrimSpace(name)
	if name == "" {
		err = errors.New("name must not be empty")
		return
	}
	name = filepath.Clean(name)
	path := filepath.Join(df.path, name)
	dir := filepath.Dir(path)
	if exist, _ := Exist(dir); !exist {
		if err = Mkdir(dir); err != nil {
			return
		}
	}
	if err = os.WriteFile(filepath.Join(df.path, name), data, 0644); err != nil {
		return
	}
	df.addScratch(name)
	return
}

func (df *DirFS) CopyFile(name string, reader io.Reader) (err error) {
	name = strings.TrimSpace(name)
	if name == "" {
		err = errors.New("name must not be empty")
		return
	}
	name = filepath.Clean(name)
	path := filepath.Join(df.path, name)
	dir := filepath.Dir(path)
	if exist, _ := Exist(dir); !exist {
		if err = Mkdir(dir); err != nil {
			return
		}
	}
	file, openErr := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if openErr != nil {
		err = openErr
		return
	}
	_, err = io.Copy(file, reader)
	_ = file.Close()
	if err != nil {
		return
	}
	df.addScratch(name)
	return
}

func (df *DirFS) CreateDir(name string) (err error) {
	name = strings.TrimSpace(name)
	if name == "" {
		err = errors.New("name must not be empty")
		return
	}
	name = filepath.Clean(name)
	path := filepath.Join(df.path, name)
	dir := filepath.Dir(path)
	if exist, _ := Exist(dir); !exist {
		if err = Mkdir(dir); err != nil {
			return
		}
		df.addScratch(name)
		return
	}
	return
}

func (df *DirFS) addScratch(filename string) {
	dir, file := filepath.Split(filename)
	if dir == "" {
		for _, entry := range df.scratches {
			if entry == file {
				return
			}
		}
		df.scratches = append(df.scratches, file)
		return
	}
	dirs := splitDirs(dir)
	for _, entry := range df.scratches {
		if entry == dirs[0] {
			return
		}
	}
	df.scratches = append(df.scratches, dirs[0])
}

func splitDirs(name string) (dirs []string) {
	name = filepath.Clean(name)
	if name == "" || name == "." {
		return
	}
	dir, file := filepath.Split(name)
	if dir != "" {
		dir = filepath.Dir(dir)
		dirs = splitDirs(dir)
	}
	dirs = append(dirs, file)
	return
}
