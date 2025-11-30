package box

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/rs/xid"
)

const (
	HeroNewMod        = "hero_new"
	HeroTweaksMod     = "hero_tweaks"
	HeroSkinsMod      = "hero_skins"
	OverhaulsMod      = "overhauls" // 大修
	TrinketsMod       = "trinkets"
	MonstersMod       = "monsters"
	LocalizationMod   = "localization"
	UIMod             = "ui"
	GameplayTweaksMod = "gameplay_tweaks"
	UnknownMod        = "unknown"
)

func moduleKey(id string) string {
	return fmt.Sprintf("mod:%s", id)
}

func Id() string {
	id := xid.New()
	h := xxhash.Sum64(id.Bytes())
	s := fmt.Sprintf("0x%s", strings.ToUpper(strconv.FormatUint(h, 16)))
	return s
}

type VersionedModule struct {
	Version              Version `json:"version"`
	PreviewIconFile      string  `json:"previewIconFile"`
	UpdateDetails        string  `json:"updateDetails"`
	ItemDescriptionShort string  `json:"itemDescriptionShort"`
	ItemDescription      string  `json:"itemDescription"`
}

// Module
/* fs
mods/
 {id}/
  {version}/
   {data_path}/
	preview_icon.png
*/
type Module struct {
	Id              string            `json:"id"`
	PublishId       string            `json:"publishId"`
	Kind            string            `json:"kind"`
	Title           string            `json:"title"`
	Remark          string            `json:"remark"`
	ModifyAT        time.Time         `json:"modifyAT"`
	PreviewIconFile string            `json:"previewIconFile"`
	Version         Version           `json:"version"`
	Versions        []VersionedModule `json:"versions"`
}

func (module *Module) Add(vm VersionedModule) {
	if idx, existed := module.ExistVersion(vm.Version); existed {
		module.Versions[idx] = vm
		return
	}
	module.Versions = append(module.Versions, vm)
	slices.SortFunc[[]VersionedModule](module.Versions, func(a, b VersionedModule) int {
		return a.Version.Compare(b.Version)
	})
	module.Version = module.Versions[len(module.Versions)-1].Version
	module.PreviewIconFile = module.Versions[len(module.Versions)-1].PreviewIconFile
}

func (module *Module) ExistVersion(target Version) (idx int, ok bool) {
	for i, v := range module.Versions {
		if v.Version.Compare(target) == 0 {
			idx = i
			ok = true
			return
		}
	}
	idx = -1
	return
}
