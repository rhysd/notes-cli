package notes

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/kballard/go-shellquote"
	"github.com/rhysd/go-fakeio"
)

func testNewConfigForListCmd(subdir string) *Config {
	cwd, err := os.Getwd()
	panicIfErr(err)
	return &Config{HomePath: filepath.Join(cwd, "testdata", "list", subdir)}
}

func TestListCmd(t *testing.T) {
	old := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = old }()

	for _, tc := range []struct {
		what   string
		cmd    *ListCmd
		want   string
		subdir string
	}{
		{
			what: "default",
			cmd:  &ListCmd{},
			want: `
			HOME/b/6.md
			HOME/c/3.md
			HOME/b/2.md
			HOME/c/5.md
			HOME/a/1.md
			HOME/a/4.md
			`,
		},
		{
			what: "sort by created",
			cmd: &ListCmd{
				SortBy: "created",
			},
			want: `
			HOME/b/6.md
			HOME/c/3.md
			HOME/b/2.md
			HOME/c/5.md
			HOME/a/1.md
			HOME/a/4.md
			`,
		},
		{
			what: "sort by filename",
			cmd: &ListCmd{
				SortBy: "filename",
			},
			want: `
			HOME/a/1.md
			HOME/b/2.md
			HOME/c/3.md
			HOME/a/4.md
			HOME/c/5.md
			HOME/b/6.md
			`,
		},
		{
			what: "sort by category",
			cmd: &ListCmd{
				SortBy: "category",
			},
			want: `
			HOME/a/1.md
			HOME/a/4.md
			HOME/b/2.md
			HOME/b/6.md
			HOME/c/3.md
			HOME/c/5.md
			`,
		},
		{
			what: "relative paths",
			cmd: &ListCmd{
				Relative: true,
			},
			want: `
			b/6.md
			c/3.md
			b/2.md
			c/5.md
			a/1.md
			a/4.md
			`,
		},
		{
			what: "relative paths sorted by file name",
			cmd: &ListCmd{
				Relative: true,
				SortBy:   "filename",
			},
			want: `
			a/1.md
			b/2.md
			c/3.md
			a/4.md
			c/5.md
			b/6.md
			`,
		},
		{
			what: "oneline",
			cmd: &ListCmd{
				Oneline: true,
			},
			want: `
			b/6.md future     text from future
			c/3.md            this is title
			b/2.md foo        this is title
			c/5.md a-bit-long this is title
			a/1.md foo,bar    this is title
			a/4.md bar        this is title this is title this is title this is title this is title this is title this is title this is title
			`,
		},
		{
			what: "oneline sorted by category",
			cmd: &ListCmd{
				Oneline: true,
				SortBy:  "category",
			},
			want: `
			a/1.md foo,bar    this is title
			a/4.md bar        this is title this is title this is title this is title this is title this is title this is title this is title
			b/2.md foo        this is title
			b/6.md future     text from future
			c/3.md            this is title
			c/5.md a-bit-long this is title
			`,
		},
		{
			what: "filter by category",
			cmd: &ListCmd{
				Category: "a",
			},
			want: `
			HOME/a/1.md
			HOME/a/4.md
			`,
		},
		{
			what: "filter by category with regex sorted by filename",
			cmd: &ListCmd{
				Category: "^(b|c)$",
				SortBy:   "filename",
			},
			want: `
			HOME/b/2.md
			HOME/c/3.md
			HOME/c/5.md
			HOME/b/6.md
			`,
		},
		{
			what: "filter by unknown category",
			cmd: &ListCmd{
				Category: "unknown-category-who-know",
			},
			want: `
			`,
		},
		{
			what: "filter by tag",
			cmd: &ListCmd{
				Tag: "foo",
			},
			want: `
			HOME/b/2.md
			HOME/a/1.md
			`,
		},
		{
			what: "filter by tag with regex sorted by filename",
			cmd: &ListCmd{
				Tag:    "^(foo|future)$",
				SortBy: "filename",
			},
			want: `
			HOME/a/1.md
			HOME/b/2.md
			HOME/b/6.md
			`,
		},
		{
			what: "filter by unknown tag",
			cmd: &ListCmd{
				Tag: "unknown-category-who-know",
			},
			want: `
			`,
		},
		{
			what: "filter by category and tag",
			cmd: &ListCmd{
				Category: "a",
				Tag:      "foo",
			},
			want: `
			HOME/a/1.md
			`,
		},
		{
			what: "full",
			cmd: &ListCmd{
				Full: true,
			},
			want: `
			HOME/b/6.md
			Category: b
			Tags:     future
			Created:  2118-10-30T11:37:45+09:00
			
			text from future
			================
			
			Lorem ipsum dolor sit amet, his no stet volumus sententiae. Usu id postea animal
			consetetur. Eum repudiare laboramus conclusionemque et, veritus tractatos dignissim
			duo ut. Ex sed quod admodum indoctum. No torquatos temporibus vis, mel tota causae
			quaestio ex, habeo laoreet adipiscing mea at.
			
			Quem necessitatibus quo et, eu ius ceteros efficiendi, ocurreret moderatius elaboraret
			no quo. Est at facilisis gubergren. Ius ea similique intellegam, quo ne soluta inermis.
			Et brute fastidii cum, sea in ferri delectus, his fastidii nominati vituperatoribus te.
			Nam ad nisl quot omittantur, graeco scripta inciderint at pro.
			
			...
			
			HOME/c/3.md
			Category: c
			Tags:     
			Created:  2018-12-30T11:37:45+09:00
			
			this is title
			=============
			
			this
			is
			newer
			this
			is
			newer
			this
			is
			newer
			this
			...
			
			HOME/b/2.md
			Category: b
			Tags:     foo
			Created:  2018-11-01T11:37:45+09:00
			
			this is title
			=============
			
			Lorem ipsum dolor sit amet, his no stet volumus sententiae. Usu id postea animal consetetur. Eum repudiare laboramus conclusionemque et, veritus tractatos dignissim duo ut. Ex sed quod admodum indoctum. No torquatos temporibus vis, mel tota causae quaestio ex, habeo laoreet adipiscing mea at.

			Quem necessitatibus quo et, eu ius ceteros efficiendi, ocurreret moderatius elaboraret no quo. Est at facilisis gubergren. Ius ea similique intellegam, quo ne soluta inermis. Et brute fastidii cum, sea in ferri delectus, his fastidii nominati vituperatoribus te. Nam ad nisl quot omittantur, graeco scripta inciderint at pro.
			
			HOME/c/5.md
			Category: c
			Tags:     a-bit-long
			Created:  2018-10-30T11:37:45+09:00
			
			this is title
			=============
			
			this
			is
			test
			this
			is
			test
			this
			is
			test
			this
			...
			
			HOME/a/1.md
			Category: a
			Tags:     foo, bar
			Created:  2018-10-30T11:17:45+09:00
			
			this is title
			=============
			
			this
			is
			test
			
			HOME/a/4.md
			Category: a
			Tags:     bar
			Created:  2017-10-30T11:37:45+09:00
			
			this is title this is title this is title this is title this is title this is title this is title this is title
			===============================================================================================================
			
			this
			is
			old text
			
			`,
		},
		{
			what: "full with filter",
			cmd: &ListCmd{
				Full:     true,
				Category: "a",
				Tag:      "foo",
			},
			want: `
			HOME/a/1.md
			Category: a
			Tags:     foo, bar
			Created:  2018-10-30T11:17:45+09:00
			
			this is title
			=============
			
			this
			is
			test
			
			`,
		},
		{
			what:   "nested categories",
			subdir: "nested",
			cmd:    &ListCmd{},
			want: `
			HOME/c/6.md
			HOME/a/5.md
			HOME/a/d/4.md
			HOME/a/d/e/3.md
			HOME/b/2.md
			HOME/b/f/1.md
			`,
		},
		{
			what:   "sort nested categories",
			subdir: "nested",
			cmd: &ListCmd{
				SortBy: "category",
			},
			want: `
			HOME/a/5.md
			HOME/a/d/4.md
			HOME/a/d/e/3.md
			HOME/b/2.md
			HOME/b/f/1.md
			HOME/c/6.md
			`,
		},
		{
			what:   "categories filtered by category",
			subdir: "nested",
			cmd: &ListCmd{
				Category: "/d",
			},
			want: `
			HOME/a/d/4.md
			HOME/a/d/e/3.md
			`,
		},
		{
			what:   "categories oneline log",
			subdir: "nested",
			cmd: &ListCmd{
				Category: "/d",
				Oneline:  true,
			},
			want: `
			a/d/4.md   d,bar  this is title
			a/d/e/3.md a,piyo this is title
			`,
		},
		{
			what:   "categories full log",
			subdir: "nested",
			cmd: &ListCmd{
				Category: "/d",
				Tag:      "piyo",
				Full:     true,
			},
			want: `
			HOME/a/d/e/3.md
			Category: a/d/e
			Tags:     a, piyo
			Created:  2018-11-03T11:17:45+09:00
			
			this is title
			=============
			
			this
			is
			test
			
			`,
		},
		{
			what:   "oneline log",
			subdir: "widechars",
			cmd: &ListCmd{
				Oneline: true,
			},
			want: `
			cat/ノート.md             タグ1,x     これはタイトル
			カテゴリ/ネスト/ノート.md タグ1,タグ2 これはタイトル
			カテゴリ/ノート.md        tag3,タグ1  これはタイトル
			`,
		},
		{
			what:   "full log with category and tag filtering",
			subdir: "widechars",
			cmd: &ListCmd{
				Category: "カテゴリ",
				Tag:      "タグ2",
				Full:     true,
			},
			want: `
			HOME/カテゴリ/ネスト/ノート.md
			Category: カテゴリ/ネスト
			Tags:     タグ1, タグ2
			Created:  2018-11-02T11:17:45+09:00
			
			これはタイトル
			==============
			
			これはマルチバイト文字のテストです
			
			`,
		},
	} {
		subdir := tc.subdir
		if subdir == "" {
			subdir = "normal"
		}

		t.Run(subdir+"_"+tc.what, func(t *testing.T) {
			cfg := testNewConfigForListCmd(subdir)

			var buf bytes.Buffer
			tc.cmd.Config = cfg
			tc.cmd.Out = &buf

			if err := tc.cmd.Do(); err != nil {
				t.Fatal(err)
			}

			have := buf.String()
			lines := strings.Split(strings.TrimPrefix(tc.want, "\n"), "\n")
			for i, s := range lines {
				// Convert only first field as file path
				ss := strings.Split(s, " ")
				ss[0] = strings.Replace(filepath.FromSlash(strings.TrimLeft(ss[0], "\t")), "HOME", cfg.HomePath, 1)
				lines[i] = strings.Join(ss, " ")
			}
			want := strings.Join(lines, "\n")

			if want != have {
				ls := strings.Split(want, "\n")
				hint := ""
				for i, l := range strings.Split(have, "\n") {
					if len(ls) <= i {
						t.Fatalf("Having text is longer than wanted text at line %d: '%s'", i, l)
					}
					if l != ls[i] {
						hint = fmt.Sprintf("first mismatch at line %d: want:%#v v.s. have:%#v", i+1, ls[i], l)
						break
					}
				}
				t.Fatalf("have:\n'%s'\n\n%s", have, hint)
			}
		})
	}
}

func TestListWriteError(t *testing.T) {
	cfg := testNewConfigForListCmd("normal")
	for _, tc := range []struct {
		what string
		cmd  *ListCmd
		want string
	}{
		{
			what: "normal",
			cmd:  &ListCmd{},
		},
		{
			what: "oneline",
			cmd: &ListCmd{
				Oneline: true,
			},
		},
		{
			what: "full",
			cmd: &ListCmd{
				Full: true,
			},
		},
	} {
		t.Run(tc.what, func(t *testing.T) {
			tc.cmd.Config = cfg
			tc.cmd.Out = alwaysErrorWriter{}
			if err := tc.cmd.Do(); err == nil || !strings.Contains(err.Error(), "Write error for test") {
				t.Fatal("Unexpected error", err)
			}
		})
	}
}

func TestListNotesVariousHomes(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	for _, tc := range []struct {
		home string
		want string
	}{
		{
			home: "with-template",
			want: "a/1.md\n",
		},
		{
			home: "hide-metadata",
			want: "a/2.md\na/1.md\n",
		},
		{
			home: "with-dot-dirs",
			want: "a/1.md\n",
		},
	} {
		t.Run(tc.home, func(t *testing.T) {
			cfg := &Config{HomePath: filepath.Join(cwd, "testdata", "list", tc.home)}

			var buf bytes.Buffer
			cmd := &ListCmd{
				Config:   cfg,
				Out:      &buf,
				Relative: true,
			}

			if err := cmd.Do(); err != nil {
				t.Fatal(err)
			}

			have := buf.String()

			ss := strings.Split(tc.want, "\n")
			for i, s := range ss {
				ss[i] = filepath.FromSlash(strings.Replace(s, "\t", "", -1))
			}
			want := strings.Join(ss, "\n")

			if want != have {
				t.Fatalf("Wanted %#v but have %#v", want, have)
			}
		})
	}
}

func TestListNoNote(t *testing.T) {
	dir := "test-for-list-empty"
	cfg := &Config{HomePath: dir}
	panicIfErr(os.Mkdir(dir, 0755))
	defer func() { panicIfErr(os.RemoveAll(dir)) }()

	for _, c := range []*ListCmd{
		&ListCmd{},
		&ListCmd{Oneline: true},
		&ListCmd{Relative: true},
		&ListCmd{Full: true},
	} {
		var b bytes.Buffer
		c.Config = cfg
		c.Out = &b
		if err := c.Do(); err != nil {
			t.Fatal(err)
		}

		out := b.String()
		if out != "" {
			t.Fatalf("should not output anything: %#v", out)
		}
	}
}

func TestListNoHome(t *testing.T) {
	cfg := &Config{HomePath: "/path/to/unknown/directory"}
	err := (&ListCmd{Config: cfg}).Do()
	if err == nil {
		t.Fatal("Error did not occur")
	}
	if !strings.Contains(err.Error(), "Cannot read home") {
		t.Fatal("Unexpected error:", err)
	}
}

func TestListBrokenCategoryRegex(t *testing.T) {
	cfg := testNewConfigForListCmd("normal")
	cmd := &ListCmd{
		Config:   cfg,
		Category: "(foo",
	}

	err := cmd.Do()
	if err == nil {
		t.Fatal("Error did not occur")
	}
	if !strings.Contains(err.Error(), "Regular expression for filtering categories is invalid") {
		t.Fatal("Unexpected error:", err)
	}
}

func TestListBrokenTagRegex(t *testing.T) {
	cfg := testNewConfigForListCmd("normal")
	cmd := &ListCmd{
		Config: cfg,
		Tag:    "(foo",
	}

	err := cmd.Do()
	if err == nil {
		t.Fatal("Error did not occur")
	}
	if !strings.Contains(err.Error(), "Regular expression for filtering tags is invalid") {
		t.Fatal("Unexpected error:", err)
	}
}

func TestListBrokenNote(t *testing.T) {
	cfg := testNewConfigForListCmd("fail")
	cmd := &ListCmd{Config: cfg}
	err := cmd.Do()
	if err == nil {
		t.Fatal("Error did not occur")
	}
	if !strings.Contains(err.Error(), "Cannot parse created date time") {
		t.Fatal("Unexpected error:", err)
	}
}

func TestListSortByModified(t *testing.T) {
	cwd, err := os.Getwd()
	panicIfErr(err)
	cfg := &Config{HomePath: filepath.Join(cwd, "testdata", "modified-order")}

	now := time.Now()
	panicIfErr(os.Chtimes(filepath.Join(cfg.HomePath, "a", "2.md"), now, now))

	var buf bytes.Buffer
	cmd := &ListCmd{
		SortBy: "modified",
		Config: cfg,
		Out:    &buf,
	}

	if err := cmd.Do(); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.Trim(buf.String(), "\n"), "\n")
	mods := []time.Time{}
	for i, l := range lines {
		s, err := os.Stat(l)
		if err != nil {
			t.Fatal("Cannot load note", l, "at line", i)
		}
		mods = append(mods, s.ModTime())
	}

	prev := mods[0]
	for i, cur := range mods[1:] {
		if prev.Before(cur) {
			t.Fatal("not sorted at index", i, "prev:", prev.Format(time.RFC3339), "cur:", cur.Format(time.RFC3339))
		}
		prev = cur
	}
}

func TestListNoteEmptyBody(t *testing.T) {
	old := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = old }()

	cfg := testNewConfigForListCmd("empty")
	for _, tc := range []struct {
		what string
		cmd  *ListCmd
		want string
	}{
		{
			what: "oneline",
			cmd: &ListCmd{
				Oneline: true,
			},
			want: `
			a<SEP>1.md foo,bar empty body
			`,
		},
		{
			what: "full",
			cmd: &ListCmd{
				Full: true,
			},
			want: `
			<HOME><SEP>a<SEP>1.md
			Category: a
			Tags:     foo, bar
			Created:  2018-10-30T11:37:45+09:00
			
			empty body
			==========
			
			`,
		},
	} {
		t.Run(tc.what, func(t *testing.T) {
			var buf bytes.Buffer
			tc.cmd.Config = cfg
			tc.cmd.Out = &buf
			if err := tc.cmd.Do(); err != nil {
				t.Fatal(err)
			}
			have := buf.String()
			want := strings.TrimPrefix(tc.want, "\n")
			want = strings.Replace(want, "\t", "", -1)
			want = strings.Replace(want, "<HOME>", cfg.HomePath, -1)
			want = strings.Replace(want, "<SEP>", string(filepath.Separator), -1)
			if want != have {
				t.Fatalf("Wanted %#v but have %#v", want, have)
			}
		})
	}
}

func TestListCmdEditOption(t *testing.T) {
	fake := fakeio.Stdout()
	defer fake.Restore()

	exe, err := exec.LookPath("echo")
	panicIfErr(err)
	exe = shellquote.Join(exe) // On Windows it may contain 'Program Files' so quoting is necessary

	cfg := testNewConfigForListCmd("normal")
	cfg.EditorCmd = exe

	var buf bytes.Buffer
	cmd := &ListCmd{
		Config: cfg,
		Out:    &buf,
		Edit:   true,
	}

	if err := cmd.Do(); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if out != "" {
		t.Fatal("Unexpected output from command itself:", out)
	}

	stdout, err := fake.String()
	panicIfErr(err)

	have := strings.Split(strings.TrimRight(stdout, "\n"), " ")
	want := []string{}
	for _, p := range []string{
		"b/6.md",
		"c/3.md",
		"b/2.md",
		"c/5.md",
		"a/1.md",
		"a/4.md",
	} {
		p = filepath.Join(cfg.HomePath, filepath.FromSlash(p))
		want = append(want, p)
	}

	if !reflect.DeepEqual(want, have) {
		t.Fatal("Args passed to editor is not expected:", have, "wanted", want)
	}
}

func TestListNoNotePrintNothing(t *testing.T) {
	cfg := testNewConfigForListCmd("no-note")
	var buf bytes.Buffer

	cmd := &ListCmd{
		Config: cfg,
		Out:    &buf,
	}

	if err := cmd.Do(); err != nil {
		t.Fatal(err)
	}

	if buf.Len() != 0 {
		t.Fatal("Some output:", buf.String())
	}
}

func TestListPagingWithPager(t *testing.T) {
	if _, err := exec.LookPath("cat"); err != nil {
		t.Skip("`cat` command is necessary to run this test")
	}

	var buf bytes.Buffer

	cfg := testNewConfigForListCmd("normal")
	cfg.PagerCmd = "cat"
	cmd := &ListCmd{
		Config: cfg,
		Out:    &buf,
	}

	if err := cmd.Do(); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.Trim(buf.String(), "\n"), "\n")
	for _, l := range lines {
		if _, err := os.Stat(l); err != nil {
			t.Fatal(l, "does not exist:", err)
		}
	}
}

func TestListPagingError(t *testing.T) {
	for _, tc := range []struct {
		cmd  string
		want string
		out  io.Writer
	}{
		{
			cmd:  "'foo",
			want: "Cannot parsing",
		},
		{
			cmd:  "/path/to/bin/unknown",
			want: "Cannot start pager command",
		},
		{
			cmd:  "cat",
			want: "Pager command did not run successfully: Write error for test",
			out:  alwaysErrorWriter{},
		},
	} {
		t.Run(tc.cmd, func(t *testing.T) {
			cfg := testNewConfigForListCmd("normal")
			cfg.PagerCmd = tc.cmd

			out := tc.out
			if out == nil {
				out = io.Discard
			}

			cmd := &ListCmd{
				Config: cfg,
				Out:    out,
			}

			if err := cmd.Do(); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatal("Error unexpected", err)
			}
		})
	}
}
