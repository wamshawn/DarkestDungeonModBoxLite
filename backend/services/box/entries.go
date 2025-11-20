package box

import (
	"fmt"
	"time"

	"DarkestDungeonModBoxLite/backend/pkg/databases"

	"github.com/tidwall/buntdb"
)

func indexes() (v []databases.Index) {
	v = append(v,
		// plan
		databases.CreateIndex("plan_index", "plan:*", buntdb.IndexJSON("index")),
		databases.CreateIndex("plan_id", "plan:*", buntdb.IndexJSON("id")),
		// mod
		databases.CreateIndex("mod_index", "mod:*", buntdb.IndexJSON("index")),
		databases.CreateIndex("mod_id", "mod:*", buntdb.IndexJSON("id")),
		databases.CreateIndex("mod_kind", "mod:*", buntdb.IndexJSON("kind")),
	)
	return
}

const (
	ClassMod   = "class"
	SkinMod    = "skin"
	TrinketMod = "trinket"
)

type Module struct {
	Id    string `json:"id"`
	Kind  string `json:"kind"`
	Title string `json:"title"`
	Index uint   `json:"index"`
}

func (mod *Module) Key() string {
	return fmt.Sprintf("mod:%s", mod.Id)
}

type Plan struct {
	Id       string    `json:"id"`
	Name     string    `json:"name"`
	Index    uint      `json:"index"`
	Deployed bool      `json:"deployed"`
	CreateAT time.Time `json:"createAT"`
}

func (plan *Plan) Key() string {
	return fmt.Sprintf("plan:%s", plan.Id)
}

func (plan *Plan) PrefixKey() string {
	return fmt.Sprintf("plan:%s:mod:*", plan.Id)
}

type PlanMod struct {
	PlanId string `json:"planId"`
	ModId  string `json:"modId"`
	Index  uint   `json:"index"`
}

func (pm *PlanMod) Key() string {
	return fmt.Sprintf("plan:%s:mod:%s", pm.PlanId, pm.ModId)
}
