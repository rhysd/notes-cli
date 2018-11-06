/*
Package notes is a library which consists notes command.

https://github.com/rhysd/notes-cli/tree/master/cmd/notes

This library is for using notes command programmatically from Go program.
It consists structs which represent each subcommands.

1. Create Config instance with NewConfig
2. Create an instance of subcommand you want to run with config
3. Run it with .Do() method. It will return an error if some error occurs

	import (
		"bytes"
		"fmt"
		"github.com/rhysd/notes-cli"
		"os"
		"strings"
	)

	var buf bytes.Buffer

	// Create user configuration
	cfg, err := notes.NewConfig()
	if err != nil {
		panic(err)
	}

	// Prepare `notes list` command
	cmd := &notes.ListCmd{
		Config: cfg,
		Relative: true,
		Out: &buf
	}

	// Runs the command
	if err := cmd.Do(); err != nil {
		fmt.Fprintln(os.Stdout, err)
	}

	paths := strings.Split(strings.Trim(buf.String(), "\n"), "\n")
	fmt.Println("Note paths:", paths)

For usage of `notes` command, please read README of the repository.

https://github.com/rhysd/notes-cli/blob/master/README.md

*/
package notes
