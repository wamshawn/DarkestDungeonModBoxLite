package archives_test

import (
	"testing"

	"DarkestDungeonModBoxLite/backend/pkg/archives"
)

func TestOption_GetPassword(t *testing.T) {
	opt := archives.Option{}
	opt.SetPassword("foo.zip", "111")
	opt.SetPassword("foo.zip/foo/bar.zip", "222")

	t.Log(opt.GetPassword(""), opt.GetPassword("foo.zip"))
	t.Log(opt.GetPassword("foo.zip/foo/bar.zip"))
	t.Log(opt.GetPassword("foo.zip/foo/bar.zip/bar/1.7z"))
	t.Log(opt.GetPassword("foo.zip/foo/bar.zip/bar/1.7z/1/2.zip"))
	t.Log(opt.GetPassword("foo.zip/foo/baz.zip"))
	t.Log(opt.GetPassword("xxx.zip"), opt.GetPassword("xxx.zip/yyy.zip"))
}

func TestOption_Discarded(t *testing.T) {
	opt := archives.Option{}
	opt.SetDiscard("foo.zip/foo/bar.zip")

	t.Log(opt.Discarded("foo.zip/foo/bar.zip"))
	t.Log(opt.Discarded("foo.zip/foo/bar.zip/bar/baz"))
	t.Log(opt.Discarded("foo.zip/foo/baz.zip"))
	t.Log(opt.Discarded("foo.zip/foo"))

}
