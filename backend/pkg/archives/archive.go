package archives

import (
	"errors"
	"io"
	"path/filepath"
	"slices"
	"strings"
)

var (
	ErrPasswordRequired = errors.New("password required")
	ErrPasswordInvalid  = errors.New("password invalid")
)

type Reader interface {
	io.ReadSeeker
	io.ReaderAt
}

func New(filename string, src Reader) (file *File, err error) {
	filename = strings.TrimSpace(filename)
	filename = filepath.Clean(filename)
	if filename == "." {
		err = errors.New("filename is invalid")
		return
	}
	if src == nil {
		err = errors.New("src is missing")
		return
	}
	file = &File{
		name:     filepath.Base(filename),
		filename: filepath.ToSlash(filename),
		option:   &Option{},
		reader:   src,
		host:     nil,
	}
	return
}

type File struct {
	name     string
	filename string
	option   *Option
	reader   Reader
	host     *File
}

func (file *File) Name() string {
	return file.name
}

func (file *File) Filename() string {
	return file.filename
}

func (file *File) Path() string {
	hosts := file.Host()
	if len(hosts) == 0 {
		return file.name
	}
	return filepath.Clean(filepath.Join(filepath.Join(hosts...), file.name))
}

func (file *File) Host() []string {
	hosts := make([]string, 0, 1)
	host := file.host
LOOP:
	if host == nil {
		slices.Reverse(hosts)
		return hosts
	}
	hosts = append(hosts, host.name)
	host = host.host
	goto LOOP
}

func (file *File) SetPassword(password string) {
	file.option.SetPassword(file.name, password)
}

func (file *File) SetEntryPassword(path string, password string) {
	path = strings.TrimSpace(path)
	path = filepath.Clean(path)
	if path == "" || path == "." {
		return
	}
	file.option.SetPassword(path, password)
	file.option.SetExtracted(path)
}

func (file *File) ExtractedEntry(path string) {
	file.option.SetExtracted(path)
}

func (file *File) DiscardEntry(path string) {
	path = strings.TrimSpace(path)
	path = filepath.Clean(path)
	if path == "" || path == "." {
		return
	}
	file.option.SetDiscard(path)
}

func (file *File) reset() (err error) {
	_, err = file.reader.Seek(0, io.SeekStart)
	return
}
