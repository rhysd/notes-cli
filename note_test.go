package notes

import (
	"errors"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func noteTestdataConfig() *Config {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	home := filepath.Join(cwd, "testdata", "note")
	return &Config{GitPath: "git", HomePath: home}
}

func TestNewNote(t *testing.T) {
	cfg := &Config{GitPath: "git", HomePath: "."}

	n, err := NewNote("cat", "foo,bar", "foo.md", "this is title", cfg)
	if err != nil {
		t.Fatal(err)
	}

	if n.Category != "cat" {
		t.Error(n.Category)
	}

	if !reflect.DeepEqual(n.Tags, []string{"foo", "bar"}) {
		t.Error(n.Tags)
	}

	if n.File != "foo.md" {
		t.Error(n.File)
	}

	if n.Title != "this is title" {
		t.Error(n.Title)
	}
}

func TestNewNoteFilenameNormalize(t *testing.T) {
	cfg := &Config{GitPath: "git", HomePath: "."}

	for _, f := range []string{"foo-bar", "foo bar.md", "foo bar"} {
		n, err := NewNote("cat", "foo,bar", f, "", cfg)
		if err != nil {
			t.Fatal(err)
		}
		if n.File != "foo-bar.md" {
			t.Error("Not normalized:", n.File)
		}
	}
}

func TestNewNoteError(t *testing.T) {
	cfg := &Config{GitPath: "git", HomePath: "."}

	if _, err := NewNote("", "", "foo.md", "", cfg); err == nil {
		t.Error("empty category should cause an error")
	}

	if _, err := NewNote("cat", "", "", "", cfg); err == nil {
		t.Error("empty file name should cause an error")
	}
}

func TestNoteDirPath(t *testing.T) {
	cfg := &Config{GitPath: "git", HomePath: "/path/to/home"}
	n, _ := NewNote("cat", "foo,bar", "foo.md", "", cfg)
	have := n.DirPath()
	if have != filepath.FromSlash("/path/to/home/cat") {
		t.Fatal(have)
	}
}

func TestNoteFilePath(t *testing.T) {
	cfg := &Config{GitPath: "git", HomePath: "/path/to/home"}
	n, _ := NewNote("cat", "foo,bar", "foo.md", "", cfg)
	have := n.FilePath()
	if have != filepath.FromSlash("/path/to/home/cat/foo.md") {
		t.Fatal(have)
	}
}

func TestNoteRelFilePath(t *testing.T) {
	cfg := &Config{GitPath: "git", HomePath: "/path/to/home"}
	n, _ := NewNote("cat", "foo,bar", "foo.md", "", cfg)
	have := n.RelFilePath()
	if have != filepath.FromSlash("cat/foo.md") {
		t.Fatal(have)
	}
}

func TestCreateNoteFile(t *testing.T) {
	cfg := noteTestdataConfig()

	check := func(n *Note, err error) *Note {
		if err != nil {
			panic(err)
		}
		return n
	}

	heredoc := func(s string) string {
		return strings.Replace(strings.Trim(s, "\n"), "\t", "", -1)
	}

	for _, tc := range []struct {
		note *Note
		want string
	}{
		{
			note: check(NewNote("cat1", "foo,bar", "create-normal", "this is title", cfg)),
			want: heredoc(`
				this is title
				=============
				- Category: cat1
				- Tags: foo, bar
				- Created: {{created}}
				
				`),
		},
		{
			note: check(NewNote("cat1", "foo,bar", "create-title-is-empty", "", cfg)),
			want: heredoc(`
				create-title-is-empty
				=====================
				- Category: cat1
				- Tags: foo, bar
				- Created: {{created}}
				
				`),
		},
		{
			note: check(NewNote("cat1", "foo", "create-one-tag", "this is title", cfg)),
			want: heredoc(`
				this is title
				=============
				- Category: cat1
				- Tags: foo
				- Created: {{created}}
				
				`),
		},
		{
			note: check(NewNote("cat1", "", "create-no-tag", "this is title", cfg)),
			want: heredoc(`
				this is title
				=============
				- Category: cat1
				- Tags: 
				- Created: {{created}}
				
				`),
		},
	} {
		t.Run(tc.note.File, func(t *testing.T) {
			n := tc.note
			p := filepath.Join(cfg.HomePath, n.Category, n.File)
			if err := n.Create(); err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(p)

			if n.Created.IsZero() {
				t.Fatal("Created date time was not set")
			}

			f, err := os.Open(p)
			if err != nil {
				t.Fatal(err)
			}

			b, err := ioutil.ReadAll(f)
			if err != nil {
				t.Fatal(err)
			}
			have := string(b)

			want := strings.Replace(tc.want, "{{created}}", n.Created.Format(time.RFC3339), -1)

			if have != want {
				t.Fatalf("have:\n%s\nwant:\n%s\nGenerated note is unexpected", have, want)
			}
		})
	}
}

func TestNoteCreateFail(t *testing.T) {
	cfg := noteTestdataConfig()

	n, err := NewNote("fail", "", "create-already-exists.md", "this is title", cfg)
	if err != nil {
		panic(err)
	}

	err = n.Create()
	if err == nil {
		t.Fatal("error did not occur")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Fatal("unexpected error", err)
	}
}

func TestNoteOpenNoEditor(t *testing.T) {
	cfg := &Config{GitPath: "git", HomePath: "."}
	n, err := NewNote("cat", "", "foo", "title", cfg)
	if err != nil {
		panic(err)
	}
	err = n.Open()
	if err == nil {
		t.Fatal("error did not occur")
	}
	if !strings.Contains(err.Error(), "Editor is not set") {
		t.Fatal("unexpected error:", err)
	}
}

func TestNoteOpenEditor(t *testing.T) {
	cfg := noteTestdataConfig()

	bin, err := exec.LookPath("true")
	if err != nil {
		panic(err)
	}
	cfg.EditorPath = bin

	n, err := NewNote("cat1", "", "foo", "title", cfg)
	if err != nil {
		panic(err)
	}
	if err := n.Open(); err != nil {
		t.Fatal(err)
	}
}

func TestNoteOpenEditorFail(t *testing.T) {
	cfg := noteTestdataConfig()

	bin, err := exec.LookPath("false")
	if err != nil {
		panic(err)
	}
	cfg.EditorPath = bin

	n, err := NewNote("cat1", "", "foo", "title", cfg)
	if err != nil {
		panic(err)
	}

	err = n.Open()
	if err == nil {
		t.Fatal("Error did not occur")
	}
	if !strings.Contains(err.Error(), "Editor command did not run successfully") {
		t.Fatal("Unexpected error:", err)
	}
}

func TestNoteReadBodyN(t *testing.T) {
	cfg := noteTestdataConfig()
	must := func(n *Note, err error) *Note {
		if err != nil {
			panic(err)
		}
		return n
	}

	for _, tc := range []struct {
		note *Note
		want string
	}{
		{
			note: must(NewNote("read-body", "", "short", "this is title", cfg)),
			want: "this\nis\ntest\n",
		},
		{
			note: must(NewNote("read-body", "", "long", "this is title", cfg)),
			want: "Lorem ipsum dolor si",
		},
		{
			note: must(NewNote("read-body", "", "newlines-before-body", "this is title", cfg)),
			want: "this\nis\ntest\n",
		},
		{
			note: must(NewNote("read-body", "", "no-body", "this is title", cfg)),
			want: "",
		},
	} {
		t.Run(tc.note.File, func(t *testing.T) {
			have, err := tc.note.ReadBodyN(20)
			if err != nil {
				t.Fatal(err)
			}
			if have != tc.want {
				t.Fatalf("have:\n%s\nwant:\n%s\nread string is unexpected", have, tc.want)
			}
		})
	}
}

func TestNoteReadBodyNFailure(t *testing.T) {
	cfg := noteTestdataConfig()
	for _, file := range []string{"missing-created", "missing-tags", "missing-category"} {
		t.Run(file, func(t *testing.T) {
			n, err := NewNote("fail", "", file, "this is title", cfg)
			if err != nil {
				panic(err)
			}
			_, err = n.ReadBodyN(20)
			if err == nil {
				t.Fatal("Error did not occur")
			}
			if !strings.Contains(err.Error(), "Some metadata may be missing") {
				t.Fatal("Unexpected error:", err)
			}
		})
	}
}

func TestLoadNote(t *testing.T) {
	cmpopt := cmpopts.IgnoreFields(Note{}, "Config", "Created")
	cfg := noteTestdataConfig()
	created, err := time.Parse(time.RFC3339, "2018-10-30T11:37:45+09:00")
	if err != nil {
		panic(err)
	}

	for _, tc := range []struct {
		file  string
		tags  string
		title string
	}{
		{
			file:  "normal",
			tags:  "foo,bar",
			title: "this is title",
		},
		{
			file:  "no-body",
			tags:  "foo,bar",
			title: "this is title",
		},
		{
			file:  "no-tag",
			tags:  "",
			title: "this is title",
		},
		{
			file:  "empty-title",
			tags:  "foo,bar",
			title: "(no title)",
		},
		{
			file:  "space-around-metadata",
			tags:  "foo,bar",
			title: "this is title",
		},
	} {
		t.Run(tc.file, func(t *testing.T) {
			want, err := NewNote("load", tc.tags, tc.file, tc.title, cfg)
			if err != nil {
				panic(err)
			}

			have, err := LoadNote(want.FilePath(), cfg)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(want, have, cmpopt) {
				t.Fatal(cmp.Diff(want, have, cmpopt))
			}

			if have.Created.IsZero() {
				t.Fatal("Created time was not loaded")
			}

			if have.Created != created {
				t.Fatal("Unexpected created datetime", have.Created.Format(time.RFC3339))
			}
		})
	}
}

func TestLoadNoteFail(t *testing.T) {
	cfg := noteTestdataConfig()
	for _, tc := range []struct {
		file string
		msg  string
	}{
		{
			file: "not-existing-file.md",
			msg:  "Cannot open note file",
		},
		{
			file: "missing-category.md",
			msg:  "Missing metadata in file",
		},
		{
			file: "missing-tags.md",
			msg:  "Missing metadata in file",
		},
		{
			file: "missing-created.md",
			msg:  "Missing metadata in file",
		},
		{
			file: "missing-title.md",
			msg:  "No title found in note",
		},
		{
			file: "timeformat-broken.md",
			msg:  "Cannot parse created date time as RFC3339 format",
		},
		{
			file: "category-mismatch.md",
			msg:  "Category does not match between file path and file content",
		},
	} {
		t.Run(tc.file, func(t *testing.T) {
			_, err := LoadNote(filepath.Join(cfg.HomePath, "fail", tc.file), cfg)
			if err == nil {
				t.Fatal("Error did not occur")
			}
			if !strings.Contains(err.Error(), tc.msg) {
				t.Fatal("Unexpected error:", err)
			}
		})
	}
}

func TestWalkNotes(t *testing.T) {
	cfg := noteTestdataConfig()
	cfg.HomePath = filepath.Join(cfg.HomePath, "walk")

	want := map[string]struct{}{
		"a/a.md": struct{}{},
		"b/b.md": struct{}{},
	}

	have := map[string]struct{}{}
	if err := WalkNotes(cfg.HomePath, cfg, func(p string, n *Note) error {
		p2 := n.FilePath()
		if p != p2 {
			t.Fatalf("'%s' v.s. '%s'", p, p2)
		}
		have[n.RelFilePath()] = struct{}{}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(want, have) {
		t.Fatal(cmp.Diff(want, have))
	}
}

func TestWalkNotesPredReturnError(t *testing.T) {
	cfg := noteTestdataConfig()
	cfg.HomePath = filepath.Join(cfg.HomePath, "walk")
	err := WalkNotes(cfg.HomePath, cfg, func(p string, n *Note) error {
		return errors.New("hello")
	})
	if err == nil {
		t.Fatal("Error did not occur")
	}
	if !strings.Contains(err.Error(), "hello") {
		t.Fatal("Unexpected error:", err)
	}
}

func TestWalkNotesBrokenNote(t *testing.T) {
	cfg := noteTestdataConfig()
	err := WalkNotes(filepath.Join(cfg.HomePath, "fail"), cfg, func(p string, n *Note) error {
		return nil // Do nothing
	})
	if err == nil {
		t.Fatal("No error occurred")
	}
}
