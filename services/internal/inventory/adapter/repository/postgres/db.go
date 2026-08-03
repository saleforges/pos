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

	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		return fmt.Errorf("check migrations: %w", err)
	}

	if count == 0 {
		// Fresh install — run full init migration (up to v1)
		migration := `
		INSERT INTO schema_migrations (version, dirty) VALUES (1, false);

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
		`
		if _, err := pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("inventory init migration: %w", err)
		}
	}

	return nil
}
