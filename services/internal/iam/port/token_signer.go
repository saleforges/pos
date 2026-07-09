package port

import "github.com/saleforge/pos/services/internal/iam/domain"

type TokenClaims struct {
	UserID      string                   `json:"user_id"`
	UserType    domain.UserType          `json:"user_type"`
	Roles       []string                 `json:"roles"`
	MerchantID  string                   `json:"merchant_id,omitempty"`
	Staff       []domain.StaffAssignment `json:"staff,omitempty"`
	Permissions []domain.Permission      `json:"permissions"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type TokenSigner interface {
	SignAccessToken(claims TokenClaims) (string, error)
	SignRefreshToken(userID string) (string, error)
	VerifyAccessToken(tokenString string) (*TokenClaims, error)
	VerifyRefreshToken(tokenString string) (string, error)
}
