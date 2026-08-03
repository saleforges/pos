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

	// Check if order v200 migration is already applied.
	// NOTE: schema_migrations is shared across services (IAM/merchant
	// golang-migrate v1, catalog 1-7, inventory v100), so we check for our
	// own version instead of the table being empty.
	var v200exists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = 200)`).Scan(&v200exists); err != nil {
		return fmt.Errorf("check order v200 migration: %w", err)
	}

	if !v200exists {
		migration := `
		INSERT INTO schema_migrations (version, dirty) VALUES (200, false);

		CREATE TABLE IF NOT EXISTS customers (
			id          BIGSERIAL    PRIMARY KEY,
			merchant_id BIGINT       NOT NULL,
			name        VARCHAR(200) NOT NULL,
			phone       VARCHAR(50),
			address     TEXT,
			note        TEXT,
			created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_customers_merchant ON customers(merchant_id);

		CREATE TABLE IF NOT EXISTS orders (
			id             BIGSERIAL    PRIMARY KEY,
			merchant_id    BIGINT       NOT NULL,
			branch_id      BIGINT       NOT NULL,
			created_by     BIGINT       NOT NULL,
			customer_id    BIGINT       REFERENCES customers(id) ON DELETE SET NULL,
			client_order_id VARCHAR(36),
			status         VARCHAR(20)  NOT NULL DEFAULT 'completed',
			subtotal       NUMERIC(14,2) NOT NULL DEFAULT 0,
			discount       NUMERIC(14,2) NOT NULL DEFAULT 0,
			tax            NUMERIC(14,2) NOT NULL DEFAULT 0,
			total          NUMERIC(14,2) NOT NULL DEFAULT 0,
			paid_amount    NUMERIC(14,2) NOT NULL DEFAULT 0,
			due_date       DATE,
			note           TEXT,
			created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_orders_merchant ON orders(merchant_id);
		CREATE INDEX IF NOT EXISTS idx_orders_branch ON orders(branch_id);
		CREATE INDEX IF NOT EXISTS idx_orders_customer ON orders(customer_id);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_client_order ON orders(client_order_id) WHERE client_order_id IS NOT NULL;

		CREATE TABLE IF NOT EXISTS order_items (
			id               BIGSERIAL    PRIMARY KEY,
			order_id         BIGINT       NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			product_item_id  BIGINT       NOT NULL,
			item_name        VARCHAR(200) NOT NULL,
			unit_price       NUMERIC(14,2) NOT NULL,
			quantity         NUMERIC(12,2) NOT NULL,
			line_total       NUMERIC(14,2) NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);

		CREATE TABLE IF NOT EXISTS payment_records (
			id          BIGSERIAL    PRIMARY KEY,
			order_id    BIGINT       NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			amount      NUMERIC(14,2) NOT NULL,
			method      VARCHAR(20)  NOT NULL,
			created_by  BIGINT       NOT NULL,
			paid_at     TIMESTAMPTZ  NOT NULL,
			created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_payment_records_order ON payment_records(order_id);
		`
		if _, err := pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("order init migration: %w", err)
		}
	}

	// Heal migrations — run on every startup so existing deployments get
	// new columns without a full re-init. Idempotent by design.
	heal := `
	ALTER TABLE orders ADD COLUMN IF NOT EXISTS client_order_id VARCHAR(36);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_client_order ON orders(client_order_id) WHERE client_order_id IS NOT NULL;
	`
	if _, err := pool.Exec(ctx, heal); err != nil {
		return fmt.Errorf("order heal migration: %w", err)
	}

	return nil
}
