package postgres

import (
	"context"

	"github.com/saleforge/pos/services/pkg/logger"
)

func SeedData(ctx context.Context, pool interface{}) error {
	logger.Info("catalog seed: no seed data needed (units seeded in migration)")
	return nil
}
