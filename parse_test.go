package notes

import (
	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"testing"
)

func TestParseArgs(t *testing.T) {
	opts := []cmp.Option{
		cmpopts.IgnoreUnexported(
			CategoriesCmd{},
			ConfigCmd{},
			ListCmd{},
			NewCmd{},
			SaveCmd{},
			TagsCmd{},
		),
		cmpopts.IgnoreTypes(&Config{}),
		cmpopts.IgnoreFields(ListCmd{}, "Out"),
	}

	for _, tc := range []struct {
		what string
		args []string
		want Cmd
	}{
		{
			what: "config command",
			args: []string{"config", "home"},
			want: &ConfigCmd{
				Name: "home",
			},
		},
		{
			what: "save command",
			args: []string{"save", "--message", "hello"},
			want: &SaveCmd{
				Message: "hello",
			},
		},
		{
			what: "tags command",
			args: []string{"tags", "dog"},
			want: &TagsCmd{
				Category: "dog",
			},
		},
		{
			what: "categories command",
			args: []string{"categories"},
			want: &CategoriesCmd{},
		},
		{
			what: "cats command",
			args: []string{"cats"},
			want: &CategoriesCmd{},
		},
		{
			what: "list command",
			args: []string{"list", "--category", "dog", "--tag", "cat", "--oneline"},
			want: &ListCmd{
				Category: "dog",
				Tag:      "cat",
				Oneline:  true,
			},
		},
		{
			what: "ls command",
			args: []string{"ls", "--category", "dog", "--tag", "cat", "--oneline"},
			want: &ListCmd{
				Category: "dog",
				Tag:      "cat",
				Oneline:  true,
			},
		},
		{
			what: "new command",
			args: []string{"new", "dog", "filename", "cat,bird", "--no-inline-input"},
			want: &NewCmd{
				Category: "dog",
				Tags:     "cat,bird",
				NoInline: true,
				Filename: "filename",
			},
		},
	} {
		t.Run(tc.what, func(t *testing.T) {
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

func TestParseGlobalFlags(t *testing.T) {
	old := color.NoColor
	defer func() {
		color.NoColor = old
	}()

	if _, err := ParseCmd([]string{"--no-color", "config"}); err != nil {
		t.Fatal(err)
	}

	if !color.NoColor {
		t.Fatal("Color was not disabled")
	}
}

func TestParseFailure(t *testing.T) {
	if _, err := ParseCmd([]string{"unknown-command"}); err == nil {
		t.Fatal("Unknown command did not cause an error")
	}
}
