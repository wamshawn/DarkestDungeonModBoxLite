package box

import "fmt"

type Version struct {
	Major uint `json:"major"`
	Minor uint `json:"minor"`
	Patch uint `json:"patch"`
}

func (v Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v Version) Compare(other Version) int {
	if v.Major > other.Major {
		return 1
	}
	if v.Major == other.Major {
		if v.Minor > other.Minor {
			return 1
		}
		if v.Minor == other.Minor {
			if v.Patch > other.Patch {
				return 1
			}
			if v.Patch == other.Patch {
				return 0
			}
		}
	}
	return -1
}
