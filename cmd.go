package notes

import (
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

// parsableCmd is an interface for subcommands of notes command parsed from command line arguments
type parsableCmd interface {
	Do() error
	defineCLI(*kingpin.Application)
	matchesCmdline(string) bool
}

// Cmd is an interface for subcommands of notes command
type Cmd interface {
	Do() error
}

// Version is version string of notes command. It conforms semantic versioning
var Version = "1.5.0"
var description = `Simple note taking tool for command line with your favorite editor.

You can manage (create/open/list) notes via this tool on terminal. notes also
optionally can save your notes thanks to Git to avoid losing your notes.

notes is intended to be used nicely with other commands such as grep (or ag, rg),
rm, filtering tools such as fzf or peco and editors which can be started from
command line.

notes is developed at https://github.com/rhysd/notes-cli. If you're seeing a bug or having a feature request,
please create a new issue. Pull requests are more than welcome.`

// ParseCmd parses given arguments as command line options and returns corresponding subcommand instance.
// When no subcommand matches or argus contains invalid argument, it returns an error
func ParseCmd(args []string) (Cmd, error) {
	cli := kingpin.New("notes", description)
	noColor := cli.Flag("no-color", "Disable color output").Bool()
	colorAlways := cli.Flag("color-always", "Enable color output always").Short('A').Bool()

	cli.Version(Version)
	cli.Author("rhysd <https://github.com/rhysd>")
	cli.HelpFlag.Short('h')

	c, err := NewConfig()
	if err != nil {
		return nil, err
	}

	colorStdout := colorable.NewColorableStdout()

	cmds := []parsableCmd{
		&NewCmd{Config: c},
		&ListCmd{Config: c, Out: colorStdout},
		&CategoriesCmd{Config: c, Out: os.Stdout},
		&TagsCmd{Config: c, Out: os.Stdout},
		&SaveCmd{Config: c},
		&ConfigCmd{Config: c, Out: os.Stdout},
		&SelfupdateCmd{Out: colorStdout},
	}

	for _, cmd := range cmds {
		cmd.defineCLI(cli)
	}

	parsed, err := cli.Parse(args)
	if err != nil {
		if ext, ok := NewExternalCmd(err, args); ok {
			return ext, nil
		}
		return nil, err
	}

	if *colorAlways {
		color.NoColor = false
	}

	if *noColor {
		color.NoColor = true
	}

	for _, cmd := range cmds {
		if cmd.matchesCmdline(parsed) {
			return cmd, nil
		}
	}

	panic("FATAL: Unknown command: " + parsed)
}
