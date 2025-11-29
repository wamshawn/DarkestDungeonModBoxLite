package box

import (
	"fmt"

	"DarkestDungeonModBoxLite/backend/pkg/databases"
	"DarkestDungeonModBoxLite/backend/pkg/failure"
)

func (bx *Box) GetModule(id string) (module *Module, err error) {
	var (
		has bool
		db  *databases.Database
	)
	if db, err = bx.database(); err != nil {
		return
	}
	module = &Module{}
	has, err = db.Get(moduleKey(id), module)
	if err != nil {
		module = nil
		err = failure.Failed("模组", fmt.Sprintf("获取 %s 失败", id)).Wrap(err)
		return
	}
	if !has {
		module = nil
		err = failure.Failed("模组", fmt.Sprintf("%s 不存在", id))
		return
	}
	return
}

func (bx *Box) ExistsModule(id string) (exists bool, err error) {
	var (
		db *databases.Database
	)
	if db, err = bx.database(); err != nil {
		return
	}
	module := &Module{}
	exists, err = db.Get(moduleKey(id), module)
	if err != nil {
		err = failure.Failed("模组", fmt.Sprintf("获取 %s 失败", id)).Wrap(err)
		return
	}
	return
}
