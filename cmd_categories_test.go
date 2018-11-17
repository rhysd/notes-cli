package notes

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCategoriesCmd(t *testing.T) {
	cwd, err := os.Getwd()
	panicIfErr(err)

	for _, tc := range []struct {
		what   string
		subdir string
		want   string
	}{
		{
			what:   "flat",
			subdir: "normal",
			want:   "a\nb\nc\n",
		},
		{
			what:   "nested",
			subdir: "nested",
			want:   "a\na/d\na/d/e\nb\nb/f\nc\n",
		},
	} {
		t.Run(tc.what, func(t *testing.T) {
			cfg := &Config{
				HomePath: filepath.Join(cwd, "testdata", "list", tc.subdir),
			}

			var buf bytes.Buffer
			cmd := CategoriesCmd{
				Config: cfg,
				Out:    &buf,
			}

			if err := cmd.Do(); err != nil {
				t.Fatal(err)
			}

			have := buf.String()
			want := tc.want
			if have != want {
				t.Fatal("wanted:", want, "but have", have)
			}
		})
	}
}

func TestCategoriesCmdError(t *testing.T) {
	cfg := &Config{
		HomePath: filepath.FromSlash("/path/to/somewhere/unknown/home"),
	}

	cmd := CategoriesCmd{
		Config: cfg,
	}

	err := cmd.Do()
	if err == nil {
		t.Fatal("Error did not occur")
	}
	if !strings.Contains(err.Error(), "Cannot read home") {
		t.Fatal("Unexpected error:", err)
	}
}
