package notes

import (
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Git represents Git command for specific repository
type Git struct {
	bin  string
	root string
}

func (git *Git) canonRoot() string {
	return canonPath(git.root)
}

// Command returns exec.Command instance which runs given Git subcommand with given arguments
func (git *Git) Command(subcmd string, args ...string) *exec.Cmd {
	// e.g. 'git diff --cached' -> 'git -C /path/to/repo diff --cached'
	a := append([]string{"-C", git.root, subcmd}, args...)
	cmd := exec.Command(git.bin, a...)
	return cmd
}

// Exec runs runs given Git subcommand with given arguments
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

// Init runs `git init` with no argument
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

// AddAll runs `git add -A`
func (git *Git) AddAll() error {
	out, err := git.Exec("add", "-A")
	if err != nil {
		return errors.Wrapf(err, "Cannot add changes to index tree at '%s': %s", git.canonRoot(), out)
	}
	return nil
}

// Commit runs `git commit` with given message
func (git *Git) Commit(msg string) error {
	out, err := git.Exec("commit", "-m", msg)
	if err != nil {
		return errors.Wrapf(err, "Cannot commit changes to repository at '%s': %s", git.canonRoot(), out)
	}
	return nil
}

// TrackingRemote returns remote name branch name. It fails when current branch does not track any branch
func (git *Git) TrackingRemote() (string, string, error) {
	s, err := git.Exec("rev-parse", "--abbrev-ref", "--symbolic", "@{u}")
	if err != nil {
		return "", "", errors.Wrapf(err, "Cannot retrieve remote name: %s", s)
	}
	// e.g. origin/master
	ss := strings.Split(s, "/")
	return ss[0], ss[1], nil
}

// Push pushes given branch of repository to the given remote
func (git *Git) Push(remote, branch string) error {
	out, err := git.Exec("push", "-u", remote, branch)
	if err != nil {
		return errors.Wrapf(err, "Cannot push changes to %s/%s at '%s': %s", remote, branch, git.canonRoot(), out)
	}
	return nil
}

// NewGit creates Git instance from Config value. Home directory is assumed to be a root of Git repository
func NewGit(c *Config) *Git {
	if c.GitPath == "" {
		// Git is optional
		return nil
	}
	return &Git{c.GitPath, c.HomePath}
}
