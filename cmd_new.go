package notes

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

// NewCmd represents `notes new` command. Each public fields represent options of the command
type NewCmd struct {
	cli    *kingpin.CmdClause
	Config *Config
	// Category is a category name of the new note. This must be a name allowed for directory name
	Category string
	// Filename is a file name of the new note
	Filename string
	// Tags is a comma-separated string of tags of the new note
	Tags string
	// NoInline is a flag equivalent to --no-inline-input
	NoInline bool
	// NoEdit is a flag equivalent to --no-edit
	NoEdit bool
}

func (cmd *NewCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("new", "Create a new note with given category and file name")
	cmd.cli.Arg("category", "Category of note. Note must belong to one category").Required().StringVar(&cmd.Category)
	cmd.cli.Arg("filename", "File name of note. It automatically adds '.md' file extension if omitted").Required().StringVar(&cmd.Filename)
	cmd.cli.Arg("tags", "Comma-separated tags of note. Zero or more tags can be specified to note").StringVar(&cmd.Tags)
	cmd.cli.Flag("no-inline-input", "Does not request inline input even if no editor command is set to $NOTES_CLI_EDITOR").BoolVar(&cmd.NoInline)
	cmd.cli.Flag("no-edit", "Does not open an editor even if an editor command is set to $NOTES_CLI_EDITOR").BoolVar(&cmd.NoEdit)
}

func (cmd *NewCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline
}

func (cmd *NewCmd) fallbackInput(note *Note) error {
	fmt.Fprintln(os.Stderr, "Input notes inline (Send EOF by Ctrl+D to stop):")
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return errors.Wrap(err, "Cannot read from stdin")
	}

	f, err := os.OpenFile(note.FilePath(), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(err, "Cannot open note file")
	}
	defer f.Close()
	if _, err := f.Write(b); err != nil {
		return errors.Wrap(err, "Cannot write to note file")
	}

	fmt.Fprintln(os.Stderr)
	fmt.Println(note.FilePath())

	return nil
}

// Do runs `notes new` command and returns an error if occurs
func (cmd *NewCmd) Do() error {
	git := NewGit(cmd.Config)

	note, err := NewNote(cmd.Category, cmd.Tags, cmd.Filename, "", cmd.Config)
	if err != nil {
		return err
	}

	if err := note.Create(); err != nil {
		return err
	}

	if git != nil {
		if err := git.Init(); err != nil {
			return err
		}
	}

	if cmd.NoEdit {
		// Falling back into only showing the path to the note. Then users can open it by themselves.
		fmt.Println(note.FilePath())
		return nil
	}

	if err := note.Open(); err != nil {
		if !cmd.NoInline {
			fmt.Fprintf(os.Stderr, "Note: %s\n", err)
		}
		if !cmd.NoInline {
			return cmd.fallbackInput(note)
		}
		// Final fallback is only showing the path to the note. Then users can open it by themselves.
		fmt.Println(note.FilePath())
	}

	return nil
}
