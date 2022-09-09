//go:build darwin
// +build darwin

package notes

import (
	"golang.org/x/text/unicode/norm"
)

func normPathNFD(path string) string {
	return norm.NFC.String(path)
}
