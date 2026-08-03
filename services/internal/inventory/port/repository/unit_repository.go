package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/inventory/domain"
)

// UnitRepository provides unit conversion reference data. Units are seeded
// from the catalog units table (same IDs); FactorToBase normalizes a unit
// into its base unit for component consumption math.
type UnitRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Unit, error)
}
