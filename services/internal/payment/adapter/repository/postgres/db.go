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

	// Check if payment v300 migration is already applied.
	// NOTE: schema_migrations is shared across services (IAM/merchant
	// golang-migrate v1, catalog 1-7, inventory v100, order v200), so we
	// check for our own version instead of the table being empty.
	var v300exists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = 300)`).Scan(&v300exists); err != nil {
		return fmt.Errorf("check payment v300 migration: %w", err)
	}

	if !v300exists {
		migration := `
		INSERT INTO schema_migrations (version, dirty) VALUES (300, false);

		CREATE TABLE IF NOT EXISTS payment_transactions (
			id          BIGSERIAL    PRIMARY KEY,
			merchant_id BIGINT       NOT NULL,
			order_id    BIGINT       NOT NULL,
			gateway     VARCHAR(20)  NOT NULL,
			status      VARCHAR(20)  NOT NULL DEFAULT 'pending',
			amount      NUMERIC(14,2) NOT NULL,
			payment_url TEXT,
			session_id  VARCHAR(100),
			payment_ref VARCHAR(100),
			created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_payment_transactions_merchant ON payment_transactions(merchant_id);
		CREATE INDEX IF NOT EXISTS idx_payment_transactions_order ON payment_transactions(order_id);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_transactions_payment_ref ON payment_transactions(payment_ref) WHERE payment_ref IS NOT NULL;
		`
		if _, err := pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("payment init migration: %w", err)
		}
	}

	// Heal migrations — run on every startup so existing deployments get
	// new columns added over time.
	heal := `
	ALTER TABLE payment_transactions ADD COLUMN IF NOT EXISTS payment_no TEXT;
	ALTER TABLE payment_transactions ADD COLUMN IF NOT EXISTS qr_string TEXT;
	ALTER TABLE payment_transactions ADD COLUMN IF NOT EXISTS qr_image TEXT;
	ALTER TABLE payment_transactions ADD COLUMN IF NOT EXISTS expired_at TEXT;
	ALTER TABLE payment_transactions ADD COLUMN IF NOT EXISTS payment_ref TEXT;
	ALTER TABLE payment_transactions ADD COLUMN IF NOT EXISTS transaction_id TEXT;
	ALTER TABLE payment_transactions ADD COLUMN IF NOT EXISTS gateway_ref TEXT;
	UPDATE payment_transactions SET payment_ref = COALESCE(transaction_id, gateway_ref) WHERE payment_ref IS NULL OR payment_ref = '';
	DROP INDEX IF EXISTS idx_payment_transactions_gateway_ref;
	CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_transactions_payment_ref ON payment_transactions(payment_ref) WHERE payment_ref IS NOT NULL;
	ALTER TABLE payment_transactions DROP COLUMN IF EXISTS transaction_id;
	ALTER TABLE payment_transactions DROP COLUMN IF EXISTS gateway_ref;
	ALTER TABLE payment_transactions ADD COLUMN IF NOT EXISTS branch_id BIGINT NOT NULL DEFAULT 0;
	CREATE INDEX IF NOT EXISTS idx_payment_transactions_branch ON payment_transactions(branch_id);

	CREATE TABLE IF NOT EXISTS payment_static_qrs (
		merchant_id BIGINT      NOT NULL,
		branch_id   BIGINT      NOT NULL DEFAULT 0,
		payment_no  VARCHAR(100) NOT NULL,
		qr_string   TEXT        NOT NULL,
		qr_image    TEXT        NOT NULL,
		updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		PRIMARY KEY (merchant_id, branch_id)
	);
	ALTER TABLE payment_static_qrs ADD COLUMN IF NOT EXISTS branch_id BIGINT NOT NULL DEFAULT 0;
	ALTER TABLE payment_static_qrs DROP CONSTRAINT IF EXISTS payment_static_qrs_pkey;
	ALTER TABLE payment_static_qrs ADD PRIMARY KEY (merchant_id, branch_id);
	`
	if _, err := pool.Exec(ctx, heal); err != nil {
		return fmt.Errorf("payment heal migration: %w", err)
	}

	return nil
}
