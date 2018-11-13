package notes

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTagsCmd(t *testing.T) {
	cwd, err := os.Getwd()
	panicIfErr(err)

	cfg := &Config{
		HomePath: filepath.Join(cwd, "testdata", "list", "normal"),
	}

	for _, tc := range []struct {
		cat  string
		want string
	}{
		{
			cat:  "",
			want: "a-bit-long\nbar\nfoo\nfuture\n",
		},
		{
			cat:  "a",
			want: "bar\nfoo\n",
		},
	} {
		t.Run(tc.cat, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := TagsCmd{
				Category: tc.cat,
				Config:   cfg,
				Out:      &buf,
			}

			if err := cmd.Do(); err != nil {
				t.Fatal(err)
			}

			have := buf.String()
			if have != tc.want {
				t.Fatal("wanted:", tc.want, "but have", have)
			}
		})
	}
}

func TestTagsInvalidHome(t *testing.T) {
	cfg := &Config{HomePath: "/unknown/path/to/home"}
	cmd := TagsCmd{Config: cfg}
	err := cmd.Do()
	if err == nil {
		t.Fatal("Error did not occur")
	}
	if !strings.Contains(err.Error(), "Cannot read home") {
		t.Fatal("Unexpected error:", err)
	}
}

func TestTagsInvalidCategory(t *testing.T) {
	cwd, err := os.Getwd()
	panicIfErr(err)

	cfg := &Config{
		HomePath: filepath.Join(cwd, "testdata", "list", "normal"),
	}

	cmd := TagsCmd{
		Category: "unknown",
		Config:   cfg,
	}

	err = cmd.Do()
	if err == nil {
		t.Fatal("Error did not occur")
	}
	if !strings.Contains(err.Error(), "Category 'unknown' does not exist") {
		t.Fatal("Unexpected error:", err)
	}
}
