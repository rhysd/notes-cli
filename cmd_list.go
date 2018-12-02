package notes

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
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

// ListCmd represents `notes list` command. Each public fields represent options of the command
// Out field represents where this command should output.
type ListCmd struct {
	cli, cliAlias *kingpin.CmdClause
	out           io.Writer
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
	c.Flag("full", "Show list of full information of note (full path, metadata, title, body (up to 10 lines)) instead of file path").Short('f').BoolVar(&cmd.Full)
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

func (cmd *ListCmd) printNoteFullTo(out *bufio.Writer, note *Note) {
	green.Fprintln(out, note.FilePath())
	yellow.Fprint(out, "Category: ")
	fmt.Fprintln(out, note.Category)
	yellow.Fprint(out, "Tags:     ")
	fmt.Fprintln(out, strings.Join(note.Tags, ", "))
	yellow.Fprint(out, "Created:  ")
	fmt.Fprintln(out, note.Created.Format(time.RFC3339))
	if note.Title != "" {
		bold.Fprintf(out, "\n%s\n%s\n\n", note.Title, strings.Repeat("=", runewidth.StringWidth(note.Title)))
	}

	body, size, err := note.ReadBodyLines(10)
	if err != nil || len(body) == 0 {
		return
	}

	fmt.Fprint(out, body)

	// Ensure body ends with newline
	if !strings.HasSuffix(body, "\n") {
		fmt.Fprintln(out)
	}

	// Body text was truncated. To indicate it, add ellipsis at the end
	if size == 10 {
		fmt.Fprintln(out, "...")
	}

	// Finally separate each note with blank line
	fmt.Fprintln(out)
}

func (cmd *ListCmd) printOnelineNotes(notes []*Note) error {
	tw := make([][2]int, len(notes))
	max := [2]int{}

	for i, note := range notes {
		tw[i][0] = runewidth.StringWidth(note.Category+note.File) + 1 // + 1 for separator
		tw[i][1] = runewidth.StringWidth(strings.Join(note.Tags, ","))
		for j := 0; j < 2; j++ {
			if tw[i][j] > max[j] {
				max[j] = tw[i][j]
			}
		}
	}

	out := bufio.NewWriter(cmd.out)
	for i, note := range notes {
		pad := strings.Repeat(" ", max[0]-tw[i][0]+1) // +1 for separator
		green.Fprint(out, filepath.FromSlash(note.Category))
		out.WriteRune(filepath.Separator)
		yellow.Fprint(out, note.File)
		out.WriteString(pad)

		pad = strings.Repeat(" ", max[1]-tw[i][1]+1) // +1 for separator
		bold.Fprint(out, strings.Join(note.Tags, ","))
		out.WriteString(pad)

		out.WriteString(note.Title)
		out.WriteRune('\n')
	}

	return out.Flush()
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
		out := bufio.NewWriter(cmd.out)
		for _, note := range notes {
			cmd.printNoteFullTo(out, note)
		}
		return out.Flush()
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
	if cmd.Relative {
		for _, note := range notes {
			b.WriteString(note.RelFilePath())
			b.WriteRune('\n')
		}
	} else {
		for _, note := range notes {
			b.WriteString(note.FilePath())
			b.WriteRune('\n')
		}
	}

	_, err := cmd.out.Write(b.Bytes())
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

	if cmd.Config.PagerCmd == "" {
		cmd.out = cmd.Out
		return cmd.printNotes(notes)
	}

	pager, err := StartPagerWriter(cmd.Config.PagerCmd, cmd.Out)
	if err != nil {
		return err
	}

	cmd.out = pager
	if err := cmd.printNotes(notes); err != nil {
		if pager.Err != nil {
			err = errors.Wrap(err, "Pager command did not run successfully")
		}
		return err
	}

	pager.Wait()
	return errors.Wrap(pager.Err, "Pager command did not run successfully")
}
