package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
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

func (r *StaffRepository) ListByUserID(ctx context.Context, userID string) ([]domain.StaffInfo, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT s.merchant_id, m.name, s.role
		 FROM staff s JOIN merchants m ON m.id = s.merchant_id
		 WHERE s.user_id = $1 AND s.status = 'active'`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.StaffInfo
	for rows.Next() {
		var info domain.StaffInfo
		if err := rows.Scan(&info.MerchantID, &info.MerchantName, &info.Role); err != nil {
			return nil, err
		}
		result = append(result, info)
	}
	return result, rows.Err()
}

func (r *StaffRepository) Create(ctx context.Context, userID, merchantID, merchantName, role string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var id string
	err = tx.QueryRow(ctx,
		`INSERT INTO staff (merchant_id, user_id, role, status)
		 VALUES ($1, $2, $3, 'active')
		 RETURNING id`, merchantID, userID, role).Scan(&id)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func scanStaffInfo(row pgx.Row) (*domain.StaffInfo, error) {
	var info domain.StaffInfo
	if err := row.Scan(&info.MerchantID, &info.MerchantName, &info.Role); err != nil {
		return nil, err
	}
	return &info, nil
}
