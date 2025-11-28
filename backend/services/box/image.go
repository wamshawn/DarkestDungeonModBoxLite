package box

import (
	"fmt"
	"time"

	"DarkestDungeonModBoxLite/backend/pkg/failure"
)

func (bx *Box) GetImage(filename string) (v string, err error) {
	v, err = bx.db.GetImage(filename, 15*24*time.Hour)
	if err != nil {
		err = failure.Failed("图片", fmt.Sprintf("无法加载 %s", filename)).Wrap(err)
		return
	}
	return
}
