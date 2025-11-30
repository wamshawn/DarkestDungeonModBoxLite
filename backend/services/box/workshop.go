package box

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"time"

	"DarkestDungeonModBoxLite/backend/pkg/databases"
	"DarkestDungeonModBoxLite/backend/pkg/failure"
	"DarkestDungeonModBoxLite/backend/pkg/files"
)

func (bx *Box) SyncWorkshopMods() (pid string, err error) {
	// todo get settings > compare mods > start task
	return
}

func (bx *Box) StopSyncWorkshopMods(pid string) (err error) {

	return
}

type WorkshopModule struct {
	Id      string   `json:"id"`
	Title   string   `json:"title"`
	Icon    string   `json:"icon"`
	Synced  bool     `json:"synced"`
	Version Version  `json:"version"`
	Tags    []string `json:"tags"`
}

func (bx *Box) ListWorkshopModules() (v []WorkshopModule, err error) {
	var (
		db *databases.Database
	)
	if db, err = bx.database(); err != nil {
		return
	}
	settings, settingsErr := bx.Settings()
	if settingsErr != nil {
		err = failure.Failed("工坊", "扫描本地存储错误").Wrap(settingsErr)
		return
	}
	dir, dirErr := files.NewDirFS(settings.Workshop)
	if dirErr != nil {
		if os.IsNotExist(dirErr) {
			return
		}
		err = failure.Failed("工坊", "加载工坊失败").Wrap(dirErr)
		return
	}
	entries, entriesErr := dir.ListDir()
	if entriesErr != nil {
		err = failure.Failed("工坊", "加载工坊失败").Wrap(entriesErr)
		return
	}
	if len(entries) == 0 {
		return
	}
	// list mods
	locals, _ := bx.moduleFS.ListDir()
	if len(locals) > 0 {
		sort.Strings(locals)
	}

	for _, entry := range entries {
		sub := dir.Dir(entry)
		projectBytes, readProjectErr := sub.ReadFile("project.xml")
		if readProjectErr != nil {
			if os.IsNotExist(readProjectErr) {
				continue
			}
			err = failure.Failed("工坊", fmt.Sprintf("读取 %s 错误", entry))
			return
		}
		if len(projectBytes) == 0 {
			continue
		}
		project := ModuleProject{}
		projectErr := xml.Unmarshal(projectBytes, &project)
		if projectErr != nil {
			err = failure.Failed("工坊", fmt.Sprintf("读取 %s 错误", entry)).Append("解析 project.xml 失败", projectErr.Error())
			return
		}
		_, found := slices.BinarySearch[[]string](locals, project.PublishedFileId)

		icon := project.PreviewIconFile
		if icon != "" {
			icon = filepath.Join(sub.Path(), icon)
			_, _ = db.GetImage(icon, 15*24*time.Hour)
		}

		version, _ := project.Version()

		v = append(v, WorkshopModule{
			Id:      project.PublishedFileId,
			Title:   project.Title,
			Icon:    icon,
			Synced:  found,
			Version: version,
			Tags:    project.ListTags(),
		})
	}
	return
}
