package jwks

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Subject    string `json:"sub"`
	SessionID  string `json:"sid"`
	RoleID     int64  `json:"rid"`
	MerchantID int64  `json:"mid"`
	BranchID   int64  `json:"bid"`
	UserID     int64  `json:"user_id"`
	UserType   string `json:"user_type"`
	jwt.RegisteredClaims
}

type jwksResponse struct {
	Keys []struct {
		Kty string `json:"kty"`
		Use string `json:"use"`
		Alg string `json:"alg"`
		Kid string `json:"kid"`
		N   string `json:"n"`
		E   string `json:"e"`
	} `json:"keys"`
}

type Verifier struct {
	jwksURL    string
	client     *http.Client
	mu         sync.RWMutex
	keys       map[string]*rsa.PublicKey
	lastFetch  time.Time
	ttl        time.Duration
}

func New(jwksURL string) *Verifier {
	return &Verifier{
		jwksURL: jwksURL,
		client:  &http.Client{Timeout: 10 * time.Second},
		keys:    make(map[string]*rsa.PublicKey),
		ttl:     5 * time.Minute,
	}
}

func (v *Verifier) Verify(tokenString string) (*Claims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("jwt: parse unverified: %w", err)
	}
	kid, _ := token.Header["kid"].(string)

	key, err := v.fetchKey(kid)
	if err != nil {
		return nil, err
	}

	parsed, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt: verify: %w", err)
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("jwt: invalid claims")
	}
	return claims, nil
}

func (v *Verifier) fetchKey(kid string) (*rsa.PublicKey, error) {
	v.mu.RLock()
	key, ok := v.keys[kid]
	age := time.Since(v.lastFetch)
	v.mu.RUnlock()

	if ok && age < v.ttl {
		return key, nil
	}

	resp, err := v.client.Get(v.jwksURL)
	if err != nil {
		return nil, fmt.Errorf("jwks: fetch: %w", err)
	}
	defer resp.Body.Close()

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("jwks: decode: %w", err)
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	for _, k := range jwks.Keys {
		pub, err := decodePublicKey(k.N, k.E)
		if err != nil {
			continue
		}
		v.keys[k.Kid] = pub
	}
	v.lastFetch = time.Now()

	key, ok = v.keys[kid]
	if !ok {
		return nil, fmt.Errorf("jwks: key %q not found", kid)
	}
	return key, nil
}

func decodePublicKey(n, e string) (*rsa.PublicKey, error) {
	nb, err := base64.RawURLEncoding.DecodeString(n)
	if err != nil {
		return nil, err
	}
	eb, err := base64.RawURLEncoding.DecodeString(e)
	if err != nil {
		return nil, err
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nb),
		E: int(new(big.Int).SetBytes(eb).Int64()),
	}, nil
}

func PublicKeyToJWKS(pub *rsa.PublicKey, kid string) ([]byte, error) {
	key := struct {
		Kty string `json:"kty"`
		Use string `json:"use"`
		Alg string `json:"alg"`
		Kid string `json:"kid"`
		N   string `json:"n"`
		E   string `json:"e"`
	}{
		Kty: "RSA",
		Use: "sig",
		Alg: "RS256",
		Kid: kid,
		N:   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}
	return json.Marshal(struct {
		Keys []interface{} `json:"keys"`
	}{Keys: []interface{}{key}})
}

func PEMToPublicKey(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return rsaKey, nil
}
