package box

import (
	"DarkestDungeonModBoxLite/backend/pkg/databases"

	"github.com/tidwall/buntdb"
)

func databaseIndexes() (v []databases.Index) {
	v = append(v,
		// plan
		databases.CreateIndex("schema_id", "schema:*", buntdb.IndexJSON("id")),
		// mod
		databases.CreateIndex("mod_id", "mod:*", buntdb.IndexJSON("id")),
		databases.CreateIndex("mod_kind", "mod:*", buntdb.IndexJSON("kind")),
	)
	return
}
