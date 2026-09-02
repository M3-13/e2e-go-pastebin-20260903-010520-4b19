package idgen

import (
	"crypto/rand"
)

const (
	alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	idLength = 22
)

// NewID returns a 22-character URL-safe identifier generated from a
// cryptographically secure random source (crypto/rand). 22 characters drawn
// from a 64-symbol alphabet carry 22*6 = 132 bits of entropy (>= 128 bits).
// Errors from rand.Read are returned to the caller.
func NewID() (string, error) {
	b := make([]byte, idLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b), nil
}
