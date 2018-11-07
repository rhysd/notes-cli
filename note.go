package notes

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	reTitleBar = regexp.MustCompile("^=+$")
	// Match to horizontal ruler of markdown: https://spec.commonmark.org/0.28/#thematic-break
	// such as:
	// ---
	// ***
	// ___
	//   -	-   - -  -
	reHorizontalRule = regexp.MustCompile(`^\s{0,3}(?:(?:-+\s*){3,}|(?:\*+\s*){3,}|(?:_+\s*){3,})$`)
)

// Note represents a note stored on filesystem or will be created
type Note struct {
	// Config is a configuration of notes command which was created by NewConfig()
	Config *Config
	// Category is a category string. It must not be empty
	Category string
	// Tags is tags of note. It can be empty and cannot contain comma
	Tags []string
	// Created is a datetime when note was created
	Created time.Time
	// File is a file name of the note
	File string
	// Title is a title string of the note. When the note is not created yet, it may be empty
	Title string
}

// DirPath returns the absolute category directory path of the note
func (note *Note) DirPath() string {
	return filepath.Join(note.Config.HomePath, note.Category)
}

// FilePath returns the absolute file path of the note
func (note *Note) FilePath() string {
	return filepath.Join(note.Config.HomePath, note.Category, note.File)
}

// RelFilePath returns the relative file path of the note from home directory
func (note *Note) RelFilePath() string {
	return filepath.Join(note.Category, note.File)
}

// Create creates a file of the note. When title is empty, file name omitting file extension is used
// for it. This function will fail when the file is already existing.
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
		return errors.Errorf("Cannot create new note since file '%s' already exists. Please edit it", note.RelFilePath())
	}

	return errors.Wrap(ioutil.WriteFile(p, b.Bytes(), 0644), "Cannot write note to file")
}

// Open opens the note using an editor command user set. When user did not set any editor command
// with $NOTES_CLI_EDITOR, this method fails. Otherwise, an editor process is spawned with argument
// of path to the note file
func (note *Note) Open() error {
	return openEditor(note.Config, note.FilePath())
}

// ReadBodyN reads body of note until maxBytes bytes and returns it as string
func (note *Note) ReadBodyN(maxBytes int64) (string, error) {
	path := note.FilePath()
	f, err := os.Open(path)
	if err != nil {
		return "", errors.Wrap(err, "Cannot open note file")
	}
	defer f.Close()

	// Skip metadata
	r := bufio.NewReader(f)
	sawCat, sawTags, sawCreated := false, false, false
	for {
		t, err := r.ReadString('\n')
		if strings.HasPrefix(t, "- Category: ") {
			sawCat = true
		} else if strings.HasPrefix(t, "- Tags:") {
			sawTags = true
		} else if strings.HasPrefix(t, "- Created: ") {
			sawCreated = true
		}
		if sawCat && sawTags && sawCreated {
			break
		}
		if err != nil {
			return "", errors.Wrapf(err, "Cannot read metadata of note file. Some metadata may be missing in '%s'", note.RelFilePath())
		}
	}

	var buf bytes.Buffer

	// Skip empty lines
	for {
		b, err := r.ReadBytes('\n')
		if err != nil {
			break
		}
		if len(b) > 1 && !reHorizontalRule.Match(b) {
			buf.Write(b)
			break
		}
	}

	len := int64(buf.Len())
	if len > maxBytes {
		return string(buf.Bytes()[:maxBytes]), nil
	}

	if _, err := io.CopyN(&buf, r, maxBytes-len); err != nil && err != io.EOF {
		return "", err
	}

	return buf.String(), nil
}

// NewNote creates a new note instance with given parameters and configuration. Category and file name
// cannot be empty. If given file name lacks file extension, it automatically adds ".md" to file name.
func NewNote(cat, tags, file, title string, cfg *Config) (*Note, error) {
	if err := validateDirname(cat); err != nil {
		return nil, errors.Wrap(err, "Invalid category as directory name")
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

	if !strings.HasSuffix(file, ".md") {
		file += ".md"
	}
	return &Note{cfg, cat, ts, time.Now(), file, title}, nil
}

// LoadNote reads note file from given path, parses it and creates Note instance. When given file path
// does not exist or when the file does note contain mandatory metadata ('Category', 'Tags' and 'Created'),
// this function returns an error
func LoadNote(path string, cfg *Config) (*Note, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot open note file")
	}
	defer f.Close()

	note := &Note{Config: cfg}

	note.File = filepath.Base(path)

	s := bufio.NewScanner(f)
	titleFound := false
	for s.Scan() {
		line := s.Text()
		// First line is title
		if !titleFound {
			if reTitleBar.MatchString(line) {
				if note.Title == "" {
					note.Title = "(no title)"
				}
				titleFound = true
			} else {
				note.Title = line
			}
		} else if strings.HasPrefix(line, "- Category: ") {
			note.Category = strings.TrimSpace(line[12:])
			if c := filepath.Base(filepath.Dir(path)); c != note.Category {
				return nil, errors.Errorf("Category does not match between file path and file content, in path '%s' v.s. in file '%s'", c, note.Category)
			}
		} else if strings.HasPrefix(line, "- Tags:") {
			tags := strings.Split(strings.TrimSpace(line[7:]), ",")
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
		if note.Category != "" && note.Tags != nil && !note.Created.IsZero() && note.Title != "" {
			break
		}
	}
	if err := s.Err(); err != nil {
		return nil, errors.Wrapf(err, "Cannot read note file '%s'", canonPath(path))
	}

	if !titleFound {
		return nil, errors.Errorf("No title found in note '%s'. Didn't you use '====' bar for h1 title?", canonPath(path))
	}

	if note.Category == "" || note.Tags == nil || note.Created.IsZero() {
		return nil, errors.Errorf("Missing metadata in file '%s'. 'Category', 'Tags', 'Created' are mandatory", canonPath(path))
	}

	return note, nil
}

// WalkNotes walks all notes with given predicate. If given category string is an empty, it traverses
// notes of all categories. Otherwise, it only traverses notes of the specified categories. When the
// category does not exist, this function returns an error. When given predicate returns an error or
// when loading a note fails, this function stops traversing and immediately returns the error
func WalkNotes(cat string, cfg *Config, pred func(path string, note *Note) error) error {
	fs, err := ioutil.ReadDir(cfg.HomePath)
	if err != nil {
		return errors.Wrap(err, "Cannot read home")
	}

	cats := make([]string, 0, len(fs))
	for _, f := range fs {
		n := f.Name()
		if f.IsDir() && n != ".git" {
			cats = append(cats, n)
		}
	}

	if cat != "" {
		found := false
		for _, c := range cats {
			if c == cat {
				cats = []string{cat}
				found = true
				break
			}
		}
		if !found {
			return errors.Errorf("Category '%s' does not exist. All categories are %s", cat, strings.Join(cats, ", "))
		}
	}

	for _, c := range cats {
		dir := filepath.Join(cfg.HomePath, c)
		es, err := ioutil.ReadDir(dir)
		if err != nil {
			return errors.Wrapf(err, "Cannot read directory for category '%s'", c)
		}

		for _, e := range es {
			f := e.Name()
			if e.IsDir() || !strings.HasSuffix(f, ".md") {
				continue
			}
			p := filepath.Join(dir, f)
			n, err := LoadNote(p, cfg)
			if err != nil {
				return err
			}
			if err := pred(p, n); err != nil {
				return err
			}
		}
	}

	return nil
}
