package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/pkg/otel"
	"github.com/saleforge/pos/services/internal/iam/domain"
)

type UserRepository struct {
	pool *otel.TracedPool
}

func NewUserRepository(pool *otel.TracedPool) *UserRepository {
	return &UserRepository{pool: pool}
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func loadUserRoles(ctx context.Context, pool *otel.TracedPool, userID string) ([]string, error) {
	rows, err := pool.Query(ctx, `SELECT role_name FROM user_roles WHERE user_id = $1 ORDER BY role_name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO users (id, username, email, password, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		user.ID, user.Username, user.Email, user.Password, user.Status, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return err
	}

	for _, role := range user.Roles {
		_, err = tx.Exec(ctx,
			`INSERT INTO user_roles (user_id, role_name) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			user.ID, role,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := scanUser(r.pool.QueryRow(ctx,
		`SELECT id, username, email, password, status, created_at, updated_at FROM users WHERE id = $1`, id,
	))
	if err != nil {
		return nil, err
	}
	user.Roles, err = loadUserRoles(ctx, r.pool, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	user, err := scanUser(r.pool.QueryRow(ctx,
		`SELECT id, username, email, password, status, created_at, updated_at FROM users WHERE username = $1`, username,
	))
	if err != nil {
		return nil, err
	}
	user.Roles, err = loadUserRoles(ctx, r.pool, user.ID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := scanUser(r.pool.QueryRow(ctx,
		`SELECT id, username, email, password, status, created_at, updated_at FROM users WHERE email = $1`, email,
	))
	if err != nil {
		return nil, err
	}
	user.Roles, err = loadUserRoles(ctx, r.pool, user.ID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, username, email, password, status, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		u.Roles, err = loadUserRoles(ctx, r.pool, u.ID)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET username = $1, email = $2, status = $3, updated_at = $4 WHERE id = $5`,
		user.Username, user.Email, user.Status, time.Now().UTC(), user.ID,
	)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) AddRole(ctx context.Context, userID, roleName string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_roles (user_id, role_name) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, roleName,
	)
	return err
}

func (r *UserRepository) RemoveRole(ctx context.Context, userID, roleName string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM user_roles WHERE user_id = $1 AND role_name = $2`,
		userID, roleName,
	)
	return err
}

type RoleRepository struct {
	pool *otel.TracedPool
}

func NewRoleRepository(pool *otel.TracedPool) *RoleRepository {
	return &RoleRepository{pool: pool}
}

func loadRolePermissions(ctx context.Context, pool *otel.TracedPool, roleName string) ([]domain.Permission, error) {
	rows, err := pool.Query(ctx,
		`SELECT permission_name FROM role_permissions WHERE role_name = $1 ORDER BY permission_name`, roleName,
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

func (r *RoleRepository) Create(ctx context.Context, role *domain.Role) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	roleName := role.Name
	_, err = tx.Exec(ctx,
		`INSERT INTO roles (name, description, is_system, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW())`,
		roleName, role.Description, role.IsSystem,
	)
	if err != nil {
		return err
	}

	for _, p := range role.Permissions {
		_, err = tx.Exec(ctx,
			`INSERT INTO role_permissions (role_name, permission_name) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			roleName, string(p),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *RoleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	err := r.pool.QueryRow(ctx,
		`SELECT name, description, is_system FROM roles WHERE name = $1`, name,
	).Scan(&role.Name, &role.Description, &role.IsSystem)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrInvalidRole
		}
		return nil, err
	}
	role.Permissions, err = loadRolePermissions(ctx, r.pool, name)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) List(ctx context.Context) ([]domain.Role, error) {
	rows, err := r.pool.Query(ctx, `SELECT name, description, is_system FROM roles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.Name, &role.Description, &role.IsSystem); err != nil {
			return nil, err
		}
		role.Permissions, err = loadRolePermissions(ctx, r.pool, role.Name)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *RoleRepository) Update(ctx context.Context, role *domain.Role) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE roles SET description = $1, updated_at = NOW() WHERE name = $2 AND is_system = false`,
		role.Description, role.Name,
	)
	return err
}

func (r *RoleRepository) Delete(ctx context.Context, name string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM roles WHERE name = $1 AND is_system = false`, name)
	return err
}

func (r *RoleRepository) AddPermission(ctx context.Context, roleName string, permission domain.Permission) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO role_permissions (role_name, permission_name) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		roleName, string(permission),
	)
	return err
}

func (r *RoleRepository) RemovePermission(ctx context.Context, roleName string, permission domain.Permission) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM role_permissions WHERE role_name = $1 AND permission_name = $2`,
		roleName, string(permission),
	)
	return err
}

func (r *RoleRepository) GetPermissions(ctx context.Context, roleName string) ([]domain.Permission, error) {
	return loadRolePermissions(ctx, r.pool, roleName)
}
