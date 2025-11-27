package box

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"DarkestDungeonModBoxLite/backend/pkg/archives"
	"DarkestDungeonModBoxLite/backend/pkg/failure"
	"DarkestDungeonModBoxLite/backend/pkg/files"
	"DarkestDungeonModBoxLite/backend/pkg/images"

	"github.com/cespare/xxhash/v2"
)

type ImportArchiveFilePassword struct {
	Path             string                      `json:"path"`
	Password         string                      `json:"password"`
	PasswordRequired bool                        `json:"passwordRequired"`
	PasswordInvalid  bool                        `json:"passwordInvalid"`
	Children         []ImportArchiveFilePassword `json:"children"`
}

func (p *ImportArchiveFilePassword) String() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf("Password: %s, required: %t, invalid: %t\n", p.Password, p.PasswordRequired, p.PasswordInvalid))
	for _, child := range p.Children {
		buf.WriteString(fmt.Sprintf("Password: %s, %s, required: %t, invalid: %t\n", child.Path, child.Password, child.PasswordRequired, child.PasswordInvalid))
	}
	return buf.String()
}

type ImportArchiveFileStats struct {
	Password ImportArchiveFilePassword `json:"password"`
}

func (stats *ImportArchiveFileStats) String() string {
	return stats.Password.String()
}

type ImportEntry struct {
	Chosen     bool          `json:"chosen"`
	Key        string        `json:"key"`
	Title      string        `json:"title"`
	IconBase64 string        `json:"iconBase64"`
	Filename   string        `json:"filename"`
	Children   []ImportEntry `json:"children"`
}

func (entry *ImportEntry) mountArchiveFileInfo(info *archives.FileInfo, chosen bool) {
	for _, child := range info.Children {
		c := ImportEntry{
			Chosen:     chosen,
			Key:        strconv.FormatUint(xxhash.Sum64String(child.Path()), 16),
			IconBase64: "",
			Filename:   child.Path(),
			Children:   nil,
		}
		if child.IsDir {
			c.mountArchiveFileInfo(child, chosen)
		}
		entry.Children = append(entry.Children, c)
	}
}

func (entry *ImportEntry) String() string {
	buf := bytes.NewBuffer(nil)
	if entry.Chosen {
		buf.WriteString("[x]")
	} else {
		buf.WriteString("[-]")
	}
	buf.WriteString(fmt.Sprintf(" [%s] %s > %s \n", entry.Key, entry.Filename, entry.Title))
	for _, child := range entry.Children {
		buf.WriteString(child.String())
	}
	return buf.String()
}

type ImportPlan struct {
	Source   string                  `json:"source"`
	Archived *ImportArchiveFileStats `json:"archived"`
	Entries  []ImportEntry           `json:"entries"`
}

func (plan *ImportPlan) String() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf("Source: %s\n", plan.Source))
	if plan.Archived != nil {
		buf.WriteString(fmt.Sprintf("%s\n", plan.Archived.String()))
	}
	for _, entry := range plan.Entries {
		buf.WriteString(fmt.Sprintf("%s\n", entry.String()))
	}
	return buf.String()
}

type MakeModuleImportPlanParam struct {
	Filename             string                    `json:"filename"`
	ArchiveFilePasswords ImportArchiveFilePassword `json:"archiveFilePasswords"`
}

func (bx *Box) MakeModuleImportPlan(param MakeModuleImportPlanParam) (plan *ImportPlan, err error) {
	filename := strings.TrimSpace(param.Filename)
	if filename == "" {
		err = failure.Failed("创建模组导入计划失败", "待导入文件不存在")
		return
	}
	if filenameExist, _ := files.Exist(filename); !filenameExist {
		err = failure.Failed("创建模组导入计划失败", "待导入文件不存在")
		return
	}
	isDir, dirErr := files.IsDir(filename)
	if dirErr != nil {
		err = failure.Failed("创建模组导入计划失败", "待导入文件错误")
		return
	}
	if isDir {
		plan, err = MakeModuleImportPlanByDir(bx.ctx, param)
	} else {
		plan, err = MakeModuleImportPlanByArchiveFile(bx.ctx, param)
	}
	return
}

func MakeModuleImportPlanByArchiveFile(ctx context.Context, param MakeModuleImportPlanParam) (plan *ImportPlan, err error) {
	// info
	src, srcErr := os.Open(param.Filename)
	if srcErr != nil {
		err = failure.Failed("导入压缩包失败", "无法打开 "+param.Filename)
		return
	}
	defer src.Close()
	file, fileErr := archives.New(param.Filename, src)
	if fileErr != nil {
		err = failure.Failed("导入压缩包失败", "无法解压 "+param.Filename)
		return
	}
	plan = &ImportPlan{
		Source:   param.Filename,
		Archived: &ImportArchiveFileStats{},
		Entries:  nil,
	}
	if param.ArchiveFilePasswords.Password != "" {
		plan.Archived.Password.Password = param.ArchiveFilePasswords.Password
		file.SetPassword(plan.Archived.Password.Password)
	}
	for _, child := range param.ArchiveFilePasswords.Children {
		if child.Path != "" && child.Password != "" {
			plan.Archived.Password.Children = append(plan.Archived.Password.Children, child)
			file.SetEntryPassword(child.Path, child.Password)
		}
	}
	// validate
	if validateErr := file.Validate(ctx); validateErr != nil {
		if errors.Is(err, archives.ErrPasswordRequired) {
			plan.Archived.Password.PasswordRequired = true
		} else if errors.Is(err, archives.ErrPasswordInvalid) {
			plan.Archived.Password.PasswordInvalid = true
		} else {
			err = failure.Failed("导入压缩包失败", "校验 "+param.Filename+" 失败")
		}
		return
	}
	// info
	info, infoErr := file.Info(ctx, "*project.xml", "*preview_icon.png")
	if infoErr != nil {
		passwordErrs, isPasswordErr := archives.IsPasswordFailed(infoErr)
		if !isPasswordErr {
			err = failure.Failed("导入压缩包失败", "扫描 "+param.Filename+" 失败")
			return
		}
		for _, passwordErr := range passwordErrs {
			matched := false
			for i, child := range plan.Archived.Password.Children {
				if child.Path == passwordErr.Filename {
					if passwordErr.PasswordRequired {
						child.PasswordRequired = true
					}
					if passwordErr.PasswordInvalid {
						child.PasswordInvalid = true
					}
					plan.Archived.Password.Children[i] = child
					matched = true
					break
				}
			}
			if !matched {
				plan.Archived.Password.Children = append(plan.Archived.Password.Children, ImportArchiveFilePassword{
					Path:             passwordErr.Filename,
					Password:         "",
					PasswordRequired: passwordErr.PasswordRequired,
					PasswordInvalid:  passwordErr.PasswordInvalid,
					Children:         nil,
				})
			}
		}
		return
	}
	projectInfos := info.Match("*project.xml")
	if len(projectInfos) == 0 {
		err = failure.Failed("导入压缩包失败", "模组不存在")
		return
	}
	for _, projectInfo := range projectInfos {
		if projectInfo.Parent == nil {
			continue
		}
		preview := projectInfo.Preview
		if len(preview) == 0 {
			continue
		}
		parent := projectInfo.Parent
		entry := ImportEntry{
			Chosen:     false,
			Key:        strconv.FormatUint(xxhash.Sum64String(projectInfo.Path()), 16),
			Title:      "",
			IconBase64: "",
			Filename:   projectInfo.Parent.Path(),
			Children:   nil,
		}
		// project.xml
		project := ModuleProject{}
		projectErr := xml.Unmarshal(preview, &project)
		if projectErr != nil {
			err = failure.Failed("导入压缩包失败", fmt.Sprintf("解析 %s 失败", projectInfo.Path()))
			return
		}
		entry.Title = project.Title
		gameStruct := GetModuleFileStruct()
		for _, child := range parent.Children {
			chosen := false
			childName := filepath.Base(child.Name)
			switch childName {
			case "preview_icon.png":
				if len(child.Preview) > 0 {
					entry.IconBase64, _ = images.EncodeBytes("preview_icon.png", child.Preview)
					chosen = true
				}
				break
			default:
				for _, structure := range gameStruct.Children {
					if structure.Name == childName {
						chosen = true
						break
					}
				}
				break
			}

			c := ImportEntry{
				Chosen:   chosen,
				Key:      strconv.FormatUint(xxhash.Sum64String(child.Path()), 16),
				Filename: child.Path(),
				Children: nil,
			}
			if child.IsDir {
				c.mountArchiveFileInfo(child, chosen)
			}
			entry.Children = append(entry.Children, c)
		}
		plan.Entries = append(plan.Entries, entry)
	}
	if len(plan.Entries) == 0 {
		err = failure.Failed("导入压缩包失败", "有效模组存在")
		return
	}
	return
}

func MakeModuleImportPlanByDir(ctx context.Context, param MakeModuleImportPlanParam) (plan *ImportPlan, err error) {

	return
}
