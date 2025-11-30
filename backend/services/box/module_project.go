package box

import (
	"errors"
	"strconv"
	"strings"
)

type ModuleProjectTags struct {
	Tags []string
}

type ModuleProject struct {
	PreviewIconFile      string
	ItemDescriptionShort string
	ModDataPath          string
	Title                string
	Language             string
	UpdateDetails        string
	Visibility           string
	UploadMode           string
	VersionMajor         string
	VersionMinor         string
	TargetBuild          string
	Tags                 ModuleProjectTags
	ItemDescription      string
	PublishedFileId      string
}

func (project *ModuleProject) ListTags() (tags []string) {
	for _, raw := range project.Tags.Tags {
		tag := strings.TrimSpace(raw)
		tags = append(tags, tag)
	}
	return
}

func (project *ModuleProject) Version() (v Version, err error) {
	var (
		major uint64
		minor uint64
		patch uint64
	)
	if s := strings.TrimSpace(project.VersionMajor); len(s) > 0 {
		if major, err = strconv.ParseUint(s, 10, 64); err != nil {
			err = errors.New("invalid major version")
			return
		}
	}
	if s := strings.TrimSpace(project.VersionMinor); len(s) > 0 {
		if minor, err = strconv.ParseUint(s, 10, 64); err != nil {
			err = errors.New("invalid minor version")
			return
		}
	}
	if s := strings.TrimSpace(project.TargetBuild); len(s) > 0 {
		if patch, err = strconv.ParseUint(s, 10, 64); err != nil {
			err = errors.New("invalid patch version")
			return
		}
	}
	v.Major = uint(major)
	v.Minor = uint(minor)
	v.Patch = uint(patch)
	return
}
