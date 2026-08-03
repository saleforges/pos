package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/inventory/domain"
)

type StockRepository interface {
	Create(ctx context.Context, stock *domain.Stock) error
	GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Stock, error)
	List(ctx context.Context, merchantID int64) ([]domain.Stock, error)
	Update(ctx context.Context, stock *domain.Stock) error
}
