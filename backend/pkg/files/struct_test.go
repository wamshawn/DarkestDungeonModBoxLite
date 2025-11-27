package files_test

import (
	"encoding/json"
	"testing"

	"DarkestDungeonModBoxLite/backend/pkg/files"
)

func TestFileStructure(t *testing.T) {
	src := `F:\games\steam\steamapps\common\DarkestDungeon`
	v, err := files.FileStructure(src)
	if err != nil {
		t.Fatal(err)
	}
	for i, child := range v.Children {
		if child.Name == "mods" {
			v.Children = append(v.Children[:i], v.Children[i+1:]...)
			break
		}
	}
	for i, child := range v.Children {
		if child.Name == "_windows" {
			v.Children = append(v.Children[:i], v.Children[i+1:]...)
			break
		}
	}
	for i, child := range v.Children {
		if child.Name == "_windowsnosteam" {
			v.Children = append(v.Children[:i], v.Children[i+1:]...)
			break
		}
	}
	for i, child := range v.Children {
		if child.Name == "user_information" {
			v.Children = append(v.Children[:i], v.Children[i+1:]...)
			break
		}
	}
	for i, child := range v.Children {
		if child.Name == "shaders" {
			v.Children = append(v.Children[:i], v.Children[i+1:]...)
			break
		}
	}
	for i, child := range v.Children {
		if child.Name == "shaders_ps4" {
			v.Children = append(v.Children[:i], v.Children[i+1:]...)
			break
		}
	}
	for i, child := range v.Children {
		if child.Name == "shaders_psv" {
			v.Children = append(v.Children[:i], v.Children[i+1:]...)
			break
		}
	}
	//p, _ := json.MarshalIndent(v, "", "  ")
	p, _ := json.Marshal(v)
	t.Log(string(p))
	//os.WriteFile("./DarkestDungeon.json", p, 0644)
}
