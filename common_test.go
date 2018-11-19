package notes

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// Precondition for tests
func init() {
	for _, env := range []string{
		"NOTES_CLI_HOME",
		"NOTES_CLI_GIT",
		"NOTES_CLI_EDITOR",
		"NOTES_CLI_PAGER",
		"EDITOR",
		"XDG_DATA_HOME",
		"PAGER",
	} {
		os.Unsetenv(env)
	}
}

// Test utilities

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func testExternalCommandBinaryDir(name string, t *testing.T) string {
	cwd, err := os.Getwd()
	panicIfErr(err)
	pkg := "./testdata/external/notes-external-" + name
	exeFile := "notes-external-" + name
	if runtime.GOOS == "windows" {
		exeFile += ".exe"
	}
	exe := filepath.Join(cwd, filepath.FromSlash(pkg), exeFile)
	if _, err := os.Stat(exe); err != nil {
		c := exec.Command("go", "build", "-o", exe, pkg)
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatal("Cannot build package", pkg, "to create executable", exe, "due to", err, "output:", string(out))
		}
		_, err = os.Stat(exe) // Verify
		panicIfErr(err)
	}
	return filepath.Dir(exe)
}

type alwaysErrorWriter struct {
}

func (w alwaysErrorWriter) Write(p []byte) (int, error) {
	return 0, errors.New("Write error for test")
}
