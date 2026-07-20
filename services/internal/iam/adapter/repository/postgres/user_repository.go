package postgres

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/pkg/otel"
)

type UserRepository struct {
	pool *otel.TracedPool
}

func NewUserRepository(pool *otel.TracedPool) *UserRepository {
	return &UserRepository{pool: pool}
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

	if err = tx.Commit(ctx); err != nil {
		return err
	}
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
	roles, err := loadScopedRoles(ctx, r.pool, user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles
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
	roles, err := loadScopedRoles(ctx, r.pool, user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles
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
	roles, err := loadScopedRoles(ctx, r.pool, user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles
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
		roles, err := loadScopedRoles(ctx, r.pool, u.ID)
		if err != nil {
			return nil, err
		}
		u.Roles = roles
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
	res, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
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
