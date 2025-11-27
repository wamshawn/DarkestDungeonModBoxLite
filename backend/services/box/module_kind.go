package box

import (
	"context"
	"encoding/xml"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"DarkestDungeonModBoxLite/backend/pkg/failure"
)

func KindOfModule(ctx context.Context, filename string) (kind string, err error) {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		err = failure.Failed("解析模组类型失败", "模组位置缺失")
		return
	}
	stat, statErr := os.Stat(filename)
	if statErr != nil {
		err = failure.Failed("解析模组类型失败", "模组文件错误")
		return
	}
	if !stat.IsDir() {
		err = failure.Failed("解析模组类型失败", "模组文件错误")
		return
	}

	filename = filepath.ToSlash(filepath.Clean(filename))
	projectBytes, projectReadErr := fs.ReadFile(os.DirFS(filename), "project.xml")
	if projectReadErr != nil {
		err = failure.Failed("解析模组类型失败", "模组文件错误")
		return
	}
	project := ModuleProject{}
	if decodeErr := xml.Unmarshal(projectBytes, &project); decodeErr != nil {
		err = failure.Failed("解析模组类型失败", "模组文件错误")
		return
	}
	// hero >>>
	// new 			新英雄 不在本体里的
	// hero_skins	纯皮肤
	// hero_tweaks  重置			存在 *.art.darkest | 存在 anim | 存在
	/*
		anim：拥有自己全新的骨骼动画。
		effects：拥有全新的技能特效。
		icons：拥有全新的技能图标、状态图标。
		shared：拥有全新的英雄肖像、地图行走图等。
		sounds：拥有全新的语音和音效。
		fx：是特效，无影响
	*/

	// hero <<<

	// trinkets >>>
	/*
		heroes 里只有 fx，没有其它改动
		trinkets 存在
	*/
	// trinkets <<<

	return
}
