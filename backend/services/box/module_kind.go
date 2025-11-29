package box

import (
	"fmt"
	"strings"

	"DarkestDungeonModBoxLite/backend/pkg/files"
)

func GetKindOfModuleByFileStructure(st files.Structure) (kind string) {
	// heroes
	heroDirNum := 0
	if heroes, hasHeroes := st.Get("heroes"); hasHeroes {
		heroDirNum = len(heroes.Children)
		if heroDirNum == 1 {
			name := heroes.Children[0].Name
			skinOnly := true
			for _, child := range heroes.Children[0].Children {
				if !strings.HasPrefix(child.Name, fmt.Sprintf("%s_", name)) {
					skinOnly = false
					break
				}
			}
			if skinOnly {
				kind = HeroSkinsMod
				return
			}
			if _, builtin := builtinHeroes[name]; builtin {
				kind = HeroTweaksMod
				return
			}
			kind = HeroNewMod
			return
		}
	}

	// trinkets
	if _, hasTrinkets := st.Get("trinkets"); hasTrinkets {
		if _, hasPanels := st.Get("panels/icons_equip/trinket"); hasPanels {
			kind = TrinketsMod
			return
		}
	}
	// others
	tweaks := 0
	ui := 0
	monsters := 0
	localization := 0
	for _, child := range st.Children {
		switch child.Name {
		case "activity_log", "colours", "cursors", "dungeons", "fe_flow", "fonts",
			"fx", "game_over", "loading_screen", "overlays", "panels", "scrolls":
			ui++
			break
		case "curios", "effects", "inventory", "loot", "maps", "modes",
			"props", "scripts", "upgrades":
			tweaks++
			break
		case "campaign":
			for _, entry := range child.Children {
				switch entry.Name {
				case "estate", "heirloom_exchange", "progression", "provision", "quest", "roster", "town_events":
					if len(child.Children) > 0 {
						tweaks++
					}
					break
				case "town":
					ui++
					break
				default:
					break
				}
			}
			break
		case "dlc":
			if child.OnlyFilesByExt(".png") {
				ui++
			} else {
				tweaks++
			}
			break
		case "raid":
			for _, entry := range child.Children {
				switch entry.Name {
				case "ai":
					tweaks++
					break
				case "camping":
					for _, camping := range entry.Children {
						switch camping.Name {
						case "default.camping_skills.json":
							tweaks++
							break
						case "skill_icons":
							if len(camping.Children) > 0 {
								ui++
							}
							break
						}
					}
					break
				case "skill_attributes":
					ui++
					break
				default:
					break
				}
			}
			break
		case "raid_results":
			for _, entry := range child.Children {
				switch entry.Name {
				case "raid_results.layout.darkest":
					tweaks++
					break
				default:
					ui++
					break
				}
			}
			break
		case "monsters":
			monsters++
			break
		case "localization":
			localization++
			break
		default:
			// shared ???
			break
		}
	}
	if tweaks == 0 {
		if monsters > 0 {
			kind = MonstersMod
			return
		}
		if localization > 0 {
			kind = LocalizationMod
			return
		}
		if ui > 0 {
			kind = UIMod
			return
		}
		kind = UnknownMod
		return
	}
	kind = GameplayTweaksMod
	return
}
