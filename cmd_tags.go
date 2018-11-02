package notes

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

type TagsCmd struct {
	cli      *kingpin.CmdClause
	Config   *Config
	Category string
	Out      io.Writer
}

func (cmd *TagsCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("tags", "List all tags")
	cmd.cli.Arg("category", "Show tags of specified category. If not specified, all tags are output").StringVar(&cmd.Category)
}

func (cmd *TagsCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline
}

func (cmd *TagsCmd) Do() error {
	saw := map[string]struct{}{}
	tags := []string{}

	// If cmd.Category is empty, it scans all notes in home
	if err := WalkNotesNew(cmd.Category, cmd.Config, func(path string, note *Note) error {
		for _, tag := range note.Tags {
			if _, ok := saw[tag]; !ok {
				tags = append(tags, tag)
				saw[tag] = struct{}{}
			}
		}
		return nil
	}); err != nil {
		return err
	}

	sort.Strings(tags)

	_, err := fmt.Fprintln(cmd.Out, strings.Join(tags, "\n"))
	return err
}
