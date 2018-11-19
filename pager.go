package notes

import (
	"github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
)

// PagerWriter is a wrapper of pager command to paging output to writer with pager command such as
// `less` or `more`. Starting process, writing input to process, waiting process finishes may cause
// an error. If one of them causes an error, later operations won't be performed. Instead, the prior
// error is returned.
type PagerWriter struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser
	// Cmdline is a string of command line which was spawned
	Cmdline string
	// Err is an error instance which occurred while paging.
	Err error
}

// Wait waits until underlying pager process finishes
func (pager *PagerWriter) Wait() error {
	if pager.Err != nil {
		return pager.Err
	}
	pager.stdin.Close()
	err := pager.cmd.Wait()
	pager.Err = err
	return err
}

// Write writes given payload to underlying pager process's stdin
func (pager *PagerWriter) Write(p []byte) (int, error) {
	if pager.Err != nil {
		return 0, pager.Err
	}
	n, err := pager.stdin.Write(p)
	pager.Err = err
	return n, err
}

// StartPagerWriter creates a new PagerWriter instance and spawns underlying pager process with given
// command. pagerCmd can contain options like "less -R". When the command cannot be parsed as shell
// arguments or starting underlying pager process fails, this function returns an error
func StartPagerWriter(pagerCmd string, stdout io.Writer) (*PagerWriter, error) {
	cmdline, err := shellquote.Split(pagerCmd)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot parsing $NOTES_CLI_PAGER as shell command line")
	}

	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = stdout

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot create pipe stdin to pager command. Please check $NOTES_CLI_PAGER")
	}

	if err = cmd.Start(); err != nil {
		err = errors.Wrapf(err, "Cannot start pager command '%s'. Please check $NOTES_CLI_PAGER is correct", pagerCmd)
	}

	return &PagerWriter{
		Cmdline: pagerCmd,
		cmd:     cmd,
		stdin:   stdin,
		Err:     err,
	}, err
}
