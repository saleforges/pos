package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

var _ repository.SellableItemRepository = (*SellableItemRepository)(nil)
var _ repository.BarcodeRepository = (*BarcodeRepository)(nil)

const siCols = `id, product_id, name, unit_id, price, track_inventory, image_url, status, created_at, updated_at, deleted_at`
const siNotDeleted = ` AND deleted_at IS NULL`

type SellableItemRepository struct {
	pool *otel.TracedPool
}

func NewSellableItemRepository(pool *otel.TracedPool) *SellableItemRepository {
	return &SellableItemRepository{pool: pool}
}

func (r *SellableItemRepository) Create(ctx context.Context, item *domain.SellableItem) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO sellable_items (product_id, name, unit_id, price, track_inventory, image_url, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		item.ProductID, item.Name, item.UnitID, item.Price, item.TrackInventory, item.ImageURL,
		item.Status, item.CreatedAt, item.UpdatedAt).Scan(&item.ID)
	return err
}

func (r *SellableItemRepository) GetByID(ctx context.Context, id int64) (*domain.SellableItem, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+siCols+` FROM sellable_items WHERE id = $1`+siNotDeleted, id)
	return scanSellableItem(row)
}

func (r *SellableItemRepository) ListByProduct(ctx context.Context, productID int64) ([]domain.SellableItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+siCols+` FROM sellable_items WHERE product_id = $1`+siNotDeleted+` ORDER BY name`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSellableItems(rows)
}

func (r *SellableItemRepository) Update(ctx context.Context, item *domain.SellableItem) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE sellable_items SET name=$1, unit_id=$2, price=$3, track_inventory=$4, image_url=$5, status=$6, updated_at=$7 WHERE id=$8 AND deleted_at IS NULL`,
		item.Name, item.UnitID, item.Price, item.TrackInventory, item.ImageURL, item.Status, item.UpdatedAt, item.ID)
	return err
}

func (r *SellableItemRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `UPDATE sellable_items SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *SellableItemRepository) Restore(ctx context.Context, id int64) (*domain.SellableItem, error) {
	_, err := r.pool.Exec(ctx, `UPDATE sellable_items SET deleted_at=NULL WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

type BarcodeRepository struct {
	pool *otel.TracedPool
}

func NewBarcodeRepository(pool *otel.TracedPool) *BarcodeRepository {
	return &BarcodeRepository{pool: pool}
}

func (r *BarcodeRepository) Create(ctx context.Context, barcode *domain.SellableItemBarcode) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO sellable_item_barcodes (sellable_item_id, barcode) VALUES ($1, $2) RETURNING id`,
		barcode.SellableItemID, barcode.Barcode).Scan(&barcode.ID)
	return err
}

func (r *BarcodeRepository) GetByBarcode(ctx context.Context, barcode string) (*domain.SellableItemBarcode, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, sellable_item_id, barcode FROM sellable_item_barcodes WHERE barcode = $1`, barcode)
	var b domain.SellableItemBarcode
	err := row.Scan(&b.ID, &b.SellableItemID, &b.Barcode)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BarcodeRepository) ListBySellableItem(ctx context.Context, sellableItemID int64) ([]domain.SellableItemBarcode, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, sellable_item_id, barcode FROM sellable_item_barcodes WHERE sellable_item_id = $1`, sellableItemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.SellableItemBarcode
	for rows.Next() {
		var b domain.SellableItemBarcode
		if err := rows.Scan(&b.ID, &b.SellableItemID, &b.Barcode); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func (r *BarcodeRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sellable_item_barcodes WHERE id = $1`, id)
	return err
}

func scanSellableItem(row pgx.Row) (*domain.SellableItem, error) {
	var s domain.SellableItem
	err := row.Scan(&s.ID, &s.ProductID, &s.Name, &s.UnitID, &s.Price, &s.TrackInventory, &s.ImageURL, &s.Status, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func scanSellableItems(rows pgx.Rows) ([]domain.SellableItem, error) {
	var result []domain.SellableItem
	for rows.Next() {
		var s domain.SellableItem
		if err := rows.Scan(&s.ID, &s.ProductID, &s.Name, &s.UnitID, &s.Price, &s.TrackInventory, &s.ImageURL, &s.Status, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}
