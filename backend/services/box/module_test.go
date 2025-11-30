package box_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"testing"
	"time"

	"DarkestDungeonModBoxLite/backend/services/box"

	"github.com/tidwall/buntdb"
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

func TestDBList(t *testing.T) {
	db, dbErr := buntdb.Open(":memory:")
	if dbErr != nil {
		t.Error(dbErr)
		return
	}
	defer db.Close()

	setErr := db.Update(func(tx *buntdb.Tx) error {
		for i := 0; i < 10; i++ {
			module := box.Module{
				Id:              fmt.Sprintf("%d", i),
				PublishId:       fmt.Sprintf("%d", i),
				Kind:            "-",
				Title:           "foo",
				Remark:          "",
				ModifyAT:        time.Now(),
				PreviewIconFile: "",
				Version:         box.Version{},
				Versions:        nil,
			}
			if i%2 == 0 {
				module.Title = "bar"
			}
			p, _ := json.Marshal(module)
			_, _, err := tx.Set(fmt.Sprintf("modules:%s", module.Id), string(p), nil)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if setErr != nil {
		t.Error(setErr)
		return
	}

	tx, txErr := db.Begin(true)
	if txErr != nil {
		t.Error(txErr)
		return
	}
	defer tx.Rollback()
	indexErr := tx.CreateIndex("modules_title", "*", buntdb.IndexJSON("title"), buntdb.IndexJSON("id"))
	if indexErr != nil {
		t.Error(indexErr)
		return
	}
	defer tx.DropIndex("modules_title")

	targets := make([]*box.Module, 0, 1)

	ascErr := tx.Ascend("modules_title", func(key, value string) bool {
		module := box.Module{}
		err := json.Unmarshal([]byte(value), &module)
		if err != nil {
			t.Error(err)
			return false
		}
		if module.Title == "foo" {
			targets = append(targets, &module)
		}
		return true
	})
	if ascErr != nil {
		t.Error(ascErr)
		return
	}

	for _, target := range targets {
		t.Log(target.Title, target.Id)
	}

}
