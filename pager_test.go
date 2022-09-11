package notes

import (
	"bytes"
	"io"
	"os/exec"
	"strings"
	"testing"
)

func TestPagerCmdParseFail(t *testing.T) {
	if _, err := StartPagerWriter("'foo", io.Discard); err == nil || !strings.Contains(err.Error(), "Cannot parsing") {
		t.Fatal("Unexpected error", err)
	}
}

func TestPagerStartErrorPropagation(t *testing.T) {
	var errs []error

	w, err := StartPagerWriter("/path/to/bin/unknown", io.Discard)
	errs = append(errs, err)
	_, err = w.Write([]byte("hello"))
	errs = append(errs, err)
	err = w.Wait()
	errs = append(errs, err)

	for i, err := range errs {
		if err == nil || !strings.Contains(err.Error(), "Cannot start pager command") {
			t.Fatal("Unexpected error", err, "at index", i)
		}
	}
}

func TestPagerRunGivenCommand(t *testing.T) {
	if _, err := exec.LookPath("less"); err != nil {
		t.Skip("`less` command is necessary to run this test")
	}

	var buf bytes.Buffer

	w, err := StartPagerWriter("less -X", &buf)
	if err != nil {
		t.Fatal(err)
	}

	if w.Cmdline != "less -X" {
		t.Fatal("Cmdline is not set correctly:", w.Cmdline)
	}

	if _, err := w.Write([]byte("hello, world!")); err != nil {
		t.Fatal(err)
	}

	if err := w.Wait(); err != nil {
		t.Fatal(err)
	}

	have := buf.String()
	if have != "hello, world!" {
		t.Fatal("Stdin to pager is not piped correctly. `cat` output:", have)
	}
}

func TestPagerRunGivenMultipleArguments(t *testing.T) {
	var buf bytes.Buffer

	w, err := StartPagerWriter("echo hello 'and goodbye' \"world\"", &buf)
	if err != nil {
		t.Fatal(err)
	}

	if err := w.Wait(); err != nil {
		t.Fatal(err)
	}

	have := buf.String()
	if have != "hello and goodbye world\n" {
		t.Fatal("Multiple arguments in pagerCmd was not parsed as intended. `echo` output:", have)
	}
}

func TestPagerCommandExitFailure(t *testing.T) {
	w, err := StartPagerWriter("false", io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Wait(); err == nil {
		t.Fatal("Error did not occur")
	}
}
