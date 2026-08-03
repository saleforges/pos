package memory

import (
	"context"

	"github.com/saleforge/pos/services/internal/inventory/domain"
	"github.com/saleforge/pos/services/internal/inventory/port/repository"
)

var _ repository.UnitRepository = (*UnitRepository)(nil)

type UnitRepository struct {
	units map[int64]*domain.Unit
}

// NewUnitRepository seeds the standard units mirroring the catalog units
// table (same IDs). FactorToBase normalizes into the base unit.
func NewUnitRepository() *UnitRepository {
	return &UnitRepository{units: map[int64]*domain.Unit{
		1: {ID: 1, Name: "Piece", Symbol: "pc", FactorToBase: 1},
		2: {ID: 2, Name: "Pack", Symbol: "pack", FactorToBase: 1},
		3: {ID: 3, Name: "Kilogram", Symbol: "kg", FactorToBase: 1000},
		4: {ID: 4, Name: "Gram", Symbol: "g", FactorToBase: 1},
		5: {ID: 5, Name: "Liter", Symbol: "L", FactorToBase: 1000},
		6: {ID: 6, Name: "Milliliter", Symbol: "ml", FactorToBase: 1},
		7: {ID: 7, Name: "Box", Symbol: "box", FactorToBase: 1},
		8: {ID: 8, Name: "Meter", Symbol: "m", FactorToBase: 1},
	}}
}

func (r *UnitRepository) GetByID(_ context.Context, id int64) (*domain.Unit, error) {
	u, ok := r.units[id]
	if !ok {
		return nil, nil
	}
	return u, nil
}
