package box

import (
	"fmt"
)

const (
	HeroMod         = "hero"
	HeroTweaksMod   = "hero_tweaks"
	HeroSkinsMod    = "hero_skins"
	TweaksMod       = "tweaks"
	OverhaulsMod    = "overhauls"
	TrinketsMod     = "trinkets"
	MonstersMod     = "monsters"
	LocalizationMod = "localization"
	UIMod           = "ui"
)

type Module struct {
	Id        string `json:"id"`
	PublishId string `json:"publishId"`
	Kind      string `json:"kind"`
	Title     string `json:"title"`
}

func (mod *Module) Key() string {
	return fmt.Sprintf("mod:%s", mod.Id)
}
