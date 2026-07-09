package port

type StaffAssignment struct {
	MerchantID string `json:"merchant_id"`
	BranchID   string `json:"branch_id,omitempty"`
	Role       string `json:"role"`
}

type TokenClaims struct {
	UserID       string            `json:"user_id"`
	UserType     string            `json:"user_type"`
	Roles        []string          `json:"roles"`
	MerchantID   string            `json:"merchant_id,omitempty"`
	Staff        []StaffAssignment `json:"staff,omitempty"`
	Permissions  []string          `json:"permissions"`
}

type TokenValidator interface {
	Validate(tokenString string) (*TokenClaims, error)
}
