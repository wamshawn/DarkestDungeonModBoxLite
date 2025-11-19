package box

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"DarkestDungeonModBoxLite/backend/pkg/databases"
	"DarkestDungeonModBoxLite/backend/pkg/files"

	"github.com/tidwall/buntdb"
)

func indexes() (v []databases.Index) {
	v = append(v,
		// plan
		databases.CreateIndex("plan_index", "plan:*", buntdb.IndexJSON("index")),
		databases.CreateIndex("plan_id", "plan:*", buntdb.IndexJSON("id")),
		// mod
		databases.CreateIndex("mod_index", "mod:*", buntdb.IndexJSON("index")),
		databases.CreateIndex("mod_id", "mod:*", buntdb.IndexJSON("id")),
		databases.CreateIndex("mod_kind", "mod:*", buntdb.IndexJSON("kind")),
	)
	return
}

const (
	userDirName = ".DarkestDungeonModBox"
)

func loadBackupDBFromWD() (backupFilename string, err error) {
	// in desktop
	if files.InDesktop() {
		return
	}
	// work dir
	wd, wdErr := os.Getwd()
	if wdErr != nil {
		err = wdErr
		return
	}
	backupFilename = filepath.Join(wd, "database", "backup.dump")
	exist, _ := files.Exist(backupFilename)
	if !exist {
		if err = files.Mkdir(filepath.Join(wd, "database")); err != nil {
			return
		}
	}
	return
}

func createDB() (db *databases.Database, err error) {
	// backup
	backupFilename, backupFilenameErr := loadBackupDBFromWD()
	if backupFilenameErr != nil {
		err = backupFilenameErr
		return
	}
	backupExist, _ := files.Exist(backupFilename)
	// main
	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		err = homeErr
		return
	}
	userDir := filepath.Join(home, userDirName)
	if exist, _ := files.Exist(userDir); !exist {
		if err = files.Mkdir(userDir); err != nil {
			return
		}
	}

	mainFilename := filepath.Join(userDir, "database.db")
	mainExist, _ := files.Exist(mainFilename)
	main, mainErr := databases.New(mainFilename, indexes()...)
	if mainErr != nil {
		err = mainErr
		return
	}

	if !backupExist {
		db = main
		return
	}

	if !mainExist {
		if err = main.Load(backupFilename); err != nil {
			main.Close()
		}
		return
	}

	// create temp db
	tempDir, tempDirErr := os.MkdirTemp("", "DarkestDungeonModBox_db_*")
	if tempDirErr != nil {
		main.Close()
		err = tempDirErr
		return
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	tempFilename := filepath.Join(tempDir, "database.db")
	temp, tempErr := databases.New(tempFilename)
	if tempErr != nil {
		main.Close()
		err = tempErr
		return
	}
	defer temp.Close()

	if err = temp.Load(backupFilename); err != nil {
		main.Close()
		return
	}
	// compare
	mainVersion := main.Version()
	tempVersion := temp.Version()
	if mainVersion < tempVersion {
		// load backup
		if err = main.Load(backupFilename); err != nil {
			main.Close()
		}
		return
	}
	// done
	db = main
	return
}

func (bx *Box) createDB() (err error) {
	bx.db, err = createDB()
	return
}

func (bx *Box) closeDB() {
	if bx.db != nil {
		// incr version
		bx.db.IncrVersion()
		// backup
		backupFilename, _ := loadBackupDBFromWD()
		if backupFilename != "" {
			_ = bx.db.Save(backupFilename)
		}
		// close
		bx.db.Close()
	}
	return
}

const (
	ClassMod   = "class"
	SkinMod    = "skin"
	TrinketMod = "trinket"
)

type Module struct {
	Id    string `json:"id"`
	Kind  string `json:"kind"`
	Title string `json:"title"`
	Index uint   `json:"index"`
}

func (mod *Module) Key() string {
	return fmt.Sprintf("mod:%s", mod.Id)
}

type Plan struct {
	Id       string    `json:"id"`
	Name     string    `json:"name"`
	Index    uint      `json:"index"`
	Deployed bool      `json:"deployed"`
	CreateAT time.Time `json:"createAT"`
}

func (plan *Plan) Key() string {
	return fmt.Sprintf("plan:%s", plan.Id)
}

func (plan *Plan) PrefixKey() string {
	return fmt.Sprintf("plan:%s:mod:*", plan.Id)
}

type PlanMod struct {
	PlanId string `json:"planId"`
	ModId  string `json:"modId"`
	Index  uint   `json:"index"`
}

func (pm *PlanMod) Key() string {
	return fmt.Sprintf("plan:%s:mod:%s", pm.PlanId, pm.ModId)
}
