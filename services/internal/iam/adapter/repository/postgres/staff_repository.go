package postgres

import (
	"context"

	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

var _ repository.StaffRepository = (*StaffRepository)(nil)

type StaffRepository struct {
	pool *otel.TracedPool
}

func NewStaffRepository(pool *otel.TracedPool) *StaffRepository {
	return &StaffRepository{pool: pool}
}

func (r *StaffRepository) ListByUserID(ctx context.Context, userID int64) ([]domain.UserRoleAssignment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT ur.id, ur.merchant_id, COALESCE(m.name, ''), ur.branch_id, COALESCE(b.name, ''), r.id, r.name, r.description, r.is_system, ur.is_default
		 FROM user_roles ur
		 JOIN roles r ON r.id = ur.role_id
		 LEFT JOIN merchants m ON m.id = ur.merchant_id
		 LEFT JOIN branches b ON b.id = ur.branch_id
		 WHERE ur.user_id = $1 AND ur.status = 'active'
		 ORDER BY m.name, b.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.UserRoleAssignment
	for rows.Next() {
		var a domain.UserRoleAssignment
		if err := rows.Scan(&a.ID, &a.MerchantID, &a.MerchantName, &a.BranchID, &a.BranchName,
			&a.Role.ID, &a.Role.Name, &a.Role.Description, &a.Role.IsSystem, &a.IsDefault); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

func (r *StaffRepository) Create(ctx context.Context, userID int64, merchantID int64, merchantName, role string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO user_roles (merchant_id, user_id, role_id, status)
		 SELECT $1, $2, id, 'active' FROM roles WHERE name = $3`,
		merchantID, userID, role)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
