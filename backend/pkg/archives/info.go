package archives

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"

	"DarkestDungeonModBoxLite/backend/pkg/archives/pkg/ioutil"
)

type FileInfo struct {
	Name            string      `json:"name"`
	IsDir           bool        `json:"isDir"`
	Archived        bool        `json:"archived"`
	Encrypted       bool        `json:"encrypted"`
	Password        string      `json:"password"`
	PasswordInvalid bool        `json:"passwordInvalid"`
	Parent          *FileInfo   `json:"-"`
	Preview         []byte      `json:"-"`
	Children        []*FileInfo `json:"children"`
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
		if child.Name == topDir {
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
	result = info.add(append(dirs[1:], file), "", nil)
	return
}

func (info *FileInfo) mountFile(filename string, preview []byte) (result *FileInfo) {
	dirs, file := ioutil.Split(filepath.Clean(filename))
	result = info.add(dirs[1:], file, preview)
	return
}

func (info *FileInfo) mountArchiveFile(filename string, encrypted bool, passwordInvalid bool, password string) (result *FileInfo) {
	result = info.mountFile(filename, nil)
	result.Archived = true
	result.Encrypted = encrypted
	result.Password = password
	result.PasswordInvalid = passwordInvalid
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

func (info *FileInfo) Match(pattern string) (targets []*FileInfo) {
	path := info.Path()
	if matched, _ := filepath.Match(pattern, path); matched {
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

func (info *FileInfo) InvalidArchivedEntries() (targets []*FileInfo) {
	for _, child := range info.Children {
		r := child.InvalidArchivedEntries()
		if len(r) > 0 {
			targets = append(targets, r...)
		}
	}
	if info.Archived && info.Encrypted && info.PasswordInvalid {
		targets = append(targets, info)
		return
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
		return info.Name
	}
	items := []string{info.Name}
	parent := info.Parent
LOOP:
	if parent != nil {
		items = append(items, parent.Name)
		parent = parent.Parent
		goto LOOP
	}
	s := ""
	for i := len(items) - 1; i > -1; i-- {
		s = filepath.Join(s, items[i])
	}
	return filepath.ToSlash(s)
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

func (file *File) Info(ctx context.Context, preview ...string) (info *FileInfo, err error) {
	encrypted, encryptedErr := file.Encrypted(ctx)
	if encryptedErr != nil {
		err = encryptedErr
		return
	}
	password := file.option.GetPassword(file.Path())
	info = &FileInfo{
		Name:      file.name,
		IsDir:     false,
		Archived:  true,
		Encrypted: encrypted,
		Password:  password,
		Parent:    nil,
		Preview:   nil,
		Children:  nil,
	}
	err = file.Extract(ctx, func(ctx context.Context, entry *Entry) (err error) {
		filename := entry.Name()
		if entry.Info().IsDir() {
			info.mountDir(filename)
			return
		}
		if ok, entryEncrypted, entryPasswordInvalid, _ := entry.Archived(); ok {
			entryPassword := ""
			if entryEncrypted {
				entryPassword = file.option.GetPassword(filename)
			}
			info.mountArchiveFile(filename, entryEncrypted, entryPasswordInvalid, entryPassword)
			if entryPasswordInvalid { // when password invalid then skip
				err = ErrSkip
				return
			}
			// extract sub entry
			file.ExtractedEntry(filename)
			return
		}
		var data []byte
		for _, pv := range preview {
			if matched, _ := filepath.Match(pv, filename); matched {
				if data, err = io.ReadAll(entry); err != nil {
					return
				}
				break
			}
		}
		info.mountFile(filename, data)
		return
	})
	if err == nil {
		entries := info.ArchiveEntries()
		for _, entry := range entries {
			if entry.PasswordInvalid {
				var entryErr error
				if entry.Password == "" {
					entryErr = ErrPasswordRequired
				} else {
					entryErr = ErrPasswordInvalid
				}
				if err == nil {
					err = FileError{
						Filename: entry.Path(),
						Err:      entryErr,
					}
				} else {
					err = errors.Join(err, FileError{
						Filename: entry.Path(),
						Err:      entryErr,
					})
				}
			}
		}
	}
	return
}
