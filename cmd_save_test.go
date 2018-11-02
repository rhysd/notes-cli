package notes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testNewConfigForSaveCmd() *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	cfg.HomePath = filepath.Join(cwd, "testdata", "save", "normal")
	return cfg
}

func TestNoGitForSave(t *testing.T) {
	cfg := testNewConfigForSaveCmd()
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
	cfg := testNewConfigForSaveCmd()
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
			if err := g.Init(); err != nil {
				panic(err)
			}
			defer os.RemoveAll(filepath.Join(cfg.HomePath, ".git"))

			if out, err := g.Exec("config", "user.name", "You"); err != nil {
				t.Fatal(out, err)
			}

			if out, err := g.Exec("config", "user.email", "you@example.com"); err != nil {
				t.Fatal(out, err)
			}

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
