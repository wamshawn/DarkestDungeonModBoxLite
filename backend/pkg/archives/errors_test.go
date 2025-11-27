package archives_test

import (
	"errors"
	"testing"

	"DarkestDungeonModBoxLite/backend/pkg/archives"
)

func TestIsPasswordFailed(t *testing.T) {
	err := errors.Join(
		errors.New("foo"),
		archives.FileError{
			Filename: "foo.zip",
			Err:      archives.ErrPasswordRequired,
		},
		archives.FileError{
			Filename: "bar.zip",
			Err:      archives.ErrPasswordInvalid,
		},
	)

	errs, ok := archives.IsPasswordFailed(err)
	if !ok {
		t.Error("not password failed")
		return
	}
	for _, e := range errs {
		t.Log(e.Filename, e.PasswordRequired, e.PasswordInvalid)
	}

}
