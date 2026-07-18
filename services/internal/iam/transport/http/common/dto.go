package common

type UserResponse struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type MeResponse struct {
	ID        int64          `json:"id"`
	Username  string         `json:"username"`
	Email     string         `json:"email"`
	Type      string         `json:"type"`
	Status    string         `json:"status"`
	Roles     []RoleResponse `json:"roles"`
	CreatedAt string         `json:"createdAt"`
	UpdatedAt string         `json:"updatedAt"`
}

type MerchantDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type BranchDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type RoleResponse struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"`
	Merchant    *MerchantDTO `json:"merchant"`
	Branch      *BranchDTO   `json:"branch"`
	BranchScope string       `json:"branchScope"`
	IsDefault   bool         `json:"isDefault"`
}
