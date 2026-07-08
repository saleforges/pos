package postgres

import (
	"context"

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

func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO products (id, merchant_id, category_id, name, sku, barcode, description, price, cost, tax_rate, unit, image_url, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		product.ID, product.MerchantID, product.CategoryID, product.Name, product.SKU, product.Barcode,
		product.Description, product.Price, product.Cost, product.TaxRate, product.Unit, product.ImageURL,
		product.Status, product.CreatedAt, product.UpdatedAt)
	return err
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, merchant_id, category_id, name, sku, barcode, description, price, cost, tax_rate, unit, image_url, status, created_at, updated_at
		 FROM products WHERE id = $1`, id)
	return scanProduct(row)
}

func (r *ProductRepository) GetBySKU(ctx context.Context, sku string, merchantID string) (*domain.Product, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, merchant_id, category_id, name, sku, barcode, description, price, cost, tax_rate, unit, image_url, status, created_at, updated_at
		 FROM products WHERE sku = $1 AND merchant_id = $2`, sku, merchantID)
	return scanProduct(row)
}

func (r *ProductRepository) List(ctx context.Context, merchantID string, offset, limit int) ([]domain.Product, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, merchant_id, category_id, name, sku, barcode, description, price, cost, tax_rate, unit, image_url, status, created_at, updated_at
		 FROM products WHERE merchant_id = $1 ORDER BY name LIMIT $2 OFFSET $3`,
		merchantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProducts(rows)
}

func (r *ProductRepository) ListByCategory(ctx context.Context, categoryID string, offset, limit int) ([]domain.Product, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, merchant_id, category_id, name, sku, barcode, description, price, cost, tax_rate, unit, image_url, status, created_at, updated_at
		 FROM products WHERE category_id = $1 ORDER BY name LIMIT $2 OFFSET $3`,
		categoryID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProducts(rows)
}

func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE products SET category_id=$1, name=$2, sku=$3, barcode=$4, description=$5, price=$6, cost=$7, tax_rate=$8, unit=$9, image_url=$10, status=$11, updated_at=$12 WHERE id=$13`,
		product.CategoryID, product.Name, product.SKU, product.Barcode, product.Description,
		product.Price, product.Cost, product.TaxRate, product.Unit, product.ImageURL,
		product.Status, product.UpdatedAt, product.ID)
	return err
}

func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	return err
}

func scanProduct(row interface{ Scan(dest ...any) error }) (*domain.Product, error) {
	var p domain.Product
	err := row.Scan(
		&p.ID, &p.MerchantID, &p.CategoryID, &p.Name, &p.SKU, &p.Barcode,
		&p.Description, &p.Price, &p.Cost, &p.TaxRate, &p.Unit, &p.ImageURL,
		&p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func scanProducts(rows interface{ Next() bool; Scan(dest ...any) error; Err() error }) ([]domain.Product, error) {
	var result []domain.Product
	for rows.Next() {
		var p domain.Product
		err := rows.Scan(
			&p.ID, &p.MerchantID, &p.CategoryID, &p.Name, &p.SKU, &p.Barcode,
			&p.Description, &p.Price, &p.Cost, &p.TaxRate, &p.Unit, &p.ImageURL,
			&p.Status, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}
