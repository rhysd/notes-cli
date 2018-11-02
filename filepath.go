package notes

import (
	"os/user"
	"path/filepath"
	"strings"
)

// canonPath canonicalizes given file path
func canonPath(path string) string {
	u, err := user.Current()
	if err != nil {
		return path // Give up
	}
	home := filepath.Clean(u.HomeDir)
	if !strings.HasPrefix(path, home) {
		return path
	}
	canon := strings.TrimPrefix(path, home)
	// Note: home went through filepath.Clean. So it does not have trailing slash and canon is
	// always prefixed with slash.
	return "~" + canon
}
