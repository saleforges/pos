package bootstrap

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/payment/adapter/client/ipaymu"
	orderclient "github.com/saleforge/pos/services/internal/payment/adapter/client/order"
	"github.com/saleforge/pos/services/internal/payment/adapter/repository/memory"
	"github.com/saleforge/pos/services/internal/payment/adapter/repository/postgres"
	por "github.com/saleforge/pos/services/internal/payment/port/repository"
	httptransport "github.com/saleforge/pos/services/internal/payment/transport/http"
	"github.com/saleforge/pos/services/internal/payment/transport/http/payment"
	"github.com/saleforge/pos/services/internal/payment/usecase"
	"github.com/saleforge/pos/services/pkg/jwks"
	"github.com/saleforge/pos/services/pkg/logger"
	"github.com/saleforge/pos/services/pkg/otel"
)

type Config struct {
	DatabaseURL     string
	IAMBaseURL      string
	OtelEndpoint    string
	InternalKey     string
	IpaymuBaseURL   string
	IpaymuVA        string
	IpaymuAPIKey    string
	IpaymuReturnURL string
	IpaymuCancelURL string
	IpaymuNotifyURL string
	OrderBaseURL    string
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
		ServiceName:  "payment-service",
		Environment:  "development",
		OtelEndpoint: cfg.OtelEndpoint,
		UseGRPC:      true,
		Insecure:     true,
	})
	if err != nil {
		logger.Warn("failed to init otel, continuing without tracing", "error", err.Error())
	}

	var paymentRepo por.PaymentRepository

	if cfg.DatabaseURL != "" {
		ctx := context.Background()
		pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, err
		}
		if err := postgres.RunMigrations(cfg.DatabaseURL); err != nil {
			return nil, err
		}
		paymentRepo = postgres.NewPaymentRepository(pool)
		logger.Info("using postgres storage")
	} else {
		paymentRepo = memory.NewPaymentRepository()
		logger.Info("using in-memory storage")
	}

	gateway := ipaymu.New(ipaymu.Config{
		BaseURL:   cfg.IpaymuBaseURL,
		VA:        cfg.IpaymuVA,
		APIKey:    cfg.IpaymuAPIKey,
		ReturnURL: cfg.IpaymuReturnURL,
		CancelURL: cfg.IpaymuCancelURL,
		NotifyURL: cfg.IpaymuNotifyURL,
	})

	orderClient := orderclient.New(orderclient.Config{
		BaseURL: cfg.OrderBaseURL,
		APIKey:  cfg.InternalKey,
	})

	paymentUC := usecase.NewPaymentUsecase(paymentRepo, gateway, orderClient)
	paymentHandler := payment.NewHandler(paymentUC, cfg.IpaymuVA)

	e := httptransport.NewRouter(paymentHandler, verifier, cfg.InternalKey)

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
