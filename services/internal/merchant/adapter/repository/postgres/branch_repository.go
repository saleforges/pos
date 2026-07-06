package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saleforge/pos/services/internal/merchant/domain"
)

type BranchRepository struct {
	pool *pgxpool.Pool
}

func NewBranchRepository(pool *pgxpool.Pool) *BranchRepository {
	return &BranchRepository{pool: pool}
}

func (r *BranchRepository) Create(ctx context.Context, branch *domain.Branch) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO branches (id, merchant_id, name, code, address, phone,
		                      status, operating_days, open_time, close_time,
		                      created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6,
		        $7, $8, $9, $10,
		        $11, $12)`,
		branch.ID, branch.MerchantID, branch.Name, branch.Code,
		branch.Address, branch.Phone,
		branch.Status, branch.OperatingDays,
		branch.OperatingHours.Open, branch.OperatingHours.Close,
		branch.CreatedAt, branch.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}
	return nil
}

func (r *BranchRepository) GetByID(ctx context.Context, id string) (*domain.Branch, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, merchant_id, name, code, address, phone,
		       status, operating_days, open_time, close_time,
		       created_at, updated_at
		FROM branches WHERE id = $1`, id)

	b := &domain.Branch{}
	b.OperatingHours = &domain.OperatingHours{}
	err := row.Scan(
		&b.ID, &b.MerchantID, &b.Name, &b.Code,
		&b.Address, &b.Phone,
		&b.Status, &b.OperatingDays,
		&b.OperatingHours.Open, &b.OperatingHours.Close,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, domain.ErrBranchNotFound
		}
		return nil, fmt.Errorf("failed to get branch: %w", err)
	}
	return b, nil
}

func (r *BranchRepository) ListByMerchant(ctx context.Context, merchantID string) ([]domain.Branch, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, merchant_id, name, code, address, phone,
		       status, operating_days, open_time, close_time,
		       created_at, updated_at
		FROM branches WHERE merchant_id = $1 ORDER BY name`, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}
	defer rows.Close()

	var result []domain.Branch
	for rows.Next() {
		var b domain.Branch
		b.OperatingHours = &domain.OperatingHours{}
		if err := rows.Scan(
			&b.ID, &b.MerchantID, &b.Name, &b.Code,
			&b.Address, &b.Phone,
			&b.Status, &b.OperatingDays,
			&b.OperatingHours.Open, &b.OperatingHours.Close,
			&b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan branch: %w", err)
		}
		result = append(result, b)
	}
	if result == nil {
		return []domain.Branch{}, nil
	}
	return result, nil
}

func (r *BranchRepository) Update(ctx context.Context, branch *domain.Branch) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE branches SET name=$1, address=$2, phone=$3, status=$4,
		                    operating_days=$5, open_time=$6, close_time=$7,
		                    updated_at=$8
		WHERE id=$9`,
		branch.Name, branch.Address, branch.Phone, branch.Status,
		branch.OperatingDays, branch.OperatingHours.Open, branch.OperatingHours.Close,
		branch.UpdatedAt, branch.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update branch: %w", err)
	}
	return nil
}

func (r *BranchRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM branches WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}
	return nil
}
