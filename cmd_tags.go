package notes

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"sort"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

// TagsCmd represents `notes tags` command. Each public fields represent options of the command
// Out field represents where this command should output.
type TagsCmd struct {
	cli    *kingpin.CmdClause
	Config *Config
	// Category is a category name of tags. If this value is empty, tags of all categories will be output
	Category string
	// Out is a writer to write output of this command. Kind of stdout is expected
	Out io.Writer
}

func (cmd *TagsCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("tags", "List all tags")
	cmd.cli.Arg("category", "Show tags of specified category. If not specified, all tags are output").StringVar(&cmd.Category)
}

func (cmd *TagsCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline
}

// Do runs `notes tags` command and returns an error if occurs
func (cmd *TagsCmd) Do() error {
	saw := map[string]struct{}{}
	tags := []string{}

	cats, err := CollectCategories(cmd.Config, 0)
	if err != nil {
		return err
	}

	if cmd.Category != "" {
		// Even if category is specified, we fetch all categories since error message requires
		// all category names for suggestion.
		cat, ok := cats[cmd.Category]
		if !ok {
			ns := cats.Names()
			return errors.Errorf("Category '%s' does not exist. All categories are %s", cmd.Category, strings.Join(ns, ", "))
		}
		cats = Categories{cmd.Category: cat}
	}

	for _, cat := range cats {
		notes, err := cat.Notes(cmd.Config)
		if err != nil {
			return err
		}
		for _, n := range notes {
			for _, tag := range n.Tags {
				if _, ok := saw[tag]; !ok {
					tags = append(tags, tag)
					saw[tag] = struct{}{}
				}
			}
		}
	}

	sort.Strings(tags)

	_, err = fmt.Fprintln(cmd.Out, strings.Join(tags, "\n"))
	return err
}
