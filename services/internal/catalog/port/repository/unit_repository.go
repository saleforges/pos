package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type UnitRepository interface {
	GetAll(ctx context.Context) ([]domain.Unit, error)
	GetByID(ctx context.Context, id int64) (*domain.Unit, error)
	GetByCode(ctx context.Context, code string) (*domain.Unit, error)
}
