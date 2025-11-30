package box_test

import (
	"testing"

	"DarkestDungeonModBoxLite/backend/services/box"
)

func TestVersion(t *testing.T) {
	a := box.Version{0, 2, 1}
	b := box.Version{0, 1, 1}
	t.Log(a.Compare(b))
}
