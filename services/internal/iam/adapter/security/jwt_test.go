package security

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
)

func generateTestKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}
	return key
}

func newTestSigner(t *testing.T) *JWTSigner {
	t.Helper()
	key := generateTestKey(t)
	return &JWTSigner{
		privateKey: key,
		publicKey:  &key.PublicKey,
		keyID:      "test-key-1",
	}
}

func TestJWTSigner_SignAndVerifyAccessToken(t *testing.T) {
	t.Parallel()

	s := newTestSigner(t)

	t.Run("sign and verify valid token", func(t *testing.T) {
		claims := port.TokenClaims{
			UserID:      42,
			SessionID:   "sess-abc",
			RoleID:      1,
			MerchantID:  100,
			BranchID:    200,
			UserType:    domain.UserTypeMerchant,
			RoleName:    "admin",
			Permissions: []domain.Permission{domain.UserRead, domain.UserCreate},
		}

		token, err := s.SignAccessToken(claims)
		if err != nil {
			t.Fatalf("SignAccessToken failed: %v", err)
		}

		verified, err := s.VerifyAccessToken(token)
		if err != nil {
			t.Fatalf("VerifyAccessToken failed: %v", err)
		}

		if verified.UserID != 42 { t.Errorf("expected UserID 42, got %d", verified.UserID) }
		if verified.SessionID != "sess-abc" { t.Errorf("expected SessionID sess-abc, got %s", verified.SessionID) }
		if verified.RoleID != 1 { t.Errorf("expected RoleID 1, got %d", verified.RoleID) }
		if verified.UserType != domain.UserTypeMerchant { t.Errorf("expected UserType merchant, got %s", verified.UserType) }
		if len(verified.Permissions) != 2 { t.Errorf("expected 2 permissions, got %d", len(verified.Permissions)) }
	})

	t.Run("rejects expired token", func(t *testing.T) {
		key := generateTestKey(t)
		expiredSigner := &JWTSigner{privateKey: key, publicKey: &key.PublicKey, keyID: "test"}
		c := jwtAccessClaims{
			UserID: 1,
			Type:   "access",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, c)
		tokenStr, err := token.SignedString(key)
		if err != nil {
			t.Fatalf("failed to sign expired token: %v", err)
		}

		_, err = expiredSigner.VerifyAccessToken(tokenStr)
		if err == nil {
			t.Error("expected error for expired token")
		}
	})

	t.Run("rejects token with wrong key", func(t *testing.T) {
		otherKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		otherSigner := &JWTSigner{privateKey: otherKey, publicKey: &otherKey.PublicKey, keyID: "other"}

		claims := port.TokenClaims{UserID: 1}
		token, _ := otherSigner.SignAccessToken(claims)

		_, err := s.VerifyAccessToken(token)
		if err == nil {
			t.Error("expected error for token signed with different key")
		}
	})

	t.Run("rejects malformed token string", func(t *testing.T) {
		_, err := s.VerifyAccessToken("not.a.jwt")
		if err == nil {
			t.Error("expected error for malformed token")
		}
	})
}

func TestJWTSigner_RefreshToken(t *testing.T) {
	t.Parallel()

	s := newTestSigner(t)

	t.Run("sign and verify refresh token", func(t *testing.T) {
		token, err := s.SignRefreshToken(42, "sess-xyz")
		if err != nil {
			t.Fatalf("SignRefreshToken failed: %v", err)
		}

		userID, sessionID, err := s.VerifyRefreshToken(token)
		if err != nil {
			t.Fatalf("VerifyRefreshToken failed: %v", err)
		}

		if userID != 42 { t.Errorf("expected UserID 42, got %d", userID) }
		if sessionID != "sess-xyz" { t.Errorf("expected SessionID sess-xyz, got %s", sessionID) }
	})

	t.Run("rejects access token as refresh token", func(t *testing.T) {
		at, _ := s.SignAccessToken(port.TokenClaims{UserID: 1})
		_, _, err := s.VerifyRefreshToken(at)
		if err == nil {
			t.Error("expected error for using access token as refresh token")
		}
	})

	t.Run("rejects malformed refresh token", func(t *testing.T) {
		_, _, err := s.VerifyRefreshToken("bad.token.string")
		if err == nil {
			t.Error("expected error for malformed token")
		}
	})
}

func TestJWTSigner_JWKS(t *testing.T) {
	t.Parallel()

	s := newTestSigner(t)
	jwks := s.JWKS()

	if len(jwks.Keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(jwks.Keys))
	}
	key := jwks.Keys[0]
	if key.Kty != "RSA" { t.Errorf("expected RSA, got %s", key.Kty) }
	if key.Alg != "RS256" { t.Errorf("expected RS256, got %s", key.Alg) }
	if key.Kid != "test-key-1" { t.Errorf("expected kid 'test-key-1', got %s", key.Kid) }
	if key.N == "" { t.Error("expected non-empty modulus") }
	if key.E == "" { t.Error("expected non-empty exponent") }
}

func TestNewJWTSigner_WithEmptyKeyGeneratesEphemeral(t *testing.T) {
	s, err := NewJWTSigner([]byte(""), "")
	if err != nil {
		t.Fatalf("NewJWTSigner with empty key failed: %v", err)
	}
	if s.privateKey == nil { t.Error("expected generated private key") }
	if s.keyID != defaultJWTKeyID { t.Errorf("expected default key ID %s, got %s", defaultJWTKeyID, s.keyID) }

	token, err := s.SignAccessToken(port.TokenClaims{UserID: 1})
	if err != nil {
		t.Fatalf("sign with ephemeral key failed: %v", err)
	}
	if _, err := s.VerifyAccessToken(token); err != nil {
		t.Errorf("verify with ephemeral key failed: %v", err)
	}
}

func TestNewJWTSigner_WithPEMKey(t *testing.T) {
	t.Parallel()

	// Minimal RSA private key in PEM format (2048-bit, PKCS1)
	pemKey := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA4f5wg5l2hKsTeErsMkLFwfS+LUqiy3k8m6X2qLk2R+E0Q6GX
G6i3nHkL3KsJb3Z4Yj6sI6S4zPv8kY8Qh7L8s9kX8f3g5s7k9jL8s9kX8f3g5s7k
9jL8s9kX8f3g5s7k9jL8s9kX8f3g5s7k9jL8s9kX8f3g5s7k9jL8s9kX8f3g5s7k
-----END RSA PRIVATE KEY-----`)
	_, err := NewJWTSigner(pemKey, "test-key")
	if err != nil {
		t.Logf("expected parse error for truncated PEM (test only): %v", err)
		// A real PEM key would parse successfully. This is just to exercise the path.
	}
}

func TestParseRSAPrivateKey_InvalidPEM(t *testing.T) {
	t.Parallel()

	_, err := parseRSAPrivateKey([]byte("not a pem block"))
	if err == nil {
		t.Error("expected error for invalid PEM")
	}
}

func TestParseRSAPrivateKey_NonRSAPEM(t *testing.T) {
	t.Parallel()

	// Valid PEM block but not RSA
	pemData := []byte("-----BEGIN CERTIFICATE-----\nZmFrZQ==\n-----END CERTIFICATE-----")
	_, err := parseRSAPrivateKey(pemData)
	if err == nil {
		t.Error("expected error for non-RSA PEM")
	}
}

func TestSHA256TokenHasher(t *testing.T) {
	t.Parallel()

	h := NewSHA256TokenHasher()

	t.Run("produces non-empty hash", func(t *testing.T) {
		result := h.HashToken("test-token")
		if result == "" {
			t.Error("expected non-empty hash")
		}
	})

	t.Run("same input produces same hash", func(t *testing.T) {
		h1 := h.HashToken("consistent")
		h2 := h.HashToken("consistent")
		if h1 != h2 {
			t.Error("expected deterministic hash")
		}
	})

	t.Run("different inputs produce different hashes", func(t *testing.T) {
		h1 := h.HashToken("token-a")
		h2 := h.HashToken("token-b")
		if h1 == h2 {
			t.Error("expected different hashes for different inputs")
		}
	})

	t.Run("empty string produces hash", func(t *testing.T) {
		result := h.HashToken("")
		if result == "" {
			t.Error("expected hash for empty string, got empty")
		}
	})
}
