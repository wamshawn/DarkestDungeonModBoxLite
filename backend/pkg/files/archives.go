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
	Name             string             `json:"name"`
	IsDir            bool               `json:"isDir"`
	Archived         bool               `json:"archived"`
	Password         string             `json:"password"`
	PasswordRequired bool               `json:"passwordRequired"`
	Children         []*ArchiveFileInfo `json:"children"`
	Parent           *ArchiveFileInfo   `json:"-"`
}

func (info *ArchiveFileInfo) add(dirs []string, file string) (result *ArchiveFileInfo) {
	if len(dirs) == 0 {
		if file == "" {
			return
		}
		result = &ArchiveFileInfo{
			Name:     file,
			IsDir:    false,
			Children: nil,
			Parent:   info,
		}
		info.Children = append(info.Children, result)
		return
	}
	topDir := dirs[0]
	for _, child := range info.Children {
		if child.IsDir && child.Name == topDir {
			result = child.add(dirs[1:], file)
			return
		}
	}
	child := &ArchiveFileInfo{
		Name:     topDir,
		IsDir:    true,
		Children: nil,
		Parent:   info,
	}
	info.Children = append(info.Children, child)
	result = child.add(dirs[1:], file)
	return
}

func (info *ArchiveFileInfo) mountDir(filename string) (result *ArchiveFileInfo) {
	dirs := splitDirs(filepath.Clean(filename))
	result = info.add(dirs, "")
	return
}

func (info *ArchiveFileInfo) mountFile(filename string) (result *ArchiveFileInfo) {
	dir, file := filepath.Split(filepath.Clean(filename))
	dirs := splitDirs(dir)
	result = info.add(dirs, file)
	return
}

func (info *ArchiveFileInfo) mountArchiveFile(filename string, child *ArchiveFileInfo) (result *ArchiveFileInfo) {
	result = info.mountFile(filename)
	result.Archived = true
	result.Password = child.Password
	result.PasswordRequired = child.PasswordRequired
	result.Children = child.Children
	for _, c := range result.Children {
		c.Parent = result
	}
	return
}

func (info *ArchiveFileInfo) get(filename string) (target *ArchiveFileInfo) {
	dir, file := filepath.Split(filepath.Clean(filename))
	dirs := splitDirs(dir)
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

func (info *ArchiveFileInfo) Root() *ArchiveFileInfo {
	parent := info.Parent
LOOP:
	if parent == nil {
		return info
	}
	parent = parent.Parent
	goto LOOP
}

func (info *ArchiveFileInfo) Path() string {
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

func (info *ArchiveFileInfo) ArchiveEntries() (entries []*ArchiveFileInfo) {
	if info.Archived {
		entries = append(entries, info)
	}
	for _, child := range info.Children {
		entries = append(entries, child.ArchiveEntries()...)
	}
	return
}

func (info *ArchiveFileInfo) String() string {
	b, _ := json.MarshalIndent(info, "", "\t")
	return string(b)
}

func ArchiveFile(filename string) *ArchiveExtractOptions {
	return &ArchiveExtractOptions{
		filename: filepath.Clean(filename),
		password: "",
		discard:  false,
		parent:   nil,
		children: nil,
	}
}

type ArchiveExtractOptions struct {
	filename string
	password string
	discard  bool
	parent   *ArchiveExtractOptions
	children []*ArchiveExtractOptions
}

func (options *ArchiveExtractOptions) SetPassword(password string) *ArchiveExtractOptions {
	options.password = password
	return options
}

func (options *ArchiveExtractOptions) SetEntryPassword(filename string, password string) *ArchiveExtractOptions {
	options.update(filename, password, false)
	return options
}

func (options *ArchiveExtractOptions) DiscardEntry(filename string) *ArchiveExtractOptions {
	options.update(filename, "", true)
	return options
}

func (options *ArchiveExtractOptions) Password(filename string) string {
	if filename == "" {
		return options.password
	}
	target := options.find(filename)
	if target == nil {
		return options.password
	}
	if target.password == "" {
	PARENT:
		if target.parent == nil {
			return ""
		}
		if target.parent.password == "" {
			goto PARENT
		}
		return target.parent.password
	}
	return target.password
}

func (options *ArchiveExtractOptions) IsDiscardEntry(filename string) bool {
	filename = filepath.Clean(filename)
	if filename == "" || filename == "." {
		return false
	}
	target := options.find(filename)
	if target == nil {
		dir := filepath.Dir(filepath.Clean(filename))
		return options.IsDiscardEntry(dir)
	}
	if target.discard {
		return true
	}
	parent := target.parent
LOOP:
	if parent == nil {
		return false
	}
	if parent.discard {
		return true
	}
	parent = parent.parent
	goto LOOP
}

func (options *ArchiveExtractOptions) find(filename string) *ArchiveExtractOptions {
	filename = filepath.Clean(filename)
	if filename == "." {
		return nil
	}
	if filename == options.filename {
		return options
	}
	dir, file := filepath.Split(filename)
	dirs := splitDirs(dir)
	if len(dirs) == 0 {
		for _, child := range options.children {
			if child.filename == file {
				return child
			}
		}
		return nil
	}
	for _, child := range options.children {
		if child.filename == dirs[0] {
			return child.find(filepath.Join(filepath.Join(dirs[1:]...), file))
		}
	}
	return nil
}

func (options *ArchiveExtractOptions) update(filename string, password string, discard bool) (ok bool) {
	filename = filepath.Clean(filename)
	if filename == "." {
		return
	}
	dir, file := filepath.Split(filename)
	dirs := splitDirs(dir)
	if len(dirs) == 0 {
		for _, child := range options.children {
			if child.filename == file {
				child.password = password
				child.discard = discard
				ok = true
				return
			}
		}
		child := &ArchiveExtractOptions{
			filename: file,
			password: password,
			discard:  discard,
			parent:   options,
			children: nil,
		}
		options.children = append(options.children, child)
		ok = true
		return
	}
	if len(options.children) == 0 {
		child := &ArchiveExtractOptions{
			filename: dirs[0],
			password: "",
			discard:  false,
			parent:   options,
			children: nil,
		}
		if ok = child.update(filepath.Join(filepath.Join(dirs[1:]...), file), password, discard); ok {
			options.children = append(options.children, child)
		}
		return
	}
	for _, child := range options.children {
		if child.filename == dirs[0] {
			ok = child.update(filepath.Join(filepath.Join(dirs[1:]...), file), password, discard)
			return
		}
	}
	return
}

var (
	ErrArchiveFileRequirePassword = errors.New("archive file require password")
	ErrArchiveFileInvalidPassword = errors.New("invalid archive file password")
)

func GetArchiveInfo(ctx context.Context, options *ArchiveExtractOptions) (info *ArchiveFileInfo, err error) {
	if options == nil {
		err = errors.Join(errors.New("failed to get archive info"), errors.New("options is nil"))
		return
	}
	filename := filepath.Clean(strings.TrimSpace(options.filename))
	if filename == "" || filename == "." {
		err = errors.Join(errors.New("failed to get archive info"), errors.New("filename is missing"))
		return
	}
	// open
	file, openErr := os.Open(filename)
	if openErr != nil {
		err = errors.Join(errors.New("failed to get archive info"), fmt.Errorf("failed to open %s", filename), openErr)
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
	info, err = getArchiveFileInfo(ctx, "", filename, file, options)
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

func getArchiveFileInfo(ctx context.Context, host string, filename string, file ReadAtSeeker, options *ArchiveExtractOptions) (archiveFileInfo *ArchiveFileInfo, err error) {
	archiveName := filepath.Clean(filename)
	if host != "" {
		archiveName = filepath.Clean(filepath.Join(filepath.Clean(filepath.Dir(host)), filename))
	}
	archiveFileInfo = &ArchiveFileInfo{
		Name:             archiveName,
		IsDir:            false,
		Archived:         true,
		Password:         "",
		PasswordRequired: false,
		Children:         nil,
		Parent:           nil,
	}
	// get extractor
	extractor, password, passwordRequired, extractorErr := getArchiveExtractor(ctx, host, filename, file, options)
	archiveFileInfo.Password = password
	archiveFileInfo.PasswordRequired = passwordRequired
	if extractorErr != nil {
		if passwordRequired {
			if password == "" {
				extractorErr = errors.Join(ErrArchiveFileRequirePassword, extractorErr)
			} else {
				extractorErr = errors.Join(ErrArchiveFileInvalidPassword, extractorErr)
			}
		}
		err = errors.Join(errors.New("failed to get archive info"), errors.New("failed to get archive extractor"), extractorErr)
		return
	}

	// extract
	err = extractor.Extract(ctx, file, func(ctx context.Context, info archives.FileInfo) (err error) {
		// check ctx
		if err = ctx.Err(); err != nil {
			return
		}
		// discard
		fileInfoPath := filepath.Clean(info.NameInArchive)
		if host != "" {
			fileInfoPath = filepath.Clean(filepath.Join(filepath.Clean(host), info.NameInArchive))
		}
		if options.IsDiscardEntry(fileInfoPath) {
			return
		}
		// mount dir
		if info.IsDir() {
			archiveFileInfo.mountDir(info.NameInArchive)
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
			archiveFileInfo.mountFile(info.NameInArchive)
			return
		}
		// handle archived item
		host = filepath.Join(host, info.NameInArchive)
		var (
			sub    *ArchiveFileInfo
			subErr error
		)
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
			sub, subErr = getArchiveFileInfo(ctx, host, info.Name(), bytes.NewReader(buf.Bytes()), options)
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
			sub, subErr = getArchiveFileInfo(ctx, host, tmpFilename, tmpFile, options)
			_ = tmpFile.Close()
			_ = tmpDir.Remove()
		}
		if subErr != nil {
			subItem := archiveFileInfo.mountFile(info.NameInArchive)
			subItem.Archived = true
			subItem.PasswordRequired = true
			err = errors.Join(fmt.Errorf("failed to extract %s", info.NameInArchive), subErr)
			return
		}
		// merge sub
		archiveFileInfo.mountArchiveFile(info.NameInArchive, sub)
		return
	})

	return
}

var (
	ErrExtractArchiveDstInvalid  = errors.New("invalid dst")
	ErrExtractArchiveDstNotEmpty = errors.New("dst not empty")
)

type ExtractArchiveHandler func(ctx context.Context, host string, filename string) (dst string, err error)

func ExtractArchive(ctx context.Context, dst string, options *ArchiveExtractOptions, handler ExtractArchiveHandler) (err error) {
	// dst
	dst = strings.TrimSpace(dst)
	if dst == "" {
		err = errors.Join(errors.New("failed to extract archive file"), errors.New("dst is missing"))
		return
	}
	if !filepath.IsAbs(dst) {
		dst, err = filepath.Abs(dst)
		if err != nil {
			err = errors.Join(errors.New("failed to extract archive file"), errors.New("failed to get abs of dst"), err)
			return
		}
	}
	dstFS, dstErr := NewDirFS(dst)
	if dstErr != nil {
		err = errors.Join(errors.New("failed to extract archive file"), ErrExtractArchiveDstInvalid, dstErr)
		return
	}
	if dstFS.Size() > 0 {
		err = errors.Join(errors.New("failed to extract archive file"), ErrExtractArchiveDstNotEmpty)
		return
	}

	if options == nil {
		err = errors.Join(errors.New("failed to extract archive file"), errors.New("options is nil"))
		return
	}

	filename := filepath.Clean(strings.TrimSpace(options.filename))
	if filename == "" || filename == "." {
		err = errors.Join(errors.New("failed to extract archive file"), errors.New("filename is missing"))
		return
	}
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
		err = errors.Join(errors.New("failed to extract archive file"), fmt.Errorf("file %s is not archived", filename))
		return
	}
	file, _ = os.Open(filename)
	err = extractArchive(ctx, dstFS, "", filename, file, options, handler)
	_ = file.Close()
	if err != nil {
		dstFS.Rollback()
		err = errors.Join(errors.New("failed to extract archive file"), err)
		return
	}
	return
}

func extractArchive(ctx context.Context, dst *DirFS, host string, filename string, file ReadAtSeeker, options *ArchiveExtractOptions, handler ExtractArchiveHandler) (err error) {
	// get extractor
	extractor, password, passwordRequired, extractorErr := getArchiveExtractor(ctx, host, filename, file, options)
	if extractorErr != nil {
		if passwordRequired {
			if password == "" {
				extractorErr = errors.Join(ErrArchiveFileRequirePassword, extractorErr)
			} else {
				extractorErr = errors.Join(ErrArchiveFileInvalidPassword, extractorErr)
			}
		}
		err = errors.Join(errors.New("failed to extract archive file"), errors.New("failed to get archive extractor"), extractorErr)
		return
	}

	// extract
	err = extractor.Extract(ctx, file, func(ctx context.Context, info archives.FileInfo) (err error) {
		// ctx
		if err = ctx.Err(); err != nil {
			return
		}
		// discard
		fileInfoPath := filepath.Clean(info.NameInArchive)
		if host != "" {
			fileInfoPath = filepath.Clean(filepath.Join(filepath.Clean(host), info.NameInArchive))
		}
		if options.IsDiscardEntry(fileInfoPath) {
			return
		}
		// dir
		if info.IsDir() { // discard
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
		if !itemArchived { // write file
			itemPath, hErr := handler(ctx, host, fileInfoPath)
			if hErr != nil {
				err = hErr
				return
			}
			itemPath = strings.TrimSpace(itemPath)
			if itemPath == "" {
				return
			}
			if errors.Is(headErr, io.EOF) { // full read
				err = dst.WriteFile(itemPath, head)
			} else { // build composite reader
				src := NewCompositeByteReader(head, item)
				err = dst.CopyFile(itemPath, src)
			}
			return
		}
		// handle archived item
		host = filepath.Join(host, info.NameInArchive)
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
			err = extractArchive(ctx, dst, host, info.Name(), bytes.NewReader(buf.Bytes()), options, handler)
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
			err = extractArchive(ctx, dst, host, tmpFilename, tmpFile, options, handler)
			_ = tmpFile.Close()
			_ = tmpDir.Remove()
		}
		return
	})
	return
}

func NewCompositeByteReader(b []byte, r io.Reader) io.Reader {
	return &CompositeByteReader{
		n: 0,
		b: b,
		r: r,
	}
}

type CompositeByteReader struct {
	n int
	b []byte
	r io.Reader
	e error
}

func (cbr *CompositeByteReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		err = io.ErrShortBuffer
		return
	}
	if cbr.n < len(cbr.b) {
		cp := copy(p, cbr.b[cbr.n:])
		cbr.n += cp
		n = cp
		if n == len(p) {
			return
		}
	}
	if cbr.e != nil {
		err = cbr.e
		return
	}
	nn, rErr := cbr.r.Read(p[n:])
	n += nn
	if errors.Is(rErr, io.EOF) {
		cbr.e = io.EOF
		if n == 0 {
			err = io.EOF
		}
		return
	}
	err = rErr
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

func getArchiveExtractor(ctx context.Context, host string, filename string, reader io.Reader, options *ArchiveExtractOptions) (v archives.Extractor, password string, passwordRequired bool, err error) {
	// identify
	format, _, identifyErr := archives.Identify(ctx, filename, reader)
	if identifyErr != nil {
		err = identifyErr
		return
	}

	var (
		extractor archives.Extractor = nil
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
		passwordRequired = true
		password = options.Password(filepath.Join(host, filename))
		if password == "" {
			return
		}
		goto EXT
	}

	return
}

func CleanArchiveFilename(host string, filename string) (out string) {
	host = filepath.ToSlash(filepath.Clean(host))
	filename = filepath.ToSlash(filepath.Clean(filename))
	if host != "." {
		dir, file := filepath.Split(host)
		if dir == "" {
			out, _ = strings.CutPrefix(filename, file+"/")
		} else {
			dir = filepath.ToSlash(filepath.Clean(dir))
			left, _ := strings.CutSuffix(host, file)
			right, _ := strings.CutPrefix(filename, host+"/")
			out = filepath.Join(left, right)
		}
	} else {
		out = filename
	}
	return
}
