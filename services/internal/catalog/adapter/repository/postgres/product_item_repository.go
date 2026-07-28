package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

var _ repository.ProductItemRepository = (*ProductItemRepository)(nil)
var _ repository.ProductItemBarcodeRepository = (*ProductItemBarcodeRepository)(nil)

const piCols = `id, product_id, merchant_id, name, sku, unit_id, track_inventory, image_url, status, created_at, updated_at, deleted_at`
const piNotDeleted = ` AND deleted_at IS NULL`

type ProductItemRepository struct {
	pool *otel.TracedPool
}

func NewProductItemRepository(pool *otel.TracedPool) *ProductItemRepository {
	return &ProductItemRepository{pool: pool}
}

func (r *ProductItemRepository) Create(ctx context.Context, item *domain.ProductItem) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO product_items (product_id, merchant_id, name, sku, unit_id, track_inventory, image_url, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
		item.ProductID, item.MerchantID, item.Name, nullIfEmpty(item.SKU), item.UnitID, item.TrackInventory, item.ImageURL,
		item.Status, item.CreatedAt, item.UpdatedAt).Scan(&item.ID)
	if err != nil {
		if isSKUConstraint(err) {
			return domain.ErrSKUDuplicate
		}
		return err
	}

	// Insert price
	_, err = r.pool.Exec(ctx,
		`INSERT INTO prices (product_item_id, amount, currency, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		item.ID, item.Price.Amount, item.Price.Currency, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *ProductItemRepository) GetByID(ctx context.Context, id int64, merchantID int64) (*domain.ProductItem, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+piCols+` FROM product_items WHERE id = $1 AND merchant_id = $2`+piNotDeleted, id, merchantID)
	item, err := scanProductItem(row)
	if err != nil {
		return nil, err
	}
	// Load price
	price, err := r.getPrice(ctx, id)
	if err != nil {
		return nil, err
	}
	item.Price = *price
	return item, nil
}

func (r *ProductItemRepository) ListByProduct(ctx context.Context, productID int64, merchantID int64) ([]domain.ProductItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+piCols+` FROM product_items WHERE product_id = $1 AND merchant_id = $2`+piNotDeleted+` ORDER BY name`, productID, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanProductItems(rows)
	if err != nil {
		return nil, err
	}
	return r.loadPrices(ctx, items)
}

func (r *ProductItemRepository) ListByMerchant(ctx context.Context, merchantID int64) ([]domain.ProductItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT pi.`+piCols+` FROM product_items pi
		 JOIN products p ON p.id = pi.product_id
		 WHERE p.merchant_id = $1 AND pi.deleted_at IS NULL
		 ORDER BY pi.name`, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanProductItems(rows)
	if err != nil {
		return nil, err
	}
	return r.loadPrices(ctx, items)
}

func (r *ProductItemRepository) Update(ctx context.Context, item *domain.ProductItem) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE product_items SET name=$1, sku=$2, unit_id=$3, track_inventory=$4, image_url=$5, status=$6, updated_at=$7 WHERE id=$8 AND merchant_id=$9 AND deleted_at IS NULL`,
		item.Name, nullIfEmpty(item.SKU), item.UnitID, item.TrackInventory, item.ImageURL, item.Status, item.UpdatedAt, item.ID, item.MerchantID)
	if err != nil {
		if isSKUConstraint(err) {
			return domain.ErrSKUDuplicate
		}
		return err
	}
	// Upsert price
	_, err = r.pool.Exec(ctx,
		`INSERT INTO prices (product_item_id, amount, currency, created_at, updated_at)
		 VALUES ($1, $2, $3, NOW(), NOW())
		 ON CONFLICT (product_item_id) DO UPDATE SET amount=$2, currency=$3, updated_at=NOW()`,
		item.ID, item.Price.Amount, item.Price.Currency)
	return err
}

func (r *ProductItemRepository) ListUpdatedAfter(ctx context.Context, merchantID int64, after time.Time) ([]domain.ProductItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT pi.`+piCols+` FROM product_items pi
		 WHERE pi.merchant_id = $1 AND (pi.updated_at > $2 OR pi.deleted_at > $2)
		 ORDER BY pi.updated_at`, merchantID, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanProductItems(rows)
	if err != nil {
		return nil, err
	}
	return r.loadPrices(ctx, items)
}

func (r *ProductItemRepository) Delete(ctx context.Context, id int64, merchantID int64) error {
	_, err := r.pool.Exec(ctx, `UPDATE product_items SET deleted_at=NOW() WHERE id=$1 AND merchant_id=$2 AND deleted_at IS NULL`, id, merchantID)
	return err
}

func (r *ProductItemRepository) Restore(ctx context.Context, id int64, merchantID int64) (*domain.ProductItem, error) {
	_, err := r.pool.Exec(ctx, `UPDATE product_items SET deleted_at=NULL WHERE id=$1 AND merchant_id=$2`, id, merchantID)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id, merchantID)
}

// internal helpers

func (r *ProductItemRepository) getPrice(ctx context.Context, productItemID int64) (*domain.Price, error) {
	var p domain.Price
	err := r.pool.QueryRow(ctx,
		`SELECT amount, currency FROM prices WHERE product_item_id = $1`, productItemID).Scan(&p.Amount, &p.Currency)
	if err != nil {
		return &domain.Price{Amount: 0, Currency: "IDR"}, nil
	}
	return &p, nil
}

func (r *ProductItemRepository) loadPrices(ctx context.Context, items []domain.ProductItem) ([]domain.ProductItem, error) {
	if len(items) == 0 {
		return items, nil
	}
	ids := make([]int64, len(items))
	for i, it := range items {
		ids[i] = it.ID
	}
	rows, err := r.pool.Query(ctx,
		`SELECT product_item_id, amount, currency FROM prices WHERE product_item_id = ANY($1)`, ids)
	if err != nil {
		return items, err
	}
	defer rows.Close()
	priceMap := make(map[int64]domain.Price)
	for rows.Next() {
		var pid int64
		var p domain.Price
		if err := rows.Scan(&pid, &p.Amount, &p.Currency); err != nil {
			return items, err
		}
		priceMap[pid] = p
	}
	if rows.Err() != nil {
		return items, rows.Err()
	}
	for i := range items {
		if p, ok := priceMap[items[i].ID]; ok {
			items[i].Price = p
		} else {
			items[i].Price = domain.Price{Currency: "IDR"}
		}
	}
	return items, nil
}

func scanProductItem(row pgx.Row) (*domain.ProductItem, error) {
	var s domain.ProductItem
	err := row.Scan(&s.ID, &s.ProductID, &s.MerchantID, &s.Name, &s.SKU, &s.UnitID, &s.TrackInventory, &s.ImageURL, &s.Status, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func scanProductItems(rows pgx.Rows) ([]domain.ProductItem, error) {
	var result []domain.ProductItem
	for rows.Next() {
		var s domain.ProductItem
		if err := rows.Scan(&s.ID, &s.ProductID, &s.MerchantID, &s.Name, &s.SKU, &s.UnitID, &s.TrackInventory, &s.ImageURL, &s.Status, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

// ProductItemBarcodeRepository

type ProductItemBarcodeRepository struct {
	pool *otel.TracedPool
}

func NewProductItemBarcodeRepository(pool *otel.TracedPool) *ProductItemBarcodeRepository {
	return &ProductItemBarcodeRepository{pool: pool}
}

func (r *ProductItemBarcodeRepository) Create(ctx context.Context, barcode *domain.ProductItemBarcode) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO product_item_barcodes (product_item_id, barcode) VALUES ($1, $2) RETURNING id`,
		barcode.ProductItemID, barcode.Barcode).Scan(&barcode.ID)
	return err
}

func (r *ProductItemBarcodeRepository) GetByBarcode(ctx context.Context, barcode string) (*domain.ProductItemBarcode, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, product_item_id, barcode FROM product_item_barcodes WHERE barcode = $1`, barcode)
	var b domain.ProductItemBarcode
	err := row.Scan(&b.ID, &b.ProductItemID, &b.Barcode)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *ProductItemBarcodeRepository) ListByProductItem(ctx context.Context, productItemID int64) ([]domain.ProductItemBarcode, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, product_item_id, barcode FROM product_item_barcodes WHERE product_item_id = $1`, productItemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.ProductItemBarcode
	for rows.Next() {
		var b domain.ProductItemBarcode
		if err := rows.Scan(&b.ID, &b.ProductItemID, &b.Barcode); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func (r *ProductItemBarcodeRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM product_item_barcodes WHERE id = $1`, id)
	return err
}

func (r *ProductItemBarcodeRepository) ListByMerchant(ctx context.Context, merchantID int64) ([]domain.ProductItemBarcode, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT b.id, b.product_item_id, b.barcode
		 FROM product_item_barcodes b
		 JOIN product_items pi ON pi.id = b.product_item_id
		 WHERE pi.merchant_id = $1`, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.ProductItemBarcode
	for rows.Next() {
		var b domain.ProductItemBarcode
		if err := rows.Scan(&b.ID, &b.ProductItemID, &b.Barcode); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func isSKUConstraint(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "idx_product_items_sku_merchant") ||
		strings.Contains(msg, "duplicate key value")
}
