package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// HMAC256TokenHasher hashes tokens using HMAC-SHA-256 with a server secret.
// This prevents brute-force attacks if the token hash store is leaked.
type HMAC256TokenHasher struct {
	secret []byte
}

func NewHMAC256TokenHasher(secret []byte) *HMAC256TokenHasher {
	if len(secret) == 0 {
		secret = []byte("default-iam-token-hasher-secret")
	}
	return &HMAC256TokenHasher{secret: secret}
}

func (h *HMAC256TokenHasher) HashToken(token string) string {
	mac := hmac.New(sha256.New, h.secret)
	mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}
