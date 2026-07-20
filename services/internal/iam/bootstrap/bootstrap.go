package bootstrap

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	iainmemory "github.com/saleforge/pos/services/internal/iam/adapter/cache/memory"
	redisadapter "github.com/saleforge/pos/services/internal/iam/adapter/cache/redis"
	"github.com/saleforge/pos/services/internal/iam/adapter/repository/memory"
	"github.com/saleforge/pos/services/internal/iam/adapter/repository/postgres"
	sessionmem "github.com/saleforge/pos/services/internal/iam/adapter/session/memory"
	sessionredis "github.com/saleforge/pos/services/internal/iam/adapter/session/redis"
	"github.com/saleforge/pos/services/internal/iam/adapter/security"
	"github.com/saleforge/pos/services/internal/iam/port"
	"github.com/saleforge/pos/services/internal/iam/port/repository"
	httpauth "github.com/saleforge/pos/services/internal/iam/transport/http/auth"
	httpperm "github.com/saleforge/pos/services/internal/iam/transport/http/permission"
	httprole "github.com/saleforge/pos/services/internal/iam/transport/http/role"
	httptransport "github.com/saleforge/pos/services/internal/iam/transport/http"
	httpuser "github.com/saleforge/pos/services/internal/iam/transport/http/user"
	"github.com/saleforge/pos/services/internal/iam/usecase"
	"github.com/saleforge/pos/services/pkg/logger"
	"github.com/saleforge/pos/services/pkg/otel"
)

type Config struct {
	JWTPrivateKeyPEM  string
	JWTKeyID          string
	DatabaseURL       string
	OtelEndpoint      string
	RedisAddr         string
	TokenHasherSecret string
	SecureCookies     bool
}

type App struct {
	router       *echo.Echo
	otelShutdown func(context.Context) error
}

type noopEventPublisher struct{}

func (n *noopEventPublisher) Publish(_ context.Context, eventName string, _ interface{}) error {
	logger.Debug("[event] WIP — no real adapter wired, event silently dropped", "event", eventName)
	return nil
}

func New(cfg Config) (*App, error) {
	otelShutdown, err := otel.Init(context.Background(), otel.Config{
		ServiceName:  "iam-service",
		Environment:  "development",
		OtelEndpoint: cfg.OtelEndpoint,
		UseGRPC:      true,
		Insecure:     true,
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
		loginAuditRepo   repository.LoginAuditRepository
		staffRepo        repository.StaffRepository
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
		loginAuditRepo = postgres.NewLoginAuditRepository(pool)
		staffRepo = postgres.NewStaffRepository(pool)

		logger.Info("PostgreSQL connected and migrated")
	} else {
		logger.Info("using in-memory storage")
		userRepo = memory.NewUserRepository()
		roleRepo = memory.NewRoleRepository()
		permissionRepo = memory.NewPermissionRepository()
		loginAuditRepo = memory.NewLoginAuditRepository()
		staffRepo = memory.NewStaffRepository()
	}

	var userCache port.UserCache

	if cfg.RedisAddr != "" {
		logger.Info("IAM cache using Redis", "addr", cfg.RedisAddr)
		userCache = redisadapter.NewUserCache(cfg.RedisAddr, 5*time.Minute)
	} else {
		logger.Info("IAM cache using in-memory (30s TTL)")
		userCache = iainmemory.NewUserCache(30*time.Second, 5*time.Minute)
	}

	var sessionStore port.SessionStore

	if cfg.RedisAddr != "" {
		logger.Info("IAM session store using Redis", "addr", cfg.RedisAddr)
		sessionStore = sessionredis.NewSessionStore(cfg.RedisAddr)
	} else {
		logger.Info("IAM session store using in-memory")
		sessionStore = sessionmem.NewSessionStore()
	}

	tokenHasher, err := security.NewHMAC256TokenHasher([]byte(cfg.TokenHasherSecret))
	if err != nil {
		return nil, err
	}
	authUsecase := usecase.NewAuthUsecase(
		userRepo,
		roleRepo,
		permissionRepo,
		loginAuditRepo,
		staffRepo,
		sessionStore,
		eventPublisher,
		passwordHasher,
		tokenSigner,
		tokenHasher,
		userCache,
	)
	authHandler := httpauth.NewHandler(authUsecase, authUsecase, cfg.SecureCookies)
	userHandler := httpuser.NewHandler(authUsecase)
	roleHandler := httprole.NewHandler(authUsecase)
	permHandler := httpperm.NewHandler(authUsecase)
	router := httptransport.NewRouter(authHandler, userHandler, roleHandler, permHandler, authUsecase, authUsecase.HasPermission, tokenSigner)

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
