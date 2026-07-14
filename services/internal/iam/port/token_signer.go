package port

import "github.com/saleforge/pos/services/internal/iam/domain"

type TokenClaims struct {
	Subject     string              `json:"sub,omitempty"`
	SessionID   string              `json:"sid,omitempty"`
	RoleID      int64               `json:"rid,omitempty"`
	MerchantID  int64               `json:"mid,omitempty"`
	BranchID    int64               `json:"bid,omitempty"`
	UserID      int64               `json:"user_id"`
	UserType    domain.UserType     `json:"user_type"`
	RoleName    string              `json:"role_name"`
	Permissions []domain.Permission `json:"permissions"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type TokenSigner interface {
	SignAccessToken(claims TokenClaims) (string, error)
	SignRefreshToken(userID int64, sessionID string) (string, error)
	VerifyAccessToken(tokenString string) (*TokenClaims, error)
	VerifyRefreshToken(tokenString string) (int64, string, error)
}
