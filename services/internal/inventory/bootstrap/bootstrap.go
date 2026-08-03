package bootstrap

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/inventory/adapter/repository/memory"
	"github.com/saleforge/pos/services/internal/inventory/adapter/repository/postgres"
	por "github.com/saleforge/pos/services/internal/inventory/port/repository"
	httptransport "github.com/saleforge/pos/services/internal/inventory/transport/http"
	productcomponent "github.com/saleforge/pos/services/internal/inventory/transport/http/product_component"
	"github.com/saleforge/pos/services/internal/inventory/transport/http/stock"
	"github.com/saleforge/pos/services/internal/inventory/usecase"
	"github.com/saleforge/pos/services/pkg/jwks"
	"github.com/saleforge/pos/services/pkg/logger"
	"github.com/saleforge/pos/services/pkg/otel"
)

type Config struct {
	DatabaseURL  string
	IAMBaseURL   string
	OtelEndpoint string
	InternalKey  string
}

type App struct {
	router       *echo.Echo
	otelShutdown func(context.Context) error
}

func New(cfg Config) (*App, error) {
	jwksURL := cfg.IAMBaseURL
	if jwksURL == "" {
		jwksURL = "http://localhost:8080"
	}
	jwksURL = fmt.Sprintf("%s/.well-known/jwks.json", jwksURL)
	verifier := jwks.New(jwksURL)

	otelShutdown, err := otel.Init(context.Background(), otel.Config{
		ServiceName:  "inventory-service",
		Environment:  "development",
		OtelEndpoint: cfg.OtelEndpoint,
		UseGRPC:      true,
		Insecure:     true,
	})
	if err != nil {
		logger.Warn("failed to init otel, continuing without tracing", "error", err.Error())
	}

	var (
		stockRepo       por.StockRepository
		stockAdjustRepo por.StockAdjustmentRepository
		componentRepo   por.ProductComponentRepository
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
		stockRepo = postgres.NewStockRepository(pool)
		stockAdjustRepo = postgres.NewStockRepository(pool)
		componentRepo = postgres.NewProductComponentRepository(pool)
		logger.Info("using postgres storage")
	} else {
		stockRepo = memory.NewStockRepository()
		stockAdjustRepo = memory.NewStockRepository()
		componentRepo = memory.NewProductComponentRepository()
		logger.Info("using in-memory storage")
	}

	stockUC := usecase.NewStockUsecase(stockRepo, stockAdjustRepo)
	componentUC := usecase.NewProductComponentUsecase(componentRepo)

	stockHandler := stock.NewHandler(stockUC)
	componentHandler := productcomponent.NewHandler(componentUC)

	e := httptransport.NewRouter(stockHandler, componentHandler, verifier, cfg.InternalKey)

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
