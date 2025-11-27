package box

import (
	"fmt"

	"DarkestDungeonModBoxLite/backend/pkg/failure"
)

func (bx *Box) GetImage(filename string) (v string, err error) {
	//takeErr := bx.imagesBucket.Take(bx.ctx, 1)
	//if takeErr != nil {
	//	err = failure.Failed("图片", fmt.Sprintf("无法加载 %s", filename)).Wrap(takeErr)
	//	return
	//}
	v, err = bx.db.GetImage(filename)
	//bx.imagesBucket.Release(1)
	if err != nil {
		err = failure.Failed("图片", fmt.Sprintf("无法加载 %s", filename)).Wrap(err)
		return
	}
	return
}
