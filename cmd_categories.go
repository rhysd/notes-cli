package notes

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"sort"
	"strings"
)

// CategoriesCmd represents `notes categories` command. Each public fields represent options of the command.
// Out field represents where this command should output.
type CategoriesCmd struct {
	cli, cliAlias *kingpin.CmdClause
	Config        *Config
	// Out is a writer to write output of this command. Kind of stdout is expected
	Out io.Writer
}

func (cmd *CategoriesCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("categories", "List all categories to stdout (alias: cats)")
	cmd.cliAlias = app.Command("cats", "List all categories to stdout. Please do not expect 🐱!").Hidden()
}

func (cmd *CategoriesCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline || cmd.cliAlias.FullCommand() == cmdline
}

// Do runs `notes categories` command and returns an error if occurs
func (cmd *CategoriesCmd) Do() error {
	cats, err := CollectCategories(cmd.Config, 0)
	if err != nil {
		return err
	}

	names := make([]string, 0, len(cats))
	for c := range cats {
		names = append(names, c)
	}

	sort.Strings(names)

	_, err = fmt.Fprintln(cmd.Out, strings.Join(names, "\n"))
	return err
}
