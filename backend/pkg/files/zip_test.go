package files_test

import (
	"errors"
	"io"
	"testing"

	"github.com/yeka/zip"
)

func TestUnzip(t *testing.T) {
	filename := `F:\games\暗黑地牢\test.zip`
	r, rErr := zip.OpenReader(filename)
	if rErr != nil {
		t.Log(rErr)
	}
	defer r.Close()
	for _, f := range r.File {
		if f == nil {
			t.Log("file is nil")
			continue
		}
		t.Log(f.Name, f.FileHeader.IsEncrypted())
		f.SetPassword("111")
		item, itemErr := f.Open()
		if itemErr != nil {
			t.Error(itemErr)
			return
		}
		b := make([]byte, 8)
		_, rErr = item.Read(b)
		item.Close()
		if rErr != nil {
			if errors.Is(rErr, io.EOF) {
				continue
			}
			t.Error(rErr)
			return
		}
	}
}
