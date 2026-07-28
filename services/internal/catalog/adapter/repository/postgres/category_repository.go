package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

var _ repository.CategoryRepository = (*CategoryRepository)(nil)

const catCols = `id, merchant_id, name, parent_id, created_at, updated_at, deleted_at`
const catNotDeleted = ` AND deleted_at IS NULL`

type CategoryRepository struct {
	pool *otel.TracedPool
}

func NewCategoryRepository(pool *otel.TracedPool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO categories (merchant_id, name, parent_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		category.MerchantID, category.Name, category.ParentID, category.CreatedAt, category.UpdatedAt).Scan(&category.ID)
	return err
}

func (r *CategoryRepository) GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Category, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+catCols+` FROM categories WHERE id = $1 AND merchant_id = $2`+catNotDeleted, id, merchantID)
	return scanCategory(row)
}

func (r *CategoryRepository) ListByMerchant(ctx context.Context, merchantID int64) ([]domain.Category, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+catCols+` FROM categories WHERE merchant_id = $1`+catNotDeleted+` ORDER BY name`, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCategories(rows)
}

func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE categories SET name=$1, parent_id=$2, updated_at=$3 WHERE id=$4 AND merchant_id=$5 AND deleted_at IS NULL`,
		category.Name, category.ParentID, category.UpdatedAt, category.ID, category.MerchantID)
	return err
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64, merchantID int64) error {
	_, err := r.pool.Exec(ctx, `UPDATE categories SET deleted_at=NOW() WHERE id=$1 AND merchant_id=$2 AND deleted_at IS NULL`, id, merchantID)
	return err
}

func (r *CategoryRepository) Restore(ctx context.Context, id int64, merchantID int64) (*domain.Category, error) {
	_, err := r.pool.Exec(ctx, `UPDATE categories SET deleted_at=NULL WHERE id=$1 AND merchant_id=$2`, id, merchantID)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id, merchantID)
}

func (r *CategoryRepository) ListUpdatedAfter(ctx context.Context, merchantID int64, after time.Time) ([]domain.Category, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+catCols+` FROM categories WHERE merchant_id = $1 AND (updated_at > $2 OR deleted_at > $2) ORDER BY updated_at`,
		merchantID, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCategories(rows)
}

func scanCategory(row pgx.Row) (*domain.Category, error) {
	var c domain.Category
	err := row.Scan(&c.ID, &c.MerchantID, &c.Name, &c.ParentID, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func scanCategories(rows pgx.Rows) ([]domain.Category, error) {
	var result []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.MerchantID, &c.Name, &c.ParentID, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}
