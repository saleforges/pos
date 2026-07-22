package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/pkg/pagination"
)

// StaffUsecase defines staff assignment operations.
type StaffUsecase interface {
	AssignStaff(ctx context.Context, input AssignStaffParams) (*domain.StaffMember, error)
	GetStaff(ctx context.Context, id int64) (*domain.StaffMember, error)
	ListStaffByBranch(ctx context.Context, branchID int64, p pagination.Params) ([]domain.StaffMember, *pagination.Metadata, error)
	ListStaffByMerchant(ctx context.Context, merchantID int64, p pagination.Params) ([]domain.StaffMember, *pagination.Metadata, error)
	GetMyStaffAssignments(ctx context.Context, userID, merchantID int64) ([]domain.StaffMember, error)
	SetMyDefaultBranch(ctx context.Context, userID, branchID int64) error
	UpdateStaff(ctx context.Context, input UpdateStaffParams) (*domain.StaffMember, error)
	RemoveStaff(ctx context.Context, id int64) error
}
