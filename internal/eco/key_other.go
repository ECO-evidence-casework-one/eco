//go:build !windows

package eco

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"os"
)

// Test/development fallback for non-Windows hosts. The Windows build uses DPAPI.
func fallbackKey() []byte {
	h := sha256.Sum256([]byte("ECO-NONWINDOWS-TEST-KEY:" + os.Getenv("USER")))
	return h[:]
}
func protectLocalKey(in []byte) ([]byte, error) {
	b, _ := aes.NewCipher(fallbackKey())
	g, _ := cipher.NewGCM(b)
	n := make([]byte, g.NonceSize())
	rand.Read(n)
	return append(n, g.Seal(nil, n, in, []byte("ECO-TEST"))...), nil
}
func unprotectLocalKey(in []byte) ([]byte, error) {
	b, _ := aes.NewCipher(fallbackKey())
	g, _ := cipher.NewGCM(b)
	if len(in) < g.NonceSize() {
		return nil, errors.New("short protected key")
	}
	return g.Open(nil, in[:g.NonceSize()], in[g.NonceSize():], []byte("ECO-TEST"))
}
