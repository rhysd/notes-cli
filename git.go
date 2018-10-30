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

func (g *Git) Command(subcmd string, args ...string) *exec.Cmd {
	// e.g. 'git diff --cached' -> 'git -C /path/to/repo diff --cached'
	a := append([]string{"-C", g.root, subcmd}, args...)
	cmd := exec.Command(g.bin, a...)
	return cmd
}

func (g *Git) Exec(subcmd string, args ...string) (string, error) {
	b, err := g.Command(subcmd, args...).CombinedOutput()
	if err != nil {
		return string(b), err
	}

	// Chop last newline
	l := len(b)
	if l > 0 && b[l-1] == '\n' {
		b = b[:l-1]
	}

	return string(b), nil
}

func (g *Git) Init() error {
	if s, err := os.Stat(filepath.Join(g.root, ".git")); err == nil && s.IsDir() {
		// Repository was already created
		return nil
	}

	out, err := g.Exec("init")
	if err != nil {
		return errors.Wrapf(err, "Cannot init Git repository at '%s': %s", g.root, out)
	}
	return nil
}

func (g *Git) AddAll() error {
	out, err := g.Exec("add", "-A")
	if err != nil {
		return errors.Wrapf(err, "Cannot add changes to index tree at '%s': %s", g.root, out)
	}
	return nil
}

func (g *Git) Commit(msg string) error {
	out, err := g.Exec("commit", "-m", msg)
	if err != nil {
		if out == "" {
			return errors.Errorf("Cannot commit changes to '%s'. Is there no change?", g.root)
		}
		return errors.Wrapf(err, "Cannot commit changes to repository at '%s': %s", g.root, out)
	}
	return nil
}

func (g *Git) TrackingRemote() (string, string, error) {
	s, err := g.Exec("rev-parse", "--abbrev-ref", "--symbolic", "@{u}")
	if err != nil {
		return "", "", errors.Wrapf(err, "Cannot retrieve remote name: %s", s)
	}
	// e.g. origin/master
	ss := strings.Split(s, "/")
	return ss[0], ss[1], nil
}

func (g *Git) Push(remote, branch string) error {
	out, err := g.Exec("push", remote, branch)
	if err != nil {
		return errors.Wrapf(err, "Cannot push changes to %s/%s at '%s': %s", remote, branch, g.root, out)
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
