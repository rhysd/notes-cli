package notes

import (
	"github.com/kballard/go-shellquote"
	"github.com/rhysd/go-tmpenv"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func testNewConfigEnvGuard() *tmpenv.Envguard {
	g, err := tmpenv.Unset(
		"NOTES_CLI_HOME",
		"XDG_DATA_HOME",
		"APPLOCALDATA",
		"NOTES_CLI_GIT",
		"NOTES_CLI_EDITOR",
		"NOTES_CLI_PAGER",
		"EDITOR",
		"PAGER",
	)
	panicIfErr(err)
	return g
}

func TestNewDefaultConfig(t *testing.T) {
	g := testNewConfigEnvGuard()
	g.Unsetenv("PATH") // Also unset $PATH to simulate git and less are not installed
	defer func() { panicIfErr(g.Restore()) }()

	c, err := NewConfig()
	if err != nil {
		t.Fatal(err)
	}
	if c.HomePath == "" {
		t.Fatal("Home is empty")
	}
	stat, err := os.Stat(c.HomePath)
	if err != nil {
		t.Fatal(err)
	}
	if !stat.IsDir() {
		t.Fatal("Directory was not created for home:", stat)
	}
	if c.GitPath != "" {
		t.Fatal("Git path should not be detected:", c.GitPath)
	}
	if c.EditorCmd != "" {
		t.Fatal("Editor path should be empty by default:", c.EditorCmd)
	}
}

func TestNewDefaultConfigWithGitAndLess(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("`git` is not available", err)
	}
	if _, err := exec.LookPath("less"); err != nil {
		t.Skip("`less` is not available", err)
	}

	g := testNewConfigEnvGuard()
	defer func() { panicIfErr(g.Restore()) }()

	c, err := NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	if c.GitPath == "" {
		t.Fatal("Git path should not be empty")
	}

	if c.PagerCmd != "less -R -F -X" {
		t.Fatal("`less -R -F -X` should be set as default", c.PagerCmd)
	}
}

func TestNewConfigCustomizeBinaryPaths(t *testing.T) {
	g := testNewConfigEnvGuard()
	defer func() { panicIfErr(g.Restore()) }()

	ls, err := exec.LookPath("ls")
	panicIfErr(err)
	qls := shellquote.Join(ls) // On Windows, it may contain 'Program Files'
	os.Setenv("NOTES_CLI_GIT", ls)
	os.Setenv("NOTES_CLI_EDITOR", qls)
	os.Setenv("NOTES_CLI_PAGER", qls)

	c, err := NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	if c.GitPath != ls {
		t.Fatal("git path is unexpected:", c.GitPath, "wanted:", ls)
	}

	if c.EditorCmd != qls {
		t.Fatal("Editor is unexpected:", c.EditorCmd, "wanted:", qls)
	}

	if c.PagerCmd != qls {
		t.Fatal("Pager is unexpected:", c.PagerCmd, "wanted:", qls)
	}

	os.Unsetenv("NOTES_CLI_EDITOR")
	os.Setenv("EDITOR", qls)
	os.Unsetenv("NOTES_CLI_PAGER")
	os.Setenv("PAGER", qls)

	c, err = NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	if c.EditorCmd != qls {
		t.Fatal("Editor is unexpected:", c.EditorCmd, "wanted:", qls)
	}
	if c.PagerCmd != qls {
		t.Fatal("Pager is unexpected:", c.PagerCmd, "wanted:", qls)
	}
}

func TestNewConfigCustomizeHome(t *testing.T) {
	for _, tc := range []struct {
		key  string
		val  string
		home string
	}{
		{
			key:  "NOTES_CLI_HOME",
			val:  "test-config-home",
			home: "test-config-home",
		},
		{
			key:  "XDG_DATA_HOME",
			val:  "test-xdg-config-home",
			home: filepath.FromSlash("test-xdg-config-home/notes-cli"),
		},
		{
			key:  "APPLOCALDATA",
			val:  "test-win-config-home",
			home: filepath.FromSlash("test-win-config-home/notes-cli"),
		},
	} {
		t.Run(tc.key, func(t *testing.T) {
			if runtime.GOOS != "windows" && tc.key == "APPLOCALDATA" {
				t.Skip("APPLOCALDATA is referred only on Windows")
			}

			g := testNewConfigEnvGuard()
			defer func() { panicIfErr(g.Restore()) }()

			panicIfErr(os.Setenv(tc.key, tc.val))

			c, err := NewConfig()
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				panicIfErr(os.RemoveAll(tc.val))
			}()

			if c.HomePath != tc.home {
				t.Fatal("Home is unexpected:", c.HomePath, "wanted:", tc.home)
			}
			stat, err := os.Stat(c.HomePath)
			if err != nil {
				t.Fatal(err, c.HomePath)
			}
			if !stat.IsDir() {
				t.Fatal("Directory was not created for home:", c.HomePath, stat)
			}
		})
	}
}

func TestNewConfigGitNotFound(t *testing.T) {
	g := testNewConfigEnvGuard()
	defer func() { panicIfErr(g.Restore()) }()

	panicIfErr(os.Setenv("NOTES_CLI_GIT", "/path/to/unknown-command"))

	c, err := NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	if c.GitPath != "" {
		t.Fatal("git path should be empty:", c.GitPath)
	}
}

func TestNewConfigCannotCreateHome(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("This test cannot be run on Windows")
	}

	g := testNewConfigEnvGuard()
	defer func() { panicIfErr(g.Restore()) }()

	panicIfErr(os.Setenv("NOTES_CLI_HOME", filepath.FromSlash("/invalid-dir-for-notes-cli-test")))

	_, err := NewConfig()
	if err == nil {
		t.Fatal("Error did not occur")
	}

}

func TestNewConfigExpandTilde(t *testing.T) {
	u, err := user.Current()
	panicIfErr(err)
	cwd, err := os.Getwd()
	panicIfErr(err)
	if !strings.HasPrefix(cwd, u.HomeDir) {
		t.Skip("because test is not running under home directory")
	}

	g := testNewConfigEnvGuard()
	defer func() { panicIfErr(g.Restore()) }()

	panicIfErr(os.Setenv("NOTES_CLI_HOME", filepath.Join("~", strings.TrimPrefix(cwd, u.HomeDir))))
	c, err := NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	if c.HomePath != cwd {
		t.Fatal("'~' was not expanded collectly. Wanted", cwd, "but got", c.HomePath)
	}
}
