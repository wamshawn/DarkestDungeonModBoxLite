package archives

import (
	"context"
	"fmt"

	"DarkestDungeonModBoxLite/backend/pkg/archives/zip"

	"github.com/mholt/archives"
)

func (file *File) identify(ctx context.Context, password string) (extractor archives.Extractor, err error) {
	format, _, identifyErr := archives.Identify(ctx, file.name, file.reader)
	if identifyErr != nil {
		err = identifyErr
		return
	}

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
			extractor = zip.CryptoZip{
				Zip:      format.(archives.Zip),
				Password: password,
			}
		case ".7z":
			ex := format.(archives.SevenZip)
			ex.Password = password
			extractor = ex
		case ".rar":
			ex := format.(archives.Rar)
			ex.Password = password
			extractor = ex
		default:
			err = fmt.Errorf("%s is not supported password", file.name)
			return
		}
	}
	return
}
