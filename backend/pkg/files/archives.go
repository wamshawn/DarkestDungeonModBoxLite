package files

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archives"
)

type ArchiveFileInfo struct {
	Name     string
	IsDir    bool
	Archived bool
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

func (info *ArchiveFileInfo) Merge(archive string, dirs []string, targets []*ArchiveFileInfo) {
	if len(dirs) == 0 {
		if len(targets) == 0 {
			return
		}
		archiveInfo := &ArchiveFileInfo{
			Name:     archive,
			IsDir:    false,
			Archived: true,
			Children: targets,
			Parent:   info,
		}
		for _, child := range archiveInfo.Children {
			child.Parent = archiveInfo
		}
		info.Children = append(info.Children, archiveInfo)
		return
	}
	topDir := dirs[0]
	for _, child := range info.Children {
		if child.IsDir && child.Name == topDir {
			child.Merge(archive, dirs[1:], targets)
			return
		}
	}
	child := &ArchiveFileInfo{
		Name:     topDir,
		IsDir:    true,
		Archived: false,
		Children: nil,
		Parent:   info,
	}
	child.Merge(archive, dirs[1:], targets)
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

func (info *ArchiveFileInfo) Path() string {
	if info.Name == "" {
		return ""
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
		s = s + "/" + items[i]
	}
	return s[1:]
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

func (info *ArchiveInfo) Merge(path string, target *ArchiveInfo) {
	dir, filename := filepath.Split(path)
	dir = filepath.Dir(dir)
	dirs := info.splitDirs(dir)
	if len(dirs) == 0 {
		info.Entries = append(info.Entries, &ArchiveFileInfo{
			Name:     filename,
			IsDir:    false,
			Archived: true,
			Children: target.Entries,
		})
		return
	}
	topDir := dirs[0]
	for _, entry := range info.Entries {
		if entry.Name == topDir {
			entry.Merge(filename, dirs[1:], target.Entries)
			return
		}
	}
	entry := &ArchiveFileInfo{
		Name:     topDir,
		IsDir:    true,
		Children: nil,
	}
	entry.Merge(filename, dirs[1:], target.Entries)
	info.Entries = append(info.Entries, entry)
	return
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
	name = filepath.Clean(name)
	if name == "" || name == "." {
		return
	}
	dir, file := filepath.Split(name)
	if dir != "" {
		dir = filepath.Dir(dir)
		dirs = info.splitDirs(dir)
	}
	dirs = append(dirs, file)
	return
}

type ctxArchivePasswordKey struct {
}

var (
	_ctxArchivePasswordKey = ctxArchivePasswordKey{}
)

func GetArchiveInfo(ctx context.Context, filename string, password string) (archive *ArchiveInfo, err error) {
	file, openErr := os.Open(filename)
	if openErr != nil {
		err = openErr
		return
	}
	if password != "" {
		ctx = context.WithValue(ctx, _ctxArchivePasswordKey, password)
	}
	archive, err = getArchiveInfo(ctx, filename, file, false)
	_ = file.Close()
	if err != nil {
		if password != "" {
			file, _ = os.Open(filename)
			archive, err = getArchiveInfo(ctx, filename, file, true)
			_ = file.Close()
			return
		}
		return
	}
	return
}

func getArchiveInfo(ctx context.Context, filename string, reader io.Reader, password bool) (archive *ArchiveInfo, err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	// identify
	format, _, identifyErr := archives.Identify(ctx, filename, reader)
	if identifyErr != nil {
		err = identifyErr
		return
	}
	// extractor
	extractor, ok := format.(archives.Extractor)
	if !ok {
		err = fmt.Errorf("%s is not supported", format.Extension())
		return
	}
	if password {
		pwd0 := ctx.Value(_ctxArchivePasswordKey)
		if pwd0 == nil {
			err = errors.New("no archive password in context")
			return
		}
		pwd := pwd0.(string)
		switch format.Extension() {
		case ".zip":
			extractor = CryptoZip{
				Zip:      extractor.(archives.Zip),
				Password: pwd,
			}
		case ".7z":
			ex := extractor.(archives.SevenZip)
			ex.Password = pwd
			extractor = ex
		case ".rar":
			ex := extractor.(archives.Rar)
			ex.Password = pwd
			extractor = ex
		default:
			err = fmt.Errorf("%s is not supported", filename)
			return
		}
	}

	archive = &ArchiveInfo{
		MediaType: format.MediaType(),
		Extension: format.Extension(),
		Entries:   nil,
	}

	err = extractor.Extract(ctx, reader, func(ctx context.Context, info archives.FileInfo) (err error) {
		if err = ctx.Err(); err != nil {
			return
		}
		if info.IsDir() {
			archive.Mount(info.NameInArchive, true)
			return
		}
		item, itemErr := info.Open()
		if itemErr != nil {
			err = itemErr
			return
		}
		defer item.Close()
		var itemReader io.Reader = item
		var tmp *TempDir
		var tmpFile *os.File
		ext := filepath.Ext(info.Name())
		archived := false
		switch strings.ToLower(ext) {
		case ".zip", ".7z", ".rar":
			archived = true
			break
		default:
			if info.Size() < 64*1024*1024 {
				b, bErr := io.ReadAll(itemReader)
				if bErr != nil {
					err = bErr
					return
				}
				itemReader = bytes.NewReader(b)
				_, archived = IsArchiveFile(bytes.NewReader(b))
				break
			}
			tmp, err = CreateTempDir("archives_*")
			if err != nil {
				return
			}
			cpErr := tmp.Copy(info.Name(), itemReader)
			if cpErr != nil {
				_ = tmp.Remove()
				err = cpErr
				return
			}
			tmpFile, err = tmp.OpenFile(info.Name())
			if err != nil {
				return
			}
			_, archived = IsArchiveFile(tmpFile)
			_, _ = tmpFile.Seek(0, io.SeekStart)
			itemReader = tmpFile
			break
		}
		if !archived {
			if tmpFile != nil {
				_ = tmpFile.Close()
			}
			if tmp != nil {
				_ = tmp.Remove()
			}
			archive.Mount(info.NameInArchive, false)
			return
		}
		if tmp == nil {
			tmp, err = CreateTempDir("archives_*")
			if err != nil {
				return
			}
		}
		defer tmp.Remove()
		if tmpFile == nil {
			cpErr := tmp.Copy(info.Name(), itemReader)
			if cpErr != nil {
				return
			}
			tmpFile, err = tmp.OpenFile(info.Name())
			if err != nil {
				return
			}
		}
		needPassword := false
	SUB:
		sub, subErr := getArchiveInfo(ctx, info.Name(), tmpFile, needPassword)
		_ = tmpFile.Close()
		if subErr != nil {
			if needPassword {
				err = subErr
				return
			}
			tmpFile, err = tmp.OpenFile(info.Name())
			if err != nil {
				return
			}
			needPassword = true
			goto SUB
		}
		archive.Merge(info.NameInArchive, sub)
		return
	})
	return
}

type ExtractArchiveHandler func(ctx context.Context, filename string, archived bool) (dst string, extract bool, err error)

func ExtractArchive(ctx context.Context, dstDir string, filename string, password string, handler ExtractArchiveHandler) (err error) {
	exist, existErr := Exist(dstDir)
	if existErr != nil {
		err = existErr
		return
	}
	if exist {
		err = errors.New("dst dir exists")
		return
	}
	if err = Mkdir(dstDir); err != nil {
		return
	}

	file, openErr := os.Open(filename)
	if openErr != nil {
		_ = os.RemoveAll(dstDir)
		err = openErr
		return
	}
	if password != "" {
		ctx = context.WithValue(ctx, _ctxArchivePasswordKey, password)
	}
	err = extractArchive(ctx, "", dstDir, filename, file, false, handler)
	_ = file.Close()
	if err != nil {
		if password != "" {
			file, _ = os.Open(filename)
			err = extractArchive(ctx, "", dstDir, filename, file, false, handler)
			if err != nil {
				_ = os.RemoveAll(dstDir)
			}
			_ = file.Close()
			return
		}
		_ = os.RemoveAll(dstDir)
		return
	}
	return
}

func extractArchive(ctx context.Context, prefix string, dstDir string, filename string, reader io.Reader, password bool, handler ExtractArchiveHandler) (err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	// identify
	format, _, identifyErr := archives.Identify(ctx, filename, reader)
	if identifyErr != nil {
		err = identifyErr
		return
	}
	// extractor
	extractor, ok := format.(archives.Extractor)
	if !ok {
		err = fmt.Errorf("%s is not supported", format.Extension())
		return
	}
	if password {
		pwd0 := ctx.Value(_ctxArchivePasswordKey)
		if pwd0 == nil {
			err = errors.New("no archive password in context")
			return
		}
		pwd := pwd0.(string)
		switch format.Extension() {
		case ".zip":
			extractor = CryptoZip{
				Zip:      extractor.(archives.Zip),
				Password: pwd,
			}
		case ".7z":
			ex := extractor.(archives.SevenZip)
			ex.Password = pwd
			extractor = ex
		case ".rar":
			ex := extractor.(archives.Rar)
			ex.Password = pwd
			extractor = ex
		default:
			err = fmt.Errorf("%s is not supported", filename)
			return
		}
	}

	err = extractor.Extract(ctx, reader, func(ctx context.Context, info archives.FileInfo) (err error) {
		if err = ctx.Err(); err != nil {
			return
		}
		target := filepath.Join(prefix, info.NameInArchive)
		if info.IsDir() {
			dst, _, hErr := handler(ctx, target, false)
			if hErr != nil {
				err = hErr
				return
			}
			if dst != "" {
				dst = filepath.Join(dstDir, dst)
				exist, existErr := Exist(dst)
				if existErr != nil {
					err = existErr
					return
				}
				if !exist {
					if err = Mkdir(dst); err != nil {
						return
					}
				}
			}
			return
		}
		item, itemErr := info.Open()
		if itemErr != nil {
			err = itemErr
			return
		}
		defer item.Close()

		var itemReader io.Reader = item
		var tmp *TempDir
		var tmpFile *os.File
		ext := filepath.Ext(info.Name())
		archived := false
		switch strings.ToLower(ext) {
		case ".zip", ".7z", ".rar":
			archived = true
			break
		default:
			if info.Size() < 64*1024*1024 {
				b, bErr := io.ReadAll(itemReader)
				if bErr != nil {
					err = bErr
					return
				}
				itemReader = bytes.NewReader(b)
				_, archived = IsArchiveFile(bytes.NewReader(b))
				break
			}
			tmp, err = CreateTempDir("archives_*")
			if err != nil {
				return
			}
			cpErr := tmp.Copy(info.Name(), itemReader)
			if cpErr != nil {
				_ = tmp.Remove()
				err = cpErr
				return
			}
			tmpFile, err = tmp.OpenFile(info.Name())
			if err != nil {
				return
			}
			_, archived = IsArchiveFile(tmpFile)
			_, _ = tmpFile.Seek(0, io.SeekStart)
			itemReader = tmpFile
			break
		}
		if !archived {
			if tmp != nil {
				defer tmp.Remove()
			}
			if tmpFile != nil {
				defer tmpFile.Close()
			}
			dst, _, hErr := handler(ctx, target, false)
			if hErr != nil {
				err = hErr
				return
			}
			dstFile, dstErr := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if dstErr != nil {
				err = dstErr
				return
			}
			defer dstFile.Close()

			_, cpErr := io.Copy(dstFile, itemReader)
			if cpErr != nil {
				err = cpErr
				return
			}
			return
		}
		// todo

		return
	})
	return
}

var (
	compressionFormats = []struct {
		magic  string
		mime   string
		format string
	}{
		{"504B0304", "application/zip", "zip"},
		{"1F8B08", "application/gzip", "gzip"},
		{"377ABCAF271C", "application/x-7z-compressed", "7z"},
		{"526172211A0700", "application/x-rar-compressed", "rar"},
		{"526172211A070100", "application/x-rar-compressed", "rar"},
		{"7573746172", "application/x-tar", "tar"},
		{"425A68", "application/x-bzip2", "bz2"},
	}
)

func IsArchiveFile(reader io.Reader) (string, bool) {
	header := make([]byte, 8)
	rn, _ := io.ReadFull(reader, header)
	if rn == 0 {
		return "", false
	}
	header = header[:rn]
	hexHeader := strings.ToUpper(hex.EncodeToString(header))
	for _, info := range compressionFormats {
		if strings.HasPrefix(hexHeader, info.magic) {
			return info.format, true
		}
	}
	return "", false
}
