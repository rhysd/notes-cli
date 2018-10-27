package notes

import (
	"bytes"
	"os"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"
)

type TagsCmd struct {
	cli      *kingpin.CmdClause
	Config   *Config
	Category string
}

func (cmd *TagsCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("tags", "List all tags")
	cmd.cli.Arg("category", "Show tags of specified category. If not specified, all tags are output").StringVar(&cmd.Category)
}

func (cmd *TagsCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline
}

func (cmd *TagsCmd) Do() error {
	var b bytes.Buffer
	saw := map[string]struct{}{}
	if err := WalkNotes(filepath.Join(cmd.Config.HomePath, cmd.Category), cmd.Config, func(path string, note *Note) error {
		for _, tag := range note.Tags {
			if _, ok := saw[tag]; !ok {
				b.WriteString(tag + "\n")
				saw[tag] = struct{}{}
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if _, err := os.Stdout.Write(b.Bytes()); err != nil {
		return err
	}
	return nil
}
