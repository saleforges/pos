package bootstrap

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/adapter/repository/memory"
	"github.com/saleforge/pos/services/internal/iam/adapter/repository/postgres"
	"github.com/saleforge/pos/services/internal/iam/adapter/security"
	"github.com/saleforge/pos/services/internal/iam/port/repository"
	httptransport "github.com/saleforge/pos/services/internal/iam/transport/http"
	"github.com/saleforge/pos/services/internal/iam/transport/http/handler"
	"github.com/saleforge/pos/services/internal/iam/usecase"
	"github.com/saleforge/pos/services/pkg/logger"
	"github.com/saleforge/pos/services/pkg/otel"
)

type Config struct {
	JWTPrivateKeyPEM string
	JWTKeyID         string
	DatabaseURL      string
}

type App struct {
	router       *echo.Echo
	otelShutdown func(context.Context) error
}

type noopEventPublisher struct{}

func (n *noopEventPublisher) Publish(_ context.Context, _ string, _ interface{}) error {
	return nil
}

func New(cfg Config) (*App, error) {
	otelShutdown, err := otel.Init(context.Background(), otel.Config{
		ServiceName: "iam-service",
		Environment: "development",
		UseGRPC:     true,
		Insecure:    true,
	})
	if err != nil {
		logger.Warn("failed to init otel, continuing without tracing", "error", err.Error())
	}

	passwordHasher := security.NewArgon2Hasher()
	tokenSigner, err := security.NewJWTSigner([]byte(cfg.JWTPrivateKeyPEM), cfg.JWTKeyID)
	if err != nil {
		return nil, err
	}
	eventPublisher := &noopEventPublisher{}

	var (
		userRepo         repository.UserRepository
		roleRepo         repository.RoleRepository
		permissionRepo   repository.PermissionRepository
		refreshTokenRepo repository.RefreshTokenRepository
		loginAuditRepo   repository.LoginAuditRepository
	)

	if cfg.DatabaseURL != "" {
		logger.Info("connecting to PostgreSQL...")
		ctx := context.Background()
		pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, err
		}

		if err := postgres.RunMigrations(cfg.DatabaseURL); err != nil {
			return nil, err
		}

		if err := postgres.SeedData(ctx, pool); err != nil {
			logger.Warn("seed data error", "error", err.Error())
		}

		userRepo = postgres.NewUserRepository(pool)
		roleRepo = postgres.NewRoleRepository(pool)
		permissionRepo = postgres.NewPermissionRepository(pool)
		refreshTokenRepo = postgres.NewRefreshTokenRepository(pool)
		loginAuditRepo = postgres.NewLoginAuditRepository(pool)

		logger.Info("PostgreSQL connected and migrated")
	} else {
		logger.Info("using in-memory storage")
		userRepo = memory.NewUserRepository()
		roleRepo = memory.NewRoleRepository()
		permissionRepo = memory.NewPermissionRepository()
		refreshTokenRepo = memory.NewRefreshTokenRepository()
		loginAuditRepo = memory.NewLoginAuditRepository()
	}

	authUsecase := usecase.NewAuthUsecase(
		userRepo,
		roleRepo,
		permissionRepo,
		refreshTokenRepo,
		loginAuditRepo,
		eventPublisher,
		passwordHasher,
		tokenSigner,
	)
	authHandler := handler.NewAuthHandler(authUsecase)
	router := httptransport.NewRouter(authHandler, authUsecase, tokenSigner)

	return &App{router: router, otelShutdown: otelShutdown}, nil
}

func (a *App) Run(addr string) error {
	return a.router.Start(addr)
}

func (a *App) Shutdown(ctx context.Context) error {
	if a.otelShutdown != nil {
		return a.otelShutdown(ctx)
	}
	return nil
}
