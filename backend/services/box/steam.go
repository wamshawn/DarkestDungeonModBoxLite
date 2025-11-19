package box

import (
	"encoding/xml"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"DarkestDungeonModBoxLite/backend/pkg/failure"
	"DarkestDungeonModBoxLite/backend/pkg/files"
	"DarkestDungeonModBoxLite/backend/pkg/images"
)

func (bx *Box) SyncSteamMods() (pid string, err error) {
	// todo get settings > compare mods > start task
	return
}

func (bx *Box) StopSyncSteamMods(pid string) (err error) {

	return
}

type WorkshopModule struct {
	Id     string   `json:"id"`
	Title  string   `json:"title"`
	Icon   string   `json:"icon"`
	Synced bool     `json:"synced"`
	Tags   []string `json:"tags"`
}

func (bx *Box) ListModules() (modules []WorkshopModule, err error) {

	settings, settingsErr := bx.Settings()
	if settingsErr != nil {
		err = failure.Failed("工坊", "扫描本地存储错误").Wrap(settingsErr)
		return
	}
	workshop := settings.Workshop()
	if workshop == "" {
		err = failure.Failed("工坊", "请先设置 steam 位置")
		return
	}
	exist, _ := files.Exist(workshop)
	if !exist {
		return
	}
	dir := os.DirFS(workshop)
	entries, dirErr := fs.ReadDir(dir, ".")
	if dirErr != nil {
		err = failure.Failed("工坊", "扫描本地存储错误").Wrap(dirErr)
		return
	}
	if len(entries) == 0 {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		projectFilename := filepath.Join(workshop, name, "project.xml")
		if exist, _ = files.Exist(projectFilename); !exist {
			continue
		}
		projectBytes, readProjectErr := os.ReadFile(projectFilename)
		if readProjectErr != nil {
			err = failure.Failed("工坊", fmt.Sprintf("读取 %s 错误", name))
			return
		}
		if len(projectBytes) == 0 {
			continue
		}
		project := ModuleProject{}
		projectErr := xml.Unmarshal(projectBytes, &project)
		if projectErr != nil {
			err = failure.Failed("工坊", fmt.Sprintf("读取 %s 错误", name)).Append("解析 project.xml 失败", projectErr.Error())
			return
		}
		module, _ := bx.Module(project.PublishedFileId)
		icon := project.PreviewIconFile
		if icon != "" {
			icon = filepath.Join(workshop, name, icon)
		}
		icon, _ = images.Base64(icon)
		//icon = "/" + icon
		modules = append(modules, WorkshopModule{
			Id:     project.PublishedFileId,
			Title:  project.Title,
			Icon:   icon,
			Synced: module.Id == project.PublishedFileId,
			Tags:   project.ListTags(),
		})
	}
	return
}
