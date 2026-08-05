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

	// Check if inventory v100 migration is already applied.
	// NOTE: schema_migrations is shared across services (IAM/merchant use
	// golang-migrate with version 1, catalog uses 1-7), so we check for our
	// own version instead of the table being empty.
	var v100exists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = 100)`).Scan(&v100exists); err != nil {
		return fmt.Errorf("check inventory v100 migration: %w", err)
	}

	if !v100exists {
		// Fresh install — run full init migration (v100)
		migration := `
		INSERT INTO schema_migrations (version, dirty) VALUES (100, false);

		CREATE TABLE IF NOT EXISTS stocks (
			id               BIGSERIAL    PRIMARY KEY,
			merchant_id      BIGINT       NOT NULL,
			branch_id        BIGINT       NOT NULL,
			product_item_id  BIGINT       NOT NULL,
			available        BIGINT       NOT NULL DEFAULT 0,
			created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_stocks_merchant ON stocks(merchant_id);
		CREATE INDEX IF NOT EXISTS idx_stocks_branch ON stocks(branch_id);
		CREATE INDEX IF NOT EXISTS idx_stocks_product_item ON stocks(product_item_id);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_stocks_branch_product ON stocks(branch_id, product_item_id);
		DROP INDEX IF EXISTS idx_stocks_branch_product;
		CREATE UNIQUE INDEX IF NOT EXISTS idx_stocks_merchant_branch_product ON stocks(merchant_id, branch_id, product_item_id);

		CREATE TABLE IF NOT EXISTS stock_movements (
			id               BIGSERIAL    PRIMARY KEY,
			merchant_id      BIGINT       NOT NULL,
			branch_id        BIGINT       NOT NULL,
			product_item_id  BIGINT       NOT NULL,
			type             VARCHAR(20)  NOT NULL,
			quantity         BIGINT       NOT NULL,
			reference_type   VARCHAR(50),
			reference_id     BIGINT,
			note             TEXT,
			created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_stock_movements_merchant ON stock_movements(merchant_id);
		CREATE INDEX IF NOT EXISTS idx_stock_movements_product_item ON stock_movements(product_item_id);
		CREATE INDEX IF NOT EXISTS idx_stock_movements_branch ON stock_movements(branch_id);

		CREATE TABLE IF NOT EXISTS product_components (
			id               BIGSERIAL    PRIMARY KEY,
			merchant_id      BIGINT       NOT NULL,
			product_item_id  BIGINT       NOT NULL,
			created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_product_components_product_item ON product_components(product_item_id);
		CREATE INDEX IF NOT EXISTS idx_product_components_merchant ON product_components(merchant_id);

		CREATE TABLE IF NOT EXISTS product_component_items (
			id                        BIGSERIAL    PRIMARY KEY,
			product_component_id      BIGINT       NOT NULL REFERENCES product_components(id) ON DELETE CASCADE,
			component_product_item_id BIGINT       NOT NULL,
			quantity                  NUMERIC(12,4) NOT NULL,
			unit_id                   BIGINT       NOT NULL,
			created_at                TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_component_items_component ON product_component_items(product_component_id);

		CREATE TABLE IF NOT EXISTS inventory_units (
			id             BIGINT       PRIMARY KEY,
			name           VARCHAR(50)  NOT NULL,
			symbol         VARCHAR(10)  NOT NULL DEFAULT '',
			factor_to_base NUMERIC(14,4) NOT NULL DEFAULT 1
		);
		INSERT INTO inventory_units (id, name, symbol, factor_to_base) VALUES
			(1, 'Piece', 'pc', 1),
			(2, 'Pack', 'pack', 1),
			(3, 'Kilogram', 'kg', 1000),
			(4, 'Gram', 'g', 1),
			(5, 'Liter', 'L', 1000),
			(6, 'Milliliter', 'ml', 1),
			(7, 'Box', 'box', 1),
			(8, 'Meter', 'm', 1)
		ON CONFLICT (id) DO NOTHING;
		`
		if _, err := pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("inventory init migration: %w", err)
		}
	}

	// Heal migrations — run on every startup so existing deployments get
	// new tables/columns without a full re-init. Idempotent by design.
	heal := `
	ALTER TABLE product_component_items DROP COLUMN IF EXISTS conversion_factor;

	DROP INDEX IF EXISTS idx_stocks_branch_product;
	CREATE UNIQUE INDEX IF NOT EXISTS idx_stocks_merchant_branch_product ON stocks(merchant_id, branch_id, product_item_id);

	CREATE TABLE IF NOT EXISTS inventory_units (
		id             BIGINT       PRIMARY KEY,
		name           VARCHAR(50)  NOT NULL,
		symbol         VARCHAR(10)  NOT NULL DEFAULT '',
		factor_to_base NUMERIC(14,4) NOT NULL DEFAULT 1
	);
	INSERT INTO inventory_units (id, name, symbol, factor_to_base) VALUES
		(1, 'Piece', 'pc', 1),
		(2, 'Pack', 'pack', 1),
		(3, 'Kilogram', 'kg', 1000),
		(4, 'Gram', 'g', 1),
		(5, 'Liter', 'L', 1000),
		(6, 'Milliliter', 'ml', 1),
		(7, 'Box', 'box', 1),
		(8, 'Meter', 'm', 1)
	ON CONFLICT (id) DO NOTHING;
	`
	if _, err := pool.Exec(ctx, heal); err != nil {
		return fmt.Errorf("inventory heal migration: %w", err)
	}

	return nil
}
