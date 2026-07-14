package bootstrap

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/catalog/adapter/repository/memory"
	"github.com/saleforge/pos/services/internal/catalog/adapter/repository/postgres"
	minioadapter "github.com/saleforge/pos/services/internal/catalog/adapter/storage/minio"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
	httptransport "github.com/saleforge/pos/services/internal/catalog/transport/http"
	"github.com/saleforge/pos/services/internal/catalog/transport/http/handler"
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

	var imgHandler *handler.ImageHandler
	if cfg.Minio.Endpoint != "" {
		store, err := minioadapter.New(cfg.Minio)
		if err != nil {
			return nil, err
		}
		imgHandler = handler.NewImageHandler(store)
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

	router := httptransport.NewRouter(catHandler, prodHandler, varHandler, imgHandler, verifier)

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
