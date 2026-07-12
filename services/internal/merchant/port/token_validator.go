package port

type StaffAssignment struct {
	MerchantID int64  `json:"merchant_id"`
	BranchID   int64  `json:"branch_id,omitempty"`
	Role       string `json:"role"`
}

type TokenClaims struct {
	UserID       int64             `json:"user_id"`
	UserType     string            `json:"user_type"`
	Roles        []string          `json:"roles"`
	MerchantID   int64             `json:"merchant_id,omitempty"`
	Staff        []StaffAssignment `json:"staff,omitempty"`
	Permissions  []string          `json:"permissions"`
}

type TokenValidator interface {
	Validate(tokenString string) (*TokenClaims, error)
}
