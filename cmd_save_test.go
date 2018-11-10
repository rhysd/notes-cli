package notes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testNewConfigForSaveCmd(subdir string) *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	cfg.HomePath = filepath.Join(cwd, "testdata", "save", subdir)
	return cfg
}

func prepareGitRepoForTestNewCmd(g *Git) {
	if err := g.Init(); err != nil {
		panic(err)
	}

	if _, err := g.Exec("config", "user.name", "You"); err != nil {
		panic(err)
	}

	if _, err := g.Exec("config", "user.email", "you@example.com"); err != nil {
		panic(err)
	}
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
			if err != nil {
				panic(err)
			}
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
	if err := os.MkdirAll(cfg.HomePath, 0755); err != nil {
		panic(err)
	}
	g := NewGit(cfg)
	prepareGitRepoForTestNewCmd(g)
	defer os.RemoveAll(filepath.Join(cfg.HomePath, ".git"))

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

	saved := os.Getenv("GIT_TERMINAL_PROMPT")
	if err := os.Setenv("GIT_TERMINAL_PROMPT", "0"); err != nil {
		panic(err)
	}
	defer func() {
		if saved == "" {
			os.Unsetenv("GIT_TERMINAL_PROMPT")
		} else {
			os.Setenv("GIT_TERMINAL_PROMPT", saved)
		}
	}()

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
