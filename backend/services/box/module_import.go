package box

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"DarkestDungeonModBoxLite/backend/pkg/failure"
	"DarkestDungeonModBoxLite/backend/pkg/files"
)

func (bx *Box) ImportModules(plan *ImportPlan) (modules []*Module, err error) {
	if plan.Source == "" || plan.Invalid || len(plan.Modules) == 0 {
		err = failure.Failed("导入模组失败", "无效导入计划")
		return
	}
	if plan.IsDir {
		modules, err = ImportModulesByDir(bx.moduleFS, plan)
	} else {
		modules, err = ImportModulesByArchiveFile(bx.moduleFS, plan)
	}
	if err != nil {
		DropImportPlan(bx.moduleFS, plan)
	}
	return
}

func ImportModulesByArchiveFile(root *files.DirFS, plan *ImportPlan) (modules []*Module, err error) {
	// todo use source not module.Filename
	return
}

func ImportModulesByDir(root *files.DirFS, plan *ImportPlan) (modules []*Module, err error) {

	temps := make([]string, 0, 1)
	for _, modulePlan := range plan.Modules {
		// src
		src, srcErr := files.NewDirFS(modulePlan.Filename)
		if srcErr != nil {
			err = failure.Failed("导入文件夹失败", fmt.Sprintf("无法打开 %s", modulePlan.Filename))
			break
		}
		srcFilenames := modulePlan.FlatEntryFilenames()
		if len(srcFilenames) == 0 {
			continue
		}

		// dst
		dstDirPath, tmpDirPath, override := getImportDst(&modulePlan)
		tmpDST := root.Dir(tmpDirPath)
		temps = append(temps, tmpDST.Path())

		project := ModuleProject{}
		wroteFailed := false
		for _, srcFilename := range srcFilenames {
			if srcFilename == "project.xml" {
				projectByte, readProjectErr := src.ReadFile("project.xml")
				if readProjectErr != nil {
					err = failure.Failed("导入文件夹失败", fmt.Sprintf("无法读取 %s", filepath.Join(modulePlan.Filename, srcFilename)))
					wroteFailed = true
					break
				}
				if decodeErr := xml.Unmarshal(projectByte, &project); decodeErr != nil {
					err = failure.Failed("导入文件夹失败", fmt.Sprintf("无法解析 %s", filepath.Join(modulePlan.Filename, srcFilename)))
					wroteFailed = true
					break
				}

			}
			file, fileErr := src.OpenFile(srcFilename)
			if fileErr != nil {
				wroteFailed = true
				err = failure.Failed("导入文件夹失败", fmt.Sprintf("无法打开 %s", filepath.Join(modulePlan.Filename, srcFilename)))
				break
			}
			cpErr := tmpDST.CopyFile(srcFilename, file)
			_ = file.Close()
			if cpErr != nil {
				err = failure.Failed("导入文件夹失败", fmt.Sprintf("无法复制 %s", filepath.Join(modulePlan.Filename, srcFilename)))
				wroteFailed = true
				break
			}
		}
		if wroteFailed {
			break
		}

		// todo not rename and save here,
		// return []tmp{fs, project, module plan}
		// use tmp.Commit to rename files.

		// rename
		if override {
			_ = os.RemoveAll(dstDirPath)
		}
		if rnErr := os.Rename(tmpDST.Path(), dstDirPath); rnErr != nil {
			err = failure.Failed("导入文件夹失败", fmt.Sprintf("无法重命名 %s", tmpDST.Path()))
			break
		}

		// save
		var module *Module
		if modulePlan.Dst != nil {
			module = modulePlan.Dst
			if override {
				if idx, ok := module.ExistVersion(*modulePlan.Override); ok {
					vm := module.Versions[idx]
					vm.PreviewIconFile = project.PreviewIconFile
					vm.ItemDescription = project.ItemDescription
					vm.ItemDescriptionShort = project.ItemDescriptionShort
					vm.UpdateDetails = project.UpdateDetails
					module.Versions[idx] = vm
					if idx+1 == len(module.Versions) {
						module.PreviewIconFile = filepath.ToSlash(filepath.Join(module.Id, module.Version.String(), module.Versions[len(module.Versions)-1].PreviewIconFile))
					}
				}
			} else {
				ver, _ := project.Version()
				module.Add(VersionedModule{
					Version:              ver,
					PreviewIconFile:      project.PreviewIconFile,
					UpdateDetails:        project.UpdateDetails,
					ItemDescriptionShort: project.ItemDescriptionShort,
					ItemDescription:      project.ItemDescription,
				})
			}
			module.ModifyAT = time.Now()
		} else {
			module = &Module{
				Id:              modulePlan.Id,
				PublishId:       modulePlan.PublishFileId,
				Kind:            modulePlan.Kind,
				Title:           modulePlan.Title,
				Remark:          "",
				ModifyAT:        time.Now(),
				PreviewIconFile: "",
				Version:         Version{},
				Versions:        nil,
			}
			ver, _ := project.Version()
			module.Add(VersionedModule{
				Version:              ver,
				PreviewIconFile:      project.PreviewIconFile,
				UpdateDetails:        project.UpdateDetails,
				ItemDescriptionShort: project.ItemDescriptionShort,
				ItemDescription:      project.ItemDescription,
			})
		}
	}
	if err != nil {
		for _, temp := range temps {
			_ = os.RemoveAll(temp)
		}
	}
	return
}

func getImportDst(plan *ModulePlan) (dst string, tmp string, override bool) {
	if !plan.Existed {
		if plan.PublishFileId == "" {
			plan.Id = Id()
		} else {
			plan.Id = plan.PublishFileId
		}
		dst = filepath.Join(plan.Id, plan.Version.String())
		tmp = filepath.Join(plan.Id, plan.Version.String()+"_tmp")
		return
	}
	if plan.Override != nil {
		dst = filepath.Join(plan.Dst.Id, plan.Override.String())
		dst = filepath.Join(plan.Dst.Id, plan.Override.String()+"_tmp")
		override = true
	} else {
		dst = filepath.Join(plan.Dst.Id, plan.Version.String())
		dst = filepath.Join(plan.Dst.Id, plan.Version.String()+"_tmp")
	}
	return
}

func DropImportPlan(root *files.DirFS, plan *ImportPlan) {

	return
}
