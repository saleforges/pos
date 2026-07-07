package bootstrap

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/merchant/adapter/iam"
	"github.com/saleforge/pos/services/internal/merchant/adapter/repository/memory"
	"github.com/saleforge/pos/services/internal/merchant/adapter/repository/postgres"
	httptransport "github.com/saleforge/pos/services/internal/merchant/transport/http"
	"github.com/saleforge/pos/services/internal/merchant/transport/http/handler"
	"github.com/saleforge/pos/services/internal/merchant/usecase"
	"github.com/saleforge/pos/services/pkg/logger"
	"github.com/saleforge/pos/services/pkg/otel"
	por "github.com/saleforge/pos/services/internal/merchant/port/repository"
)

type Config struct {
	DatabaseURL string
	IAMBaseURL  string
}

type App struct {
	router    *echo.Echo
	otelShutdown func(context.Context) error
}

func New(cfg Config) (*App, error) {
	iamBaseURL := cfg.IAMBaseURL
	if iamBaseURL == "" {
		iamBaseURL = "http://localhost:8080"
	}
	tokenValidator := iam.NewTokenValidator(iamBaseURL)

	otelShutdown, err := otel.Init(context.Background(), otel.Config{
		ServiceName: "merchant-service",
		Environment: "development",
		UseGRPC:     true,
		Insecure:    true,
	})
	if err != nil {
		logger.Warn("failed to init otel, continuing without tracing", "error", err.Error())
	}

	var merchantRepo por.MerchantRepository
	var branchRepo por.BranchRepository
	var staffRepo por.StaffRepository

	if cfg.DatabaseURL != "" {
		ctx := context.Background()
		pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, err
		}
		if err := postgres.RunMigrations(cfg.DatabaseURL); err != nil {
			return nil, err
		}
		merchantRepo = postgres.NewMerchantRepository(pool)
		branchRepo = postgres.NewBranchRepository(pool)
		staffRepo = postgres.NewStaffRepository(pool)
		logger.Info("using postgres storage")
	} else {
		merchantRepo = memory.NewMerchantRepository()
		branchRepo = memory.NewBranchRepository()
		staffRepo = memory.NewStaffRepository()
		logger.Info("using in-memory storage")
	}

	uc := usecase.NewMerchantUsecase(merchantRepo, branchRepo, staffRepo)
	merchantHandler := handler.NewMerchantHandler(uc)
	branchHandler := handler.NewBranchHandler(uc)
	staffHandler := handler.NewStaffHandler(uc)
	e := httptransport.NewRouter(merchantHandler, branchHandler, staffHandler, tokenValidator, staffRepo)

	return &App{router: e, otelShutdown: otelShutdown}, nil
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