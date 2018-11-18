package notes

import (
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

// Note: External command name must consist of alphabets, numbers, dash '-' and underscore '_'.
var (
	reExtractCmdName = regexp.MustCompile(`^expected command but got "([[:alnum:]_-]+)"$`)
)

// ExternalCmd represents user-defined subcommand
type ExternalCmd struct {
	// ExePath is a path to executable of the external subcommand
	ExePath string
	// Args is arguments passed to external subcommand. Arguments specified to `notes` are forwarded
	Args []string
	// NotesPath is an executable path of the `notes` command. This is passed to the first argument of external subcommand
	NotesPath string
}

// Do invokes external subcommand with exec. If it did not exit successfully this function returns an error
func (cmd *ExternalCmd) Do() error {
	args := append([]string{cmd.NotesPath}, cmd.Args...)
	c := exec.Command(cmd.ExePath, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	if err := c.Run(); err != nil {
		name := filepath.Base(cmd.ExePath)
		return errors.Wrapf(err, "External command '%s' did not exit successfully", name)
	}
	return nil
}

// NewExternalCmd creates ExternalCmd instance from given error and arguments. The error must be parse
// error of kingpin.Parse(). When the missing subcommand is not detected in the error message, this
// function returns false as 2nd return value
func NewExternalCmd(fromErr error, args []string) (*ExternalCmd, bool) {
	match := reExtractCmdName.FindSubmatch([]byte(fromErr.Error()))
	if match == nil {
		return nil, false
	}

	exe, err := exec.LookPath("notes-" + string(match[1]))
	if err != nil {
		return nil, false
	}

	notes, err := os.Executable()
	if err != nil {
		return nil, false
	}

	return &ExternalCmd{
		ExePath:   exe,
		Args:      args,
		NotesPath: notes,
	}, true
}
