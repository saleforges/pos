package postgres

import (
	"context"
	"fmt"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

var _ repository.CategoryRepository = (*CategoryRepository)(nil)

type CategoryRepository struct {
	pool *otel.TracedPool
}

func NewCategoryRepository(pool *otel.TracedPool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO categories (merchant_id, name, slug, description, parent_id, sort_order, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		category.MerchantID, category.Name, category.Slug, category.Description,
		category.ParentID, category.SortOrder, category.Status, category.CreatedAt, category.UpdatedAt).Scan(&category.ID)
	return err
}

func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*domain.Category, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, merchant_id, name, slug, description, parent_id, sort_order, status, created_at, updated_at
		 FROM categories WHERE id = $1`, id)
	return scanCategory(row)
}

func (r *CategoryRepository) List(ctx context.Context, merchantID int64, search string, offset, limit int) ([]domain.Category, error) {
	query := `SELECT id, merchant_id, name, slug, description, parent_id, sort_order, status, created_at, updated_at
		 FROM categories WHERE merchant_id = $1`
	args := []interface{}{merchantID}
	argIdx := 2
	if search != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR slug ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx+1, argIdx+2)
		searchParam := "%" + search + "%"
		args = append(args, searchParam, searchParam, searchParam)
		argIdx += 3
	}
	query += " ORDER BY sort_order, name"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Category
	for rows.Next() {
		var cat domain.Category
		err := rows.Scan(&cat.ID, &cat.MerchantID, &cat.Name, &cat.Slug, &cat.Description, &cat.ParentID, &cat.SortOrder, &cat.Status, &cat.CreatedAt, &cat.UpdatedAt)
		if err != nil {
			return nil, err
		}
		result = append(result, cat)
	}
	return result, rows.Err()
}

func (r *CategoryRepository) Count(ctx context.Context, merchantID int64, search string) (int, error) {
	query := `SELECT COUNT(*) FROM categories WHERE merchant_id = $1`
	args := []interface{}{merchantID}
	if search != "" {
		query += ` AND (name ILIKE $2 OR slug ILIKE $3 OR description ILIKE $4)`
		searchParam := "%" + search + "%"
		args = append(args, searchParam, searchParam, searchParam)
	}
	var count int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE categories SET name=$1, slug=$2, description=$3, parent_id=$4, sort_order=$5, status=$6, updated_at=$7 WHERE id=$8`,
		category.Name, category.Slug, category.Description, category.ParentID, category.SortOrder, category.Status, category.UpdatedAt, category.ID)
	return err
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	return err
}

func scanCategory(row interface{ Scan(dest ...any) error }) (*domain.Category, error) {
	var cat domain.Category
	err := row.Scan(
		&cat.ID, &cat.MerchantID, &cat.Name, &cat.Slug, &cat.Description,
		&cat.ParentID, &cat.SortOrder, &cat.Status, &cat.CreatedAt, &cat.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &cat, nil
}
