package box_test

import (
	"context"
	"sort"
	"testing"

	"DarkestDungeonModBoxLite/backend/services/box"
)

func TestMakeModuleImportPlanByArchiveFile(t *testing.T) {
	ctx := context.Background()
	param := box.MakeModuleImportPlanParam{
		Filename: `F:\games\暗黑地牢\tests\montsers\3500984490.7z`,
		ArchiveFilePasswords: box.ImportArchiveFilePassword{
			Path:     "",
			Password: "111",
			Invalid:  false,
			//Children: []box.ImportArchiveFilePassword{
			//	{
			//		Path:     "test.zip/test/foo.7z/foo/ZIMIK Arbalest skin.7z",
			//		Password: "222",
			//	},
			//},
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
		Filename: `F:\games\暗黑地牢\mods\heros\antiquarian\skin\001`,
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

func TestModuleId(t *testing.T) {
	ids := make([]string, 0, 8)
	ids = append(ids, "2106165672", "2836305813", "2852433413")
	for i := 0; i < 5; i++ {
		ids = append(ids, box.Id())
	}

	sort.Strings(ids)

	t.Log(ids)

}
