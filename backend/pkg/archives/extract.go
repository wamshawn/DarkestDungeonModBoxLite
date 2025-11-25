package archives

import (
	"context"
	"io"
	"path/filepath"
)

type Entry struct {
	name   string
	host   string
	header any
	reader io.Reader
}

func (e *Entry) Name() string {
	if e.host == "" {
		return e.name
	}
	return filepath.Join(e.host, e.name)
}

func (e *Entry) Read(p []byte) (n int, err error) {
	n, err = e.reader.Read(p)
	return
}

func (e *Entry) Header() any {
	return e.header
}

type ExtractHandler func(ctx context.Context, entry *Entry) (err error)

func (file *File) Extract(ctx context.Context, handler ExtractHandler) (err error) {

	return
}
