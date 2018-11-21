// +build !windows

package notes

import (
	"os"
	"path/filepath"
)

func Example() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	cfg := &Config{
		HomePath: filepath.Join(cwd, "example", "notes-cli"),
	}

	cmd := ListCmd{
		Config:  cfg,
		Oneline: true,
		Out:     os.Stdout,
	}

	// Shows oneline notes (relative file path, category, tags, title)
	if err := cmd.Do(); err != nil {
		panic(err)
	}
	// Output:
	// blog/daily/dialy-2018-11-20.md                         dialy-2018-11-20
	// blog/daily/dialy-2018-11-18.md             notes       dialy-2018-11-18
	// memo/tasks.md                                          My tasks
	// memo/notes-urls.md                         notes       URLs for notes
	// blog/tech/introduction-to-notes-command.md notes       introduction-to-notes-command
	// blog/tech/how-to-handle-files.md           golang,file How to hanle files in Go
}
