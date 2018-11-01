package main

import (
	"fmt"
	"github.com/rhysd/notes-cli"
	"os"
)

func exit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "notes: error: %s\n", err.Error())
		os.Exit(110)
	}
	os.Exit(0)
}

func main() {
	c, err := notes.ParseCmd(os.Args[1:])
	if err != nil {
		exit(err)
	}
	exit(c.Do())
}
