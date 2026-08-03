package bootstrap

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/order/adapter/client/inventory"
	"github.com/saleforge/pos/services/internal/order/adapter/repository/memory"
	"github.com/saleforge/pos/services/internal/order/adapter/repository/postgres"
	por "github.com/saleforge/pos/services/internal/order/port/repository"
	httptransport "github.com/saleforge/pos/services/internal/order/transport/http"
	"github.com/saleforge/pos/services/internal/order/transport/http/customer"
	"github.com/saleforge/pos/services/internal/order/transport/http/order"
	"github.com/saleforge/pos/services/internal/order/usecase"
	"github.com/saleforge/pos/services/pkg/jwks"
	"github.com/saleforge/pos/services/pkg/logger"
	"github.com/saleforge/pos/services/pkg/otel"
)

type Config struct {
	DatabaseURL      string
	IAMBaseURL       string
	OtelEndpoint     string
	InventoryBaseURL string
	InventoryAPIKey  string
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
		ServiceName:  "order-service",
		Environment:  "development",
		OtelEndpoint: cfg.OtelEndpoint,
		UseGRPC:      true,
		Insecure:     true,
	})
	if err != nil {
		logger.Warn("failed to init otel, continuing without tracing", "error", err.Error())
	}

	var (
		orderRepo    por.OrderRepository
		customerRepo por.CustomerRepository
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
		orderRepo = postgres.NewOrderRepository(pool)
		customerRepo = postgres.NewCustomerRepository(pool)
		logger.Info("using postgres storage")
	} else {
		orderRepo = memory.NewOrderRepository()
		customerRepo = memory.NewCustomerRepository()
		logger.Info("using in-memory storage")
	}

	orderUC := usecase.NewOrderUsecase(orderRepo, customerRepo, inventory.New(inventory.Config{
		BaseURL: cfg.InventoryBaseURL,
		APIKey:  cfg.InventoryAPIKey,
	}))
	customerUC := usecase.NewCustomerUsecase(customerRepo)

	orderHandler := order.NewHandler(orderUC)
	customerHandler := customer.NewHandler(customerUC)

	e := httptransport.NewRouter(orderHandler, customerHandler, verifier)

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
