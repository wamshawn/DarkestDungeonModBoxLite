package box

import (
	"fmt"

	"DarkestDungeonModBoxLite/backend/pkg/databases"
	"DarkestDungeonModBoxLite/backend/pkg/failure"
)

func (bx *Box) SaveModule(module *Module) (err error) {
	var (
		db *databases.Database
	)
	if db, err = bx.database(); err != nil {
		return
	}
	err = db.Update(moduleKey(module.Id), module)
	if err != nil {
		err = failure.Failed("模组", fmt.Sprintf("保存 %s 失败", module.Id)).Wrap(err)
		return
	}
	return
}
