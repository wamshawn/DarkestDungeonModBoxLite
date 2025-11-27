package box

import (
	"fmt"
	"slices"
	"time"
)

func ReverseSortSchemaByCreateAT(schemas []Schema) {
	slices.SortFunc[[]Schema](schemas, func(a, b Schema) int {
		return int(a.CreateAT.UnixMilli() - b.CreateAT.UnixMilli())
	})
	slices.Reverse[[]Schema](schemas)
}

type Schema struct {
	Id       string    `json:"id"`
	Name     string    `json:"name"`
	Deployed bool      `json:"deployed"`
	CreateAT time.Time `json:"createAT"`
}

func (schema *Schema) Key() string {
	return fmt.Sprintf("schema:%s", schema.Id)
}

func (schema *Schema) PrefixKey() string {
	return fmt.Sprintf("schema:%s:mod:*", schema.Id)
}

type SchemaModule struct {
	PlanId string `json:"planId"`
	ModId  string `json:"modId"`
	Index  uint   `json:"index"`
}

func (module *SchemaModule) Key() string {
	return fmt.Sprintf("plan:%s:mod:%s", module.PlanId, module.ModId)
}
