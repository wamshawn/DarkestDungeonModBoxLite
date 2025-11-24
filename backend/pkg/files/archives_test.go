package files_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"DarkestDungeonModBoxLite/backend/pkg/files"
)

func TestDir(t *testing.T) {
	t.Log(filepath.Join("", "foo.txt"))
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
	t.Log(filepath.Clean(""))
	t.Log(filepath.Split(filepath.Clean("./foo")))
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
