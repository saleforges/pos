package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/internal/inventory/domain"
	"github.com/saleforge/pos/services/internal/inventory/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

var _ repository.UnitRepository = (*UnitRepository)(nil)

type UnitRepository struct {
	pool *otel.TracedPool
}

func NewUnitRepository(pool *otel.TracedPool) *UnitRepository {
	return &UnitRepository{pool: pool}
}

func (r *UnitRepository) GetByID(ctx context.Context, id int64) (*domain.Unit, error) {
	var u domain.Unit
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, symbol, factor_to_base FROM units WHERE id = $1`, id,
	).Scan(&u.ID, &u.Name, &u.Symbol, &u.FactorToBase)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}
