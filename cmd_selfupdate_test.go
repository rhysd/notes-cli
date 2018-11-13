package notes

import (
	"bytes"
	"github.com/fatih/color"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

func testCheckOnCI() bool {
	if _, ok := os.LookupEnv("TRAVIS"); ok {
		return true
	}
	if _, ok := os.LookupEnv("APPVEYOR"); ok {
		return true
	}
	return false
}

func maySkipTestForSelfupdate(t *testing.T) {
	if testCheckOnCI() && os.Getenv("GITHUB_TOKEN") == "" {
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

func TestSelfupdateUpdateSuccessfully(t *testing.T) {
	maySkipTestForSelfupdate(t)

	if !testCheckOnCI() {
		t.Skip("Run test for successful selfupdate only on CI since it's one-shot due to replacing test binary")
	}

	exe, err := os.Executable()
	panicIfErr(err)

	var buf bytes.Buffer
	cmd := &SelfupdateCmd{
		Slug: "rhysd-test/notes-cli.test",
		Out:  &buf,
	}

	if err := cmd.Do(); err != nil {
		t.Fatal(err)
	}

	r := regexp.MustCompile(`New version v(\d+\.\d+\.\d+)`)
	sub := r.FindSubmatch(buf.Bytes())
	if sub == nil {
		t.Fatal("Version did not update:", buf.String())
	}
	v := string(sub[1])

	b, err := exec.Command(exe, "--version").CombinedOutput()
	out := string(b)
	if err != nil {
		t.Fatal(err, out)
	}

	if !strings.Contains(out, v) {
		t.Fatal("Unexpected version. Wanted:", v, "but have output:", out)
	}
}

func TestSelfupdateUpdateError(t *testing.T) {
	maySkipTestForSelfupdate(t)

	if !testCheckOnCI() {
		t.Skip("Run test for selfupdate failure only on CI since takes time")
	}

	oldV := Version
	defer func() {
		Version = oldV
	}()
	Version = "0.0.1"

	var buf bytes.Buffer
	cmd := &SelfupdateCmd{Out: &buf}
	err := cmd.Do()
	if err == nil {
		t.Fatal("Error did not occur")
	}
	if !strings.Contains(err.Error(), "the command is not found in") {
		t.Fatal("Unexpected error:", err)
	}
}
