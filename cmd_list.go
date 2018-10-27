package notes

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	yellow = color.New(color.FgYellow)
	bold   = color.New(color.Bold)
)

type ListCmd struct {
	cli, cliAlias *kingpin.CmdClause
	Config        *Config
	Full          bool
	Category      string
	Tag           string
	Out           io.Writer
}

func (cmd *ListCmd) defineListCLI(c *kingpin.CmdClause) {
	c.Flag("full", "Show full information of note instead of path").Short('f').BoolVar(&cmd.Full)
	c.Flag("category", "Filter category name by regular expression").Short('c').StringVar(&cmd.Category)
	c.Flag("tag", "Filter tag name by regular expression").Short('t').StringVar(&cmd.Tag)
}

func (cmd *ListCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("list", "List note paths with filtering by categories and/or tags with regular expressions")
	cmd.defineListCLI(cmd.cli)
	cmd.cliAlias = app.Command("ls", "List note paths with filtering by categories and/or tags with regular expressions").Hidden()
	cmd.defineListCLI(cmd.cliAlias)
}

func (cmd *ListCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline || cmd.cliAlias.FullCommand() == cmdline
}

func (cmd *ListCmd) printNoteFull(note *Note) {
	yellow.Fprintln(cmd.Out, note.FilePath())
	yellow.Fprint(cmd.Out, "Category: ")
	fmt.Fprintln(cmd.Out, note.Category)
	yellow.Fprint(cmd.Out, "Tags: ")
	fmt.Fprintln(cmd.Out, note.Tags)
	yellow.Fprint(cmd.Out, "Created: ")
	fmt.Fprintln(cmd.Out, note.Created.Format(time.RFC3339))
	bold.Fprintf(cmd.Out, "\n%s\n\n", note.Title)
}

func (cmd *ListCmd) doCategories(cats []string) error {
	var r *regexp.Regexp
	if cmd.Tag != "" {
		var err error
		if r, err = regexp.Compile(cmd.Tag); err != nil {
			return errors.Wrap(err, "Regular expression for filtering tags is invalid")
		}
	}

	notes := []*Note{}
	for _, cat := range cats {
		dir := filepath.Join(cmd.Config.HomePath, cat)
		fs, err := ioutil.ReadDir(dir)
		if err != nil {
			return errors.Wrapf(err, "Cannot read category directory for '%s'", cat)
		}
		for _, f := range fs {
			n := f.Name()
			if f.IsDir() || !strings.HasSuffix(n, ".md") {
				continue
			}
			note, err := LoadNote(filepath.Join(dir, n), cmd.Config)
			if err != nil {
				return err
			}
			if r == nil {
				notes = append(notes, note)
				continue
			}
			for _, tag := range note.Tags {
				if r.MatchString(tag) {
					notes = append(notes, note)
					break
				}
			}
		}
	}

	if !cmd.Full {
		var b bytes.Buffer
		for _, note := range notes {
			b.WriteString(note.FilePath() + "\n")
		}
		_, err := cmd.Out.Write(b.Bytes())
		return err
	}

	for _, note := range notes {
		cmd.printNoteFull(note)
	}

	return nil
}

func (cmd *ListCmd) Do() error {
	fs, err := ioutil.ReadDir(cmd.Config.HomePath)
	if err != nil {
		return errors.Wrap(err, "Cannot read note-cli home")
	}

	if cmd.Category == "" {
		cs := make([]string, 0, len(fs))
		for _, f := range fs {
			if f.IsDir() {
				cs = append(cs, f.Name())
			}
		}
		return cmd.doCategories(cs)
	}

	r, err := regexp.Compile(cmd.Category)
	if err != nil {
		return errors.Wrap(err, "Regular expression for filtering categories is invalid")
	}

	cs := make([]string, 0, len(fs))
	for _, f := range fs {
		if !f.IsDir() {
			continue
		}
		n := f.Name()
		if r.MatchString(n) {
			cs = append(cs, f.Name())
		}
	}
	return cmd.doCategories(cs)
}
