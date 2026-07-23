package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

var _ repository.UnitRepository = (*UnitRepository)(nil)

type UnitRepository struct {
	pool *otel.TracedPool
}

func NewUnitRepository(pool *otel.TracedPool) *UnitRepository {
	return &UnitRepository{pool: pool}
}

func (r *UnitRepository) GetAll(ctx context.Context) ([]domain.Unit, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, code, name FROM units ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUnits(rows)
}

func (r *UnitRepository) GetByID(ctx context.Context, id int64) (*domain.Unit, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, code, name FROM units WHERE id = $1`, id)
	return scanUnit(row)
}

func (r *UnitRepository) GetByCode(ctx context.Context, code string) (*domain.Unit, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, code, name FROM units WHERE code = $1`, code)
	return scanUnit(row)
}

func scanUnit(row pgx.Row) (*domain.Unit, error) {
	var u domain.Unit
	err := row.Scan(&u.ID, &u.Code, &u.Name)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func scanUnits(rows pgx.Rows) ([]domain.Unit, error) {
	var result []domain.Unit
	for rows.Next() {
		var u domain.Unit
		if err := rows.Scan(&u.ID, &u.Code, &u.Name); err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	return result, rows.Err()
}
