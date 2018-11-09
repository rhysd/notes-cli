package notes

import (
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

// Cmd is an interface for subcommands of notes command
type Cmd interface {
	Do() error
	defineCLI(*kingpin.Application)
	matchesCmdline(string) bool
}

// Version is version string of notes command. It conforms semantic versioning
var Version = "1.2.0"

// ParseCmd parses given arguments as command line options and returns corresponding subcommand instance.
// When no subcommand matches or argus contains invalid argument, it returns an error
func ParseCmd(args []string) (Cmd, error) {
	cli := kingpin.New("notes", "Simple note taking tool for command line with your favorite editor")
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

	cmds := []Cmd{
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
