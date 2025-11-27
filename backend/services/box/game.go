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

func GetGameFileStruct() *files.Structure {
	_gameFileStructOnce.Do(func() {
		s := &files.Structure{}
		if err := json.Unmarshal(resources.GameResStruct, s); err != nil {
			panic(fmt.Errorf("decode game file struct failed, %v", err))
		}
		_gameFileStruct = s
	})
	return _gameFileStruct
}

// ------------------
