package notes

import (
	"fmt"
	"github.com/blang/semver"
	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/rhysd/go-fakeio"
	"github.com/rhysd/go-tmpenv"
	"os"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	if _, err := semver.Parse(Version); err != nil {
		t.Fatal(err)
	}
}

func TestParseArgs(t *testing.T) {
	opts := []cmp.Option{
		cmpopts.IgnoreUnexported(
			CategoriesCmd{},
			ConfigCmd{},
			ListCmd{},
			NewCmd{},
			SaveCmd{},
			TagsCmd{},
			SelfupdateCmd{},
		),
		cmpopts.IgnoreTypes(&Config{}),
		cmpopts.IgnoreFields(ListCmd{}, "Out"),
		cmpopts.IgnoreFields(ConfigCmd{}, "Out"),
		cmpopts.IgnoreFields(TagsCmd{}, "Out"),
		cmpopts.IgnoreFields(CategoriesCmd{}, "Out"),
		cmpopts.IgnoreFields(SelfupdateCmd{}, "Out"),
	}

	for _, tc := range []struct {
		args []string
		want parsableCmd
	}{
		{
			args: []string{"config", "home"},
			want: &ConfigCmd{
				Name: "home",
			},
		},
		{
			args: []string{"save", "--message", "hello"},
			want: &SaveCmd{
				Message: "hello",
			},
		},
		{
			args: []string{"tags", "dog"},
			want: &TagsCmd{
				Category: "dog",
			},
		},
		{
			args: []string{"categories"},
			want: &CategoriesCmd{},
		},
		{
			args: []string{"cats"},
			want: &CategoriesCmd{},
		},
		{
			args: []string{"list", "--category", "dog", "--tag", "cat", "--oneline", "--edit"},
			want: &ListCmd{
				Category: "dog",
				Tag:      "cat",
				Oneline:  true,
				Edit:     true,
			},
		},
		{
			args: []string{"ls", "--category", "dog", "--tag", "cat", "--oneline", "--edit"},
			want: &ListCmd{
				Category: "dog",
				Tag:      "cat",
				Oneline:  true,
				Edit:     true,
			},
		},
		{
			args: []string{"new", "dog", "filename", "cat,bird", "--no-inline-input"},
			want: &NewCmd{
				Category: "dog",
				Tags:     "cat,bird",
				NoInline: true,
				Filename: "filename",
			},
		},
		{
			args: []string{"selfupdate", "--dry"},
			want: &SelfupdateCmd{
				Dry: true,
			},
		},
	} {
		t.Run(tc.args[0]+" command", func(t *testing.T) {
			have, err := ParseCmd(tc.args)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(tc.want, have, opts...) {
				t.Fatal(cmp.Diff(tc.want, have, opts...))
			}
		})
	}
}

func TestParseGlobalColorFlags(t *testing.T) {
	old := color.NoColor
	defer func() {
		color.NoColor = old
	}()

	for _, tc := range []struct {
		arg  string
		want bool
	}{
		{
			arg:  "--no-color",
			want: true,
		},
		{
			arg:  "--color-always",
			want: false,
		},
		{
			arg:  "-A",
			want: false,
		},
	} {
		t.Run(tc.arg, func(t *testing.T) {
			if _, err := ParseCmd([]string{tc.arg, "config"}); err != nil {
				t.Fatal(err)
			}
			if color.NoColor != tc.want {
				t.Fatal("Color config is unexpected:", color.NoColor, "wanted", tc.want)
			}
		})
	}
}

func TestParseFailure(t *testing.T) {
	_, err := ParseCmd([]string{"--unknown-flag"})
	if err == nil {
		t.Fatal("Unknown command did not cause an error")
	}
	if !strings.Contains(err.Error(), "unknown long flag") {
		t.Fatal("Unexpected error:", err)
	}
}

func TestExternalCommand(t *testing.T) {
	bindir := testExternalCommandBinaryDir("test", t)
	tmp := tmpenv.New("PATH")
	defer tmp.Restore()

	panicIfErr(os.Setenv("PATH", os.Getenv("PATH")+string(os.PathListSeparator)+bindir))

	for _, tc := range []struct {
		what string
		args []string
	}{
		{
			what: "no arg",
			args: []string{"external-test"},
		},
		{
			what: "with args",
			args: []string{"external-test", "--foo", "xxx", "-b"},
		},
		{
			what: "with global args",
			args: []string{"--no-color", "external-test", "--foo", "xxx", "-b"},
		},
	} {
		t.Run(tc.what, func(t *testing.T) {
			cmd, err := ParseCmd(tc.args)
			if err != nil {
				t.Fatal(err)
			}

			ext, ok := cmd.(*ExternalCmd)
			if !ok {
				t.Fatalf("Did not resolve to external command: %#v", cmd)
			}

			fake := fakeio.Stdout().Stderr()
			defer fake.Restore()

			if err := ext.Do(); err != nil {
				t.Fatal(err)
			}

			output, err := fake.String()
			panicIfErr(err)

			if !strings.Contains(output, "Output from stdout") {
				t.Fatal("Output to stdout is unexpected:", output)
			}

			if !strings.Contains(output, "Output from stderr") {
				t.Fatal("Output to stderr is unexpected:", output)
			}

			want := fmt.Sprintln(tc.args)
			if !strings.Contains(output, want) {
				t.Fatal("Passed arguments to external command is unexpected. Wanted", want, "in output but have output", output)
			}
		})
	}
}
