package domain

type UserType string

const (
	UserTypePlatform UserType = "platform"
	UserTypeMerchant UserType = "merchant"
)

type StaffInfo struct {
	MerchantID   string `json:"merchant_id"`
	MerchantName string `json:"merchant_name"`
	Role         string `json:"role"`
}

type StaffAssignment struct {
	MerchantID string `json:"merchant_id"`
	BranchID   string `json:"branch_id,omitempty"`
	Role       string `json:"role"`
}
