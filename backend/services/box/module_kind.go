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
	// hero

	//

	return
}
