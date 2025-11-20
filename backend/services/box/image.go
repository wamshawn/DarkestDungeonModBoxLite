package box

import (
	"fmt"

	"DarkestDungeonModBoxLite/backend/pkg/failure"
)

func (bx *Box) GetImage(filename string) (v string, err error) {
	v, err = bx.db.GetImage(filename)
	if err != nil {
		err = failure.Failed("图片", fmt.Sprintf("无法加载 %s", filename)).Wrap(err)
		return
	}
	return
}
