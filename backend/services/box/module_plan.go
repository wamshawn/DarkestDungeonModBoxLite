package box

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io/fs"
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
	Path     string                      `json:"path"`
	Password string                      `json:"password"`
	Invalid  bool                        `json:"invalid"`
	Children []ImportArchiveFilePassword `json:"children"`
}

func (p *ImportArchiveFilePassword) String() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf("password: %s, invalid: %t\n", p.Password, p.Invalid))
	for _, child := range p.Children {
		buf.WriteString(fmt.Sprintf("password: %s, filename: %s, invalid: %t\n", child.Password, child.Path, child.Invalid))
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
	Chosen   bool          `json:"chosen"`
	Key      string        `json:"key"`
	Filename string        `json:"filename"`
	Children []ImportEntry `json:"children"`
}

func (entry *ImportEntry) mountArchiveFileInfo(info *archives.FileInfo, chosen bool) {
	for _, child := range info.Children {
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
}

func (entry *ImportEntry) mountDir(filename string, chosen bool) {
	items, dirErr := fs.ReadDir(os.DirFS(filename), ".")
	if dirErr != nil {
		return
	}
	for _, item := range items {
		itemPath := filepath.ToSlash(filepath.Join(filename, item.Name()))
		c := ImportEntry{
			Chosen:   chosen,
			Key:      strconv.FormatUint(xxhash.Sum64String(itemPath), 16),
			Filename: itemPath,
			Children: nil,
		}
		if item.IsDir() {
			c.mountDir(itemPath, chosen)
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
	buf.WriteString(fmt.Sprintf(" [%s] %s \n", entry.Key, entry.Filename))
	for _, child := range entry.Children {
		buf.WriteString(child.String())
	}
	return buf.String()
}

type ModulePlan struct {
	Title      string        `json:"title"`
	IconBase64 string        `json:"iconBase64"`
	Filename   string        `json:"filename"`
	IsDir      bool          `json:"isDir"`
	Entries    []ImportEntry `json:"entries"`
}

func (module *ModulePlan) String() string {
	buf := bytes.NewBuffer(nil)
	if module.IsDir {
		buf.WriteString(fmt.Sprintf("[DIRECTORY] %s > %s \n", module.Title, module.Filename))
	} else {
		buf.WriteString(fmt.Sprintf("[ ARCHIVED] %s > %s \n", module.Title, module.Filename))
	}
	for _, entry := range module.Entries {
		buf.WriteString(entry.String())
	}
	return buf.String()
}

type ImportPlan struct {
	Source   string                  `json:"source"`
	Archived *ImportArchiveFileStats `json:"archived"`
	Invalid  bool                    `json:"invalid"`
	Modules  []ModulePlan            `json:"modules"`
}

func (plan *ImportPlan) String() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf("Source: %s\n", plan.Source))
	if plan.Archived != nil {
		buf.WriteString(fmt.Sprintf("%s\n", plan.Archived.String()))
	}
	for _, entry := range plan.Modules {
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
		Source:   filepath.ToSlash(param.Filename),
		Archived: &ImportArchiveFileStats{},
		Modules:  nil,
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
			plan.Archived.Password.Invalid = true
		} else if errors.Is(err, archives.ErrPasswordInvalid) {
			plan.Archived.Password.Invalid = true
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
						child.Invalid = true
					}
					if passwordErr.PasswordInvalid {
						child.Invalid = true
					}
					plan.Archived.Password.Children[i] = child
					matched = true
					break
				}
			}
			if !matched {
				plan.Archived.Password.Children = append(plan.Archived.Password.Children, ImportArchiveFilePassword{
					Path:     passwordErr.Filename,
					Password: "",
					Invalid:  passwordErr.PasswordInvalid || passwordErr.PasswordRequired,
					Children: nil,
				})
			}
		}
		plan.Invalid = true
		return
	}

	projectInfos := info.Match("*project.xml")
	if len(projectInfos) == 0 {
		for _, invalid := range info.InvalidArchivedEntries() {
			if invalid.Encrypted {
				if invalid.Password == "" {
					err = failure.Failed("导入压缩包失败", "内涵有密码的压缩包").Append("需要密码", invalid.Path())
				} else if invalid.PasswordInvalid {
					err = failure.Failed("导入压缩包失败", "内涵有密码的压缩包").Append("密码错误", invalid.Path())
				}
			}
			return
		}
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
		module := ModulePlan{
			Title:      "",
			IconBase64: "",
			Filename:   filepath.ToSlash(projectInfo.Parent.Path()),
			IsDir:      false,
			Entries:    nil,
		}
		// project.xml
		project := ModuleProject{}
		projectErr := xml.Unmarshal(preview, &project)
		if projectErr != nil {
			err = failure.Failed("导入压缩包失败", fmt.Sprintf("解析 %s 失败", projectInfo.Path()))
			return
		}
		module.Title = project.Title
		gameStruct := GetModuleFileStruct()
		for _, child := range parent.Children {
			chosen := false
			childName := filepath.Base(child.Name)
			switch childName {
			case "preview_icon.png":
				if len(child.Preview) > 0 {
					module.IconBase64, _ = images.EncodeBytes("preview_icon.png", child.Preview)
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
			entry := ImportEntry{
				Chosen:   chosen,
				Key:      strconv.FormatUint(xxhash.Sum64String(child.Path()), 16),
				Filename: child.Path(),
				Children: nil,
			}
			if child.IsDir {
				entry.mountArchiveFileInfo(child, chosen)
			}
			module.Entries = append(module.Entries, entry)
		}
		plan.Modules = append(plan.Modules, module)
	}
	if len(plan.Modules) == 0 {
		err = failure.Failed("导入压缩包失败", "有效模组存在")
		return
	}
	return
}

func MakeModuleImportPlanByDir(_ context.Context, param MakeModuleImportPlanParam) (plan *ImportPlan, err error) {
	dir := os.DirFS(param.Filename)
	// project.xml
	projectBytes, readProjectErr := fs.ReadFile(dir, "project.xml")
	if readProjectErr != nil {
		if os.IsNotExist(readProjectErr) {
			err = failure.Failed("导入压缩包失败", fmt.Sprintf("%s 内缺失 project.xml", param.Filename))
			return
		}
		err = failure.Failed("导入压缩包失败", fmt.Sprintf("读取 %s 中 project.xml 失败", param.Filename))
		return
	}
	project := ModuleProject{}
	projectErr := xml.Unmarshal(projectBytes, &project)
	if projectErr != nil {
		err = failure.Failed("导入压缩包失败", fmt.Sprintf("解析 %s 失败", filepath.Join(param.Filename, "project.xml")))
		return
	}
	// icon
	iconBytes, readIconErr := fs.ReadFile(dir, "preview_icon.png")
	if readIconErr != nil {
		if os.IsNotExist(readIconErr) {
			err = failure.Failed("导入压缩包失败", fmt.Sprintf("%s 内缺失 preview_icon.png", param.Filename))
			return
		}
		err = failure.Failed("导入压缩包失败", fmt.Sprintf("读取 %s 中 preview_icon.png 失败", param.Filename))
		return
	}
	iconBase64, iconBase64Err := images.EncodeBytes("preview_icon.png", iconBytes)
	if iconBase64Err != nil {
		err = failure.Failed("导入压缩包失败", fmt.Sprintf("解析 %s 失败", filepath.Join(param.Filename, "preview_icon.png")))
		return
	}

	entries, dirErr := fs.ReadDir(dir, ".")
	if dirErr != nil {
		err = failure.Failed("导入压缩包失败", fmt.Sprintf("读取 %s 失败", param.Filename))
		return
	}

	module := ModulePlan{
		Title:      project.Title,
		IconBase64: iconBase64,
		Filename:   filepath.ToSlash(param.Filename),
		IsDir:      true,
		Entries:    nil,
	}

	gameStruct := GetModuleFileStruct()
	for _, item := range entries {
		chosen := false
		for _, structure := range gameStruct.Children {
			if structure.Name == item.Name() {
				chosen = true
				break
			}
		}
		itemPath := filepath.ToSlash(filepath.Join(param.Filename, item.Name()))
		entry := ImportEntry{
			Chosen:   chosen,
			Key:      strconv.FormatUint(xxhash.Sum64String(itemPath), 16),
			Filename: itemPath,
			Children: nil,
		}
		if item.IsDir() {
			entry.mountDir(itemPath, chosen)
		}
		module.Entries = append(module.Entries, entry)
	}
	plan = &ImportPlan{
		Source:   filepath.ToSlash(param.Filename),
		Archived: &ImportArchiveFileStats{},
		Modules:  []ModulePlan{module},
	}
	return
}
