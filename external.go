package notes

import (
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"regexp"
)

var (
	reExtractCmdName = regexp.MustCompile(`^expected command but got "([[:alnum:]_-]+)"$`)
)

// ExternalCmd represents user-defined subcommand
type ExternalCmd struct {
	// ExePath is a path to executable of the external subcommand
	ExePath string
	// Args is arguments passed to external subcommand. Arguments specified to `notes` are forwarded
	Args []string
}

// Do invokes executable command with exec. If it did not exit successfully this function returns an error
func (cmd *ExternalCmd) Do() error {
	c := exec.Command(cmd.ExePath, cmd.Args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return errors.Wrapf(c.Run(), "External command '%s' did not exit successfully", cmd.ExePath)
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

	return &ExternalCmd{
		ExePath: exe,
		Args:    args,
	}, true
}
