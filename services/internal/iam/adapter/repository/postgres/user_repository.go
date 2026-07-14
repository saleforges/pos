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
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Type, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func loadUserSystemRole(ctx context.Context, pool *otel.TracedPool, userID int64) *domain.Role {
	var r domain.Role
	err := pool.QueryRow(ctx,
		`SELECT r.id, r.name, r.description, r.is_system
		 FROM user_roles ur JOIN roles r ON ur.role_id = r.id
		 WHERE ur.user_id = $1 AND ur.merchant_id IS NULL
		 LIMIT 1`, userID,
	).Scan(&r.ID, &r.Name, &r.Description, &r.IsSystem)
	if err != nil {
		return nil
	}
	r.Permissions, _ = loadRolePermissions(ctx, pool, r.ID)
	return &r
}

func loadScopedRoles(ctx context.Context, pool *otel.TracedPool, userID int64) []domain.UserRoleAssignment {
	var result []domain.UserRoleAssignment

	rows, err := pool.Query(ctx,
		`SELECT ur.id, ur.merchant_id, COALESCE(m.name, ''), ur.branch_id, COALESCE(b.name, ''), r.id, r.name, r.description, r.is_system, ur.is_default
		 FROM user_roles ur
		 JOIN roles r ON r.id = ur.role_id
		 LEFT JOIN merchants m ON m.id = ur.merchant_id
		 LEFT JOIN branches b ON b.id = ur.branch_id
		 WHERE ur.user_id = $1 AND ur.status = 'active' AND ur.merchant_id IS NOT NULL
		 ORDER BY m.name, b.name`, userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var a domain.UserRoleAssignment
			if err := rows.Scan(&a.ID, &a.MerchantID, &a.MerchantName, &a.BranchID, &a.BranchName,
				&a.Role.ID, &a.Role.Name, &a.Role.Description, &a.Role.IsSystem, &a.IsDefault); err == nil {
				if a.BranchID != nil {
					a.BranchScope = domain.BranchScopeAssigned
				} else if a.MerchantID != 0 {
					a.BranchScope = domain.BranchScopeAll
				}
				result = append(result, a)
			}
		}
	}
	if result == nil {
		result = []domain.UserRoleAssignment{}
	}

	return result
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO users (username, email, password, type, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		user.Username, user.Email, user.Password, user.Type, user.Status, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)
	if err != nil {
		return err
	}

	_ = tx.Commit(ctx)
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	user, err := scanUser(r.pool.QueryRow(ctx,
		`SELECT id, username, email, password, type, status, created_at, updated_at FROM users WHERE id = $1`, id,
	))
	if err != nil {
		return nil, err
	}
	user.SystemRole = loadUserSystemRole(ctx, r.pool, user.ID)
	user.Roles = loadScopedRoles(ctx, r.pool, user.ID)
	if user.Roles == nil {
		user.Roles = []domain.UserRoleAssignment{}
	}
	return user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	user, err := scanUser(r.pool.QueryRow(ctx,
		`SELECT id, username, email, password, type, status, created_at, updated_at FROM users WHERE username = $1`, username,
	))
	if err != nil {
		return nil, err
	}
	user.SystemRole = loadUserSystemRole(ctx, r.pool, user.ID)
	user.Roles = loadScopedRoles(ctx, r.pool, user.ID)
	if user.Roles == nil {
		user.Roles = []domain.UserRoleAssignment{}
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := scanUser(r.pool.QueryRow(ctx,
		`SELECT id, username, email, password, type, status, created_at, updated_at FROM users WHERE email = $1`, email,
	))
	if err != nil {
		return nil, err
	}
	user.SystemRole = loadUserSystemRole(ctx, r.pool, user.ID)
	user.Roles = loadScopedRoles(ctx, r.pool, user.ID)
	if user.Roles == nil {
		user.Roles = []domain.UserRoleAssignment{}
	}
	return user, nil
}

func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, username, email, password, type, status, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Type, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		u.SystemRole = loadUserSystemRole(ctx, r.pool, u.ID)
		u.Roles = loadScopedRoles(ctx, r.pool, u.ID)
		if u.Roles == nil {
			u.Roles = []domain.UserRoleAssignment{}
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

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) AddRole(ctx context.Context, userID int64, roleName string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_roles (user_id, role_id) SELECT $1, id FROM roles WHERE name = $2 ON CONFLICT DO NOTHING`,
		userID, roleName,
	)
	return err
}

func (r *UserRepository) RemoveRole(ctx context.Context, userID int64, roleName string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM user_roles USING roles WHERE user_roles.role_id = roles.id AND user_roles.user_id = $1 AND roles.name = $2`,
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
		_, err = tx.Exec(ctx,
			`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, (SELECT id FROM permissions WHERE name = $2)) ON CONFLICT DO NOTHING`,
			role.ID, string(p),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func scanRole(row pgx.Row) (*domain.Role, error) {
	var role domain.Role
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrInvalidRole
		}
		return nil, err
	}
	return &role, nil
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


