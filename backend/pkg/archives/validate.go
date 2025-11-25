package archives

import (
	"context"
	"errors"
	"io"
	"path/filepath"

	"github.com/mholt/archives"
)

func (file *File) Validate(ctx context.Context) (err error) {
	password := file.option.GetPassword(filepath.Base(file.name))

	extractor, identifyErr := file.identify(ctx, password)
	if identifyErr != nil {
		err = identifyErr
		return
	}

	err = extractor.Extract(ctx, file.reader, func(ctx context.Context, info archives.FileInfo) (err error) {
		if info.IsDir() {
			return
		}
		item, openErr := info.Open()
		if openErr != nil {
			err = openErr
			return
		}
		b := make([]byte, 8)
		_, readErr := item.Read(b)
		_ = item.Close()
		if readErr != nil {
			if readErr != io.EOF {
				err = readErr
				return
			}
			return
		}
		return
	})

	resetErr := file.reset()

	if err != nil {
		if resetErr != nil {
			err = errors.Join(err, resetErr)
			return
		}
		return
	}
	if resetErr != nil {
		err = resetErr
		return
	}
	return
}
