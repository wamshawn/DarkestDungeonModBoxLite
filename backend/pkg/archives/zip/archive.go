package zip

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"DarkestDungeonModBoxLite/backend/pkg/archives/pkg/seeker"

	"github.com/mholt/archives"
	"github.com/yeka/zip"
)

type CryptoZip struct {
	archives.Zip
	Password string
}

func (z CryptoZip) Extract(ctx context.Context, archive io.Reader, handleFile archives.FileHandler) (err error) {
	src, srcOk := archive.(io.ReaderAt)
	if !srcOk {
		return errors.New("input archive is not a io.ReaderAt")
	}
	sk, skOk := archive.(io.Seeker)
	if !skOk {
		return errors.New("input archive is not a io.ReadSeeker")
	}

	size, sizeErr := seeker.Size(sk)
	if sizeErr != nil {
		return fmt.Errorf("determining stream size: %w", sizeErr)
	}

	r, rErr := zip.NewReader(src, size)
	if rErr != nil {
		return rErr
	}
	skipDirs := skipList{}

	for i, f := range r.File {
		if err = ctx.Err(); err != nil {
			return err // honor context cancellation
		}

		if fileIsIncluded(skipDirs, f.Name) {
			continue
		}

		if z.Password != "" && f.IsEncrypted() {
			f.SetPassword(z.Password)
		}
		info := f.FileInfo()
		linkTarget, linkTargetErr := z.getLinkTarget(f)
		if linkTargetErr != nil {
			return fmt.Errorf("getting link target for file %d: %s: %w", i, f.Name, linkTargetErr)
		}

		file := archives.FileInfo{
			FileInfo:      info,
			Header:        f.FileHeader,
			NameInArchive: f.Name,
			LinkTarget:    linkTarget,
			Open: func() (fs.File, error) {
				openedFile, openErr := f.Open()
				if openErr != nil {
					return nil, openErr
				}
				return fileInArchive{openedFile, info}, nil
			},
		}
		err = handleFile(ctx, file)
		if errors.Is(err, fs.SkipAll) {
			break
		} else if errors.Is(err, fs.SkipDir) && file.IsDir() {
			skipDirs.add(f.Name)
		} else if err != nil {
			if z.ContinueOnError {
				//log.Printf("[ERROR] %s: %v", f.Name, err)
				continue
			}
			return fmt.Errorf("handling file %d: %s: %w", i, f.Name, err)
		}
	}
	return
}

func (z CryptoZip) getLinkTarget(f *zip.File) (string, error) {
	info := f.FileInfo()
	// Exit early if not a symlink
	if info.Mode()&os.ModeSymlink == 0 {
		return "", nil
	}

	// Open the file and read the link target
	file, err := f.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	const maxLinkTargetSize = 32768
	linkTargetBytes, err := io.ReadAll(io.LimitReader(file, maxLinkTargetSize))
	if err != nil {
		return "", err
	}

	if len(linkTargetBytes) == maxLinkTargetSize {
		return "", fmt.Errorf("link target is too large: %d bytes", len(linkTargetBytes))
	}

	return string(linkTargetBytes), nil
}

type fileInArchive struct {
	io.ReadCloser
	info fs.FileInfo
}

func (af fileInArchive) Stat() (fs.FileInfo, error) { return af.info, nil }

type skipList []string

func (s *skipList) add(dir string) {
	trimmedDir := strings.TrimSuffix(dir, "/")
	var dontAdd bool
	for i := 0; i < len(*s); i++ {
		trimmedElem := strings.TrimSuffix((*s)[i], "/")
		if trimmedDir == trimmedElem {
			return
		}
		// don't add dir if a broader path already exists in the list
		if strings.HasPrefix(trimmedDir, trimmedElem+"/") {
			dontAdd = true
			continue
		}
		// if dir is broader than a path in the list, remove more specific path in list
		if strings.HasPrefix(trimmedElem, trimmedDir+"/") {
			*s = append((*s)[:i], (*s)[i+1:]...)
			i--
		}
	}
	if !dontAdd {
		*s = append(*s, dir)
	}
}

func fileIsIncluded(filenameList []string, filename string) bool {
	// include all files if there is no specific list
	if filenameList == nil {
		return true
	}
	for _, fn := range filenameList {
		// exact matches are of course included
		if filename == fn {
			return true
		}
		// also consider the file included if its parent folder/path is in the list
		if strings.HasPrefix(filename, strings.TrimSuffix(fn, "/")+"/") {
			return true
		}
	}
	return false
}
