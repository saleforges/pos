package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type UnitUsecase interface {
	List(ctx context.Context) ([]domain.Unit, error)
}
