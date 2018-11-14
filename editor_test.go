package notes

import (
	"github.com/kballard/go-shellquote"
	"github.com/rhysd/go-fakeio"
	"os/exec"
	"strings"
	"testing"
)

func TestOpenEditor(t *testing.T) {
	echo, err := exec.LookPath("echo")
	panicIfErr(err)

	for _, tc := range []struct {
		what   string
		editor string
		want   string
	}{
		{
			what:   "executable",
			editor: "echo",
			want:   "foo bar\n",
		},
		{
			what:   "executable path",
			editor: shellquote.Join(echo), // On Windows, path would contain 'Program Files'
			want:   "foo bar\n",
		},
		{
			what:   "multiple args",
			editor: "echo 'aaa' bbb",
			want:   "aaa bbb foo bar\n",
		},
		{
			what:   "arg contains white space",
			editor: "echo 'aaa bbb' 'ccc\tddd'",
			want:   "aaa bbb ccc\tddd foo bar\n",
		},
		{
			what:   "arg contains multiple white space",
			editor: "echo     aaa               bbb   ",
			want:   "aaa bbb foo bar\n",
		},
	} {
		t.Run(tc.what, func(t *testing.T) {
			fake := fakeio.Stdout()
			defer fake.Restore()

			cfg := &Config{
				EditorCmd: tc.editor,
				HomePath:  ".",
			}

			if err := openEditor(cfg, "foo", "bar"); err != nil {
				t.Fatal(err)
			}

			have, err := fake.String()
			panicIfErr(err)

			if have != tc.want {
				t.Fatalf("Output from %q is unexpected. Wanted %q but have %q", tc.editor, tc.want, have)
			}
		})
	}
}

func TestEditorInvalidEditor(t *testing.T) {
	exe, err := exec.LookPath("false")
	panicIfErr(err)
	exe = shellquote.Join(exe)

	for _, tc := range []struct {
		what  string
		value string
		want  string
	}{
		{
			what:  "empty",
			value: "",
			want:  "Editor is not set",
		},
		{
			what:  "empty command",
			value: "''",
			want:  "did not exit successfully",
		},
		{
			what:  "unterminated single quote",
			value: "vim '-g",
			want:  "Cannot parse editor command line",
		},
		{
			what:  "unterminated double quote",
			value: "vim \"-g",
			want:  "Cannot parse editor command line",
		},
		{
			what:  "executable exit failure",
			value: exe,
			want:  "did not exit successfully",
		},
		{
			what:  "executable does not exist",
			value: "unknown-command-which-does-not-exist",
			want:  "did not exit successfully",
		},
	} {
		t.Run(tc.want, func(t *testing.T) {
			cfg := &Config{EditorCmd: tc.value}
			err := openEditor(cfg)
			if err == nil {
				t.Fatal("Error did not occur")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatal("Unexpected error:", err)
			}
		})
	}
}
