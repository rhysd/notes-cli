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

	for _, tc := range []struct {
		what   string
		cat    string
		want   string
		subdir string
	}{
		{
			what:   "flat and all categories",
			cat:    "",
			want:   "a-bit-long\nbar\nfoo\nfuture\n",
			subdir: "normal",
		},
		{
			what:   "flat and specific category",
			cat:    "a",
			want:   "bar\nfoo\n",
			subdir: "normal",
		},
		{
			what:   "nested and all categories",
			cat:    "",
			want:   "a\nb\nbar\nc\nd\nfoo\npiyo\n",
			subdir: "nested",
		},
		{
			what:   "nested and specific category",
			cat:    "b",
			want:   "b\nfoo\n",
			subdir: "nested",
		},
		{
			what:   "nested and specific nested category",
			cat:    "a/d/e",
			want:   "a\npiyo\n",
			subdir: "nested",
		},
	} {
		t.Run(tc.cat, func(t *testing.T) {
			cfg := &Config{
				HomePath: filepath.Join(cwd, "testdata", "list", tc.subdir),
			}

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

func TestTagsLoadNoteError(t *testing.T) {
	cwd, err := os.Getwd()
	panicIfErr(err)

	cfg := &Config{
		HomePath: filepath.Join(cwd, "testdata", "list", "fail"),
	}

	cmd := TagsCmd{Config: cfg}
	err = cmd.Do()
	if err == nil {
		t.Fatal("Error did not occur")
	}
	if !strings.Contains(err.Error(), "Cannot parse created date time as RFC3339") {
		t.Fatal("Unexpected error:", err)
	}
}
