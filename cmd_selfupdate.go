package notes

import (
	"fmt"
	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"os"
)

// SelfupdateCmd represents `notes selfupdate` subcommand.
type SelfupdateCmd struct {
	cli *kingpin.CmdClause
	// Dry is a flag equivalent to --dry
	Dry bool
	// Slug is a "user/repo" string where releases are put. This field is useful when you forked
	// notes-cli into your own repository.
	Slug string
	// Out is a writer to write output of this command. Kind of stdout is expected
	Out io.Writer
}

func (cmd *SelfupdateCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("selfupdate", "Update myself to the latest version")
	cmd.cli.Flag("dry", "Dry run update. Only check the newer version is available").Short('d').BoolVar(&cmd.Dry)
}

func (cmd *SelfupdateCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline
}

// Do method checks the newer version binary. If new version is available, it updates myself with
// the latest binary.
func (cmd *SelfupdateCmd) Do() error {
	slug := cmd.Slug
	if slug == "" {
		slug = "rhysd/notes-cli"
	}

	latest, found, err := selfupdate.DetectLatest(slug)
	if err != nil {
		return errors.Wrap(err, "Cannot detect version from GitHub")
	}

	v := semver.MustParse(Version)
	if !found || latest.Version.LTE(v) {
		fmt.Fprintln(cmd.Out, "Current version is the latest")
		return nil
	}

	if !cmd.Dry {
		exe, err := os.Executable()
		if err != nil {
			return errors.Wrap(err, "Cannot get path to executable to update")
		}
		if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
			return err
		}
	}

	yellow.Fprintf(cmd.Out, "New version v%s\n\n", latest.Version)
	fmt.Fprintf(cmd.Out, "Release Note:\n%s\n", latest.ReleaseNotes)
	return nil
}
