package domain

import "time"

type StaffRole string

const (
	StaffRoleManager    StaffRole = "manager"
	StaffRoleSupervisor StaffRole = "supervisor"
	StaffRoleCashier    StaffRole = "cashier"
	StaffRoleViewer     StaffRole = "viewer"
)

type StaffStatus string

const (
	StaffStatusActive   StaffStatus = "active"
	StaffStatusInactive StaffStatus = "inactive"
)

type StaffMember struct {
	ID         int64       `json:"id"`
	MerchantID int64       `json:"merchantId"`
	BranchID   int64       `json:"branchId"`
	UserID     int64       `json:"userId"`
	Role       StaffRole   `json:"role"`
	Status     StaffStatus `json:"status"`
	IsDefault  bool        `json:"isDefault"`
	CreatedAt  time.Time   `json:"createdAt"`
	UpdatedAt  time.Time   `json:"updatedAt"`
}
