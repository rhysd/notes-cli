package notes

import (
	"fmt"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
)

type NewCmd struct {
	cli      *kingpin.CmdClause
	Config   *Config
	Category string
	Filename string
	Tags     string
	NoInline bool
}

func (cmd *NewCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("new", "Create a new note")
	cmd.cli.Arg("category", "Category of memo").Required().StringVar(&cmd.Category)
	cmd.cli.Arg("filename", "Name of memo").Required().StringVar(&cmd.Filename)
	cmd.cli.Arg("tags", "Comma-separated tags of memo").StringVar(&cmd.Tags)
	cmd.cli.Flag("no-inline-input", "Does not request inline input even if no editor is set").BoolVar(&cmd.NoInline)
}

func (cmd *NewCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline
}

func (cmd *NewCmd) fallbackInput(note *Note) error {
	fmt.Fprintln(os.Stderr, "Input notes inline (Ctrl+D to stop):")
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return errors.Wrap(err, "Cannot read from stdin")
	}

	f, err := os.OpenFile(note.FilePath(), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(err, "Cannot open note file")
	}
	if _, err := f.Write(b); err != nil {
		return errors.Wrap(err, "Cannot write to note file")
	}

	fmt.Fprintln(os.Stderr)
	fmt.Println(note.FilePath())

	return nil
}

func (cmd *NewCmd) Do() error {
	git := NewGit(cmd.Config)

	if git != nil {
		if err := git.Init(); err != nil {
			return err
		}
	}

	note, err := NewNote(cmd.Category, cmd.Tags, cmd.Filename, "", cmd.Config)
	if err != nil {
		return err
	}

	if err := note.Create(); err != nil {
		return err
	}

	if err := note.Open(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		fd := os.Stdin.Fd()
		if !cmd.NoInline && (isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)) {
			return cmd.fallbackInput(note)
		}
		// Final fallback is only showing the path to the note. Then users can open it by themselves.
		fmt.Println(note.FilePath())
	}

	return nil
}
