package notes

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"path/filepath"
	"time"
)

// SaveCmd represents `notes save` command. Each public fields represent options of the command
type SaveCmd struct {
	cli    *kingpin.CmdClause
	Config *Config
	// Message is a message of Git commit which will be created to save notes. If this value is empty,
	// automatically generated message will be used.
	Message string
}

func (cmd *SaveCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("save", "Save notes using Git. It adds all notes and creates a commit to Git repository at home directory")
	cmd.cli.Flag("message", "Commit message on save. If omitted, an automatic message will be used").Short('m').StringVar(&cmd.Message)
}

func (cmd *SaveCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline
}

// Do runs `notes save` command and returns an error if occurs
func (cmd *SaveCmd) Do() error {
	git := NewGit(cmd.Config)
	if git == nil {
		return errors.New("'save' command cannot work without Git. Please check Git command listed in output of 'config' command is available")
	}

	if _, err := os.Stat(filepath.Join(cmd.Config.HomePath, ".git")); err != nil {
		return errors.New("'.git' directory does not exist in home. Please create a new note with `notes new` at first")
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
			return errors.Wrap(err, "Cannot push to 'origin' remote")
		}
	}

	return nil
}
