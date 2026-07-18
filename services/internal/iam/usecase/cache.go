package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/iam/domain"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (uc *authUsecase) cacheGet(ctx context.Context, id int64) (*domain.User, error) {
	span := trace.SpanFromContext(ctx)

	if uc.userCache != nil {
		if u, ok := uc.userCache.Get(ctx, id); ok {
			span.AddEvent("cache.hit", trace.WithAttributes(
				attribute.Int64("cache.key", id),
			))
			return u, nil
		}
		span.AddEvent("cache.miss", trace.WithAttributes(
			attribute.Int64("cache.key", id),
		))
	}

	u, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if uc.userCache != nil {
		uc.userCache.Set(ctx, u, 0)
	}
	return u, nil
}

func (uc *authUsecase) cacheSet(ctx context.Context, u *domain.User) {
	if uc.userCache != nil {
		uc.userCache.Set(ctx, u, 0)
	}
}

func (uc *authUsecase) cacheDel(ctx context.Context, id int64) {
	if uc.userCache != nil {
		uc.userCache.Delete(ctx, id)
	}
}
