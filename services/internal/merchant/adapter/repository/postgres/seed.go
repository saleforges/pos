package postgres

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/pkg/otel"
)

func SeedData(ctx context.Context, pool *otel.TracedPool) error {
	now := time.Now().UTC()

	_, err := pool.Exec(ctx,
		`INSERT INTO merchants (id, name, email, created_at, updated_at)
		 VALUES (1, 'Warung Makmur', 'owner@merchant.com', $1, $1) ON CONFLICT DO NOTHING`, now)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx,
		`INSERT INTO branches (merchant_id, name, code, created_at, updated_at)
		 VALUES (1, 'Cabang A', 'A001', $1, $1) ON CONFLICT DO NOTHING`, now)
	if err != nil {
		return err
	}

	return nil
}
