package notes

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"io/ioutil"
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
	fs, err := ioutil.ReadDir(cmd.Config.HomePath)
	if err != nil {
		return errors.Wrap(err, "Cannot read notes-cli home")
	}

	cats := []string{}

	for _, f := range fs {
		n := f.Name()
		if !f.IsDir() || strings.HasPrefix(n, ".") {
			continue
		}
		cats = append(cats, n)
	}

	sort.Strings(cats)

	_, err = fmt.Fprintln(cmd.Out, strings.Join(cats, "\n"))
	return err
}
