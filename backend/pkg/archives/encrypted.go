package archives

import (
	"context"
	"errors"
	"io"

	"github.com/bodgit/sevenzip"
	czip "github.com/klauspost/compress/zip"
	"github.com/mholt/archives"
	"github.com/nwaples/rardecode/v2"
	"github.com/yeka/zip"
)

var (
	errIsEncrypted = errors.New("file is encrypted")
)

func (file *File) Encrypted(ctx context.Context) (ok bool, err error) {
	extractor, identifyErr := file.identify(ctx, "")
	if identifyErr != nil {
		err = identifyErr
		return
	}

	exErr := extractor.Extract(ctx, file.reader, func(ctx context.Context, info archives.FileInfo) (err error) {
		if info.IsDir() {
			return
		}
		switch header := info.Header.(type) {
		case zip.FileHeader:
			if header.IsEncrypted() {
				err = errIsEncrypted
				return
			}
			break
		case czip.FileHeader:
			if header.Flags&0x1 == 1 {
				err = errIsEncrypted
				return
			}
			break
		case *rardecode.FileHeader:
			if header.HeaderEncrypted {
				err = errIsEncrypted
				return
			}
			break
		case sevenzip.FileHeader:
			break
		default:
			break
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
	ok = exErr != nil

	resetErr := file.reset()
	if resetErr != nil {
		err = resetErr
		return
	}
	return
}
