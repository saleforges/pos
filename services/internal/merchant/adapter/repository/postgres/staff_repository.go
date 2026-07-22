package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/pkg/otel"
)

type StaffRepository struct {
	pool *otel.TracedPool
}

func NewStaffRepository(pool *otel.TracedPool) *StaffRepository {
	return &StaffRepository{pool: pool}
}

func (r *StaffRepository) Create(ctx context.Context, staff *domain.StaffMember) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO staff (merchant_id, branch_id, user_id, role, status, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		staff.MerchantID, staff.BranchID, staff.UserID,
		staff.Role, staff.Status, staff.IsDefault, staff.CreatedAt, staff.UpdatedAt,
	).Scan(&staff.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrStaffExists
		}
		return fmt.Errorf("failed to create staff: %w", err)
	}
	return nil
}

func (r *StaffRepository) GetByID(ctx context.Context, id int64) (*domain.StaffMember, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, merchant_id, branch_id, user_id, role, status, is_default, created_at, updated_at
		FROM staff WHERE id = $1`, id)

	s := &domain.StaffMember{}
	err := row.Scan(&s.ID, &s.MerchantID, &s.BranchID, &s.UserID,
		&s.Role, &s.Status, &s.IsDefault, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrStaffNotFound
		}
		return nil, fmt.Errorf("failed to get staff: %w", err)
	}
	return s, nil
}

func (r *StaffRepository) ListByBranch(ctx context.Context, branchID int64, offset, limit int) ([]domain.StaffMember, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM staff WHERE branch_id = $1`, branchID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count staff by branch: %w", err)
	}

	if limit == -1 {
		limit = int(total)
		if limit == 0 { limit = 1 }
		offset = 0
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, merchant_id, branch_id, user_id, role, status, is_default, created_at, updated_at
		FROM staff WHERE branch_id = $1 ORDER BY created_at OFFSET $2 LIMIT $3`, branchID, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list staff by branch: %w", err)
	}
	defer rows.Close()

	var result []domain.StaffMember
	for rows.Next() {
		var s domain.StaffMember
		if err := rows.Scan(&s.ID, &s.MerchantID, &s.BranchID, &s.UserID,
			&s.Role, &s.Status, &s.IsDefault, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan staff: %w", err)
		}
		result = append(result, s)
	}
	if result == nil {
		return []domain.StaffMember{}, total, nil
	}
	return result, total, nil
}

func (r *StaffRepository) ListByMerchant(ctx context.Context, merchantID int64, offset, limit int) ([]domain.StaffMember, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM staff WHERE merchant_id = $1`, merchantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count staff by merchant: %w", err)
	}

	if limit == -1 {
		limit = int(total)
		if limit == 0 { limit = 1 }
		offset = 0
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, merchant_id, branch_id, user_id, role, status, is_default, created_at, updated_at
		FROM staff WHERE merchant_id = $1 ORDER BY created_at OFFSET $2 LIMIT $3`, merchantID, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list staff by merchant: %w", err)
	}
	defer rows.Close()

	var result []domain.StaffMember
	for rows.Next() {
		var s domain.StaffMember
		if err := rows.Scan(&s.ID, &s.MerchantID, &s.BranchID, &s.UserID,
			&s.Role, &s.Status, &s.IsDefault, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan staff: %w", err)
		}
		result = append(result, s)
	}
	if result == nil {
		return []domain.StaffMember{}, total, nil
	}
	return result, total, nil
}

func (r *StaffRepository) GetByUserAndBranch(ctx context.Context, userID, branchID int64) (*domain.StaffMember, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, merchant_id, branch_id, user_id, role, status, is_default, created_at, updated_at
		FROM staff WHERE user_id = $1 AND branch_id = $2`, userID, branchID)

	s := &domain.StaffMember{}
	err := row.Scan(&s.ID, &s.MerchantID, &s.BranchID, &s.UserID,
		&s.Role, &s.Status, &s.IsDefault, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrStaffNotFound
		}
		return nil, fmt.Errorf("failed to get staff by user and branch: %w", err)
	}
	return s, nil
}

func (r *StaffRepository) ListByUserAndMerchant(ctx context.Context, userID, merchantID int64) ([]domain.StaffMember, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, merchant_id, branch_id, user_id, role, status, is_default, created_at, updated_at
		FROM staff WHERE user_id = $1 AND merchant_id = $2 ORDER BY created_at`, userID, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list staff by user and merchant: %w", err)
	}
	defer rows.Close()

	var result []domain.StaffMember
	for rows.Next() {
		var s domain.StaffMember
		if err := rows.Scan(&s.ID, &s.MerchantID, &s.BranchID, &s.UserID,
			&s.Role, &s.Status, &s.IsDefault, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan staff: %w", err)
		}
		result = append(result, s)
	}
	if result == nil {
		return []domain.StaffMember{}, nil
	}
	return result, nil
}

func (r *StaffRepository) SetDefaultBranch(ctx context.Context, userID, branchID int64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE staff SET is_default = (branch_id = $2) WHERE user_id = $1`, userID, branchID)
	if err != nil {
		return fmt.Errorf("failed to set default branch: %w", err)
	}
	return nil
}

func (r *StaffRepository) Update(ctx context.Context, staff *domain.StaffMember) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE staff SET role=$1, status=$2, is_default=$3, updated_at=$4 WHERE id=$5`,
		staff.Role, staff.Status, staff.IsDefault, staff.UpdatedAt, staff.ID)
	if err != nil {
		return fmt.Errorf("failed to update staff: %w", err)
	}
	return nil
}

func (r *StaffRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM staff WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete staff: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrStaffNotFound
	}
	return nil
}
