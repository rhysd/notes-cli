package notes

import (
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

func openEditor(config *Config, args ...string) error {
	if config.EditorPath == "" {
		return errors.New("Editor is not set. To open note in editor, please set $NOTES_CLI_EDITOR")
	}
	c := exec.Command(config.EditorPath, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	c.Dir = config.HomePath
	return errors.Wrap(c.Run(), "Editor command did not exit successfully")
}
