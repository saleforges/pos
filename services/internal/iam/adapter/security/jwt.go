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
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
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
	SessionID   string              `json:"sid,omitempty"`
	UserID      int64               `json:"user_id"`
	RoleName    string              `json:"role_name,omitempty"`
	UserType    string              `json:"user_type,omitempty"`
	Permissions []domain.Permission `json:"permissions,omitempty"`
	Type        string              `json:"type"`
	jwt.RegisteredClaims
}

type jwtRefreshClaims struct {
	UserID    int64  `json:"user_id"`
	SessionID string `json:"sid"`
	Type      string `json:"type"`
	jwt.RegisteredClaims
}

func (s *JWTSigner) SignAccessToken(claims port.TokenClaims) (string, error) {
	c := jwtAccessClaims{
		SessionID:   claims.SessionID,
		UserID:      claims.UserID,
		RoleName:    claims.RoleName,
		UserType:    string(claims.UserType),
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

func (s *JWTSigner) SignRefreshToken(userID int64, sessionID string) (string, error) {
	c := jwtRefreshClaims{
		UserID:    userID,
		SessionID: sessionID,
		Type:      "refresh",
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
		SessionID:   claims.SessionID,
		UserID:      claims.UserID,
		RoleName:    claims.RoleName,
		UserType:    domain.UserType(claims.UserType),
		Permissions: claims.Permissions,
	}, nil
}

func (s *JWTSigner) VerifyRefreshToken(tokenString string) (int64, string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtRefreshClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, domain.ErrInvalidRefreshToken
		}
		if kid, _ := t.Header["kid"].(string); kid != "" && kid != s.keyID {
			return nil, domain.ErrInvalidRefreshToken
		}
		return s.publicKey, nil
	})
	if err != nil {
		return 0, "", domain.ErrInvalidRefreshToken
	}

	claims, ok := token.Claims.(*jwtRefreshClaims)
	if !ok || !token.Valid {
		return 0, "", domain.ErrInvalidRefreshToken
	}

	if claims.Type != "refresh" {
		return 0, "", domain.ErrInvalidRefreshToken
	}

	return claims.UserID, claims.SessionID, nil
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
