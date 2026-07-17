package security

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/argon2"
)

var errPasswordMismatch = errors.New("password mismatch")

type Argon2Hasher struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
	saltLen int
}

func NewArgon2Hasher() *Argon2Hasher {
	return &Argon2Hasher{
		time:    3,
		memory:  64 * 1024,
		threads: 4,
		keyLen:  32,
		saltLen: 16,
	}
}

func (h *Argon2Hasher) Hash(password string) (string, error) {
	salt := make([]byte, h.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, h.time, h.memory, h.threads, h.keyLen)

	var buf bytes.Buffer
	buf.Write(salt)
	buf.Write(hash)

	return base64.RawStdEncoding.EncodeToString(buf.Bytes()), nil
}

func (h *Argon2Hasher) Compare(hashedPassword, password string) error {
	decoded, err := base64.RawStdEncoding.DecodeString(hashedPassword)
	if err != nil {
		return err
	}

	if len(decoded) < h.saltLen {
		return errors.New("invalid hash format")
	}

	salt := decoded[:h.saltLen]
	expectedHash := decoded[h.saltLen:]

	computedHash := argon2.IDKey([]byte(password), salt, h.time, h.memory, h.threads, h.keyLen)

	if !bytes.Equal(expectedHash, computedHash) {
		return errPasswordMismatch
	}

	return nil
}
