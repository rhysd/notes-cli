package notes

import (
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Git struct {
	bin  string
	root string
}

func (git *Git) canonRoot() string {
	return canonPath(git.root)
}

func (git *Git) Command(subcmd string, args ...string) *exec.Cmd {
	// e.g. 'git diff --cached' -> 'git -C /path/to/repo diff --cached'
	a := append([]string{"-C", git.root, subcmd}, args...)
	cmd := exec.Command(git.bin, a...)
	return cmd
}

func (git *Git) Exec(subcmd string, args ...string) (string, error) {
	b, err := git.Command(subcmd, args...).CombinedOutput()

	// Chop last newline
	l := len(b)
	if l > 0 && b[l-1] == '\n' {
		b = b[:l-1]
	}

	// Make output in oneline in error cases
	if err != nil {
		for i := range b {
			if b[i] == '\n' {
				b[i] = ' '
			}
		}
	}

	return string(b), err
}

func (git *Git) Init() error {
	if s, err := os.Stat(filepath.Join(git.root, ".git")); err == nil && s.IsDir() {
		// Repository was already created
		return nil
	}

	out, err := git.Exec("init")
	if err != nil {
		return errors.Wrapf(err, "Cannot init Git repository at '%s': %s", git.canonRoot(), out)
	}
	return nil
}

func (git *Git) AddAll() error {
	out, err := git.Exec("add", "-A")
	if err != nil {
		return errors.Wrapf(err, "Cannot add changes to index tree at '%s': %s", git.canonRoot(), out)
	}
	return nil
}

func (git *Git) Commit(msg string) error {
	out, err := git.Exec("commit", "-m", msg)
	if err != nil {
		return errors.Wrapf(err, "Cannot commit changes to repository at '%s': %s", git.canonRoot(), out)
	}
	return nil
}

func (git *Git) TrackingRemote() (string, string, error) {
	s, err := git.Exec("rev-parse", "--abbrev-ref", "--symbolic", "@{u}")
	if err != nil {
		return "", "", errors.Wrapf(err, "Cannot retrieve remote name: %s", s)
	}
	// e.g. origin/master
	ss := strings.Split(s, "/")
	return ss[0], ss[1], nil
}

func (git *Git) Push(remote, branch string) error {
	out, err := git.Exec("push", remote, branch)
	if err != nil {
		return errors.Wrapf(err, "Cannot push changes to %s/%s at '%s': %s", remote, branch, git.canonRoot(), out)
	}
	return nil
}

func NewGit(c *Config) *Git {
	if c.GitPath == "" {
		// Git is optional
		return nil
	}
	return &Git{c.GitPath, c.HomePath}
}
