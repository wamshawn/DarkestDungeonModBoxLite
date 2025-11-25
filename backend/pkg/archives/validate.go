package archives

import (
	"context"
	"encoding/hex"
	"errors"
	"io"
	"strings"

	"github.com/mholt/archives"
)

func (file *File) Validate(ctx context.Context) (err error) {
	password := file.option.GetPassword(file.name)

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

func TryValidate(reader io.Reader) (string, bool) {
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
