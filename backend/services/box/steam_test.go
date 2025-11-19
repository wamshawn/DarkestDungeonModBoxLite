package box_test

import (
	"encoding/xml"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"DarkestDungeonModBoxLite/backend/pkg/failure"
	"DarkestDungeonModBoxLite/backend/services/box"
)

func TestDir(t *testing.T) {
	workshop := `F:\games\steam\steamapps\workshop\content\262060`
	t.Log(workshop)
	dir := os.DirFS(workshop)

	entries, dirErr := fs.ReadDir(dir, ".")
	if dirErr != nil {
		err := failure.Failed("工坊", "扫描本地存储错误").Wrap(dirErr)
		t.Log(err)
		return
	}
	if len(entries) == 0 {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		t.Log(entry.Name())
		moduleFP := filepath.Join(workshop, entry.Name())
		moduleDir := os.DirFS(moduleFP)
		projectBytes, readProjectErr := fs.ReadFile(moduleDir, "project.xml")
		if readProjectErr != nil {
			err := failure.Failed("工坊", fmt.Sprintf("读取 %s 错误", entry.Name()))
			t.Log(err)
			return
		}
		if len(projectBytes) == 0 {
			continue
		}
		project := box.ModuleProject{}
		projectErr := xml.Unmarshal(projectBytes, &project)
		if projectErr != nil {
			err := failure.Failed("工坊", fmt.Sprintf("读取 %s 错误", entry.Name())).Wrap(projectErr)
			t.Log(err)
			return
		}
		t.Log(project.Title, project.Tags)
	}
}
