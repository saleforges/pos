package bootstrap

import (
	"bytes"
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

	"github.com/golang-jwt/jwt/v5"
	"github.com/saleforge/pos/services/internal/iam/port"
)

func generateTestRSAPrivateKey() string {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	return string(pem.EncodeToMemory(block))
}

func testConfig() Config {
	return Config{
		JWTPrivateKeyPEM:  generateTestRSAPrivateKey(),
		TokenHasherSecret: "test-hasher-secret",
	}
}

type apiResponse struct {
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func TestApp_HTTPFlow(t *testing.T) {
	t.Parallel()

	app, err := New(testConfig())
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}

	registerBody := []byte(`{"username":"bootstrap_user","email":"bootstrap_user@example.com","password":"Password1","roles":["admin"]}`)
	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
	registerReq.Header.Set("Content-Type", "application/json")
	registerRec := httptest.NewRecorder()

	app.router.ServeHTTP(registerRec, registerReq)

	if registerRec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, registerRec.Code)
	}

	var wrap apiResponse
	if err := json.NewDecoder(registerRec.Body).Decode(&wrap); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	if wrap.Message != "success" {
		t.Fatalf("expected success message, got %q", wrap.Message)
	}

	var regResp authResponse
	if err := json.Unmarshal(wrap.Data, &regResp); err != nil {
		t.Fatalf("decode auth data: %v", err)
	}
	if regResp.AccessToken == "" {
		t.Fatal("expected access token in register response")
	}

	jwksReq := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	jwksRec := httptest.NewRecorder()
	app.router.ServeHTTP(jwksRec, jwksReq)

	if jwksRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, jwksRec.Code)
	}

	var jwks port.JSONWebKeySet
	if err := json.NewDecoder(jwksRec.Body).Decode(&jwks); err != nil {
		t.Fatalf("decode jwks response: %v", err)
	}
	if len(jwks.Keys) != 1 {
		t.Fatalf("expected 1 jwk, got %d", len(jwks.Keys))
	}

	pub, err := jwkToRSAPublicKey(jwks.Keys[0])
	if err != nil {
		t.Fatalf("build rsa public key: %v", err)
	}

	token, _, err := new(jwt.Parser).ParseUnverified(regResp.AccessToken, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("parse token header: %v", err)
	}
	if token.Method.Alg() != "RS256" {
		t.Fatalf("expected RS256, got %s", token.Method.Alg())
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+regResp.AccessToken)
	meRec := httptest.NewRecorder()

	app.router.ServeHTTP(meRec, meReq)

	if meRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, meRec.Code)
	}

	verified, err := jwt.ParseWithClaims(regResp.AccessToken, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		return pub, nil
	})
	if err != nil {
		t.Fatalf("verify token with jwks key: %v", err)
	}
	if !verified.Valid {
		t.Fatal("expected token to be valid")
	}
}

func TestApp_LoginAndAuthFailures(t *testing.T) {
	t.Parallel()

	app, err := New(testConfig())
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}

	registerBody := []byte(`{"username":"login_user","email":"login_user@example.com","password":"Password1","roles":["viewer"]}`)
	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
	registerReq.Header.Set("Content-Type", "application/json")
	registerRec := httptest.NewRecorder()
	app.router.ServeHTTP(registerRec, registerReq)

	loginBody := []byte(`{"username":"login_user","password":"Password1"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	app.router.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, loginRec.Code)
	}

	badReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	badRec := httptest.NewRecorder()
	app.router.ServeHTTP(badRec, badReq)

	if badRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, badRec.Code)
	}
}

func TestNew_FailsWithEmptyTokenHasherSecret(t *testing.T) {
	_, err := New(Config{JWTPrivateKeyPEM: generateTestRSAPrivateKey()})
	if err == nil {
		t.Fatal("expected error for empty TokenHasherSecret, got nil")
	}
}

func jwkToRSAPublicKey(key port.JSONWebKey) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, err
	}

	pub := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(new(big.Int).SetBytes(eBytes).Int64()),
	}
	return pub, nil
}
