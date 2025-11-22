package files_test

import (
	"testing"

	"DarkestDungeonModBoxLite/backend/pkg/files"
)

func TestTempDir_CreateDir(t *testing.T) {

	tmp, tmpErr := files.CreateTempDir("foo_*")
	if tmpErr != nil {
		t.Errorf("Error creating temp dir: %v", tmpErr)
		return
	}
	defer tmp.Remove()

	t.Log(tmp.IsDir(), tmp.Path())

	sub, subErr := tmp.CreateDir("sub")
	if subErr != nil {
		t.Errorf("Error creating sub dir: %v", subErr)
		return
	}
	wErr := sub.WriteFile("foo.txt", []byte("hello world"))
	if wErr != nil {
		t.Errorf("Error writing file: %v", wErr)
		return
	}
	sub1, sub1Err := tmp.Sub("sub")
	if sub1Err != nil {
		t.Errorf("Error get sub dir: %v", sub1Err)
		return
	}
	b, bErr := sub1.ReadFile("foo.txt")
	if bErr != nil {
		t.Errorf("Error reading file: %v", bErr)
		return
	}
	if string(b) != "hello world" {
		t.Errorf("Error reading file: %v", string(b))
		return
	}
	t.Log(string(b))
}
