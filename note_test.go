package notes

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kballard/go-shellquote"
)

func noteTestdataConfig() *Config {
	cwd, err := os.Getwd()
	panicIfErr(err)
	home := filepath.Join(cwd, "testdata", "note")
	return &Config{GitPath: "git", HomePath: home}
}

func TestNewNoteOK(t *testing.T) {
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

	for _, f := range []string{"foo-bar", "foo-bar.md"} {
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

	for _, tc := range []struct {
		cat  string
		file string
		want string
	}{
		{
			cat:  "",
			file: "foo.md",
			want: "Cannot be empty",
		},
		{
			cat:  "cat",
			file: "",
			want: "File name cannot be empty",
		},
		{
			cat:  "cat",
			file: ".foo",
			want: "cannot start with '.'",
		},
		{
			cat:  "foo|bar",
			file: "",
			want: "Invalid category part 'foo|bar' as directory name",
		},
	} {
		t.Run(tc.want, func(t *testing.T) {
			_, err := NewNote(tc.cat, "", tc.file, "", cfg)
			if err == nil {
				t.Fatal("Error did not occur")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatal("Unexpected error:", err)
			}
		})
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
		panicIfErr(err)
		return n
	}

	heredoc := func(s string) string {
		return strings.Replace(strings.Trim(s, "\n"), "\t", "", -1)
	}

	for _, tc := range []struct {
		note       *Note
		want       string
		nestedHome string
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
		{
			note: check(NewNote("with-template", "", "create-with-template", "this is title", cfg)),
			want: heredoc(`
				this is title
				=============
				- Category: with-template
				- Tags: 
				- Created: {{created}}
				!!!!!!!
				This text was inserted via template
				!!!!!!!
				`),
		},
		{
			note: check(NewNote("with-template2", "", "create-with-template", "this is title", cfg)),
			want: heredoc(`
				this is title
				=============
				- Category: with-template2
				- Tags: 
				- Created: {{created}}
				-------------
				
				!!!!!!!
				This text was inserted via template
				!!!!!!!
				`),
		},
		{
			note: check(NewNote("with-template-comment-metadata", "", "commentout-metadata-with-template", "this is title", cfg)),
			want: heredoc(`
				this is title
				=============
				<!--
				- Category: with-template-comment-metadata
				- Tags: 
				- Created: {{created}}
				-->
				-------------
				
				Metadata is commented out since template starts with '-->'
				`),
		},
		{
			nestedHome: "template-at-home",
			note:       check(NewNote("cat1", "", "create-with-template-at-home", "this is title", cfg)),
			want: heredoc(`
				this is title
				=============
				- Category: cat1
				- Tags: 
				- Created: {{created}}
				------
				
				Template at root was inserted
				`),
		},
		{
			nestedHome: "template-at-home",
			note:       check(NewNote("cat2", "", "prioritize-category-local-template", "this is title", cfg)),
			want: heredoc(`
				this is title
				=============
				- Category: cat2
				- Tags: 
				- Created: {{created}}
				------
				
				Template under cat2 was inserted
				`),
		},
	} {
		t.Run(tc.note.File, func(t *testing.T) {
			if tc.nestedHome != "" {
				old := cfg.HomePath
				defer func() { cfg.HomePath = old }()
				cfg.HomePath = filepath.Join(cfg.HomePath, tc.nestedHome)
			}
			n := tc.note
			p := filepath.Join(cfg.HomePath, n.Category, n.File)
			if err := n.Create(); err != nil {
				t.Fatal(err)
			}
			defer func() { panicIfErr(os.RemoveAll(p)) }()

			if n.Created.IsZero() {
				t.Fatal("Created date time was not set")
			}

			f, err := os.Open(p)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			b, err := io.ReadAll(f)
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
	panicIfErr(err)

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
	panicIfErr(err)
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
	panicIfErr(err)
	bin = shellquote.Join(bin) // On Windows, it may contain 'Program Files'
	cfg.EditorCmd = bin

	n, err := NewNote("cat1", "", "foo", "title", cfg)
	panicIfErr(err)
	if err := n.Open(); err != nil {
		t.Fatal(err)
	}
}

func TestNoteOpenEditorFail(t *testing.T) {
	cfg := noteTestdataConfig()

	bin, err := exec.LookPath("false")
	panicIfErr(err)
	bin = shellquote.Join(bin) // On Windows, it may contain 'Program Files'
	cfg.EditorCmd = bin

	n, err := NewNote("cat1", "", "foo", "title", cfg)
	panicIfErr(err)

	err = n.Open()
	if err == nil {
		t.Fatal("Error did not occur")
	}
	if !strings.Contains(err.Error(), "did not exit successfully") {
		t.Fatal("Unexpected error:", err)
	}
}

func TestNoteReadBodyN(t *testing.T) {
	cfg := noteTestdataConfig()
	must := func(n *Note, err error) *Note {
		panicIfErr(err)
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
			want: "Lorem ipsum dolor sit amet, his no stet volumus sententiae.\nUsu id postea animal consetetur.\nEum repudiare laboramus conclusionemque et, veritus tractatos dignissim duo ut.\nEx sed quod admodum indoctum.\n",
		},
		{
			note: must(NewNote("read-body", "", "newlines-before-body", "this is title", cfg)),
			want: "this\nis\ntest\n",
		},
		{
			note: must(NewNote("read-body", "", "no-body", "this is title", cfg)),
			want: "",
		},
		{
			note: must(NewNote("read-body", "", "ignore-horizontal-rules", "this is title", cfg)),
			want: "text\n\n---\n^ not ignored\n",
		},
		{
			note: must(NewNote("hide-metadata", "", "1.md", "this is title", cfg)),
			want: "this\nis\ntest\n",
		},
		{
			note: must(NewNote("hide-metadata", "", "2.md", "this is title", cfg)),
			want: "this\nis\ntest\n",
		},
	} {
		t.Run(tc.note.File, func(t *testing.T) {
			have, size, err := tc.note.ReadBodyLines(4)
			if err != nil {
				t.Fatal(err)
			}
			if size > 4 {
				t.Fatal("Returned size is over max size (=4):", size)
			}
			if have != tc.want {
				t.Fatalf("have:\n%s\nwant:\n%s\nread string is unexpected", have, tc.want)
			}
			c := strings.Count(have, "\n")
			if size != c {
				t.Fatal("Returned size is", size, "but having", c, "newlines in body")
			}
		})
	}
}

func TestNoteReadBodyNFailure(t *testing.T) {
	cfg := noteTestdataConfig()
	for _, file := range []string{"missing-created", "missing-tags", "missing-category"} {
		t.Run(file, func(t *testing.T) {
			n, err := NewNote("fail", "", file, "this is title", cfg)
			panicIfErr(err)
			_, _, err = n.ReadBodyLines(20)
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
	panicIfErr(err)

	for _, tc := range []struct {
		file  string
		tags  string
		title string
		cat   string
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
			file:  "1.md",
			tags:  "foo,bar",
			title: "this is title",
			cat:   "hide-metadata",
		},
		{
			file:  "2.md",
			tags:  "foo,bar",
			title: "this is title",
			cat:   "hide-metadata",
		},
	} {
		t.Run(tc.file, func(t *testing.T) {
			cat := tc.cat
			if cat == "" {
				cat = "load"
			}

			want, err := NewNote(cat, tc.tags, tc.file, tc.title, cfg)
			panicIfErr(err)

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

			if have.Created.Unix() != created.Unix() {
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

func TestLoadNoteMismatchCategory(t *testing.T) {
	cfg := noteTestdataConfig()
	_, err := LoadNote(filepath.Join(cfg.HomePath, "fail", "category-mismatch.md"), cfg)
	if err == nil {
		t.Fatal("Error did not occur")
	}
	if !errors.Is(err, &MismatchCategoryError{}) {
		t.Fatal("Unexpected error:", err)
	}
}
