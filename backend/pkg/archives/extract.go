package archives

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"DarkestDungeonModBoxLite/backend/pkg/archives/pkg/ioutil"

	"github.com/mholt/archives"
)

func (file *File) fork(filename string, r Reader) (sub *File) {
	sub = &File{
		name:   filepath.ToSlash(filename),
		option: file.option,
		reader: r,
		host:   file,
	}
	return
}

type Entry struct {
	hosts  []string
	name   string
	info   fs.FileInfo
	header any
	reader io.Reader
}

func (e *Entry) Name() string {
	return e.name
}

func (e *Entry) Hosts() []string {
	return e.hosts
}

func (e *Entry) Path() string {
	return filepath.Join(filepath.Join(e.hosts...), e.name)
}

func (e *Entry) Info() fs.FileInfo {
	return e.info
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
	password := file.option.GetPassword(file.name)

	extractor, identifyErr := file.identify(ctx, password)
	if identifyErr != nil {
		err = identifyErr
		return
	}

	err = extractor.Extract(ctx, file.reader, func(ctx context.Context, info archives.FileInfo) (err error) {
		filename := filepath.ToSlash(filepath.Join(filepath.Join(file.Host()...), info.NameInArchive))
		// discard
		if file.option.Discarded(filename) {
			return
		}
		reader, openErr := info.Open()
		if openErr != nil {
			err = openErr
			return
		}
		defer reader.Close()
		// dir
		if info.IsDir() {
			err = handler(ctx, &Entry{
				hosts:  file.Host(),
				name:   info.NameInArchive,
				info:   info.FileInfo,
				header: info.Header,
				reader: reader,
			})
			return
		}
		// file >>>
		// header
		head := make([]byte, 64)
		headN, headErr := io.ReadFull(reader, head)
		if headN == 0 {
			if errors.Is(headErr, io.EOF) {
				// empty file
				err = handler(ctx, &Entry{
					hosts:  file.Host(),
					name:   info.NameInArchive,
					info:   info.FileInfo,
					header: info.Header,
					reader: bytes.NewReader(nil),
				})
				return
			}
			err = errors.Join(fmt.Errorf("failed to read %s", info.NameInArchive), headErr)
			return
		}
		head = head[:headN]
		// check archived
		_, archived := TryValidate(bytes.NewReader(head))
		if !archived { // not archived
			err = handler(ctx, &Entry{
				hosts:  file.Host(),
				name:   info.NameInArchive,
				info:   info.FileInfo,
				header: info.Header,
				reader: ioutil.NewCompositeByteReader(head, reader),
			})
			return
		}
		// try extract entry
		if !file.option.Extracted(filename) { // not extract
			err = handler(ctx, &Entry{
				hosts:  file.Host(),
				name:   info.NameInArchive,
				info:   info.FileInfo,
				header: info.Header,
				reader: ioutil.NewCompositeByteReader(head, reader),
			})
			return
		}
		// extract
		if info.Size() < 64*1024*1024 { // use memory
			buf := bytes.NewBuffer(head)
			cp, cpErr := io.Copy(buf, reader)
			if cp+int64(headN) != info.Size() {
				if errors.Is(cpErr, io.EOF) {
					err = errors.Join(fmt.Errorf("failed to read %s", info.NameInArchive))
				} else {
					err = errors.Join(fmt.Errorf("failed to read %s", info.NameInArchive), cpErr)
				}
				return
			}
			sub := file.fork(info.NameInArchive, bytes.NewReader(buf.Bytes()))
			if err = sub.Extract(ctx, handler); err != nil {
				return
			}
		} else { // use tmp file
			// tmp dir
			tmpDir, createTmpDirErr := os.MkdirTemp("", "DarkestDungeonModBox_archives_*")
			if createTmpDirErr != nil {
				err = createTmpDirErr
				return
			}
			defer os.RemoveAll(tmpDir)
			// tmp file
			tmpFile, tmpFileErr := os.OpenFile(filepath.Join(tmpDir, info.Name()), os.O_RDONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if tmpFileErr != nil {
				err = errors.Join(fmt.Errorf("failed to open %s", filepath.Join(tmpDir, info.Name())), tmpFileErr)
				return
			}
			defer tmpFile.Close()
			// cp header
			if _, cpErr := io.Copy(tmpFile, bytes.NewBuffer(head)); cpErr != nil {
				err = errors.Join(fmt.Errorf("failed to write %s", filepath.Join(tmpDir, info.Name())), cpErr)
				return
			}
			// cp body
			if _, cpErr := io.Copy(tmpFile, reader); cpErr != nil {
				err = errors.Join(fmt.Errorf("failed to write %s", filepath.Join(tmpDir, info.Name())), cpErr)
				return
			}
			// fork
			sub := file.fork(info.NameInArchive, tmpFile)
			if err = sub.Extract(ctx, handler); err != nil {
				return
			}
		}
		// file <<<
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
