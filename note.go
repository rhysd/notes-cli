package notes

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	reTitleBar = regexp.MustCompile("^=+$")
)

func canonPath(path string) string {
	u, err := user.Current()
	if err != nil {
		return path // Give up
	}
	if !strings.HasPrefix(path, u.HomeDir) {
		return path
	}
	canon := strings.TrimPrefix(path, u.HomeDir)
	sep := string(filepath.Separator)
	if !strings.HasPrefix(canon, sep) {
		canon = sep + canon
	}
	return "~" + canon
}

type Note struct {
	Config   *Config
	Category string
	Tags     []string
	Created  time.Time
	File     string
	Title    string
}

func (note *Note) DirPath() string {
	return filepath.Join(note.Config.HomePath, note.Category)
}

func (note *Note) FilePath() string {
	return filepath.Join(note.Config.HomePath, note.Category, note.File)
}

func (note *Note) RelFilePath() string {
	return filepath.Join(note.Category, note.File)
}

func (note *Note) Create() error {
	var b bytes.Buffer

	// Write title
	title := note.Title
	if title == "" {
		title = strings.TrimSuffix(note.File, filepath.Ext(note.File))
	}
	b.WriteString(title + "\n")
	b.WriteString(strings.Repeat("=", len(title)) + "\n")

	// Write metadata
	fmt.Fprintf(&b, "- Category: %s\n", note.Category)
	fmt.Fprintf(&b, "- Tags: %s\n", strings.Join(note.Tags, ", "))
	fmt.Fprintf(&b, "- Created: %s\n\n", note.Created.Format(time.RFC3339))

	d := note.DirPath()
	if err := os.MkdirAll(d, 0755); err != nil {
		return errors.Wrapf(err, "Could not create category directory '%s'", d)
	}

	p := filepath.Join(d, note.File)
	if _, err := os.Stat(p); err == nil {
		return errors.Errorf("Cannot create new note since file '%s' already exists. Please use 'edit' command to edit it", filepath.Join(note.Category, note.File))
	}

	return errors.Wrap(ioutil.WriteFile(p, b.Bytes(), 0644), "Cannot write note to file")
}

func (note *Note) Open() error {
	if note.Config.EditorPath == "" {
		return errors.New("Note: To open note in editor, please set $NOTES_CLI_EDITOR environment variable")
	}
	c := exec.Command(note.Config.EditorPath, note.FilePath())
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	c.Dir = note.DirPath()
	return errors.Wrap(c.Run(), "Editor command did not run successfully")
}

func (note *Note) ReadBodyLines(reader func(line string) (stop bool, err error)) error {
	path := note.FilePath()
	f, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "Cannot open note file")
	}
	defer f.Close()

	// Skip metadata
	s := bufio.NewScanner(f)
	sawCat, sawTags, sawCreated := false, false, false
	for s.Scan() {
		t := s.Text()
		if strings.HasPrefix(t, "- Category: ") {
			sawCat = true
		} else if strings.HasPrefix(t, "- Tags:") {
			sawTags = true
		} else if strings.HasPrefix(t, "- Created:") {
			sawCreated = true
		}
		if sawCat && sawTags && sawCreated {
			break
		}
	}
	if err := s.Err(); err != nil {
		return errors.Wrap(err, "Cannot read metadata of note")
	}
	if !sawCat || !sawTags || !sawCreated {
		return errors.Errorf("Some metadata is missing in %s", path)
	}

	for s.Scan() {
		t := s.Text()
		stop, err := reader(t)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	if err := s.Err(); err != nil {
		return errors.Wrap(err, "Error while reading lines of body of note")
	}

	return nil
}

func NewNote(cat, tags, file, title string, cfg *Config) (*Note, error) {
	if cat == "" {
		return nil, errors.New("Category cannot be empty")
	}
	if file == "" {
		return nil, errors.New("File name cannot be empty")
	}
	ts := []string{}
	for _, t := range strings.Split(tags, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			ts = append(ts, t)
		}
	}

	file = strings.Replace(file, " ", "-", -1)
	if !strings.HasSuffix(file, ".md") {
		file += ".md"
	}
	return &Note{cfg, cat, ts, time.Now(), file, ""}, nil
}

func LoadNote(path string, cfg *Config) (*Note, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot open note file")
	}
	defer f.Close()

	note := &Note{Config: cfg}

	p, md := filepath.Split(path)
	c := filepath.Base(p)
	note.File = md
	note.Category = c

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		// First line is title
		if note.Title == "" {
			if reTitleBar.MatchString(line) {
				note.Title = "(no title)"
			} else {
				note.Title = line
			}
		} else if strings.HasPrefix(line, "- Category: ") {
			if c := strings.TrimSpace(line[12:]); c != note.Category {
				return nil, errors.Errorf("Category does not match between file path and file content; in path '%s' v.s. in file '%s'", note.Category, c)
			}
		} else if strings.HasPrefix(line, "- Tags: ") {
			tags := strings.Split(strings.TrimSpace(line[8:]), ",")
			note.Tags = make([]string, 0, len(tags))
			for _, t := range tags {
				t = strings.TrimSpace(t)
				if t != "" {
					note.Tags = append(note.Tags, t)
				}
			}
		} else if strings.HasPrefix(line, "- Created: ") {
			t, err := time.Parse(time.RFC3339, strings.TrimSpace(line[11:]))
			if err != nil {
				return nil, errors.Wrapf(err, "Cannot parse created date time as RFC3339 format: %s", line)
			}
			note.Created = t
		}
	}
	if err := s.Err(); err != nil {
		return nil, errors.Wrapf(err, "Cannot read note file '%s'", canonPath(path))
	}

	if note.Title == "" {
		return nil, errors.Errorf("No title found in note '%s'. Didn't you use '====' bar for h1 title?", canonPath(path))
	}

	if note.Category == "" || note.Tags == nil || note.Created.IsZero() {
		return nil, errors.Errorf("Missing metadata in file '%s'. 'Category', 'Tags', 'Created' are mandatory", canonPath(path))
	}

	return note, nil
}

func WalkNotes(path string, cfg *Config, pred func(path string, note *Note) error) error {
	return errors.Wrap(
		filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() || !strings.HasSuffix(path, ".md") {
				// Skip
				return nil
			}

			n, err := LoadNote(path, cfg)
			if err != nil {
				return err
			}

			return pred(path, n)
		}),
		"Cannot read directory to traverse notes. Directory for category or note-cli home does not exist",
	)
}
