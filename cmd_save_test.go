package notes

import (
	"github.com/rhysd/go-tmpenv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testNewConfigForSaveCmd(subdir string) *Config {
	cfg, err := NewConfig()
	panicIfErr(err)
	cwd, err := os.Getwd()
	panicIfErr(err)
	cfg.HomePath = filepath.Join(cwd, "testdata", "save", subdir)
	return cfg
}

func prepareGitRepoForTestNewCmd(g *Git) {
	panicIfErr(g.Init())
	_, err := g.Exec("config", "user.name", "You")
	panicIfErr(err)
	_, err = g.Exec("config", "user.email", "you@example.com")
	panicIfErr(err)
}

func TestNoGitForSave(t *testing.T) {
	cfg := testNewConfigForSaveCmd("normal")
	cfg.GitPath = ""
	cmd := &SaveCmd{Config: cfg}
	err := cmd.Do()
	if err == nil {
		t.Fatal("No error occurred")
	}
	if !strings.Contains(err.Error(), "'save' command cannot work without Git") {
		t.Fatal("Unexpected error:", err)
	}
}

func TestSaveCmd(t *testing.T) {
	cfg := testNewConfigForSaveCmd("normal")
	g := NewGit(cfg)
	for _, tc := range []struct {
		what string
		msg  string
		want string
	}{
		{
			what: "no message",
			msg:  "",
			want: "Saved by notes CLI at",
		},
		{
			what: "with message",
			msg:  "this is custom message",
			want: "this is custom message",
		},
	} {
		t.Run(tc.what, func(t *testing.T) {
			prepareGitRepoForTestNewCmd(g)
			defer os.RemoveAll(filepath.Join(cfg.HomePath, ".git"))

			cmd := &SaveCmd{
				Config:  cfg,
				Message: tc.msg,
			}

			if err := cmd.Do(); err != nil {
				t.Fatal(err)
			}

			log, err := g.Exec("log", "--oneline")
			panicIfErr(err)
			lines := strings.Split(strings.TrimSuffix(log, "\n"), "\n")
			if len(lines) != 1 {
				t.Fatal("Number of logs is not match. log:", log)
			}
			if !strings.Contains(lines[0], tc.want) {
				t.Fatal("Unexpected log", lines[0], "log:", log)
			}
		})
	}
}

func TestSaveCmdNoGitInitYet(t *testing.T) {
	cfg := testNewConfigForSaveCmd("normal")
	cmd := &SaveCmd{Config: cfg}
	err := cmd.Do()
	if err == nil {
		t.Fatal("Error did not occur")
	}
	if !strings.Contains(err.Error(), "'.git' directory does not exist in home") {
		t.Fatal("Unexpected error:", err)
	}
}

func TestSaveCmdAddNothing(t *testing.T) {
	cfg := testNewConfigForSaveCmd("empty")
	panicIfErr(os.MkdirAll(cfg.HomePath, 0755))
	g := NewGit(cfg)
	prepareGitRepoForTestNewCmd(g)
	defer func() {
		panicIfErr(os.RemoveAll(filepath.Join(cfg.HomePath, ".git")))
	}()

	cmd := &SaveCmd{Config: cfg}
	err := cmd.Do()

	if err == nil {
		t.Fatal("No error occurred")
	}

	if !strings.Contains(err.Error(), "nothing to commit") {
		t.Fatal("Unexpected output:", err)
	}
}

func TestSaveCmdCannotPush(t *testing.T) {
	if _, ok := os.LookupEnv("APPVEYOR"); ok {
		t.Skip("Pushing to not permitted repository hangs on Appveyor")
	}

	tmp := tmpenv.New()
	defer func() { panicIfErr(tmp.Restore()) }()
	panicIfErr(tmp.Setenv("GIT_TERMINAL_PROMPT", "0"))

	cfg := testNewConfigForSaveCmd("normal")
	g := NewGit(cfg)
	prepareGitRepoForTestNewCmd(g)
	defer os.RemoveAll(filepath.Join(cfg.HomePath, ".git"))

	if out, err := g.Exec("remote", "add", "origin", "https://github.com/rhysd-test/empty.git"); err != nil {
		t.Fatal(err, out)
	}
	if out, err := g.Exec("fetch"); err != nil {
		t.Fatal(err, out)
	}
	if out, err := g.Exec("commit", "--allow-empty", "--allow-empty-message", "-m", ""); err != nil {
		t.Fatal(err, out)
	}
	if out, err := g.Exec("branch", "--set-upstream-to", "origin/master", "master"); err != nil {
		t.Fatal(err, out)
	}

	cmd := &SaveCmd{
		Config: cfg,
	}

	err := cmd.Do()
	if err == nil {
		t.Fatal("No error occurred")
	}
	if !strings.Contains(err.Error(), "Cannot push to 'origin' remote") {
		t.Fatal("Unexpected output:", err)
	}
}
