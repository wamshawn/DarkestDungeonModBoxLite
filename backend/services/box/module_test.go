package box_test

import (
	"context"
	"testing"

	"DarkestDungeonModBoxLite/backend/services/box"
)

func TestMakeModuleImportPlanByArchiveFile(t *testing.T) {
	ctx := context.Background()
	param := box.MakeModuleImportPlanParam{
		Filename: `F:\games\暗黑地牢\test.zip`,
		ArchiveFilePasswords: box.ImportArchiveFilePassword{
			Path:     "",
			Password: "111",
			Invalid:  false,
			Children: []box.ImportArchiveFilePassword{
				{
					Path:     "test.zip/test/foo.7z/foo/ZIMIK Arbalest skin.7z",
					Password: "222",
				},
			},
		},
	}
	plan, makeErr := box.MakeModuleImportPlanByArchiveFile(ctx, param)
	if makeErr != nil {
		t.Error(makeErr.Error())
		return
	}
	if plan.Invalid {
		t.Error("Invalid")
	}
	t.Log(plan.String())
}

func TestMakeModuleImportPlanByDir(t *testing.T) {
	ctx := context.Background()
	param := box.MakeModuleImportPlanParam{
		Filename: `F:\games\暗黑地牢\test_out\Arbalest`,
	}
	plan, makeErr := box.MakeModuleImportPlanByDir(ctx, param)
	if makeErr != nil {
		t.Error(makeErr.Error())
		return
	}
	if plan.Invalid {
		t.Error("Invalid")
	}
	t.Log(plan.String())
}
