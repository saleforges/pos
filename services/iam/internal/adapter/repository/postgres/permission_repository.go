package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saleforge/pos/services/iam/internal/domain"
)

type PermissionRepository struct {
	pool *pgxpool.Pool
}

func NewPermissionRepository(pool *pgxpool.Pool) *PermissionRepository {
	return &PermissionRepository{pool: pool}
}

func (r *PermissionRepository) Create(ctx context.Context, permission domain.Permission) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO permissions (name, created_at) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		string(permission), time.Now().UTC(),
	)
	return err
}

func (r *PermissionRepository) GetAll(ctx context.Context) ([]domain.Permission, error) {
	rows, err := r.pool.Query(ctx, `SELECT name FROM permissions ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Permission
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		result = append(result, domain.Permission(name))
	}
	return result, rows.Err()
}

func (r *PermissionRepository) Delete(ctx context.Context, permission domain.Permission) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM permissions WHERE name = $1`, string(permission))
	return err
}
