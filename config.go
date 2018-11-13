package notes

import (
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
)

// Config represents user configuration of notes command
type Config struct {
	// HomePath is a file path to directory of home of notes command. If $NOTES_CLI_HOME is set, it is used.
	// Otherwise, notes-cli directory in XDG data directory is used. This directory is automatically created
	// when config is created
	HomePath string
	// GitPath is a file path to `git` executable. If $NOTES_CLI_GIT is set, it is used.
	// Otherwise, `git` is used by default. This is optional and can be empty. When empty, some command
	// and functionality which require Git don't work
	GitPath string
	// EditorPath is a file path to executable of your favorite editor. If $NOTES_CLI_EDITOR is set, it is used.
	// Otherwise, this value will be empty. When empty, some functionality which requires an editor to open note
	// doesn't work
	EditorPath string
}

func homePath() (string, error) {
	if env := os.Getenv("NOTES_CLI_HOME"); env != "" {
		return filepath.Clean(env), nil
	}

	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "notes-cli"), nil
	}

	if runtime.GOOS == "windows" {
		if env := os.Getenv("APPLOCALDATA"); env != "" {
			return filepath.Join(env, "notes-cli"), nil
		}
	}

	u, err := user.Current()
	if err != nil {
		return "", errors.Wrap(err, "Cannot locate home directory. Please set $NOTES_CLI_HOME")
	}
	return filepath.Join(u.HomeDir, ".local", "share", "notes-cli"), nil
}

func gitPath() string {
	c := "git"
	if env := os.Getenv("NOTES_CLI_GIT"); env != "" {
		c = filepath.Clean(env)
	}

	exe, err := exec.LookPath(c)
	if err != nil {
		// Git is optional
		return ""
	}

	return exe
}

func editorPath() string {
	for _, key := range []string{"NOTES_CLI_EDITOR", "EDITOR"} {
		if env := os.Getenv(key); env != "" {
			if exe, err := exec.LookPath(env); err == nil {
				return exe
			}
		}
	}
	return ""
}

// NewConfig creates a new Config instance by looking the user's environment. GitPath and EditorPath
// may be empty when proper configuration is not found. When home directory path cannot be located,
// this function returns an error
func NewConfig() (*Config, error) {
	h, err := homePath()
	if err != nil {
		return nil, err
	}

	// Ensure home directory exists
	if err := os.MkdirAll(h, 0755); err != nil {
		return nil, errors.Wrapf(err, "Could not create home '%s'", h)
	}

	return &Config{h, gitPath(), editorPath()}, nil
}
