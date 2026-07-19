package security

import (
	"testing"
)

func TestArgon2Hasher_HashAndCompare(t *testing.T) {
	t.Parallel()

	h := NewArgon2Hasher()

	t.Run("hash produces valid output", func(t *testing.T) {
		hash, err := h.Hash("SecurePass123")
		if err != nil {
			t.Fatalf("Hash failed: %v", err)
		}
		if hash == "" {
			t.Fatal("expected non-empty hash")
		}
		t.Logf("hash length: %d", len(hash))
	})

	t.Run("compare with correct password succeeds", func(t *testing.T) {
		password := "MyP@ssw0rd!"
		hash, err := h.Hash(password)
		if err != nil {
			t.Fatalf("Hash failed: %v", err)
		}

		if err := h.Compare(hash, password); err != nil {
			t.Errorf("Compare failed for correct password: %v", err)
		}
	})

	t.Run("compare with wrong password fails", func(t *testing.T) {
		hash, err := h.Hash("RealPass123")
		if err != nil {
			t.Fatalf("Hash failed: %v", err)
		}

		if err := h.Compare(hash, "WrongPass456"); err == nil {
			t.Error("expected error for wrong password, got nil")
		}
	})

	t.Run("different hashes for same password (random salt)", func(t *testing.T) {
		h1, _ := h.Hash("SamePassword")
		h2, _ := h.Hash("SamePassword")
		if h1 == h2 {
			t.Error("expected different hashes due to random salt")
		}
	})

	t.Run("compare with invalid base64 fails", func(t *testing.T) {
		if err := h.Compare("invalid-base64!!!", "password"); err == nil {
			t.Error("expected error for invalid hash format")
		}
	})

	t.Run("compare with too-short hash fails", func(t *testing.T) {
		if err := h.Compare("c2hvcnQ=", "password"); err == nil {
			t.Error("expected error for too-short hash")
		}
	})
}
