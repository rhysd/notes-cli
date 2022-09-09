package notes

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/rhysd/go-fakeio"
	"github.com/rhysd/go-tmpenv"
)

func TestNewExternalCmdOK(t *testing.T) {
	cwd, err := os.Getwd()
	panicIfErr(err)
	bindir := filepath.Join(cwd, "testdata", "external", "bin-name")
	tmp := tmpenv.New("PATH")
	defer tmp.Restore()

	panicIfErr(os.Setenv("PATH", os.Getenv("PATH")+string(os.PathListSeparator)+bindir))

	exe, err := os.Executable()
	panicIfErr(err)

	for _, name := range []string{
		"foo", "foo-bar", "foo_bar", "-", "_", "-foo", "bar_",
	} {
		args := []string{"--foo", "xxx", "-b"}
		c, ok := NewExternalCmd(fmt.Errorf(`expected command but got "%s"`, name), args)
		if !ok {
			t.Fatal("subcommand was not extracted", name)
		}

		want := "notes-" + name
		if runtime.GOOS == "windows" {
			want += ".bat"
		}

		if have := filepath.Base(c.ExePath); have != want {
			t.Fatal("Wanted command name", want, "but have", have)
		}
		if !reflect.DeepEqual(args, c.Args) {
			t.Fatal("Passed args are unexpected:", c.Args)
		}
		if !strings.HasSuffix(c.NotesPath, exe) {
			t.Fatal("`notes` full path is unexpected:", c.NotesPath, "wanted full path of", exe)
		}
	}
}

func TestNewExternalCmdSubcmdNotFound(t *testing.T) {
	for _, tc := range []struct {
		what string
		msg  string
	}{
		{
			what: "subcommand is not contained in parse error message",
			msg:  "unknown long flag '--foo'",
		},
		{
			what: "unknown subcommand",
			msg:  `expected command but got "unknown-subcommand-specified"`,
		},
		{
			what: "invalid subcommand name",
			msg:  `expected command but got "./foo"`,
		},
	} {
		t.Run(tc.what, func(t *testing.T) {
			c, ok := NewExternalCmd(errors.New(tc.msg), []string{})
			if ok {
				t.Fatalf("Error did not occur: %#v", c)
			}
		})
	}
}

func TestRunExternalCommandOK(t *testing.T) {
	bindir := testExternalCommandBinaryDir("test", t)
	tmp := tmpenv.New("PATH")
	defer tmp.Restore()

	panicIfErr(os.Setenv("PATH", os.Getenv("PATH")+string(os.PathListSeparator)+bindir))

	args := []string{"-A", "external-test", "--foo", "xxx", "-b"}
	cmd, ok := NewExternalCmd(errors.New(`expected command but got "external-test"`), args)
	if !ok {
		t.Fatal("Subcommand was not found")
	}

	fake := fakeio.Stdout().Stderr()
	defer fake.Restore()

	if err := cmd.Do(); err != nil {
		t.Fatal(err)
	}

	out, err := fake.String()
	panicIfErr(err)

	if !strings.Contains(out, "Output from stdout") {
		t.Fatal("Output to stdout is unexpected:", out)
	}

	if !strings.Contains(out, "Output from stderr") {
		t.Fatal("Output to stderr is unexpected:", out)
	}

	// First argument is a executable path of `notes`. Ignore it by strings.HasSuffix()
	exe, err := os.Executable()
	panicIfErr(err)
	want := fmt.Sprintln(append([]string{exe}, args...))
	if !strings.Contains(out, want) {
		t.Fatal("Passed arguments to external command is unexpected. Wanted", want, "in output but have output", out)
	}
}

func TestRunExternalCommandExitFailure(t *testing.T) {
	bindir := testExternalCommandBinaryDir("error", t)
	tmp := tmpenv.New("PATH")
	defer tmp.Restore()

	panicIfErr(os.Setenv("PATH", os.Getenv("PATH")+string(os.PathListSeparator)+bindir))

	cmd, ok := NewExternalCmd(errors.New(`expected command but got "external-error"`), []string{})
	if !ok {
		t.Fatal("Subcommand was not found")
	}

	err := cmd.Do()
	if err == nil {
		t.Fatal("Error did not occur")
	}
	exe := "notes-external-error"
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	want := fmt.Sprintf("External command '%s' did not exit successfully", exe)
	if !strings.Contains(err.Error(), want) {
		t.Fatal("Unexpected error:", err)
	}
}
