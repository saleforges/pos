package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
)

type unitUsecase struct {
	unitRepo repository.UnitRepository
}

func NewUnitUsecase(unitRepo repository.UnitRepository) UnitUsecase {
	return &unitUsecase{unitRepo: unitRepo}
}

func (uc *unitUsecase) List(ctx context.Context) ([]domain.Unit, error) {
	return uc.unitRepo.GetAll(ctx)
}

var _ UnitUsecase = (*unitUsecase)(nil)
