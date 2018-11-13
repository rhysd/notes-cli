package notes

import (
	"github.com/rhysd/go-fakeio"
	"os/exec"
	"strings"
	"testing"
)

func TestOpenEditor(t *testing.T) {
	fake := fakeio.Stdout()
	defer fake.Restore()

	exe, err := exec.LookPath("echo")
	panicIfErr(err)

	cfg := &Config{
		EditorPath: exe,
		HomePath:   ".",
	}

	if err := openEditor(cfg, "foo", "bar"); err != nil {
		t.Fatal(err)
	}

	have, err := fake.String()
	if err != nil {
		panic(err)
	}

	if have != "foo bar\n" {
		t.Fatal("Arguments are unexpected:", have)
	}
}

func TestEditorPathNotSet(t *testing.T) {
	err := openEditor(&Config{})
	if err == nil {
		t.Fatal("Error did not occur")
	}
	if !strings.Contains(err.Error(), "Editor is not set") {
		t.Fatal("Unexpected error:", err)
	}
}

func TestEditorExitError(t *testing.T) {
	exe, err := exec.LookPath("false")
	panicIfErr(err)

	err = openEditor(&Config{EditorPath: exe})
	if err == nil {
		t.Fatal("Error did not occur")
	}
	if !strings.Contains(err.Error(), "Editor command did not exit successfully") {
		t.Fatal("Unexpected error:", err)
	}
}
