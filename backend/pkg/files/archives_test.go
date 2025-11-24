package files_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"DarkestDungeonModBoxLite/backend/pkg/files"
)

func TestDir(t *testing.T) {
	t.Log(filepath.Join("", "foo.txt"))
}

func TestRM(t *testing.T) {
	t.Log(os.RemoveAll(`F:\games\暗黑地牢\test\xxx.txt`))
}

func TestNewCompositeByteReader(t *testing.T) {
	r := files.NewCompositeByteReader([]byte("hello"), bytes.NewReader([]byte(" world")))

	buf := bytes.NewBuffer(nil)
	b := make([]byte, 4)
	for {
		n, err := r.Read(b)
		buf.Write(b[:n])
		t.Log(n, string(b[:n]), err)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Error(err)
			return
		}
	}
	t.Log(buf.String())

	r = files.NewCompositeByteReader([]byte("hello"), bytes.NewReader([]byte(" world")))
	buf.Reset()
	cp, cpErr := io.Copy(buf, r)
	t.Log(cp, buf.String(), cpErr)
	if cpErr != nil {
		t.Error(cpErr)
	}
	return
}

func TestGetArchiveInfo(t *testing.T) {
	ctx := context.Background()
	//filename := `F:\games\暗黑地牢\test.7z`
	filename := `F:\games\暗黑地牢\test.zip`
	//filename := `F:\games\暗黑地牢\test_inner\test_inner.7z`
	//filename := `F:\games\暗黑地牢\test\ZIMIK Arbalest skin.7z`

	options := files.ArchiveFile(filename).
		SetPassword("111").
		SetEntryPassword(`test/ZIMIK Arbalest skin.7z`, "222").
		DiscardEntry(`test\empty`)

	t.Log(options.IsDiscardEntry(`test\empty`))

	info, err := files.GetArchiveInfo(ctx, options)
	if err != nil {
		t.Error(err)
		t.Log(info)
		return
	}
	targets := info.Find("project.xml")
	t.Log(len(targets))
	if len(targets) > 0 {
		for _, target := range targets {
			t.Log(target.Name, target.Path())
		}
	}
	entries := info.ArchiveEntries()
	for _, entry := range entries {
		t.Log(entry.Name, entry.PasswordRequired, entry.Password, entry.Path())
	}
	t.Log(info)
}

type ModuleFS struct {
	root string
}

func TestExtractArchive(t *testing.T) {
	ctx := context.Background()
	dst := `F:\games\暗黑地牢\test_out`
	filename := `F:\games\暗黑地牢\test.zip`
	options := files.ArchiveFile(filename).
		SetPassword("111").
		SetEntryPassword(`test/ZIMIK Arbalest skin.7z`, "222").
		DiscardEntry(`test\empty`)

	err := files.ExtractArchive(ctx, dst, options, func(ctx context.Context, host string, filename string) (dst string, err error) {
		dst = files.CleanArchiveFilename(host, filename)
		return
	})
	if err != nil {
		t.Error(err)
	}
}

func TestCut(t *testing.T) {
	host := ``
	filename := `ZIMIK Arbalest skin.7z/ZIMIK Arbalest skin/foo.txt`
	t.Log(files.CleanArchiveFilename(host, filename))

}

func TestArchiveExtractOptions_IsDiscardEntry(t *testing.T) {

	options := files.ArchiveFile(`F:\games\暗黑地牢\test.zip`)
	options.SetPassword("111")
	options.SetEntryPassword(`foo\bar.zip`, `222`)
	t.Log(
		options.Password(``),
		options.Password(`F:\games\暗黑地牢\test.zip`),
		options.Password(`foo\bar.zip`),
		options.Password(`foo\baz.zip`),
	)

	options.DiscardEntry(`foo\baz.zip`)
	t.Log(
		options.IsDiscardEntry(``),
		options.IsDiscardEntry(`foo\bar.zip`),
		options.IsDiscardEntry(`foo`),
		options.IsDiscardEntry(`foo\baz.zip`),
		options.IsDiscardEntry(`foo\baz.zip\abc`),
	)
	t.Log(options)
}

func TestClean(t *testing.T) {
	t.Log(filepath.Clean("foo/xxx/"))
	t.Log(filepath.Split(filepath.Clean("xxx/foo")))
	t.Log(filepath.Dir(""))
	t.Log(filepath.Join(filepath.Join("foo", "bar"), "bar.txt"))
	t.Log(filepath.Join("", "baz.txt"))
}

func TestIsArchiveFile(t *testing.T) {
	ss := []string{
		`F:\games\暗黑地牢\test.zip`,
		`F:\games\暗黑地牢\test.7z`,
		`F:\games\暗黑地牢\test.rar`,
	}

	for _, s := range ss {
		t.Log(s)
		file, openErr := os.Open(s)
		if openErr != nil {
			t.Error(openErr)
			return
		}
		t.Log(files.IsArchiveFile(file))
		file.Close()
	}

}
