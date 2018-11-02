package notes

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCategoriesCmd(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	cfg := &Config{
		HomePath: filepath.Join(cwd, "testdata", "list", "normal"),
	}

	var buf bytes.Buffer
	cmd := CategoriesCmd{
		Config: cfg,
		Out:    &buf,
	}

	if err := cmd.Do(); err != nil {
		t.Fatal(err)
	}

	have := buf.String()
	want := "a\nb\nc\nd\n"
	if have != want {
		t.Fatal("wanted:", want, "but have", have)
	}
}

func TestCategoriesCmdError(t *testing.T) {
	cfg := &Config{
		HomePath: filepath.FromSlash("/path/to/somewhere/unknown/home"),
	}

	cmd := CategoriesCmd{
		Config: cfg,
	}

	err := cmd.Do()
	if err == nil {
		t.Fatal("Error did not occur")
	}
	if !strings.Contains(err.Error(), "Cannot read notes-cli home") {
		t.Fatal("Unexpected error:", err)
	}
}
