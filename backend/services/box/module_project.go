package box

import "strings"

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
	VersionMajor         int
	VersionMinor         int
	TargetBuild          int
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
