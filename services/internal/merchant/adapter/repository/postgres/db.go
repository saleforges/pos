package postgres

import (
	"context"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saleforge/pos/services/pkg/logger"
	"github.com/saleforge/pos/services/pkg/otel"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

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
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	upErr := m.Up()
	if upErr != nil && upErr != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", upErr)
	}

	if upErr == migrate.ErrNoChange {
		pool, poolErr := pgxpool.New(context.Background(), databaseURL)
		if poolErr == nil {
			var exists bool
			poolErr = pool.QueryRow(context.Background(),
				"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'merchants')",
			).Scan(&exists)
			if poolErr == nil && !exists {
				logger.Warn("migration tracking corrupted, forcing re-run")
				m.Force(-1)
				if err := m.Up(); err != nil && err != migrate.ErrNoChange {
					return fmt.Errorf("failed to re-run migrations: %w", err)
				}
			}
			pool.Close()
		}
	}

	srcErr, dbErr := m.Close()
	if srcErr != nil {
		return fmt.Errorf("migration source close error: %w", srcErr)
	}
	if dbErr != nil {
		return fmt.Errorf("migration database close error: %w", dbErr)
	}

	return nil
}
