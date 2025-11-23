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

func ArchivePassword(password string) *ArchivePasswords {
	return &ArchivePasswords{
		password: password,
	}
}

type ArchivePasswords struct {
	password  string
	passwords map[string]string
}

func (options *ArchivePasswords) AddPassword(filename string, password string) {
	options.passwords[filename] = password
}

func (options *ArchivePasswords) GetPassword(filename string) string {
	if len(options.passwords) == 0 {
		return ""
	}
	password, _ := options.passwords[filename]
	return password
}

func (options *ArchivePasswords) Password() string {
	return options.password
}

func (options *ArchivePasswords) Sub(filename string) *ArchivePasswords {
	if len(options.passwords) == 0 {
		return options
	}
	password := options.GetPassword(filename)
	if password == "" {
		return options
	}
	return &ArchivePasswords{
		password:  password,
		passwords: options.passwords,
	}
}

type ctxArchivePasswordKey struct{}

var (
	_ctxArchivePasswordKey = ctxArchivePasswordKey{}
)

func getArchivePasswords(ctx context.Context) *ArchivePasswords {
	v := ctx.Value(_ctxArchivePasswordKey)
	if v != nil {
		return v.(*ArchivePasswords)
	}
	return nil
}

func withArchivePassword(ctx context.Context, password *ArchivePasswords) context.Context {
	if password != nil {
		return context.WithValue(ctx, _ctxArchivePasswordKey, password)
	}
	return ctx
}

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

func GetArchiveInfo(ctx context.Context, filename string, passwords *ArchivePasswords) (archive *ArchiveInfo, err error) {
	// open
	file, openErr := os.Open(filename)
	if openErr != nil {
		err = errors.Join(errors.New("failed to get archive info"), openErr)
		return
	}
	// validate
	_, isArchived := IsArchiveFile(file)
	_ = file.Close()
	if !isArchived {
		err = errors.Join(errors.New("failed to get archive info"), fmt.Errorf("file %s is not archived", filename))
		return
	}
	// get info
	file, _ = os.Open(filename)
	archive, err = getArchiveInfo(ctx, filename, file, passwords)
	_ = file.Close()
	if err != nil {
		err = errors.Join(errors.New("failed to get archive info"), err)
		return
	}
	return
}

type ReadAtSeeker interface {
	io.ReadSeeker
	io.ReaderAt
}

func getArchiveInfo(ctx context.Context, filename string, file ReadAtSeeker, passwords *ArchivePasswords) (archive *ArchiveInfo, err error) {
	// get extractor
	extractor, mime, ext, extractorErr := getArchiveExtractor(ctx, filename, file, passwords)
	if extractorErr != nil {
		err = errors.Join(errors.New("failed to get archive info"), errors.New("failed to get archive extractor"), extractorErr)
		return
	}

	// ctx
	ctx = withArchivePassword(ctx, passwords)

	// walk
	archive = &ArchiveInfo{
		MediaType: mime,
		Extension: ext,
		Entries:   nil,
	}

	err = extractor.Extract(ctx, file, func(ctx context.Context, info archives.FileInfo) (err error) {
		// check ctx
		if err = ctx.Err(); err != nil {
			return
		}
		// mount dir
		if info.IsDir() {
			archive.Mount(info.NameInArchive, true)
			return
		}
		// file
		item, itemErr := info.Open()
		if itemErr != nil {
			err = itemErr
			return
		}
		defer item.Close()
		// buf
		head := make([]byte, 64)
		headN, headErr := io.ReadFull(item, head)
		if headN == 0 {
			if errors.Is(headErr, io.EOF) {
				// empty file
				return
			}
			err = errors.Join(fmt.Errorf("failed to read %s", info.NameInArchive), headErr)
			return
		}
		head = head[:headN]

		// check archived
		_, itemArchived := IsArchiveFile(bytes.NewReader(head))
		if !itemArchived { // mount file
			archive.Mount(info.NameInArchive, false)
			return
		}
		// handle archived item
		var (
			sub          *ArchiveInfo
			subPasswords *ArchivePasswords
			subErr       error
		)
		if passwords != nil {
			subPasswords = passwords.Sub(info.NameInArchive)
		}
		if info.Size() < 64*1024*1024 { // use memory
			buf := bytes.NewBuffer(head)
			cp, cpErr := io.Copy(buf, item)
			if cp+int64(headN) != info.Size() {
				if errors.Is(cpErr, io.EOF) {
					err = errors.Join(fmt.Errorf("failed to read %s", info.NameInArchive))
				} else {
					err = errors.Join(fmt.Errorf("failed to read %s", info.NameInArchive), cpErr)
				}
				return
			}
			sub, subErr = getArchiveInfo(ctx, info.Name(), bytes.NewReader(buf.Bytes()), subPasswords)
		} else { // use tmp file
			// create tmp dir
			tmpDir, tmpDirErr := CreateTempDir("archives_*")
			if tmpDirErr != nil {
				err = errors.Join(errors.New("failed to create temp dir"), tmpDirErr)
				return
			}
			// copy file
			tmpFilename := info.Name()
			tmpErr := tmpDir.WriteFile(tmpFilename, head)
			if tmpErr != nil {
				_ = tmpDir.Remove()
				err = errors.Join(errors.New("failed to write tmp file"), tmpErr)
				return
			}
			tmpErr = tmpDir.AppendFile(tmpFilename, item)
			if tmpErr != nil {
				_ = tmpDir.Remove()
				err = errors.Join(errors.New("failed to write tmp file"), tmpErr)
				return
			}
			tmpFile, tmpFileErr := tmpDir.OpenFile(tmpFilename)
			if tmpFileErr != nil {
				_ = tmpDir.Remove()
				err = errors.Join(errors.New("failed to open temp file"), tmpFileErr)
				return
			}
			sub, subErr = getArchiveInfo(ctx, tmpFilename, tmpFile, subPasswords)
			_ = tmpFile.Close()
			_ = tmpDir.Remove()
		}
		if subErr != nil {
			err = errors.Join(fmt.Errorf("failed to extract %s", info.NameInArchive), subErr)
			return
		}
		// merge sub
		archive.Merge(info.NameInArchive, sub)
		return
	})

	return
}

type ExtractArchiveHandler func(ctx context.Context, filename string, archived bool) (dst string, extract bool, err error)

func ExtractArchive(ctx context.Context, filename string, passwords *ArchivePasswords, handler ExtractArchiveHandler) (err error) {
	// open
	file, openErr := os.Open(filename)
	if openErr != nil {
		err = openErr
		return
	}
	// validate
	_, isArchived := IsArchiveFile(file)
	_ = file.Close()
	if !isArchived {
		err = errors.Join(errors.New("failed to get archive info"), fmt.Errorf("file %s is not archived", filename))
		return
	}
	file, _ = os.Open(filename)
	err = extractArchive(ctx, "", filename, file, passwords, handler)
	_ = file.Close()
	if err != nil {
		err = errors.Join(fmt.Errorf("failed to extract %s", filename), err)
		return
	}
	return
}

func extractArchive(ctx context.Context, prefix string, filename string, reader io.Reader, passwords *ArchivePasswords, handler ExtractArchiveHandler) (err error) {
	// get extractor
	extractor, _, _, extractorErr := getArchiveExtractor(ctx, filename, reader, passwords)
	if extractorErr != nil {
		err = errors.Join(errors.New("failed to get archive info"), errors.New("failed to get archive extractor"), extractorErr)
		return
	}
	// ctx
	ctx = withArchivePassword(ctx, passwords)
	// extract
	err = extractor.Extract(ctx, reader, func(ctx context.Context, info archives.FileInfo) (err error) {
		// check ctx
		if err = ctx.Err(); err != nil {
			return
		}
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

func getArchiveExtractor(ctx context.Context, filename string, reader io.Reader, passwords *ArchivePasswords) (v archives.Extractor, mime string, ext string, err error) {
	// identify
	format, _, identifyErr := archives.Identify(ctx, filename, reader)
	if identifyErr != nil {
		err = identifyErr
		return
	}
	mime = format.MediaType()
	ext = format.Extension()

	var (
		extractor archives.Extractor = nil
		password  string             = ""
	)
EXT:
	if password == "" {
		ok := false
		extractor, ok = format.(archives.Extractor)
		if !ok {
			err = fmt.Errorf("%s is not supported", format.Extension())
			return
		}
	} else {
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
			err = fmt.Errorf("%s is not supported password", filename)
			return
		}
	}

	err = extractor.Extract(ctx, reader, func(ctx context.Context, info archives.FileInfo) (err error) {
		file, openErr := info.Open()
		if openErr != nil {
			err = openErr
			return
		}
		defer file.Close()
		b := make([]byte, 8)
		_, readErr := file.Read(b)
		if readErr != nil {
			if readErr != io.EOF {
				err = readErr
				return
			}
			return
		}
		return
	})
	if err == nil {
		v = extractor
		return
	}
	if password == "" {
		password = passwords.Password()
		if password == "" {
			return
		}
		goto EXT
	}

	return
}
