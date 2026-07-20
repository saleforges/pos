package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/pkg/otel"
)

func scanRole(row pgx.Row) (*domain.Role, error) {
	var role domain.Role
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func loadRolePermissions(ctx context.Context, pool *otel.TracedPool, roleID int64) ([]domain.Permission, error) {
	rows, err := pool.Query(ctx,
		`SELECT p.name FROM role_permissions rp JOIN permissions p ON p.id = rp.permission_id WHERE rp.role_id = $1 ORDER BY p.name`, roleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		perms = append(perms, domain.Permission(p))
	}
	return perms, rows.Err()
}
