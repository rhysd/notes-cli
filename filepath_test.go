package notes

import (
	"os/user"
	"path/filepath"
	"strings"
	"testing"
)

func TestCanonPathCanonicalize(t *testing.T) {
	u, err := user.Current()
	panicIfErr(err)
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
	panicIfErr(err)
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

func TestValidateDirnameOK(t *testing.T) {
	if err := validateDirname("foo-bar is.ok"); err != nil {
		t.Fatal(err)
	}
}

func TestValidateDirnameInvalid(t *testing.T) {
	for _, tc := range []struct {
		name string
		want string
	}{
		{
			name: ".foo",
			want: "Cannot start from '.'",
		},
		{
			name: "foo*bar",
			want: "Cannot contain",
		},
		{
			name: "",
			want: "Cannot be empty",
		},
	} {
		t.Run(tc.want, func(t *testing.T) {
			err := validateDirname(tc.name)
			if err == nil {
				t.Fatal("Error did not occur")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatal("Unexpected error:", err)
			}
		})
	}
}
