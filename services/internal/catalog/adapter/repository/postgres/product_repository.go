package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

var _ repository.ProductRepository = (*ProductRepository)(nil)

type ProductRepository struct {
	pool *otel.TracedPool
}

func NewProductRepository(pool *otel.TracedPool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

const productCols = `p.id, p.merchant_id, p.category_id, p.name, p.description, p.image_url, p.status, p.created_at, p.updated_at, p.deleted_at`
const productNotDeleted = ` AND p.deleted_at IS NULL`

func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO products (merchant_id, category_id, name, description, image_url, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		product.MerchantID, product.CategoryID, product.Name, product.Description, product.ImageURL,
		product.Status, product.CreatedAt, product.UpdatedAt).Scan(&product.ID)
	return err
}

func (r *ProductRepository) GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Product, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+productCols+` FROM products p WHERE p.id = $1 AND p.merchant_id = $2`+productNotDeleted, id, merchantID)
	return scanProduct(row)
}

func (r *ProductRepository) List(ctx context.Context, merchantID int64, search string, offset, limit int) ([]domain.Product, int, error) {
	baseQuery := `SELECT ` + productCols + ` FROM products p WHERE p.merchant_id = $1` + productNotDeleted
	countQuery := `SELECT COUNT(*) FROM products p WHERE p.merchant_id = $1` + productNotDeleted
	var args []any
	args = append(args, merchantID)
	paramIdx := 2
	if search != "" {
		pattern := "%" + search + "%"
		baseQuery += fmt.Sprintf(` AND p.name ILIKE $%d`, paramIdx)
		countQuery += fmt.Sprintf(` AND p.name ILIKE $%d`, paramIdx)
		args = append(args, pattern)
		paramIdx++
	}
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := baseQuery + fmt.Sprintf(` ORDER BY p.name LIMIT $%d OFFSET $%d`, paramIdx, paramIdx+1)
	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := scanProducts(rows)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE products SET category_id=$1, name=$2, description=$3, image_url=$4, status=$5, updated_at=$6 WHERE id=$7 AND merchant_id=$8 AND deleted_at IS NULL`,
		product.CategoryID, product.Name, product.Description, product.ImageURL, product.Status, product.UpdatedAt, product.ID, product.MerchantID)
	return err
}

func (r *ProductRepository) Delete(ctx context.Context, id int64, merchantID int64) error {
	_, err := r.pool.Exec(ctx, `UPDATE products SET deleted_at=NOW() WHERE id=$1 AND merchant_id=$2 AND deleted_at IS NULL`, id, merchantID)
	return err
}

func (r *ProductRepository) Restore(ctx context.Context, id int64, merchantID int64) (*domain.Product, error) {
	_, err := r.pool.Exec(ctx, `UPDATE products SET deleted_at=NULL WHERE id=$1 AND merchant_id=$2`, id, merchantID)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id, merchantID)
}

func (r *ProductRepository) ListUpdatedAfter(ctx context.Context, merchantID int64, after time.Time) ([]domain.Product, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+productCols+` FROM products p WHERE p.merchant_id = $1 AND (p.updated_at > $2 OR p.deleted_at > $2) ORDER BY p.updated_at`,
		merchantID, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProducts(rows)
}

func scanProduct(row pgx.Row) (*domain.Product, error) {
	var (
		p           domain.Product
		description sql.NullString
		imageURL    sql.NullString
	)
	err := row.Scan(&p.ID, &p.MerchantID, &p.CategoryID, &p.Name, &description, &imageURL, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		return nil, err
	}
	p.Description = description.String
	p.ImageURL = imageURL.String
	return &p, nil
}

func scanProducts(rows pgx.Rows) ([]domain.Product, error) {
	var result []domain.Product
	for rows.Next() {
		var (
			p           domain.Product
			description sql.NullString
			imageURL    sql.NullString
		)
		if err := rows.Scan(&p.ID, &p.MerchantID, &p.CategoryID, &p.Name, &description, &imageURL, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt); err != nil {
			return nil, err
		}
		p.Description = description.String
		p.ImageURL = imageURL.String
		result = append(result, p)
	}
	return result, rows.Err()
}
