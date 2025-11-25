package ioutil

import (
	"path/filepath"
	"strings"
)

func Split(filename string) (dirs []string, file string) {
	filename = strings.TrimSpace(filename)
	filename = filepath.Clean(filename)
	if filename == "." {
		return
	}
	dir := ""
	dir, file = filepath.Split(filename)
	if dir != "" {
		subDirs, subFile := Split(dir)
		if subFile != "" {
			if len(subDirs) > 0 {
				dirs = append(dirs, subDirs...)
			}
			dirs = append(dirs, subFile)
		}
	}
	return
}
