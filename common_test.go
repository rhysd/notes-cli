package notes

import (
	"os"
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
