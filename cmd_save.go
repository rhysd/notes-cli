package notes

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"time"
)

type SaveCmd struct {
	cli     *kingpin.CmdClause
	Config  *Config
	Message string
}

func (cmd *SaveCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("save", "Save memo data with Git")
	cmd.cli.Flag("message", "Commit message on save").Short('m').StringVar(&cmd.Message)
}

func (cmd *SaveCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline
}

func (cmd *SaveCmd) Do() error {
	git := NewGit(cmd.Config)
	if git == nil {
		return errors.New("'save' command cannot work without Git. Please check Git command listed in output of 'config' command is available")
	}

	if err := git.AddAll(); err != nil {
		return err
	}

	msg := cmd.Message
	if msg == "" {
		// TODO: More helpful commit message for future git-grep
		msg = fmt.Sprintf("Saved by notes CLI at %s", time.Now().Format(time.RFC3339))
	}
	if err := git.Commit(msg); err != nil {
		return err
	}

	if remote, branch, err := git.TrackingRemote(); err == nil && remote == "origin" {
		if err := git.Push(remote, branch); err != nil {
			return err
		}
	}

	return nil
}
