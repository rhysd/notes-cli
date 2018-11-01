package notes

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

type testConfigEnvState map[string]string

func testNewConfigEnvState() testConfigEnvState {
	s := testConfigEnvState{}

	for _, key := range []string{
		"NOTES_CLI_HOME",
		"XDG_DATA_HOME",
		"APPLOCALDATA",
		"NOTES_CLI_GIT",
		"NOTES_CLI_EDITOR",
	} {
		if e, ok := os.LookupEnv(key); ok {
			s[key] = e
			os.Unsetenv(key)
		}
	}
	return s
}

func (state testConfigEnvState) restore() {
	for k, v := range state {
		if err := os.Setenv(k, v); err != nil {
			panic(err)
		}
	}
}

func TestNewDefaultConfig(t *testing.T) {
	s := testNewConfigEnvState()
	defer s.restore()

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
	if _, err := exec.LookPath("git"); err == nil {
		if c.GitPath == "" {
			t.Fatal("Git path was not detected")
		}
	} else {
		if c.GitPath != "" {
			t.Fatal("Git path should not be detected:", c.GitPath)
		}
	}
	if c.EditorPath != "" {
		t.Fatal("Editor path should be empty by default:", c.EditorPath)
	}
}

func TestNewConfigCustomizeBinaryPaths(t *testing.T) {
	s := testNewConfigEnvState()
	defer s.restore()

	ls, err := exec.LookPath("ls")
	if err != nil {
		panic(err)
	}
	os.Setenv("NOTES_CLI_GIT", ls)
	os.Setenv("NOTES_CLI_EDITOR", ls)

	c, err := NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	if c.GitPath != ls {
		t.Fatal("git path is unexpected:", c.GitPath, "wanted:", ls)
	}

	if c.EditorPath != ls {
		t.Fatal("Editor is unexpected:", c.EditorPath, "wanted:", ls)
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
		if runtime.GOOS != "windows" && tc.key == "APPLOCALDATA" {
			continue
		}

		t.Run(tc.key, func(t *testing.T) {
			s := testNewConfigEnvState()
			defer s.restore()

			os.Setenv(tc.key, tc.val)

			c, err := NewConfig()
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tc.val)

			if c.HomePath != tc.home {
				t.Fatal("Home is unexpected:", c.HomePath, "wanted:", tc.home)
			}
			stat, err := os.Stat(c.HomePath)
			if err != nil {
				t.Fatal(err)
			}
			if !stat.IsDir() {
				t.Fatal("Directory was not created for home:", stat)
			}
		})
	}
}
