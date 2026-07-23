package bootstrap

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/catalog/adapter/repository/memory"
	"github.com/saleforge/pos/services/internal/catalog/adapter/repository/postgres"
	minioadapter "github.com/saleforge/pos/services/internal/catalog/adapter/storage/minio"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
	h "github.com/saleforge/pos/services/internal/catalog/transport/http"
	"github.com/saleforge/pos/services/internal/catalog/transport/http/category"
	"github.com/saleforge/pos/services/internal/catalog/transport/http/image"
	"github.com/saleforge/pos/services/internal/catalog/transport/http/product"
	"github.com/saleforge/pos/services/internal/catalog/transport/http/product_item"
	"github.com/saleforge/pos/services/internal/catalog/transport/http/unit"
	"github.com/saleforge/pos/services/internal/catalog/usecase"
	"github.com/saleforge/pos/services/pkg/jwks"
	"github.com/saleforge/pos/services/pkg/logger"
	"github.com/saleforge/pos/services/pkg/otel"
)

type Config struct {
	DatabaseURL  string
	IAMBaseURL   string
	OtelEndpoint string
	Minio        minioadapter.Config
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
		prodRepo  repository.ProductRepository
		itemRepo  repository.ProductItemRepository
		catRepo   repository.CategoryRepository
		unitRepo  repository.UnitRepository
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
		prodRepo = postgres.NewProductRepository(pool)
		itemRepo = postgres.NewProductItemRepository(pool)
		catRepo = postgres.NewCategoryRepository(pool)
		unitRepo = postgres.NewUnitRepository(pool)

		if err := postgres.SeedData(ctx, pool); err != nil {
			logger.Warn("seed failed", "error", err.Error())
		}

		logger.Info("using postgres storage")
	} else {
		prodRepo = memory.NewProductRepository()
		itemRepo = memory.NewProductItemRepository()
		catRepo = memory.NewCategoryRepository()
		unitRepo = memory.NewUnitRepository()
		logger.Info("using in-memory storage")
	}

	prodUC := usecase.NewProductUsecase(prodRepo, catRepo, unitRepo)
	itemUC := usecase.NewProductItemUsecase(itemRepo, prodRepo, unitRepo)
	catUC := usecase.NewCategoryUsecase(catRepo)
	unitUC := usecase.NewUnitUsecase(unitRepo)

	prodHandler := product.NewHandler(prodUC, catRepo, itemRepo, unitRepo)
	itemHandler := productitem.NewHandler(itemUC)
	catHandler := category.NewHandler(catUC)
	unitHandler := unit.NewHandler(unitUC)

	var imgHandler *image.Handler
	if cfg.Minio.Endpoint != "" {
		store, err := minioadapter.New(cfg.Minio)
		if err != nil {
			return nil, err
		}
		imgHandler = image.NewHandler(store)
		logger.Info("minio storage enabled")
	} else {
		logger.Warn("minio not configured, image upload disabled")
	}

	jwksURL := cfg.IAMBaseURL
	if jwksURL == "" {
		jwksURL = "http://iam-service:8080"
	}
	jwksURL = fmt.Sprintf("%s/.well-known/jwks.json", jwksURL)
	verifier := jwks.New(jwksURL)

	router := h.NewRouter(prodHandler, itemHandler, catHandler, unitHandler, imgHandler, verifier)

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
