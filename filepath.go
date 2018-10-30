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
	if !strings.HasPrefix(path, u.HomeDir) {
		return path
	}
	canon := strings.TrimPrefix(path, u.HomeDir)
	sep := string(filepath.Separator)
	if !strings.HasPrefix(canon, sep) {
		canon = sep + canon
	}
	return "~" + canon
}
