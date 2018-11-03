package notes

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

// TagsCmd represents `notes tags` command. Each public fields represent options of the command
// Out field represents where this command should output.
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

// Do runs `notes tags` command and returns an error if occurs
func (cmd *TagsCmd) Do() error {
	saw := map[string]struct{}{}
	tags := []string{}

	// If cmd.Category is empty, it scans all notes in home
	if err := WalkNotes(cmd.Category, cmd.Config, func(path string, note *Note) error {
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
