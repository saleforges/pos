package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/saleforge/pos/services/internal/order/bootstrap"
	"github.com/saleforge/pos/services/pkg/logger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		logger.Warn("no .env file loaded", "error", err.Error())
	}

	logger.Init(logger.Config{
		Level:  os.Getenv("LOG_LEVEL"),
		Format: os.Getenv("LOG_FORMAT"),
	})

	addr := getEnv("ORDER_PORT", ":8084")
	logger.Info("Order service starting", "addr", addr)

	app, err := bootstrap.New(bootstrap.Config{
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		IAMBaseURL:       os.Getenv("IAM_BASE_URL"),
		OtelEndpoint:     os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		InventoryBaseURL: getEnv("INVENTORY_BASE_URL", "http://inventory-service:8083"),
		InventoryAPIKey:  os.Getenv("INTERNAL_API_KEY"),
	})
	if err != nil {
		logger.Error("bootstrap failed", "error", err.Error())
		os.Exit(1)
	}
	if err := app.Run(addr); err != nil {
		logger.Error("server failed", "error", err.Error())
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
