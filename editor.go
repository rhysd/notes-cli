package notes

import (
	"os"
	"os/exec"

	"github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
)

func openEditor(config *Config, paths ...string) error {
	if config.EditorCmd == "" {
		return errors.New("Editor is not set. To open note in editor, please set $NOTES_CLI_EDITOR or $EDITOR")
	}

	cmdline, err := shellquote.Split(config.EditorCmd)
	if err != nil {
		return errors.Wrap(err, "Cannot parse editor command line. Please check $NOTES_CLI_EDITOR or $EDITOR")
	}

	editor := cmdline[0]
	args := make([]string, 0, len(cmdline)-1+len(paths))
	args = append(args, cmdline[1:]...)
	args = append(args, paths...)

	c := exec.Command(editor, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	c.Dir = config.HomePath

	return errors.Wrapf(c.Run(), "Editor command '%s' did not exit successfully", editor)
}
