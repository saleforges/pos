package port

type TokenClaims struct {
	UserID       string   `json:"user_id"`
	UserType     string   `json:"user_type"`
	Roles        []string `json:"roles"`
	MerchantID   string   `json:"merchant_id,omitempty"`
	MerchantRole string   `json:"merchant_role,omitempty"`
	Permissions  []string `json:"permissions"`
}

type TokenValidator interface {
	Validate(tokenString string) (*TokenClaims, error)
}
