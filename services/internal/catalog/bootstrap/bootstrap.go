package bootstrap

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/catalog/adapter/repository/memory"
	"github.com/saleforge/pos/services/internal/catalog/adapter/repository/postgres"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
	httptransport "github.com/saleforge/pos/services/internal/catalog/transport/http"
	"github.com/saleforge/pos/services/internal/catalog/transport/http/handler"
	"github.com/saleforge/pos/services/internal/catalog/usecase"
	"github.com/saleforge/pos/services/pkg/logger"
	"github.com/saleforge/pos/services/pkg/otel"
)

type Config struct {
	DatabaseURL  string
	OtelEndpoint string
}

type App struct {
	router       *echo.Echo
	otelShutdown func(context.Context) error
}

func New(cfg Config) (*App, error) {
	otelShutdown, err := otel.Init(context.Background(), otel.Config{
		ServiceName:  "catalog-service",
		Environment:  "development",
		OtelEndpoint: cfg.OtelEndpoint,
		UseGRPC:      true,
		Insecure:     true,
	})
	if err != nil {
		logger.Warn("failed to init otel, continuing without tracing", "error", err.Error())
	}

	var (
		catRepo  repository.CategoryRepository
		prodRepo repository.ProductRepository
		varRepo  repository.VariantRepository
	)

	if cfg.DatabaseURL != "" {
		ctx := context.Background()
		pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, err
		}
		if err := postgres.RunMigrations(cfg.DatabaseURL); err != nil {
			return nil, err
		}
		catRepo = postgres.NewCategoryRepository(pool)
		prodRepo = postgres.NewProductRepository(pool)
		varRepo = postgres.NewVariantRepository(pool)
		logger.Info("using postgres storage")
	} else {
		catRepo = memory.NewCategoryRepository()
		prodRepo = memory.NewProductRepository()
		varRepo = memory.NewVariantRepository()
		logger.Info("using in-memory storage")
	}

	catUC := usecase.NewCategoryUsecase(catRepo)
	prodUC := usecase.NewProductUsecase(prodRepo, catRepo)
	varUC := usecase.NewVariantUsecase(varRepo, prodRepo)

	catHandler := handler.NewCategoryHandler(catUC)
	prodHandler := handler.NewProductHandler(prodUC)
	varHandler := handler.NewVariantHandler(varUC)

	router := httptransport.NewRouter(catHandler, prodHandler, varHandler)

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
