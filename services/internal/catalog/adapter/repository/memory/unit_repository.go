package memory

import (
	"context"
	"sync"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
)

var _ repository.UnitRepository = (*UnitRepository)(nil)

type UnitRepository struct {
	mu    sync.RWMutex
	units map[int64]*domain.Unit
}

func NewUnitRepository() *UnitRepository {
	return &UnitRepository{
		units: map[int64]*domain.Unit{
			1: {ID: 1, Code: "PCS", Name: "Piece"},
			2: {ID: 2, Code: "PACK", Name: "Pack"},
			3: {ID: 3, Code: "KG", Name: "Kilogram"},
			4: {ID: 4, Code: "GRAM", Name: "Gram"},
			5: {ID: 5, Code: "LITER", Name: "Liter"},
			6: {ID: 6, Code: "ML", Name: "Milliliter"},
			7: {ID: 7, Code: "BOX", Name: "Box"},
			8: {ID: 8, Code: "METER", Name: "Meter"},
		},
	}
}

func (r *UnitRepository) GetAll(_ context.Context) ([]domain.Unit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Unit
	for _, u := range r.units {
		result = append(result, *u)
	}
	if result == nil {
		return []domain.Unit{}, nil
	}
	return result, nil
}

func (r *UnitRepository) GetByID(_ context.Context, id int64) (*domain.Unit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.units[id]
	if !ok {
		return nil, domain.ErrUnitNotFound
	}
	return u, nil
}

func (r *UnitRepository) GetByCode(_ context.Context, code string) (*domain.Unit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.units {
		if u.Code == code {
			return u, nil
		}
	}
	return nil, domain.ErrUnitNotFound
}
