package archives_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"DarkestDungeonModBoxLite/backend/pkg/archives"
)

func TestFile_Extract(t *testing.T) {
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

	targetsDir := `F:\games\暗黑地牢\test_out`
	targets := map[string]string{
		`test.zip/test/foo.7z/foo/ZIMIK Arbalest skin.7z/ZIMIK Arbalest skin`: `Arbalest`,
	}

	err := file.Extract(ctx, func(ctx context.Context, entry *archives.Entry) (err error) {
		t.Log(entry.Name(), entry.Info().Name(), entry.Info().Size())
		if entry.Info().IsDir() {
			return
		}
		path := filepath.ToSlash(entry.Name())
		t.Log(path)
		for srcPrefix, dstPrefix := range targets {
			if dst, cut := strings.CutPrefix(path, srcPrefix); cut {
				dst = filepath.Join(targetsDir, dstPrefix, dst)
				dst = filepath.ToSlash(filepath.Clean(dst))
				dstDir := filepath.Dir(dst)
				if dstDir != "" {
					_, statErr := os.Stat(dstDir)
					if statErr != nil {
						if os.IsNotExist(statErr) {
							mkErr := os.MkdirAll(dstDir, 0644)
							if mkErr != nil {
								err = mkErr
								return
							}
						}
					}
				}
				df, dfErr := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
				if dfErr != nil {
					err = dfErr
					return
				}
				cp, cpErr := io.Copy(df, entry)
				_ = df.Close()
				if cpErr != nil {
					err = cpErr
					return
				}
				t.Log("extracted:", path, "->", dst, cp)
			}
		}
		return
	})
	if err != nil {
		t.Error(err)
		return
	}
}
