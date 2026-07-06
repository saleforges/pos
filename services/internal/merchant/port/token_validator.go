package port

type TokenClaims struct {
	UserID      string   `json:"user_id"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

type TokenValidator interface {
	Validate(tokenString string) (*TokenClaims, error)
}
