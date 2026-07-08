package postgres

import (
	"context"

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
	_, err := r.pool.Exec(ctx,
		`INSERT INTO categories (id, merchant_id, name, slug, description, parent_id, sort_order, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		category.ID, category.MerchantID, category.Name, category.Slug, category.Description,
		category.ParentID, category.SortOrder, category.Status, category.CreatedAt, category.UpdatedAt)
	return err
}

func (r *CategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, merchant_id, name, slug, description, parent_id, sort_order, status, created_at, updated_at
		 FROM categories WHERE id = $1`, id)
	return scanCategory(row)
}

func (r *CategoryRepository) List(ctx context.Context, merchantID string, offset, limit int) ([]domain.Category, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, merchant_id, name, slug, description, parent_id, sort_order, status, created_at, updated_at
		 FROM categories WHERE merchant_id = $1 ORDER BY sort_order, name LIMIT $2 OFFSET $3`,
		merchantID, limit, offset)
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

func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE categories SET name=$1, slug=$2, description=$3, parent_id=$4, sort_order=$5, status=$6, updated_at=$7 WHERE id=$8`,
		category.Name, category.Slug, category.Description, category.ParentID, category.SortOrder, category.Status, category.UpdatedAt, category.ID)
	return err
}

func (r *CategoryRepository) Delete(ctx context.Context, id string) error {
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
