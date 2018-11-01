package notes

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

type Cmd interface {
	Do() error
	defineCLI(*kingpin.Application)
	matchesCmdline(string) bool
}
