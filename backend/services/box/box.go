package box

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"DarkestDungeonModBoxLite/backend/pkg/databases"
	"DarkestDungeonModBoxLite/backend/pkg/failure"
	"DarkestDungeonModBoxLite/backend/pkg/files"
)

type Process struct {
	Id string
}

type Box struct {
	ctx       context.Context
	cancel    context.CancelFunc
	db        *databases.Database
	moduleFS  *files.DirFS
	processes sync.Map
	err       error
}

func (bx *Box) startup(ctx context.Context) {
	// desktop
	if files.InDesktop() {
		bx.err = failure.Failed("错误", "不能在桌面运行")
		return
	}
	// work
	wd, wdErr := os.Getwd()
	if wdErr != nil {
		bx.err = failure.Failed("错误", "无法获取当前运行位置").Wrap(wdErr)
		return
	}
	// mods
	moduleDirPath := filepath.Join(wd, "mods")
	if exist, _ := files.Exist(moduleDirPath); !exist {
		if err := files.Mkdir(moduleDirPath); err != nil {
			bx.err = failure.Failed("错误", "无法创建目录").Append("位置", moduleDirPath).Wrap(err)
			return
		}
	}
	module, moduleErr := files.NewDirFS(moduleDirPath)
	if moduleErr != nil {
		bx.err = failure.Failed("错误", "无法加载模组目录").Append("位置", moduleDirPath).Wrap(moduleErr)
		return
	}
	bx.moduleFS = module

	// database
	databaseDirPath := filepath.Join(wd, "database")
	if exist, _ := files.Exist(databaseDirPath); !exist {
		if err := files.Mkdir(databaseDirPath); err != nil {
			bx.err = failure.Failed("错误", "无法创建目录").Append("位置", databaseDirPath).Wrap(err)
			return
		}
	}
	db, dbErr := databases.New(filepath.Join(databaseDirPath, "database.db"), databaseIndexes()...)
	if dbErr != nil {
		bx.err = failure.Failed("错误", "无法打开数据库").Wrap(dbErr)
		return
	}
	bx.db = db

	// ctx
	bx.ctx, bx.cancel = context.WithCancel(ctx)
	return
}

func (bx *Box) shutdown(_ context.Context) {
	if bx.err != nil {
		return
	}
	bx.cancel()
	bx.db.Close()
	return
}

func (bx *Box) database() (*databases.Database, error) {
	if bx.db == nil {
		return nil, failure.Failed("错误", "数据库未打开")
	}
	return bx.db, nil
}
