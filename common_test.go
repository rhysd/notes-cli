package notes

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Precondition for tests
func init() {
	for _, env := range []string{"NOTES_CLI_HOME", "NOTES_CLI_GIT", "NOTES_CLI_EDITOR", "EDITOR", "XDG_DATA_HOME"} {
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
	exe := filepath.Join(cwd, filepath.FromSlash(pkg), "notes-external-test")
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
