package archives_test

import (
	"context"
	"os"
	"testing"

	"DarkestDungeonModBoxLite/backend/pkg/archives"
)

func TestFile_Encrypted(t *testing.T) {
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

	ctx := context.Background()

	t.Log(file.Encrypted(ctx))
	t.Log(file.Filename(), file.Name(), file.Host())
}
