package archives

import (
	"context"
	"encoding/json"
	"path/filepath"

	"DarkestDungeonModBoxLite/backend/pkg/archives/pkg/ioutil"

	"github.com/tidwall/match"
)

type FileInfo struct {
	Name             string      `json:"name"`
	IsDir            bool        `json:"isDir"`
	Archived         bool        `json:"archived"`
	Password         string      `json:"password"`
	PasswordRequired bool        `json:"passwordRequired"`
	Parent           *FileInfo   `json:"-"`
	Preview          []byte      `json:"-"`
	Children         []*FileInfo `json:"children"`
}

func (info *FileInfo) add(dirs []string, file string, preview []byte) (result *FileInfo) {
	if len(dirs) == 0 {
		if file == "" {
			return
		}
		result = &FileInfo{
			Name:     file,
			IsDir:    false,
			Children: nil,
			Parent:   info,
			Preview:  preview,
		}
		info.Children = append(info.Children, result)
		return
	}
	topDir := dirs[0]
	for _, child := range info.Children {
		if child.IsDir && child.Name == topDir {
			result = child.add(dirs[1:], file, preview)
			return
		}
	}
	child := &FileInfo{
		Name:     topDir,
		IsDir:    true,
		Children: nil,
		Parent:   info,
	}
	info.Children = append(info.Children, child)
	result = child.add(dirs[1:], file, preview)
	return
}

func (info *FileInfo) mountDir(filename string) (result *FileInfo) {
	dirs, file := ioutil.Split(filepath.Clean(filename))
	result = info.add(append(dirs, file), "", nil)
	return
}

func (info *FileInfo) mountFile(filename string, preview []byte) (result *FileInfo) {
	dirs, file := ioutil.Split(filepath.Clean(filename))
	result = info.add(dirs, file, preview)
	return
}

func (info *FileInfo) mountArchiveFile(filename string, child *FileInfo) (result *FileInfo) {
	result = info.mountFile(filename, nil)
	result.Archived = true
	result.Password = child.Password
	result.PasswordRequired = child.PasswordRequired
	result.Children = child.Children
	for _, c := range result.Children {
		c.Parent = result
	}
	return
}

func (info *FileInfo) get(filename string) (target *FileInfo) {
	dirs, file := ioutil.Split(filepath.Clean(filename))
	if len(dirs) == 0 {
		for _, child := range info.Children {
			if child.Name == file {
				return child
			}
		}
		return
	}
	for _, child := range info.Children {
		if child.Name == dirs[0] {
			return child.get(filepath.Join(filepath.Join(dirs[1:]...), file))
		}
	}
	return
}

func (info *FileInfo) Find(name string) (targets []*FileInfo) {
	if info.Name == name {
		targets = append(targets, info)
		return
	}
	for _, child := range info.Children {
		r := child.Find(name)
		if len(r) > 0 {
			targets = append(targets, r...)
		}
	}
	return
}

func (info *FileInfo) Match(pattern string) (targets []*FileInfo) {
	path := info.Path()
	if match.Match(path, pattern) {
		targets = append(targets, info)
	}
	for _, child := range info.Children {
		r := child.Match(pattern)
		if len(r) > 0 {
			targets = append(targets, r...)
		}
	}
	return
}

func (info *FileInfo) Root() *FileInfo {
	parent := info.Parent
LOOP:
	if parent == nil {
		return info
	}
	parent = parent.Parent
	goto LOOP
}

func (info *FileInfo) Path() string {
	if info.Name == "" {
		return ""
	}
	if info.Parent == nil {
		return ""
	}
	items := []string{info.Name}
	parent := info.Parent
LOOP:
	if parent != nil {
		if parent.Parent != nil {
			items = append(items, parent.Name)
		}
		parent = parent.Parent
		goto LOOP
	}
	s := ""
	for i := len(items) - 1; i > -1; i-- {
		s = filepath.Join(s, items[i])
	}
	return s
}

func (info *FileInfo) ArchiveEntries() (entries []*FileInfo) {
	if info.Archived {
		entries = append(entries, info)
	}
	for _, child := range info.Children {
		entries = append(entries, child.ArchiveEntries()...)
	}
	return
}

func (info *FileInfo) String() string {
	b, _ := json.MarshalIndent(info, "", "\t")
	return string(b)
}

func (file *File) Info(ctx context.Context) (info *FileInfo, err error) {

	return
}
