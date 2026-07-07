package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/merchant/domain"
)

type StaffRepository interface {
	Create(ctx context.Context, staff *domain.StaffMember) error
	GetByID(ctx context.Context, id string) (*domain.StaffMember, error)
	ListByBranch(ctx context.Context, branchID string) ([]domain.StaffMember, error)
	ListByMerchant(ctx context.Context, merchantID string) ([]domain.StaffMember, error)
	GetByUserAndBranch(ctx context.Context, userID, branchID string) (*domain.StaffMember, error)
	ListByUserAndMerchant(ctx context.Context, userID, merchantID string) ([]domain.StaffMember, error)
	SetDefaultBranch(ctx context.Context, userID, branchID string) error
	Update(ctx context.Context, staff *domain.StaffMember) error
	Delete(ctx context.Context, id string) error
}
