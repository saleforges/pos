package repository

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/order/domain"
)

type CustomerRepository interface {
	Create(ctx context.Context, customer *domain.Customer) error
	GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Customer, error)
	List(ctx context.Context, merchantID int64, search string) ([]domain.Customer, error)
	ListChangedSince(ctx context.Context, merchantID int64, since *time.Time) ([]domain.Customer, error)
	Update(ctx context.Context, customer *domain.Customer) error
	Delete(ctx context.Context, id int64, merchantID int64) error
}
