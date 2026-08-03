package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/saleforge/pos/services/internal/payment/bootstrap"
	"github.com/saleforge/pos/services/pkg/logger"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	if err := godotenv.Load(); err != nil {
		logger.Warn("no .env file loaded", "error", err.Error())
	}

	addr := getEnv("PAYMENT_PORT", ":8085")
	logger.Info("Payment service starting", "addr", addr)

	app, err := bootstrap.New(bootstrap.Config{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		IAMBaseURL:      os.Getenv("IAM_BASE_URL"),
		OtelEndpoint:    os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		InternalKey:     os.Getenv("INTERNAL_API_KEY"),
		IpaymuBaseURL:   getEnv("IPAYMU_BASE_URL", "https://sandbox.ipaymu.com"),
		IpaymuVA:        os.Getenv("IPAYMU_VA"),
		IpaymuAPIKey:    os.Getenv("IPAYMU_API_KEY"),
		IpaymuReturnURL: getEnv("IPAYMU_RETURN_URL", ""),
		IpaymuCancelURL: getEnv("IPAYMU_CANCEL_URL", ""),
		IpaymuNotifyURL: getEnv("IPAYMU_NOTIFY_URL", ""),
		OrderBaseURL:    getEnv("ORDER_BASE_URL", "http://order-service:8084"),
	})
	if err != nil {
		logger.Error("bootstrap failed", "error", err.Error())
		os.Exit(1)
	}

	if err := app.Run(addr); err != nil {
		logger.Error("server stopped", "error", err.Error())
		os.Exit(1)
	}
}
