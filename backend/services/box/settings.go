package box

import (
	"os"
	"path/filepath"

	"DarkestDungeonModBoxLite/backend/pkg/databases"
	"DarkestDungeonModBoxLite/backend/pkg/files"
	"DarkestDungeonModBoxLite/backend/pkg/programs"
)

type Settings struct {
	Game  string `json:"game"`
	Steam string `json:"steam"`
	Mods  string `json:"mods"`
}

func (settings *Settings) Workshop() string {
	if settings.Steam == "" {
		return ""
	}
	workshop := filepath.Join(settings.Steam, "steamapps", "workshop", "content", "262060")
	return workshop
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
		err = ErrGetSettings
		return
	}
	if v.Game == "" && v.Steam == "" {
		changed := 0
		game, steam, ok := getGameFromSteam()
		if ok {
			v.Game = game
			v.Steam = steam
			changed++
		}
		if v.Mods == "" {
			if !files.InDesktop() {
				wd, _ := os.Getwd()
				if wd != "" {
					v.Mods = filepath.Join(wd, "mods")
					if exist, _ := files.Exist(v.Mods); !exist {
						if err := files.Mkdir(v.Mods); err != nil {
							v.Mods = ""
						}
					}
					changed++
				}
			}
		}
		if changed > 0 {
			_ = bx.SetSettings(v)
		}
	}
	return
}

func (bx *Box) SetSettings(v Settings) (err error) {
	var (
		db *databases.Database
	)
	if db, err = bx.database(); err != nil {
		return
	}
	if err = db.Update(settingsKey, v); err != nil {
		err = ErrSetSettings
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
