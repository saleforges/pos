package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saleforge/pos/services/pkg/otel"
)

func Connect(ctx context.Context, databaseURL string) (*otel.TracedPool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}
	return otel.NewTracedPool(pool), nil
}

func RunMigrations(databaseURL string) error {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("unable to connect for migrations: %w", err)
	}
	defer pool.Close()

	schema := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version BIGINT PRIMARY KEY,
		dirty   BOOLEAN NOT NULL
	);
	`
	if _, err := pool.Exec(ctx, schema); err != nil {
		return fmt.Errorf("schema_migrations table: %w", err)
	}
	// Check if catalog v7 migration is already applied.
	// NOTE: schema_migrations is shared across services (IAM/merchant use
	// golang-migrate with version 1), so we check for our own version instead
	// of the table being empty.
	var v7exists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = 7)`).Scan(&v7exists); err != nil {
		return fmt.Errorf("check catalog v7 migration: %w", err)
	}
	if !v7exists {
		// Fresh install — run full init migration (up to v7)
		migration := `
		INSERT INTO schema_migrations (version, dirty) VALUES (1, false) ON CONFLICT DO NOTHING;
		INSERT INTO schema_migrations (version, dirty) VALUES (2, false) ON CONFLICT DO NOTHING;
		INSERT INTO schema_migrations (version, dirty) VALUES (3, false) ON CONFLICT DO NOTHING;
		INSERT INTO schema_migrations (version, dirty) VALUES (4, false) ON CONFLICT DO NOTHING;
		INSERT INTO schema_migrations (version, dirty) VALUES (5, false) ON CONFLICT DO NOTHING;
		INSERT INTO schema_migrations (version, dirty) VALUES (6, false) ON CONFLICT DO NOTHING;
		INSERT INTO schema_migrations (version, dirty) VALUES (7, false) ON CONFLICT DO NOTHING;

		CREATE TABLE IF NOT EXISTS units (
			id   BIGSERIAL PRIMARY KEY,
			code VARCHAR(10)  NOT NULL UNIQUE,
			name VARCHAR(100) NOT NULL
		);

		INSERT INTO units (code, name) VALUES
			('PCS', 'Piece'), ('PACK', 'Pack'), ('KG', 'Kilogram'),
			('GRAM', 'Gram'), ('LITER', 'Liter'), ('ML', 'Milliliter'),
			('BOX', 'Box'), ('METER', 'Meter')
		ON CONFLICT DO NOTHING;

		CREATE TABLE IF NOT EXISTS categories (
			id          BIGSERIAL    PRIMARY KEY,
			merchant_id BIGINT       NOT NULL,
			name        VARCHAR(100) NOT NULL,
			parent_id   BIGINT       REFERENCES categories(id),
			created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_categories_merchant ON categories(merchant_id);

		CREATE TABLE IF NOT EXISTS products (
			id          BIGSERIAL    PRIMARY KEY,
			merchant_id BIGINT       NOT NULL,
			category_id BIGINT       NOT NULL REFERENCES categories(id),
			name        VARCHAR(200) NOT NULL,
			description TEXT,
			image_url   TEXT,
			status      VARCHAR(20)  NOT NULL DEFAULT 'active',
			created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_products_merchant ON products(merchant_id);
		CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);

		CREATE TABLE IF NOT EXISTS product_items (
			id               BIGSERIAL    PRIMARY KEY,
			product_id       BIGINT       NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			merchant_id      BIGINT       NOT NULL,
			name             VARCHAR(200) NOT NULL,
			sku              VARCHAR(100),
			unit_id          BIGINT       REFERENCES units(id),
			track_inventory  BOOLEAN      NOT NULL DEFAULT FALSE,
			image_url        TEXT,
			status           VARCHAR(20)  NOT NULL DEFAULT 'active',
			created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_product_items_product ON product_items(product_id);
		CREATE INDEX IF NOT EXISTS idx_product_items_merchant ON product_items(merchant_id);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_product_items_sku_merchant ON product_items(merchant_id, sku) WHERE sku IS NOT NULL;

		CREATE TABLE IF NOT EXISTS prices (
			id               BIGSERIAL    PRIMARY KEY,
			product_item_id  BIGINT       NOT NULL REFERENCES product_items(id) ON DELETE CASCADE UNIQUE,
			amount           NUMERIC(12,2) NOT NULL DEFAULT 0,
			currency         VARCHAR(3)   NOT NULL DEFAULT 'IDR',
			created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS product_item_barcodes (
			id               BIGSERIAL    PRIMARY KEY,
			product_item_id  BIGINT       NOT NULL REFERENCES product_items(id) ON DELETE CASCADE,
			barcode          VARCHAR(100) NOT NULL
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_product_item_barcodes_unique ON product_item_barcodes(barcode);
		CREATE INDEX IF NOT EXISTS idx_product_item_barcodes_item ON product_item_barcodes(product_item_id);
		`
		if _, err := pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("catalog init migration: %w", err)
		}
	}

	return nil
}
