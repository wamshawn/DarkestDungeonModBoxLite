package files

import (
	"context"
	"io"
	"io/fs"
	"os"

	"github.com/mholt/archives"
	"github.com/yeka/zip"
)

type CryptoZip struct {
	archives.Zip
	Password string
}

func (z CryptoZip) Extract(ctx context.Context, archive io.Reader, handleFile archives.FileHandler) (err error) {
	src := archive.(*os.File)
	fi, fiErr := src.Stat()
	if fiErr != nil {
		return fiErr
	}
	r, rErr := zip.NewReader(src, fi.Size())
	if rErr != nil {
		return rErr
	}
	for _, f := range r.File {
		if z.Password != "" {
			f.SetPassword(z.Password)
		}
		info := f.FileInfo()
		file := archives.FileInfo{
			FileInfo:      info,
			Header:        f.FileHeader,
			NameInArchive: f.Name,
			LinkTarget:    "",
			Open: func() (fs.File, error) {
				openedFile, openErr := f.Open()
				if openErr != nil {
					return nil, openErr
				}
				return fileInArchive{openedFile, info}, nil
			},
		}
		if hErr := handleFile(ctx, file); hErr != nil {
			return hErr
		}
	}
	return
}

type fileInArchive struct {
	io.ReadCloser
	info fs.FileInfo
}

func (af fileInArchive) Stat() (fs.FileInfo, error) { return af.info, nil }
