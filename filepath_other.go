//go:build !darwin
// +build !darwin

package notes

func normPathNFD(path string) string {
	// No need to normalize path string other than macOS
	return path
}
