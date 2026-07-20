package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/pkg/otel"
)

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Type, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func loadUserSystemRole(ctx context.Context, pool *otel.TracedPool, userID int64) (*domain.Role, error) {
	var r domain.Role
	err := pool.QueryRow(ctx,
		`SELECT r.id, r.name, r.description, r.is_system
		 FROM user_roles ur JOIN roles r ON ur.role_id = r.id
		 WHERE ur.user_id = $1 AND ur.merchant_id IS NULL
		 LIMIT 1`, userID,
	).Scan(&r.ID, &r.Name, &r.Description, &r.IsSystem)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	perms, err := loadRolePermissions(ctx, pool, r.ID)
	if err != nil {
		return nil, err
	}
	r.Permissions = perms
	return &r, nil
}

func loadScopedRoles(ctx context.Context, pool *otel.TracedPool, userID int64) ([]domain.UserRoleAssignment, error) {
	var result []domain.UserRoleAssignment

	rows, err := pool.Query(ctx,
		`SELECT ur.id, ur.merchant_id, COALESCE(m.name, ''), ur.branch_id, COALESCE(b.name, ''), r.id, r.name, r.description, r.is_system, ur.is_default
		 FROM user_roles ur
		 JOIN roles r ON r.id = ur.role_id
		 LEFT JOIN merchants m ON m.id = ur.merchant_id
		 LEFT JOIN branches b ON b.id = ur.branch_id
		 WHERE ur.user_id = $1 AND ur.status = 'active' AND ur.merchant_id IS NOT NULL
		 ORDER BY m.name, b.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var a domain.UserRoleAssignment
		if err := rows.Scan(&a.ID, &a.MerchantID, &a.MerchantName, &a.BranchID, &a.BranchName,
			&a.Role.ID, &a.Role.Name, &a.Role.Description, &a.Role.IsSystem, &a.IsDefault); err != nil {
			return nil, err
		}
		if a.BranchID != nil {
			a.BranchScope = domain.BranchScopeAssigned
		} else if a.MerchantID != 0 {
			a.BranchScope = domain.BranchScopeAll
		}
		result = append(result, a)
	}
	if result == nil {
		result = []domain.UserRoleAssignment{}
	}

	return result, rows.Err()
}
