package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saleforge/pos/services/pkg/logger"
	"github.com/saleforge/pos/services/pkg/otel"
)

func SeedData(ctx context.Context, pool *otel.TracedPool) error {
	logger.Info("seeding catalog data...")

	// Check if seed already exists
	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM categories`).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		logger.Info("catalog seed already exists, skipping")
		return nil
	}

	now := time.Now().UTC()

	// Create seed category
	var catID int64
	err = pool.QueryRow(ctx,
		`INSERT INTO categories (merchant_id, name, created_at, updated_at)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		1, "Makanan Ringan", now, now).Scan(&catID)
	if err != nil {
		return err
	}
	logger.Info("created seed category", "id", catID)

	// Create seed product: Marning
	var prodID int64
	err = pool.QueryRow(ctx,
		`INSERT INTO products (merchant_id, category_id, name, description, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		1, catID, "Marning", "Marning original", "active", now, now).Scan(&prodID)
	if err != nil {
		return err
	}
	logger.Info("created seed product", "id", prodID)

	// Create product items for Marning
	var itemID int64
	err = pool.QueryRow(ctx,
		`INSERT INTO product_items (product_id, merchant_id, name, sku, unit_id, track_inventory, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		prodID, 1, "Marning Curah", "MRN-KG-001", 3, true, "active", now, now).Scan(&itemID)
	if err != nil {
		return err
	}
	// Insert price
	_, err = pool.Exec(ctx,
		`INSERT INTO prices (product_item_id, amount, currency, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		itemID, 40000, "IDR", now, now)
	if err != nil {
		return err
	}

	err = pool.QueryRow(ctx,
		`INSERT INTO product_items (product_id, merchant_id, name, sku, unit_id, track_inventory, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		prodID, 1, "Marning Pack", "MRN-PACK-001", 2, true, "active", now, now).Scan(&itemID)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx,
		`INSERT INTO prices (product_item_id, amount, currency, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		itemID, 1000, "IDR", now, now)
	if err != nil {
		return err
	}

	// Create seed product: Es Teh
	err = pool.QueryRow(ctx,
		`INSERT INTO products (merchant_id, category_id, name, description, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		1, catID, "Es Teh", "Teh manis", "active", now, now).Scan(&prodID)
	if err != nil {
		return err
	}
	logger.Info("created seed product", "id", prodID)

	// Simple product — auto-create item
	err = pool.QueryRow(ctx,
		`INSERT INTO product_items (product_id, merchant_id, name, track_inventory, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		prodID, 1, "Es Teh", false, "active", now, now).Scan(&itemID)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx,
		`INSERT INTO prices (product_item_id, amount, currency, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		itemID, 5000, "IDR", now, now)
	if err != nil {
		return err
	}

	logger.Info("catalog seed complete")
	return nil
}

// SeedDataSchema is kept for backward compatibility — called by the migrate cmd
// with a raw pgxpool.Pool via the interface{} parameter.
func SeedDataSchema(ctx context.Context, poolInterface interface{}) error {
	p, ok := poolInterface.(*pgxpool.Pool)
	if !ok {
		logger.Warn("catalog seed: expected *pgxpool.Pool, got different type, skipping")
		return nil
	}
	return SeedData(ctx, otel.NewTracedPool(p))
}
