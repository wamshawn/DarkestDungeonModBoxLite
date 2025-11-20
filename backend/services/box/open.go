package box

import (
	"DarkestDungeonModBoxLite/backend/pkg/failure"
)

func (bx *Box) Open() (settings Settings, err error) {
	if bx.err != nil {
		err = bx.err
		return
	}
	// settings
	settings, err = bx.Settings()
	if err != nil {
		err = failure.Failed("错误", "启动失败").Wrap(err)
		return
	}
	return
}
