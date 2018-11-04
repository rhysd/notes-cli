package notes

import (
	"github.com/blang/semver"
	"testing"
)

func TestVersion(t *testing.T) {
	_, err := semver.Parse(Version)
	if err != nil {
		t.Fatal(err)
	}
}
