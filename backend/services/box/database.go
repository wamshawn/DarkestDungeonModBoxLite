package box

import (
	"DarkestDungeonModBoxLite/backend/pkg/databases"

	"github.com/tidwall/buntdb"
)

const (
	schemaIdIndex    = "schema_id"
	moduleIdIndex    = "module_id"
	moduleKindIndex  = "module_kind"
	moduleTitleIndex = "module_title"
)

func databaseIndexes() (v []databases.Index) {
	v = append(v,
		// schema
		databases.CreateIndex(schemaIdIndex, "schema:*", buntdb.IndexJSON("id")),
		// module
		databases.CreateIndex(moduleIdIndex, "module:*", buntdb.IndexJSON("id")),
		databases.CreateIndex(moduleKindIndex, "module:*", buntdb.IndexJSON("kind"), buntdb.IndexJSON("id")),
		databases.CreateIndex(moduleTitleIndex, "module:*", buntdb.IndexJSON("title"), buntdb.IndexJSON("id")),
	)
	return
}
