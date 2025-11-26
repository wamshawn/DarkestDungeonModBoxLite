package archives_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"DarkestDungeonModBoxLite/backend/pkg/archives"
)

func TestFile_Info(t *testing.T) {
	//filename := `F:\games\暗黑地牢\test.7z`
	filename := `F:\games\暗黑地牢\test.zip`
	src, openErr := os.Open(filename)
	if openErr != nil {
		t.Error(openErr)
		return
	}
	defer src.Close()

	file, fileErr := archives.New(filename, src)
	if fileErr != nil {
		t.Error(fileErr)
		return
	}
	file.SetPassword("111")
	file.SetEntryPassword(`test.zip/test/foo.7z/foo/ZIMIK Arbalest skin.7z`, "222")

	ctx := context.Background()

	info, infoErr := file.Info(ctx, "*/project.xml")
	if infoErr != nil {
		t.Error(infoErr)
		t.Log(info.String())
		return
	}

	targets := info.Match("*/project.xml")
	for _, target := range targets {
		t.Log("matched", target.Path())
		t.Log(string(target.Preview))
	}

	t.Log(info.String())
}

func TestMatch(t *testing.T) {

	t.Log(filepath.Match("bbb/project.xml", "foo/project.xml"))
	t.Log(filepath.Match("777/*/project.xml", "777/555/project.xml"))

}
