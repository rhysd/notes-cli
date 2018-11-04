package notes

import (
	"bytes"
	"github.com/fatih/color"
	"os"
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

	onCI := false
	if _, ok := os.LookupEnv("TRAVIS"); ok {
		onCI = true
	}
	if _, ok := os.LookupEnv("APPVEYOR"); ok {
		onCI = true
	}
	if onCI && os.Getenv("GITHUB_TOKEN") == "" {
		// Skip this test since GitHub API token is not set.
		// On CI, GitHub API almost expires without API token.
		return
	}

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
