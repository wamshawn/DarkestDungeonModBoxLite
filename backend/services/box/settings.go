package box

import (
	"path/filepath"

	"DarkestDungeonModBoxLite/backend/pkg/databases"
	"DarkestDungeonModBoxLite/backend/pkg/failure"
	"DarkestDungeonModBoxLite/backend/pkg/files"
	"DarkestDungeonModBoxLite/backend/pkg/programs"
)

type Settings struct {
	Game     string `json:"game"`
	Workshop string `json:"workshop"`
}

func (settings *Settings) GameModDir() string {
	if settings.Game == "" {
		return ""
	}
	return filepath.Join(settings.Game, "mods")
}

const (
	settingsKey = "settings"
)

func (bx *Box) Settings() (v Settings, err error) {
	var (
		db *databases.Database
	)
	if db, err = bx.database(); err != nil {
		return
	}
	if _, err = db.Get(settingsKey, &v); err != nil {
		err = failure.Failed("获取设置失败", err.Error())
		return
	}
	if v.Game == "" && v.Workshop == "" {
		game, steam, ok := getGameFromSteam()
		if ok {
			v.Game = game
			v.Workshop = filepath.Join(steam, "steamapps", "workshop", "content", "262060")
			_ = bx.UpdateSettings(v)
		}
	}
	return
}

func (bx *Box) UpdateSettings(v Settings) (err error) {
	var (
		db *databases.Database
	)
	if db, err = bx.database(); err != nil {
		return
	}
	if err = db.Update(settingsKey, v); err != nil {
		err = failure.Failed("保存设置失败", err.Error())
		return
	}
	return
}

func getGameFromSteam() (game string, steam string, ok bool) {
	steam, _ = programs.FindSteam()
	if steam == "" {
		return
	}
	game = filepath.Join(steam, "steamapps", "common", "DarkestDungeon")
	if gameExist, _ := files.Exist(game); !gameExist {
		return
	}
	ok = true
	return
}
