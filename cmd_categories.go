package notes

import (
	"bytes"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
	"strings"
)

type CategoriesCmd struct {
	cli, cliAlias *kingpin.CmdClause
	Config        *Config
}

func (cmd *CategoriesCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("categories", "List all categories")
	cmd.cliAlias = app.Command("cats", "List all categories. Please do not expect 🐱!").Hidden()
}

func (cmd *CategoriesCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline || cmd.cliAlias.FullCommand() == cmdline
}

func (cmd *CategoriesCmd) Do() error {
	fs, err := ioutil.ReadDir(cmd.Config.HomePath)
	if err != nil {
		return errors.Wrap(err, "Cannot read notes-cli home")
	}

	var b bytes.Buffer
	for _, f := range fs {
		n := f.Name()
		if !f.IsDir() || strings.HasPrefix(n, ".") {
			continue
		}
		b.WriteString(n + "\n")
	}
	if _, err := os.Stdout.Write(b.Bytes()); err != nil {
		return err
	}

	return nil
}
