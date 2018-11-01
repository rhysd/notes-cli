package notes

import (
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"gopkg.in/alecthomas/kingpin.v2"
)

func ParseCmd(args []string) (Cmd, error) {
	cli := kingpin.New("notes", "Simple note taking tool for command line with your favorite editor")
	noColor := cli.Flag("no-color", "Disable color output").Bool()

	cli.Version("0.2.1")
	cli.Author("rhysd <https://github.com/rhysd>")
	cli.HelpFlag.Short('h')

	c, err := NewConfig()
	if err != nil {
		return nil, err
	}

	out := colorable.NewColorableStdout()
	cmds := []Cmd{
		&NewCmd{Config: c},
		&ListCmd{Config: c, Out: out},
		&CategoriesCmd{Config: c},
		&TagsCmd{Config: c},
		&SaveCmd{Config: c},
		&ConfigCmd{Config: c},
	}

	for _, cmd := range cmds {
		cmd.defineCLI(cli)
	}

	parsed, err := cli.Parse(args)
	if err != nil {
		return nil, err
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
