package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/pkg/otel"
)

type RoleRepository struct {
	pool *otel.TracedPool
}

func NewRoleRepository(pool *otel.TracedPool) *RoleRepository {
	return &RoleRepository{pool: pool}
}

func (r *RoleRepository) Create(ctx context.Context, role *domain.Role) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO roles (name, description, is_system, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id`,
		role.Name, role.Description, role.IsSystem,
	).Scan(&role.ID)
	if err != nil {
		return err
	}

	for _, p := range role.Permissions {
		var permissionID int64
		err = tx.QueryRow(ctx,
			`SELECT id FROM permissions WHERE name = $1`, string(p),
		).Scan(&permissionID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("permission %q does not exist", string(p))
			}
			return err
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			role.ID, permissionID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *RoleRepository) GetByID(ctx context.Context, id int64) (*domain.Role, error) {
	role, err := scanRole(r.pool.QueryRow(ctx,
		`SELECT id, name, description, is_system FROM roles WHERE id = $1`, id,
	))
	if err != nil {
		return nil, err
	}
	role.Permissions, err = loadRolePermissions(ctx, r.pool, role.ID)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RoleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	role, err := scanRole(r.pool.QueryRow(ctx,
		`SELECT id, name, description, is_system FROM roles WHERE name = $1`, name,
	))
	if err != nil {
		return nil, err
	}
	role.Permissions, err = loadRolePermissions(ctx, r.pool, role.ID)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RoleRepository) List(ctx context.Context, merchantID *int64) ([]domain.Role, error) {
	var rows pgx.Rows
	var err error
	if merchantID != nil {
		rows, err = r.pool.Query(ctx,
			`SELECT DISTINCT r.id, r.name, r.description, r.is_system
			 FROM roles r
			 JOIN user_roles ur ON ur.role_id = r.id
			 WHERE ur.merchant_id = $1
			 ORDER BY r.id`, *merchantID)
	} else {
		rows, err = r.pool.Query(ctx, `SELECT id, name, description, is_system FROM roles ORDER BY id`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem); err != nil {
			return nil, err
		}
		role.Permissions, err = loadRolePermissions(ctx, r.pool, role.ID)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *RoleRepository) Update(ctx context.Context, role *domain.Role) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE roles SET description = $1, updated_at = NOW() WHERE id = $2 AND is_system = false`,
		role.Description, role.ID,
	)
	return err
}

func (r *RoleRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM roles WHERE id = $1 AND is_system = false`, id)
	return err
}

func (r *RoleRepository) AddPermission(ctx context.Context, roleID int64, permission domain.Permission) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, (SELECT id FROM permissions WHERE name = $2)) ON CONFLICT DO NOTHING`,
		roleID, string(permission),
	)
	return err
}

func (r *RoleRepository) RemovePermission(ctx context.Context, roleID int64, permission domain.Permission) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = (SELECT id FROM permissions WHERE name = $2)`,
		roleID, string(permission),
	)
	return err
}

func (r *RoleRepository) GetPermissions(ctx context.Context, roleID int64) ([]domain.Permission, error) {
	return loadRolePermissions(ctx, r.pool, roleID)
}
