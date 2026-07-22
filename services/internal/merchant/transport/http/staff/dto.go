package staff

import "github.com/saleforge/pos/services/internal/merchant/domain"

type assignStaffReq struct {
	BranchID  int64           `json:"branchId"`
	UserID    int64           `json:"userId"`
	Role      domain.StaffRole `json:"role"`
	IsDefault bool            `json:"isDefault"`
}

type updateStaffReq struct {
	Role   *domain.StaffRole   `json:"role,omitempty"`
	Status *domain.StaffStatus `json:"status,omitempty"`
}

type setDefaultBranchReq struct {
	BranchID int64 `json:"branchId"`
}
