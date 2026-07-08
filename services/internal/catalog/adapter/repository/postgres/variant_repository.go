package postgres

import (
	"context"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

var _ repository.VariantRepository = (*VariantRepository)(nil)

type VariantRepository struct {
	pool *otel.TracedPool
}

func NewVariantRepository(pool *otel.TracedPool) *VariantRepository {
	return &VariantRepository{pool: pool}
}

func (r *VariantRepository) Create(ctx context.Context, variant *domain.Variant) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO variants (id, product_id, name, sku, barcode, price, cost, image_url, sort_order, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		variant.ID, variant.ProductID, variant.Name, variant.SKU, variant.Barcode,
		variant.Price, variant.Cost, variant.ImageURL, variant.SortOrder,
		variant.CreatedAt, variant.UpdatedAt)
	return err
}

func (r *VariantRepository) GetByID(ctx context.Context, id string) (*domain.Variant, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, product_id, name, sku, barcode, price, cost, image_url, sort_order, created_at, updated_at
		 FROM variants WHERE id = $1`, id)
	return scanVariant(row)
}

func (r *VariantRepository) ListByProduct(ctx context.Context, productID string) ([]domain.Variant, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, product_id, name, sku, barcode, price, cost, image_url, sort_order, created_at, updated_at
		 FROM variants WHERE product_id = $1 ORDER BY sort_order`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Variant
	for rows.Next() {
		var v domain.Variant
		err := rows.Scan(
			&v.ID, &v.ProductID, &v.Name, &v.SKU, &v.Barcode,
			&v.Price, &v.Cost, &v.ImageURL, &v.SortOrder,
			&v.CreatedAt, &v.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (r *VariantRepository) Update(ctx context.Context, variant *domain.Variant) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE variants SET name=$1, sku=$2, barcode=$3, price=$4, cost=$5, image_url=$6, sort_order=$7, updated_at=$8 WHERE id=$9`,
		variant.Name, variant.SKU, variant.Barcode, variant.Price, variant.Cost,
		variant.ImageURL, variant.SortOrder, variant.UpdatedAt, variant.ID)
	return err
}

func (r *VariantRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM variants WHERE id = $1`, id)
	return err
}

func scanVariant(row interface{ Scan(dest ...any) error }) (*domain.Variant, error) {
	var v domain.Variant
	err := row.Scan(
		&v.ID, &v.ProductID, &v.Name, &v.SKU, &v.Barcode,
		&v.Price, &v.Cost, &v.ImageURL, &v.SortOrder,
		&v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
