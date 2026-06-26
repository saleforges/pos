package bootstrap

import (
	"context"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/iam/internal/adapter/repository/memory"
	"github.com/saleforge/pos/services/iam/internal/adapter/repository/postgres"
	"github.com/saleforge/pos/services/iam/internal/adapter/security"
	"github.com/saleforge/pos/services/iam/internal/port/repository"
	httptransport "github.com/saleforge/pos/services/iam/internal/transport/http"
	"github.com/saleforge/pos/services/iam/internal/transport/http/handler"
	"github.com/saleforge/pos/services/iam/internal/usecase"
)

type Config struct {
	JWTPrivateKeyPEM string
	JWTKeyID         string
	DatabaseURL      string
}

type App struct {
	router *echo.Echo
}

type noopEventPublisher struct{}

func (n *noopEventPublisher) Publish(_ context.Context, _ string, _ interface{}) error {
	return nil
}

func New(cfg Config) (*App, error) {
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
		log.Println("connecting to PostgreSQL...")
		ctx := context.Background()
		pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, err
		}

		if err := postgres.RunMigrations(cfg.DatabaseURL); err != nil {
			return nil, err
		}

		if err := postgres.SeedData(ctx, pool); err != nil {
			log.Printf("warning: seed data error: %v", err)
		}

		userRepo = postgres.NewUserRepository(pool)
		roleRepo = postgres.NewRoleRepository(pool)
		permissionRepo = postgres.NewPermissionRepository(pool)
		refreshTokenRepo = postgres.NewRefreshTokenRepository(pool)
		loginAuditRepo = postgres.NewLoginAuditRepository(pool)

		log.Println("PostgreSQL connected and migrated")
	} else {
		log.Println("using in-memory storage")
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

	return &App{router: router}, nil
}

func (a *App) Run(addr string) error {
	return a.router.Start(addr)
}
