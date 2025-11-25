package ioutil_test

import (
	"testing"

	"DarkestDungeonModBoxLite/backend/pkg/archives/pkg/ioutil"
)

func TestSplit(t *testing.T) {
	ss := []string{
		``,
		`hello.txt`,
		`./hello.txt`,
		`foo/bar/baz/hello.txt`,
		`./foo/bar/baz/hello.txt`,
	}

	for _, s := range ss {
		t.Log(ioutil.Split(s))
	}

}
