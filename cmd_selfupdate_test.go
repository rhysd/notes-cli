package notes

import (
	"bytes"
	"github.com/fatih/color"
	"os"
	"strings"
	"testing"
)

func maySkipTestForSelfupdate(t *testing.T) {
	onCI := false
	if _, ok := os.LookupEnv("TRAVIS"); ok {
		onCI = true
	}
	if _, ok := os.LookupEnv("APPVEYOR"); ok {
		onCI = true
	}
	if onCI && os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("Skipping tests for selfupdate on CI without $GITHUB_TOKEN since GitHub API almost expires on CI without API token")
	}
}

func TestSelfupdateUpdateToLatest(t *testing.T) {
	maySkipTestForSelfupdate(t)

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

func TestSelfupdateCurrentIsLatest(t *testing.T) {
	maySkipTestForSelfupdate(t)

	var buf bytes.Buffer
	cmd := &SelfupdateCmd{
		Dry: true,
		Out: &buf,
	}

	if err := cmd.Do(); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Current version is the latest") {
		t.Fatal("Unexpected output:", out)
	}
}
