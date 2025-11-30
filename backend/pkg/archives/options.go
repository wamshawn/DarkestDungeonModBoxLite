package archives

import (
	"path/filepath"
	"strings"

	"DarkestDungeonModBoxLite/backend/pkg/archives/pkg/ioutil"
)

type Option struct {
	filename string
	password string
	discard  bool
	extract  bool
	parent   *Option
	children []*Option
}

func (option *Option) GetPassword(filename string) string {
	filename = strings.TrimSpace(filename)
	filename = filepath.Clean(filename)
	if filename == "" || filename == "." {
		return option.password
	}
	target, _ := option.get(filename)
	if target == nil {
		return ""
	}
	if target.password != "" {
		return target.password
	}
	parent := target.parent
LOOP:
	if parent != nil {
		if parent.password != "" {
			return parent.password
		}
		parent = parent.parent
		goto LOOP
	}
	return ""
}

func (option *Option) Discarded(filename string) bool {
	filename = strings.TrimSpace(filename)
	filename = filepath.Clean(filename)
	if filename == "" || filename == "." {
		return option.discard
	}
	target, _ := option.get(filename)
	if target == nil {
		return false
	}
	if target.discard {
		return true
	}
	parent := target.parent
LOOP:
	if parent != nil {
		if parent.discard {
			return true
		}
		parent = parent.parent
		goto LOOP
	}
	return false
}

func (option *Option) isExtracted() bool {
	if option.extract {
		return true
	}
	for _, child := range option.children {
		return child.isExtracted()
	}
	return false
}

func (option *Option) Extracted(filename string) bool {
	filename = strings.TrimSpace(filename)
	filename = filepath.Clean(filename)
	if filename == "" || filename == "." {
		return false
	}
	target, _ := option.get(filename)
	if target == nil {
		return false
	}
	return target.isExtracted()
}

func (option *Option) SetPassword(filename string, password string) {
	extracted := option.Extracted(filename)
	option.update(filename, password, false, extracted)
}

func (option *Option) SetDiscard(filename string) {
	option.update(filename, "", true, false)
}

func (option *Option) SetExtracted(filename string) {
	password := option.GetPassword(filename)
	option.update(filename, password, false, true)
}

func (option *Option) get(filename string) (target *Option, leaf bool) {
	dirs, file := ioutil.Split(filename)
	if file == "" {
		return
	}
	if len(dirs) == 0 {
		if option.filename == file {
			target = option
			leaf = true
			return
		}
		for _, child := range option.children {
			if child.filename == file {
				target = child
				leaf = true
				return
			}
		}
		return
	}
	if dirs[0] != option.filename {
		return
	}

	if len(dirs) == 1 {
		for _, child := range option.children {
			if child.filename == file {
				target = child
				leaf = true
				return
			}
		}
		target = option
		return
	}

	current := option
	dirs = dirs[1:]
	matched := false
MATCH:
	for _, child := range current.children {
		if child.filename == dirs[0] {
			matched = true
			current = child
			dirs = dirs[1:]
			break
		}
	}
	if matched {
		if len(dirs) > 0 {
			matched = false
			goto MATCH
		}
	}
	if current == option {
		return
	}

	target, leaf = current.get(filepath.Join(filepath.Join(dirs...), file))
	if target == nil {
		target = current
	}
	return
}

func (option *Option) update(target string, password string, discard bool, extracted bool) {
	dirs, file := ioutil.Split(target)
	if file == "" {
		return
	}
	if len(dirs) == 0 {
		if option.filename == "" || option.filename == file {
			option.filename = file
			option.password = password
			option.discard = discard
			option.extract = extracted
			return
		}
		for _, child := range option.children {
			if child.filename == file {
				child.password = password
				child.discard = discard
				child.extract = extracted
				return
			}
		}
		child := &Option{
			filename: file,
			password: password,
			discard:  discard,
			extract:  extracted,
			parent:   option,
			children: nil,
		}
		option.children = append(option.children, child)
		return
	}
	if option.filename != dirs[0] {
		if option.filename == "" {
			option.filename = dirs[0]
		} else {
			//panic(fmt.Errorf("%s is not in tree", target))
			return
		}
	}
	dirs = dirs[1:]
	if len(dirs) == 0 {
		option.update(file, password, discard, extracted)
		return
	}
	for _, child := range option.children {
		if child.filename == dirs[0] {
			child.update(filepath.Join(filepath.Join(dirs...), file), password, discard, extracted)
			return
		}
	}

	child := &Option{
		filename: dirs[0],
		password: "",
		discard:  false,
		parent:   option,
		children: nil,
	}
	option.children = append(option.children, child)
	child.update(filepath.Join(filepath.Join(dirs...), file), password, discard, extracted)
	return
}

func (option *Option) root() *Option {
	if option.parent == nil {
		return option
	}
	return option.parent.root()
}

func (option *Option) path() string {
	if option.filename == "" {
		return ""
	}
	if option.parent == nil {
		return option.filename
	}
	items := []string{option.filename}
	parent := option.parent
LOOP:
	if parent != nil {
		if option.parent != nil {
			items = append(items, parent.filename)
		}
		parent = parent.parent
		goto LOOP
	}
	s := ""
	for i := len(items) - 1; i > -1; i-- {
		s = filepath.Join(s, items[i])
	}
	return s
}
