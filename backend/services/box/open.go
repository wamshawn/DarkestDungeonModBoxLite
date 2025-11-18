package box

import (
	"sync"

	"DarkestDungeonModBoxLite/backend/pkg/failure"
)

var openOnce sync.Once

func (bx *Box) Open() (settings Settings, err error) {
	// open database
	openOnce.Do(func() {
		dbErr := bx.createDB()
		if dbErr != nil {
			err = failure.Failed("错误", "打开数据库失败").Append("数据库错误", dbErr.Error())
			return
		}
	})
	// settings
	settings, err = bx.Settings()
	return
}
