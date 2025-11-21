package files_test

import (
	"context"
	"testing"

	"DarkestDungeonModBoxLite/backend/pkg/files"
)

func TestWalkArchiveInfo(t *testing.T) {
	ctx := context.Background()
	filename := `F:\games\暗黑地牢\test.zip`
	info, err := files.WalkArchiveInfo(ctx, filename, "111")
	if err != nil {
		t.Error(err)
		return
	}
	targets := info.Find("project.xml")
	t.Log(len(targets))
	if len(targets) > 0 {
		for _, target := range targets {
			t.Log(target.Name, target.Parent.Name)
		}
	}
}
