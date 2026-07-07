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
	ID         string      `json:"id"`
	MerchantID string      `json:"merchant_id"`
	BranchID   string      `json:"branch_id"`
	UserID     string      `json:"user_id"`
	Role       StaffRole   `json:"role"`
	Status     StaffStatus `json:"status"`
	IsDefault  bool        `json:"is_default"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}
