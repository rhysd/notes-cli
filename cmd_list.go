package notes

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"regexp"
	"strings"
	"time"
)

var (
	yellow = color.New(color.FgYellow)
	bold   = color.New(color.Bold)
	green  = color.New(color.FgGreen)
)

// ListCmd represents `notes list` command. Each public fields represent options of the command
// Out field represents where this command should output.
type ListCmd struct {
	cli, cliAlias *kingpin.CmdClause
	Config        *Config
	// Full is a flag equivalent to --full
	Full bool
	// Category is a regex string equivalent to --cateogry
	Category string
	// Tag is a regex string equivalent to --tag
	Tag string
	// Relative is a flag equivalent to --relative
	Relative bool
	// Oneline is a flag equivalent to --oneline
	Oneline bool
	// Tag is a string indicating how to sort the list. This value is equivalent to --sort option
	SortBy string
	// Edit is a flag equivalent to --edit
	Edit bool
	// Out is a writer to write output of this command. Kind of stdout is expected
	Out io.Writer
}

func (cmd *ListCmd) defineListCLI(c *kingpin.CmdClause) {
	c.Flag("full", "Show list of full information of note (full path, metadata, title, body) instead of file path").Short('f').BoolVar(&cmd.Full)
	c.Flag("category", "Filter list by category name with regular expression").Short('c').StringVar(&cmd.Category)
	c.Flag("tag", "Filter list by tag name with regular expression").Short('t').StringVar(&cmd.Tag)
	c.Flag("relative", "Show relative paths from $NOTES_CLI_HOME directory").Short('r').BoolVar(&cmd.Relative)
	c.Flag("oneline", "Show oneline information of note (relative path, category, tags, title) instead of file path").Short('o').BoolVar(&cmd.Oneline)
	c.Flag("sort", "Sort list by 'modified', 'created', 'filename' or 'category'. Default is 'created'").Short('s').EnumVar(&cmd.SortBy, "modified", "created", "filename", "category")
	c.Flag("edit", "Open listed notes with your favorite editor. $NOTES_CLI_EDITOR must be set. Paths of listed notes are passed to the editor command's arguments").Short('e').BoolVar(&cmd.Edit)
}

func (cmd *ListCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("list", "List notes with filtering by categories and/or tags with regular expressions. By default, it shows full path of notes (alias: ls)")
	cmd.defineListCLI(cmd.cli)
	cmd.cliAlias = app.Command("ls", "List notes with filtering by categories and/or tags with regular expressions. By default, it shows full path of notes ").Hidden()
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

	out := bufio.NewWriter(cmd.Out)
	for _, data := range table {
		for i := 0; i < lenCols; i++ {
			last := i == lenCols-1
			c := colors[i]
			d := data[i]
			max := maxLen[i]
			pad := strings.Repeat(" ", max-len(d))

			sep := " "
			if last {
				pad = ""
				sep = "\n"
			}

			s := d + pad + sep

			// No need to check errors returned from Fprint since the error will be caughted
			// at out.Flush() call.
			if c == nil {
				fmt.Fprint(out, s)
			} else {
				c.Fprint(out, s)
			}
		}
	}

	return out.Flush()
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

func (cmd *ListCmd) printNotes(notes []*Note) error {
	switch strings.ToLower(cmd.SortBy) {
	case "filename":
		sortByFilename(notes)
	case "category":
		sortByCategory(notes)
	case "modified":
		if err := sortByModified(notes); err != nil {
			return err
		}
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

	if cmd.Edit {
		args := make([]string, 0, len(notes))
		for _, n := range notes {
			args = append(args, n.FilePath())
		}
		return openEditor(cmd.Config, args...)
	}

	var b bytes.Buffer
	for _, note := range notes {
		if cmd.Relative {
			b.WriteString(note.RelFilePath())
		} else {
			b.WriteString(note.FilePath())
		}
		b.WriteRune('\n')
	}
	_, err := cmd.Out.Write(b.Bytes())
	return err
}

// Do runs `notes list` command and returns an error if occurs
func (cmd *ListCmd) Do() error {
	cats, err := CollectCategories(cmd.Config)
	if err != nil {
		return err
	}

	var catReg *regexp.Regexp
	if cmd.Category != "" {
		if catReg, err = regexp.Compile(cmd.Category); err != nil {
			return errors.Wrap(err, "Regular expression for filtering categories is invalid")
		}
	}

	numNotes := 0
	for n, c := range cats {
		if catReg != nil && !catReg.MatchString(n) {
			delete(cats, n)
			continue
		}
		numNotes += len(c.NotePaths)
	}

	var tagReg *regexp.Regexp
	if cmd.Tag != "" {
		if tagReg, err = regexp.Compile(cmd.Tag); err != nil {
			return errors.Wrap(err, "Regular expression for filtering tags is invalid")
		}
	}

	notes := make([]*Note, 0, numNotes)
	for _, cat := range cats {
		for _, p := range cat.NotePaths {
			note, err := LoadNote(p, cmd.Config)
			if err != nil {
				return err
			}
			if tagReg == nil {
				notes = append(notes, note)
				continue
			}
			for _, tag := range note.Tags {
				if tagReg.MatchString(tag) {
					notes = append(notes, note)
					break
				}
			}
			// When no tag is matched to tag regex, the note is ignored
		}
	}

	if len(notes) == 0 {
		return nil
	}
	return cmd.printNotes(notes)
}
