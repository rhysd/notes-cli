package notes

import (
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
)

type Config struct {
	HomePath   string
	GitPath    string
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
	env := os.Getenv("NOTES_CLI_EDITOR")
	if env == "" {
		return ""
	}

	exe, err := exec.LookPath(env)
	if err != nil {
		return ""
	}

	return exe
}

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
