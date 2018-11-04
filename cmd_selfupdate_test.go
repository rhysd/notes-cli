package notes

import (
	"bytes"
	"github.com/fatih/color"
	"strings"
	"testing"
)

func TestSelfupdate(t *testing.T) {
	oldV := Version
	oldC := color.NoColor
	defer func() {
		Version = oldV
		color.NoColor = oldC
	}()
	Version = "0.0.1"
	color.NoColor = true

	var buf bytes.Buffer
	cmd := &SelfupdateCmd{
		Dry: true,
		Out: &buf,
	}

	// Check only dry-run since selfupdate replaces test executable and it maybe have some
	// unexpected side effects.

	if err := cmd.Do(); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "New version v") {
		t.Error("New version is not output", out)
	}
	if !strings.Contains(out, "Release Note:") {
		t.Error("Release note is not output", out)
	}
}
