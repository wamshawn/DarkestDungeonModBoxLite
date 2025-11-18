package files

import (
	"os"
	"path/filepath"
)

func Exist(path string) (exist bool, err error) {
	if path == "" {
		return
	}
	if _, err = os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			err = nil
			return
		}
		return
	}
	exist = true
	return
}

func Mkdir(path string) (err error) {
	exist, existErr := Exist(path)
	if existErr != nil {
		err = existErr
		return
	}
	if !exist {
		err = os.MkdirAll(path, 0644)
	}
	return
}

func IsDir(path string) (ok bool, err error) {
	info, statErr := os.Stat(path)
	if statErr != nil {
		err = statErr
		return
	}
	ok = info.IsDir()
	return
}

func IsFile(path string) (ok bool, err error) {
	info, statErr := os.Stat(path)
	if statErr != nil {
		err = statErr
		return
	}
	ok = !info.IsDir()
	return
}

func IsEmpty(path string) (ok bool, err error) {
	info, statErr := os.Stat(path)
	if statErr != nil {
		err = statErr
		return
	}
	ok = info.Size() == 0
	return
}

func InDesktop() (ok bool) {
	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return
	}
	desktop := filepath.Join(home, "Desktop")
	wd, wdErr := os.Getwd()
	if wdErr != nil {
		return
	}
	ok = wd == desktop
	return
}
