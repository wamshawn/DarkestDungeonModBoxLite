package files

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mholt/archives"
)

type ArchiveFileInfo struct {
	Name     string
	IsDir    bool
	Children []*ArchiveFileInfo
	Parent   *ArchiveFileInfo `json:"-"`
}

func (info *ArchiveFileInfo) Add(dirs []string, file string) {
	if len(dirs) == 0 {
		if file == "" {
			return
		}
		info.Children = append(info.Children, &ArchiveFileInfo{
			Name:     file,
			IsDir:    false,
			Children: nil,
			Parent:   info,
		})
		return
	}
	topDir := dirs[0]
	for _, child := range info.Children {
		if child.IsDir && child.Name == topDir {
			child.Add(dirs[1:], file)
			return
		}
	}
	child := &ArchiveFileInfo{
		Name:     topDir,
		IsDir:    true,
		Children: nil,
		Parent:   info,
	}
	child.Add(dirs[1:], file)
	info.Children = append(info.Children, child)
}

func (info *ArchiveFileInfo) Find(name string) (targets []*ArchiveFileInfo) {
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

type ArchiveInfo struct {
	MediaType string
	Extension string
	Entries   []*ArchiveFileInfo
}

func (info *ArchiveInfo) Mount(name string, isDir bool) {
	dir := ""
	file := ""
	if !isDir {
		dir, file = filepath.Split(name)
		dir = filepath.Dir(dir)
	} else {
		dir = filepath.Dir(name)
	}
	dirs := info.splitDirs(dir)
	if len(dirs) == 0 {
		if isDir {
			info.Entries = append(info.Entries, &ArchiveFileInfo{
				Name:     dir,
				IsDir:    true,
				Children: nil,
			})
		} else {
			info.Entries = append(info.Entries, &ArchiveFileInfo{
				Name:     file,
				IsDir:    false,
				Children: nil,
			})
		}
		return
	}
	topDir := dirs[0]
	for _, entry := range info.Entries {
		if entry.Name == topDir {
			entry.Add(dirs[1:], file)
			return
		}
	}
	entry := &ArchiveFileInfo{
		Name:     topDir,
		IsDir:    true,
		Children: nil,
	}
	entry.Add(dirs[1:], file)
	info.Entries = append(info.Entries, entry)
}

func (info *ArchiveInfo) Find(name string) (targets []*ArchiveFileInfo) {
	for _, entry := range info.Entries {
		r := entry.Find(name)
		if len(r) > 0 {
			targets = append(targets, r...)
		}
	}
	return
}

func (info *ArchiveInfo) String() string {
	b, _ := json.MarshalIndent(info, "", "\t")
	return string(b)
}

func (info *ArchiveInfo) splitDirs(name string) (dirs []string) {
	dir, file := filepath.Split(name)
	if dir != "" {
		dir = filepath.Dir(dir)
		dirs = info.splitDirs(dir)
	}
	dirs = append(dirs, file)
	return
}

func WalkArchiveInfo(ctx context.Context, filename string, password string) (archive *ArchiveInfo, err error) {
	file, openErr := os.Open(filename)
	if openErr != nil {
		err = openErr
		return
	}
	defer file.Close()

	format, _, identifyErr := archives.Identify(ctx, ".", file)
	if identifyErr != nil {
		err = identifyErr
		return
	}

	extractor, ok := format.(archives.Extractor)
	if !ok {
		err = fmt.Errorf("%s is not supported", format.Extension())
		return
	}
	if password != "" {
		switch format.Extension() {
		case ".zip":
			extractor = CryptoZip{
				Zip:      extractor.(archives.Zip),
				Password: password,
			}
		case ".7z":
			ex := extractor.(archives.SevenZip)
			ex.Password = password
			extractor = ex
		case ".rar":
			ex := extractor.(archives.Rar)
			ex.Password = password
			extractor = ex
		default:
			err = fmt.Errorf("%s is not supported", format.Extension())
			return
		}
	}

	archive = &ArchiveInfo{
		MediaType: format.MediaType(),
		Extension: format.Extension(),
		Entries:   nil,
	}

	err = extractor.Extract(ctx, file, func(ctx context.Context, info archives.FileInfo) (err error) {
		if info.IsDir() {
			archive.Mount(info.NameInArchive, true)
			return
		}
		item, itemErr := info.Open()
		if itemErr != nil {
			err = itemErr
			return
		}
		b := make([]byte, 64)
		rn, rErr := item.Read(b)

		if rn == 0 && rErr != nil {
			if errors.Is(rErr, io.EOF) {

			}
			_ = item.Close()
			err = rErr
			return
		}
		_ = item.Close()
		archive.Mount(info.NameInArchive, false)
		return
	})

	return
}
