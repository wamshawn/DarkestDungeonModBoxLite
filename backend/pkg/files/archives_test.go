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
	info, err := files.GetArchiveInfo(ctx, filename, files.ArchivePassword("111"))
	if err != nil {
		t.Error(err)
		return
	}
	targets := info.Find("project.xml")
	t.Log(len(targets))
	if len(targets) > 0 {
		for _, target := range targets {
			t.Log(target.Name, target.Path())
		}
	}
	t.Log(info)
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
