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

func TestOption_Extracted(t *testing.T) {
	opt := archives.Option{}
	opt.SetPassword("test.zip/test/foo.7z/foo/ZIMIK Arbalest skin.7z", "222")
	opt.SetExtracted("test.zip/test/foo.7z/foo/ZIMIK Arbalest skin.7z")
	t.Log(opt.Extracted("test.zip/test/foo.7z"))
	t.Log(opt.Extracted("test.zip/test/foo.7z/foo/ZIMIK Arbalest skin.7z"))
	t.Log(opt.GetPassword("test.zip/test/foo.7z/foo/ZIMIK Arbalest skin.7z"))
}
