package notes

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"strings"
)

type ConfigCmd struct {
	cli    *kingpin.CmdClause
	Config *Config
	Name   string
}

func (cmd *ConfigCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("config", "Output config value to stdout")
	cmd.cli.Arg("name", "One of 'home', 'git', 'editor'").StringVar(&cmd.Name)
}

func (cmd *ConfigCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline
}

func (cmd *ConfigCmd) Do() error {
	switch strings.ToLower(cmd.Name) {
	case "":
		fmt.Printf(
			"EDITOR=%s\nGIT=%s\nEDITOR=%s\n",
			cmd.Config.HomePath,
			cmd.Config.GitPath,
			cmd.Config.EditorPath,
		)
	case "home":
		fmt.Println(cmd.Config.HomePath)
	case "git":
		fmt.Println(cmd.Config.GitPath)
	case "editor":
		fmt.Println(cmd.Config.EditorPath)
	default:
		return errors.Errorf("Unknown config name '%s'", cmd.Name)
	}
	return nil
}
