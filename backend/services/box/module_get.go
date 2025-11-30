package box

import (
	"encoding/json"
	"fmt"
	"strings"

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

func (bx *Box) ListModuleByTitle(title string) (modules []*Module, err error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return
	}
	var (
		db *databases.Database
	)
	if db, err = bx.database(); err != nil {
		return
	}

	modules = make([]*Module, 0, 1)
	err = db.Ascend(moduleTitleIndex, &modules, func(_, value string) bool {
		module := Module{}
		decodeErr := json.Unmarshal([]byte(value), &module)
		if decodeErr != nil {
			return false
		}
		if module.Title == title {
			return true
		}
		return false
	})
	if err != nil {
		err = failure.Failed("模组", "获取模组列表失败").Wrap(err)
		return
	}
	return
}
