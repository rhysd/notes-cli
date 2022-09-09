//go:build darwin
// +build darwin

package notes

import (
	"testing"
)

func TestNormalizePathNFD(t *testing.T) {
	// 'カテゴリ' in NFD
	nfd := string([]byte{227, 130, 171, 227, 131, 134, 227, 130, 179, 227, 130, 153, 227, 131, 170})
	// 'カテゴリ' in NFC
	want := string([]byte{227, 130, 171, 227, 131, 134, 227, 130, 180, 227, 131, 170})
	have := normPathNFD(nfd)
	if want != have {
		t.Fatal("String was not normalized:", []byte(have))
	}
}
