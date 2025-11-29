package box

import (
	"fmt"
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

type Module struct {
	Id        string    `json:"id"`
	PublishId string    `json:"publishId"`
	Kind      string    `json:"kind"`
	Title     string    `json:"title"`
	Filename  string    `json:"filename"`
	Remark    string    `json:"remark"`
	ModifyAT  time.Time `json:"modifyAT"`
}
