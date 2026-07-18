package security

import (
	"crypto/sha256"
	"fmt"
)

// SHA256TokenHasher hashes tokens using SHA-256.
type SHA256TokenHasher struct{}

func NewSHA256TokenHasher() *SHA256TokenHasher {
	return &SHA256TokenHasher{}
}

func (h *SHA256TokenHasher) HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum)
}
