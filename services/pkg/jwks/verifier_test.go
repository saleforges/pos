package jwks

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func generateTestKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return key
}

// signToken creates a signed JWT with the given claims and key.
func signToken(t *testing.T, key *rsa.PrivateKey, kid string, claims jwt.Claims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = kid
	s, err := tok.SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return s
}

// jwkFromPublicKey builds a single JWK entry from an RSA public key.
func jwkFromPublicKey(pub *rsa.PublicKey, kid string) map[string]string {
	return map[string]string{
		"kty": "RSA",
		"use": "sig",
		"alg": "RS256",
		"kid": kid,
		"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}
}

// jwksBody returns the JSON body for a set of JWK entries.
func jwksBody(keys ...map[string]string) []byte {
	b, _ := json.Marshal(map[string]interface{}{"keys": keys})
	return b
}

// newJWKSServer starts an httptest.Server that serves the given JWKS body.
func newJWKSServer(t *testing.T, body []byte) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}))
}

// ---------------------------------------------------------------------------
// Claims that match the production struct
// ---------------------------------------------------------------------------

type testClaims struct {
	Subject    string `json:"sub,omitempty"`
	UserID     int64  `json:"user_id"`
	Type       string `json:"type"`
	jwt.RegisteredClaims
}

// ---------------------------------------------------------------------------
// Verify
// ---------------------------------------------------------------------------

func TestVerifier_Verify_ValidToken(t *testing.T) {
	key := generateTestKey(t)
	kid := "test-key-1"
	pub := &key.PublicKey

	srv := newJWKSServer(t, jwksBody(jwkFromPublicKey(pub, kid)))
	defer srv.Close()

	v := New(srv.URL)
	v.keys[kid] = pub
	v.lastFetch = time.Now()

	token := signToken(t, key, kid, testClaims{
		UserID: 42,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})

	claims, err := v.Verify(token)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("expected UserID 42, got %d", claims.UserID)
	}
}

func TestVerifier_Verify_RejectsNonRSASigningMethod(t *testing.T) {
	key := generateTestKey(t)
	kid := "test-key-1"
	pub := &key.PublicKey

	srv := newJWKSServer(t, jwksBody(jwkFromPublicKey(pub, kid)))
	defer srv.Close()

	v := New(srv.URL)
	v.keys[kid] = pub
	v.lastFetch = time.Now()

	// Sign with HMAC instead of RSA
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, testClaims{
		UserID: 1,
		Type:   "access",
	})
	token, err := tok.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign hmac token: %v", err)
	}

	_, err = v.Verify(token)
	if err == nil {
		t.Fatal("expected error for non-RSA signing method")
	}
}

func TestVerifier_Verify_RejectsExpiredToken(t *testing.T) {
	key := generateTestKey(t)
	kid := "test-key-1"
	pub := &key.PublicKey

	srv := newJWKSServer(t, jwksBody(jwkFromPublicKey(pub, kid)))
	defer srv.Close()

	v := New(srv.URL)
	v.keys[kid] = pub
	v.lastFetch = time.Now()

	token := signToken(t, key, kid, testClaims{
		UserID: 1,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	})

	_, err := v.Verify(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestVerifier_Verify_RejectsMissingTypeClaim(t *testing.T) {
	key := generateTestKey(t)
	kid := "test-key-1"
	pub := &key.PublicKey

	srv := newJWKSServer(t, jwksBody(jwkFromPublicKey(pub, kid)))
	defer srv.Close()

	v := New(srv.URL)
	v.keys[kid] = pub
	v.lastFetch = time.Now()

	// Token without the "type" claim — the struct has `Type string` so it defaults to ""
	token := signToken(t, key, kid, testClaims{
		UserID: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	})

	_, err := v.Verify(token)
	if err == nil {
		t.Fatal("expected error for missing type claim")
	}
}

func TestVerifier_Verify_RejectsRefreshTokenType(t *testing.T) {
	key := generateTestKey(t)
	kid := "test-key-1"
	pub := &key.PublicKey

	srv := newJWKSServer(t, jwksBody(jwkFromPublicKey(pub, kid)))
	defer srv.Close()

	v := New(srv.URL)
	v.keys[kid] = pub
	v.lastFetch = time.Now()

	token := signToken(t, key, kid, testClaims{
		UserID: 1,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	})

	_, err := v.Verify(token)
	if err == nil {
		t.Fatal("expected error for refresh token type")
	}
}

func TestVerifier_Verify_RejectsWrongKey(t *testing.T) {
	key := generateTestKey(t)
	otherKey := generateTestKey(t)
	kid := "test-key-1"

	// Server serves the wrong key
	pub := &otherKey.PublicKey
	srv := newJWKSServer(t, jwksBody(jwkFromPublicKey(pub, kid)))
	defer srv.Close()

	v := New(srv.URL)

	// Pre-populate the verifier with the WRONG key so the HTTP fetch is skipped
	v.keys[kid] = pub
	v.lastFetch = time.Now()

	// Sign with the correct key — the verifier has the wrong one
	token := signToken(t, key, kid, testClaims{
		UserID: 1,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	})

	_, err := v.Verify(token)
	if err == nil {
		t.Fatal("expected error for token signed with wrong key")
	}
}

func TestVerifier_Verify_MalformedTokenString(t *testing.T) {
	key := generateTestKey(t)
	kid := "test-key-1"
	pub := &key.PublicKey

	srv := newJWKSServer(t, jwksBody(jwkFromPublicKey(pub, kid)))
	defer srv.Close()

	v := New(srv.URL)
	v.keys[kid] = pub
	v.lastFetch = time.Now()

	_, err := v.Verify("not.a.jwt")
	if err == nil {
		t.Fatal("expected error for malformed token")
	}
}

// ---------------------------------------------------------------------------
// fetchKey — resilient cache update
// ---------------------------------------------------------------------------

func TestFetchKey_ZeroValidKeysReturnsError(t *testing.T) {
	// JWKS response with a key that has invalid base64 modulus
	body := jwksBody(map[string]string{
		"kty": "RSA",
		"kid": "bad-key",
		"n":   "!!!not-valid-base64!!!",
		"e":   "AQAB",
	})
	srv := newJWKSServer(t, body)
	defer srv.Close()

	v := New(srv.URL)

	// Pre-populate a valid key to prove rollback
	oldKey := generateTestKey(t)
	v.keys["old-key"] = &oldKey.PublicKey
	oldFetch := time.Now().Add(-10 * time.Minute)
	v.lastFetch = oldFetch

	_, err := v.fetchKey("bad-key")
	if err == nil {
		t.Fatal("expected error for zero valid keys")
	}

	// Verify old cache is preserved
	if _, ok := v.keys["old-key"]; !ok {
		t.Error("old key should still be in cache after failed fetch")
	}

	// Verify lastFetch was NOT updated (should still be ~10 min ago, not recent)
	if v.lastFetch.After(time.Now().Add(-1 * time.Minute)) {
		t.Error("lastFetch should not have been updated after failed fetch")
	}
}

func TestFetchKey_RefreshesFromHTTPWhenStale(t *testing.T) {
	key := generateTestKey(t)
	kid := "live-key"
	pub := &key.PublicKey

	srv := newJWKSServer(t, jwksBody(jwkFromPublicKey(pub, kid)))
	defer srv.Close()

	v := New(srv.URL)
	v.lastFetch = time.Now().Add(-10 * time.Minute) // stale cache

	// Trigger HTTP fetch by requesting a key not in cache
	fetchedKey, err := v.fetchKey(kid)
	if err != nil {
		t.Fatalf("fetchKey failed: %v", err)
	}
	if fetchedKey == nil {
		t.Fatal("expected non-nil key")
	}
}

func TestFetchKey_UsesCachedKeyWhenFresh(t *testing.T) {
	key := generateTestKey(t)
	kid := "cached-key"
	pub := &key.PublicKey

	v := New("http://localhost:19999/does-not-exist") // would fail if HTTP called
	v.keys[kid] = pub
	v.lastFetch = time.Now()

	fetchedKey, err := v.fetchKey(kid)
	if err != nil {
		t.Fatalf("fetchKey failed: %v", err)
	}
	if fetchedKey != pub {
		t.Error("expected cached key to be returned")
	}
}

// ---------------------------------------------------------------------------
// PublicKeyToJWKS
// ---------------------------------------------------------------------------

func TestPublicKeyToJWKS(t *testing.T) {
	key := generateTestKey(t)
	kid := "my-key"

	data, err := PublicKeyToJWKS(&key.PublicKey, kid)
	if err != nil {
		t.Fatalf("PublicKeyToJWKS failed: %v", err)
	}

	var parsed struct {
		Keys []map[string]string `json:"keys"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal jwks: %v", err)
	}
	if len(parsed.Keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(parsed.Keys))
	}
	k := parsed.Keys[0]
	if k["kid"] != kid {
		t.Errorf("expected kid %q, got %q", kid, k["kid"])
	}
	if k["kty"] != "RSA" {
		t.Errorf("expected kty RSA, got %s", k["kty"])
	}
	if k["alg"] != "RS256" {
		t.Errorf("expected alg RS256, got %s", k["alg"])
	}
	if k["n"] == "" {
		t.Error("expected non-empty modulus")
	}
}

// ---------------------------------------------------------------------------
// PEMToPublicKey
// ---------------------------------------------------------------------------

func TestPEMToPublicKey(t *testing.T) {
	key := generateTestKey(t)
	pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	pemData := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	pub, err := PEMToPublicKey(pemData)
	if err != nil {
		t.Fatalf("PEMToPublicKey failed: %v", err)
	}
	if pub.N.Cmp(key.PublicKey.N) != 0 || pub.E != key.PublicKey.E {
		t.Error("decoded public key does not match original")
	}
}

func TestPEMToPublicKey_InvalidPEM(t *testing.T) {
	_, err := PEMToPublicKey([]byte("not a PEM block"))
	if err == nil {
		t.Fatal("expected error for invalid PEM")
	}
}

func TestPEMToPublicKey_NonPublicKeyPEM(t *testing.T) {
	key := generateTestKey(t)
	privBytes := x509.MarshalPKCS1PrivateKey(key)
	pemData := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})

	_, err := PEMToPublicKey(pemData)
	if err == nil {
		t.Fatal("expected error for private key PEM")
	}
}
