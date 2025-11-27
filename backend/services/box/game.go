package box

import (
	"encoding/json"
	"fmt"
	"sync"

	"DarkestDungeonModBoxLite/backend/pkg/files"
	"DarkestDungeonModBoxLite/backend/services/box/internal/resources"
)

var (
	_gameFileStructOnce sync.Once
	_gameFileStruct     *files.Structure
)

func GetModuleFileStruct() *files.Structure {
	_gameFileStructOnce.Do(func() {
		s := &files.Structure{}
		if err := json.Unmarshal(resources.GameResStruct, s); err != nil {
			panic(fmt.Errorf("decode game file struct failed, %v", err))
		}
		s.Children = append(s.Children,
			files.Structure{
				Name:     "project.xml",
				IsDir:    false,
				Children: nil,
			},
			files.Structure{
				Name:     "preview_icon.png",
				IsDir:    false,
				Children: nil,
			},
		)
		_gameFileStruct = s
	})
	return _gameFileStruct
}

// ------------------
