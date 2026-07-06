package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saleforge/pos/services/internal/merchant/domain"
)

type StaffRepository struct {
	pool *pgxpool.Pool
}

func NewStaffRepository(pool *pgxpool.Pool) *StaffRepository {
	return &StaffRepository{pool: pool}
}

func (r *StaffRepository) Create(ctx context.Context, staff *domain.StaffMember) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO staff (id, merchant_id, branch_id, user_id, role, status, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		staff.ID, staff.MerchantID, staff.BranchID, staff.UserID,
		staff.Role, staff.Status, staff.IsDefault, staff.CreatedAt, staff.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create staff: %w", err)
	}
	return nil
}

func (r *StaffRepository) GetByID(ctx context.Context, id string) (*domain.StaffMember, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, merchant_id, branch_id, user_id, role, status, is_default, created_at, updated_at
		FROM staff WHERE id = $1`, id)

	s := &domain.StaffMember{}
	err := row.Scan(&s.ID, &s.MerchantID, &s.BranchID, &s.UserID,
		&s.Role, &s.Status, &s.IsDefault, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, domain.ErrStaffNotFound
		}
		return nil, fmt.Errorf("failed to get staff: %w", err)
	}
	return s, nil
}

func (r *StaffRepository) ListByBranch(ctx context.Context, branchID string) ([]domain.StaffMember, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, merchant_id, branch_id, user_id, role, status, is_default, created_at, updated_at
		FROM staff WHERE branch_id = $1 ORDER BY created_at`, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to list staff by branch: %w", err)
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

func (r *StaffRepository) ListByMerchant(ctx context.Context, merchantID string) ([]domain.StaffMember, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, merchant_id, branch_id, user_id, role, status, is_default, created_at, updated_at
		FROM staff WHERE merchant_id = $1 ORDER BY created_at`, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list staff by merchant: %w", err)
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

func (r *StaffRepository) GetByUserAndBranch(ctx context.Context, userID, branchID string) (*domain.StaffMember, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, merchant_id, branch_id, user_id, role, status, is_default, created_at, updated_at
		FROM staff WHERE user_id = $1 AND branch_id = $2`, userID, branchID)

	s := &domain.StaffMember{}
	err := row.Scan(&s.ID, &s.MerchantID, &s.BranchID, &s.UserID,
		&s.Role, &s.Status, &s.IsDefault, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, domain.ErrStaffNotFound
		}
		return nil, fmt.Errorf("failed to get staff by user and branch: %w", err)
	}
	return s, nil
}

func (r *StaffRepository) ListByUserAndMerchant(ctx context.Context, userID, merchantID string) ([]domain.StaffMember, error) {
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

func (r *StaffRepository) SetDefaultBranch(ctx context.Context, userID, branchID string) error {
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

func (r *StaffRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM staff WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete staff: %w", err)
	}
	return nil
}
