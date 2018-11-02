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
	green  = color.New(color.FgGreen)
)

type ListCmd struct {
	cli, cliAlias *kingpin.CmdClause
	Config        *Config
	Full          bool
	Category      string
	Tag           string
	Relative      bool
	Oneline       bool
	SortBy        string
	Out           io.Writer
}

func (cmd *ListCmd) defineListCLI(c *kingpin.CmdClause) {
	c.Flag("full", "Show full information of note instead of path").Short('f').BoolVar(&cmd.Full)
	c.Flag("category", "Filter category name by regular expression").Short('c').StringVar(&cmd.Category)
	c.Flag("tag", "Filter tag name by regular expression").Short('t').StringVar(&cmd.Tag)
	c.Flag("relative", "Show relative paths from $NOTES_CLI_HOME directory").Short('r').BoolVar(&cmd.Relative)
	c.Flag("oneline", "Show oneline information of note instead of path").Short('o').BoolVar(&cmd.Oneline)
	c.Flag("sort", "Sort results by 'created', 'filename' or 'category'. 'created' is default").Short('s').EnumVar(&cmd.SortBy, "created", "filename", "category")
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
	green.Fprintln(cmd.Out, note.FilePath())
	yellow.Fprint(cmd.Out, "Category: ")
	fmt.Fprintln(cmd.Out, note.Category)
	yellow.Fprint(cmd.Out, "Tags:     ")
	fmt.Fprintln(cmd.Out, strings.Join(note.Tags, ", "))
	yellow.Fprint(cmd.Out, "Created:  ")
	fmt.Fprintln(cmd.Out, note.Created.Format(time.RFC3339))
	if note.Title != "" {
		bold.Fprintf(cmd.Out, "\n%s\n%s\n\n", note.Title, strings.Repeat("=", len(note.Title)))
	}

	body, err := note.ReadBodyN(200)
	if err != nil || len(body) == 0 {
		return
	}

	fmt.Fprint(cmd.Out, body)

	// Adjust end of body. Ensure it ends with \n\n
	if !strings.HasSuffix(body, "\n") {
		fmt.Fprint(cmd.Out, "\n\n")
	} else if !strings.HasSuffix(body, "\n\n") {
		fmt.Fprintln(cmd.Out)
	}
}

func (cmd *ListCmd) writeTable(colors []*color.Color, table [][]string) error {
	if len(table) == 0 {
		return nil
	}

	lenCols := len(colors)

	maxLen := make([]int, lenCols)
	for i := 0; i < lenCols; i++ {
		max := len(table[0][i])
		for _, d := range table[1:] {
			l := len(d[i])
			if l > max {
				max = l
			}
		}
		maxLen[i] = max
	}

	for _, data := range table {
		for i := 0; i < lenCols; i++ {
			last := i == lenCols-1
			c := colors[i]
			d := data[i]
			max := maxLen[i]
			pad := strings.Repeat(" ", max-len(d))

			sep := " "
			if last {
				sep = "\n"
			}

			s := d + pad + sep

			var err error
			if c == nil {
				_, err = fmt.Fprint(cmd.Out, s)
			} else {
				_, err = c.Fprint(cmd.Out, s)
			}
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (cmd *ListCmd) printOnelineNotes(notes []*Note) error {
	colors := []*color.Color{bold, yellow, green, nil}
	data := make([][]string, 0, len(notes))

	for _, note := range notes {
		title := note.Title
		if len(title) > 73 {
			title = title[:70] + "..."
		}

		data = append(data, []string{
			note.RelFilePath(),
			note.Category,
			strings.Join(note.Tags, ","),
			title,
		})
	}

	return cmd.writeTable(colors, data)
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

	switch strings.ToLower(cmd.SortBy) {
	case "filename":
		sortByFilename(notes)
	case "category":
		sortByCategory(notes)
	default:
		sortByCreated(notes)
	}

	if cmd.Full {
		for _, note := range notes {
			cmd.printNoteFull(note)
		}
		return nil
	}

	if cmd.Oneline {
		return cmd.printOnelineNotes(notes)
	}

	var b bytes.Buffer
	for _, note := range notes {
		if cmd.Relative {
			b.WriteString(note.RelFilePath() + "\n")
		} else {
			b.WriteString(note.FilePath() + "\n")
		}
	}
	_, err := cmd.Out.Write(b.Bytes())
	return err
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
