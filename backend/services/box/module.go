package box

import (
	"fmt"
	"strings"

	"DarkestDungeonModBoxLite/backend/pkg/databases"
	"DarkestDungeonModBoxLite/backend/pkg/failure"
)

type ModuleProjectTags struct {
	Tags []string `xml:"Tags"`
}

type ModuleProject struct {
	PreviewIconFile      string `xml:"PreviewIconFile"`
	ItemDescriptionShort string `xml:"ItemDescriptionShort"`
	ModDataPath          string
	Title                string `xml:"Title"`
	Language             string
	UpdateDetails        string
	Visibility           string
	UploadMode           string
	VersionMajor         int
	VersionMinor         int
	TargetBuild          int
	Tags                 ModuleProjectTags `xml:"Tags"`
	ItemDescription      string            `xml:"ItemDescription"`
	PublishedFileId      string
}

func (project *ModuleProject) ListTags() (tags []string) {
	for _, raw := range project.Tags.Tags {
		tag := strings.TrimSpace(raw)
		tags = append(tags, tag)
	}
	return
}

func (bx *Box) GetModule(id string) (module *Module, err error) {
	var (
		has bool
		db  *databases.Database
	)
	if db, err = bx.database(); err != nil {
		return
	}
	module = &Module{Id: id}
	has, err = db.Get(module.Key(), module)
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
