package notes

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/rhysd/go-fakeio"
)

func testNewConfigForNewCmd(subdir string) *Config {
	cwd, err := os.Getwd()
	panicIfErr(err)
	return &Config{
		GitPath:  "git",
		HomePath: filepath.Join(cwd, "testdata", "new", subdir),
	}
}

func TestNewCmdNewNoteInlineFallbackInput(t *testing.T) {
	cfg := testNewConfigForNewCmd("empty")
	fake := fakeio.Stdout().Stdin("this\nis\ntest").CloseStdin()
	defer fake.Restore()

	cmd := &NewCmd{
		Config:   cfg,
		Category: "cat",
		Filename: "test",
		Tags:     "foo, bar",
	}

	if err := cmd.Do(); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(filepath.Join(cfg.HomePath, "cat"))

	p := filepath.Join(cfg.HomePath, "cat", "test.md")
	n, err := LoadNote(p, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if n.Category != "cat" {
		t.Error(n.Category)
	}

	if !reflect.DeepEqual(n.Tags, []string{"foo", "bar"}) {
		t.Error("Tags are not correct", n.Tags)
	}

	if n.Title != "test" {
		t.Error(n.Title)
	}

	if n.Created.After(time.Now()) {
		t.Error("Created date invalid", n.Created.Format(time.RFC3339))
	}

	f, err := os.Open(p)
	if err != nil {
		t.Fatal("File was not created", err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	panicIfErr(err)
	s := string(b)

	if !strings.Contains(s, "this\nis\ntest") {
		t.Fatal("Note body is not correct:", s)
	}

	if cfg.GitPath != "" {
		dotgit := filepath.Join(cfg.HomePath, ".git")
		if s, err := os.Stat(dotgit); err != nil || !s.IsDir() {
			t.Fatal(".git directory was not created. `git init` did not run:", err)
		}
		os.RemoveAll(dotgit)
	}

	stdout, err := fake.String()
	panicIfErr(err)
	stdout = strings.TrimSuffix(stdout, "\n")
	if stdout != p {
		t.Error("Output is not path to the file:", stdout)
	}
}

func TestNewCmdNewNoteWithNoInlineInput(t *testing.T) {

	cfg := testNewConfigForNewCmd("empty")
	defer os.RemoveAll(filepath.Join(cfg.HomePath, ".git"))

	for _, tc := range []struct {
		cat   string
		title string
	}{
		{
			cat:   "cat",
			title: "inline-1",
		},
		{
			cat:   "cat",
			title: "inline-2",
		},
		{
			cat:   "cat",
			title: "inline-3",
		},
		{
			cat:   "cat2",
			title: "inline-1",
		},
		{
			cat:   "cat3",
			title: "inline-1",
		},
		{
			cat:   "nested/cat",
			title: "inline-1",
		},
		{
			cat:   "nested/cat",
			title: "inline-2",
		},
		{
			cat:   "nested/more/cat",
			title: "inline-1",
		},
		{
			cat:   "morenested/more/more/more/cat",
			title: "inline-1",
		},
		{
			cat:   "カテゴリ",
			title: "ノート",
		},
	} {
		for _, c := range []struct {
			opt string
			cmd *NewCmd
		}{
			{
				"noinline",
				&NewCmd{
					Config:   cfg,
					Category: tc.cat,
					Tags:     "foo, bar",
					NoInline: true,
				},
			},
			{
				"noedit",
				&NewCmd{
					Config:   cfg,
					Category: tc.cat,
					Tags:     "foo, bar",
					NoEdit:   true,
				},
			},
		} {
			title := tc.title + "-" + c.opt
			c.cmd.Filename = title + ".md"
			defer os.RemoveAll(filepath.Join(cfg.HomePath, strings.Split(tc.cat, "/")[0]))
			t.Run(tc.cat+"_"+title, func(t *testing.T) {
				fake := fakeio.Stdout()
				defer fake.Restore()

				if err := c.cmd.Do(); err != nil {
					t.Fatal(err)
				}

				p := filepath.Join(cfg.HomePath, filepath.FromSlash(tc.cat), title+".md")
				n, err := LoadNote(p, cfg)
				if err != nil {
					t.Fatal(err)
				}

				if n.Category != tc.cat {
					t.Error("Note category mismatch", n.Category)
				}

				if !reflect.DeepEqual(n.Tags, []string{"foo", "bar"}) {
					t.Error("Tags are not correct", n.Tags)
				}

				if n.Title != title {
					t.Error(n.Title)
				}

				if n.Created.After(time.Now()) {
					t.Error("Created date invalid", n.Created.Format(time.RFC3339))
				}

				stdout, err := fake.String()
				panicIfErr(err)
				stdout = strings.TrimSuffix(stdout, "\n")
				if stdout != p {
					t.Error("Output is not path to the file:", stdout)
				}
			})
		}
	}
}

func TestNewCmdNoteAlreadyExists(t *testing.T) {
	cfg := testNewConfigForNewCmd("fail")
	cmd := &NewCmd{
		Config:   cfg,
		Category: "cat",
		Filename: "already-exists",
		Tags:     "",
		NoInline: true,
	}

	err := cmd.Do()
	if err == nil {
		t.Fatal("No error occurred")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatal("Unexpected error:", err)
	}
}

func TestNewCmdNoteInvalidInput(t *testing.T) {
	cfg := testNewConfigForNewCmd("fail")
	cmd := &NewCmd{
		Config:   cfg,
		Category: "", // Empty category is not permitted
		Filename: "test",
		Tags:     "",
		NoInline: true,
	}

	err := cmd.Do()
	if err == nil {
		t.Fatal("No error occurred")
	}
	if !strings.Contains(err.Error(), "Invalid category part '' as directory name") {
		t.Fatal("Unexpected error:", err)
	}
}
