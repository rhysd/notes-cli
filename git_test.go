package notes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitExecSuccess(t *testing.T) {
	g := NewGit(&Config{GitPath: "git", HomePath: "."})

	out, err := g.Exec("status", "-b")
	if err != nil {
		t.Fatal(err, out)
	}
	if !strings.HasPrefix(out, "On branch ") && !strings.HasPrefix(out, "HEAD detached at ") {
		t.Fatal("Unexpected output:", out)
	}
}

func TestGitExecFailure(t *testing.T) {
	g := NewGit(&Config{GitPath: "git", HomePath: "."})
	out, err := g.Exec("hoge-fuga")
	if err == nil {
		t.Fatal("Unknown Git subcommand did not cause an error")
	}
	if !strings.Contains(out, "'hoge-fuga' is not a git command") {
		t.Fatal("Unexpected error message:", out)
	}
}

func TestGitIsOptional(t *testing.T) {
	g := NewGit(&Config{GitPath: "", HomePath: "."})
	if g != nil {
		t.Fatal("Expected nil:", g)
	}
}

func TestGitInitAddCommit(t *testing.T) {
	dir := "test-tmp-dir-git"
	panicIfErr(os.Mkdir(dir, 0755))
	defer func() { panicIfErr(os.RemoveAll(dir)) }()

	g := NewGit(&Config{GitPath: "git", HomePath: dir})

	if err := g.Init(); err != nil {
		t.Fatal(err)
	}
	s, err := os.Stat(filepath.Join(dir, ".git"))
	if err != nil {
		t.Fatal(".git was not created", err)
	}
	if !s.IsDir() {
		t.Fatal(".git is not directory")
	}

	if out, err := g.Exec("config", "user.name", "You"); err != nil {
		t.Fatal(out, err)
	}

	if out, err := g.Exec("config", "user.email", "you@example.com"); err != nil {
		t.Fatal(out, err)
	}

	f, err := os.Create(filepath.Join(dir, "tmp.txt"))
	panicIfErr(err)
	f.WriteString("hello\n")
	f.Close()

	if err := g.AddAll(); err != nil {
		t.Fatal(err)
	}

	out, err := g.Exec("status")
	panicIfErr(err)
	if !strings.Contains(out, "new file:   tmp.txt") {
		t.Fatal("file was not added. Status:", out)
	}

	if err := g.Commit("hello hello"); err != nil {
		t.Fatal(err)
	}

	out, err = g.Exec("log", "--oneline")
	panicIfErr(err)

	if !strings.Contains(out, "hello hello") {
		t.Fatal("The commit does not exist in log:", out)
	}
}

func TestGitTrackingRemote(t *testing.T) {
	// This test cannot be run on CI since they use detached HEAD for clone
	wantErr := false
	if _, ok := os.LookupEnv("GITHUB_ACTIONS"); ok {
		wantErr = true
	}

	g := NewGit(&Config{GitPath: "git", HomePath: "."})
	re, br, err := g.TrackingRemote()
	if err != nil {
		if wantErr {
			return
		}
		t.Fatal(err)
	}
	if re != "origin" {
		t.Error("Unexpected remote name", re)
	}
	out, err := g.Exec("rev-parse", "--verify", br)
	if err != nil {
		t.Error("Branch is not correct", br, out)
	}
}

func TestGitOpFails(t *testing.T) {
	g := NewGit(&Config{GitPath: "git", HomePath: "/path/to/not/existing/dir"})

	for _, tc := range []struct {
		what string
		do   func() error
	}{
		{
			what: "init",
			do:   func() error { return g.Init() },
		},
		{
			what: "add -A",
			do:   func() error { return g.AddAll() },
		},
		{
			what: "commit",
			do:   func() error { return g.Commit("hello") },
		},
		{
			what: "tracking remote",
			do:   func() error { _, _, err := g.TrackingRemote(); return err },
		},
		{
			what: "push",
			do:   func() error { return g.Push("origin", "master") },
		},
	} {
		t.Run(tc.what, func(t *testing.T) {
			if err := tc.do(); err == nil {
				t.Fatal("Error expected")
			}
		})
	}
}

func TestGitInitTwice(t *testing.T) {
	dir := "test-tmp-dir-git-init"
	panicIfErr(os.Mkdir(dir, 0755))
	defer func() { panicIfErr(os.RemoveAll(dir)) }()

	g := NewGit(&Config{GitPath: "git", HomePath: dir})
	if err := g.Init(); err != nil {
		t.Fatal(err)
	}
	if err := g.Init(); err != nil {
		t.Fatal(err)
	}
}
