package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	"github.com/saleforge/pos/services/pkg/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// cacheGet retrieves a user by ID, trying cache first then DB.
// Cache errors (Redis down, corrupt entry) are treated as cache misses;
// the DB is the authoritative source of user data.
func (uc *authUsecase) cacheGet(ctx context.Context, id int64) (*domain.User, error) {
	span := trace.SpanFromContext(ctx)
	if uc.userCache != nil {
		u, err := uc.userCache.Get(ctx, id)
		if err == nil && u != nil {
			span.AddEvent("cache.hit", trace.WithAttributes(attribute.Int64("cache.key", id)))
			return u, nil
		}
		span.AddEvent("cache.miss", trace.WithAttributes(attribute.Int64("cache.key", id)))
	}
	u, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if uc.userCache != nil {
		if err := uc.userCache.Set(ctx, u, 0); err != nil {
			logger.Warn("cache: set failed", "error", err.Error())
		}
	}
	return u, nil
}

func cacheSet(ctx context.Context, cache port.UserCache, u *domain.User) {
	if cache != nil {
		if err := cache.Set(ctx, u, 0); err != nil {
			logger.Warn("cache: set failed", "error", err.Error())
		}
	}
}

func cacheDel(ctx context.Context, cache port.UserCache, id int64) {
	if cache != nil {
		if err := cache.Delete(ctx, id); err != nil {
			logger.Warn("cache: delete failed", "error", err.Error())
		}
	}
}

func eventPublish(publisher port.EventPublisher, ctx context.Context, eventName string, payload interface{}) {
	if err := publisher.Publish(ctx, eventName, payload); err != nil {
		logger.Warn("event: publish failed", "event", eventName, "error", err.Error())
	}
}

// auditLogin persists a login audit event. Errors are logged internally;
// callers intentionally discard the error — audit failures never block login.
// Returns error for testing purposes only.
func (uc *authUsecase) auditLogin(ctx context.Context, userID int64, email string, success bool, ip, userAgent, reason string) error {
	audit := &domain.LoginAudit{
		UserID:    userID,
		Email:     email,
		Success:   success,
		IPAddress: ip,
		UserAgent: userAgent,
		Reason:    reason,
		CreatedAt: time.Now().UTC(),
	}
	if err := uc.loginAuditRepo.Create(ctx, audit); err != nil {
		logger.Warn("audit: login audit failed", "error", err.Error())
		return err
	}
	return nil
}
