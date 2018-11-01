package notes

import (
	"os/user"
	"path/filepath"
	"testing"
)

func TestCanonPathCanonicalize(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	h := u.HomeDir

	p := filepath.Join(h, "foo", "bar")
	p = canonPath(p)

	if p != filepath.FromSlash("~/foo/bar") {
		t.Error("home dir was not canonicalized", p)
	}

	p = canonPath(filepath.Join(h, "foo", "bar") + string(filepath.Separator))
	if p != filepath.FromSlash("~/foo/bar/") {
		t.Error("home dir was not canonicalized", p)
	}
}

func TestCanonPathDoesNotCanonicalize(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	h := u.HomeDir

	p := filepath.Join("foo", "bar")
	c := canonPath(p)
	if p != c {
		t.Error("unexpectedly modified", c)
	}

	p = filepath.Dir(h)
	c = canonPath(p)
	if p != c {
		t.Error("unexpectedly modified", c)
	}
}
