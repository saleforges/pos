package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/saleforge/pos/services/iam/internal/domain"
	"github.com/saleforge/pos/services/iam/internal/port"
)

const (
	defaultJWTKeyID = "iam-key-1"
	jwtIssuer       = "pos-iam"
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 30 * 24 * time.Hour
)

type JWTSigner struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	keyID      string
}

func NewJWTSigner(privateKeyPEM []byte, keyID string) (*JWTSigner, error) {
	privateKeyPEM = []byte(strings.ReplaceAll(string(privateKeyPEM), "\\n", "\n"))

	if keyID == "" {
		keyID = defaultJWTKeyID
	}

	if len(strings.TrimSpace(string(privateKeyPEM))) == 0 {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, err
		}
		return &JWTSigner{
			privateKey: privateKey,
			publicKey:  &privateKey.PublicKey,
			keyID:      keyID,
		}, nil
	}

	privateKey, err := parseRSAPrivateKey(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	return &JWTSigner{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		keyID:      keyID,
	}, nil
}

type jwtAccessClaims struct {
	UserID      string              `json:"user_id"`
	Roles       []string            `json:"roles"`
	Permissions []domain.Permission `json:"permissions"`
	Type        string              `json:"type"`
	jwt.RegisteredClaims
}

type jwtRefreshClaims struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

func (s *JWTSigner) SignAccessToken(claims port.TokenClaims) (string, error) {
	c := jwtAccessClaims{
		UserID:      claims.UserID,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
		Type:        "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    jwtIssuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, c)
	token.Header["kid"] = s.keyID
	return token.SignedString(s.privateKey)
}

func (s *JWTSigner) SignRefreshToken(userID string) (string, error) {
	c := jwtRefreshClaims{
		UserID: userID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    jwtIssuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, c)
	token.Header["kid"] = s.keyID
	return token.SignedString(s.privateKey)
}

func (s *JWTSigner) VerifyAccessToken(tokenString string) (*port.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtAccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, domain.ErrInvalidToken
		}
		if kid, _ := t.Header["kid"].(string); kid != "" && kid != s.keyID {
			return nil, domain.ErrInvalidToken
		}
		return s.publicKey, nil
	})
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwtAccessClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	if claims.Type != "access" {
		return nil, domain.ErrInvalidToken
	}

	return &port.TokenClaims{
		UserID:      claims.UserID,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
	}, nil
}

func (s *JWTSigner) VerifyRefreshToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtRefreshClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return "", domain.ErrInvalidRefreshToken
		}
		if kid, _ := t.Header["kid"].(string); kid != "" && kid != s.keyID {
			return "", domain.ErrInvalidRefreshToken
		}
		return s.publicKey, nil
	})
	if err != nil {
		return "", domain.ErrInvalidRefreshToken
	}

	claims, ok := token.Claims.(*jwtRefreshClaims)
	if !ok || !token.Valid {
		return "", domain.ErrInvalidRefreshToken
	}

	if claims.Type != "refresh" {
		return "", domain.ErrInvalidRefreshToken
	}

	return claims.UserID, nil
}

func (s *JWTSigner) JWKS() port.JSONWebKeySet {
	return port.JSONWebKeySet{
		Keys: []port.JSONWebKey{
			{
				Kty: "RSA",
				Use: "sig",
				Alg: "RS256",
				Kid: s.keyID,
				N:   base64.RawURLEncoding.EncodeToString(s.publicKey.N.Bytes()),
				E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(s.publicKey.E)).Bytes()),
			},
		},
	}
}

func parseRSAPrivateKey(privateKeyPEM []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, errors.New("invalid rsa private key pem")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("pem is not an rsa private key")
	}

	return rsaKey, nil
}
