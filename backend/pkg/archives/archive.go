package archives

import (
	"errors"
	"io"
	"path/filepath"
	"strings"
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

func (file *File) Host() []string {
	if file.host == nil {
		return []string{file.name}
	}
	p := file.host.Host()
	return append(p, file.name)
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
	path = filepath.Join(file.name, path)
	file.option.SetPassword(path, password)
}

func (file *File) DiscardEntry(path string) {
	path = strings.TrimSpace(path)
	path = filepath.Clean(path)
	if path == "" || path == "." {
		return
	}
	path = filepath.Join(file.name, path)
	file.option.SetDiscard(path)
}

func (file *File) reset() (err error) {
	_, err = file.reader.Seek(0, io.SeekStart)
	return
}
