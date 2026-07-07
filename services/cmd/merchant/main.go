package main

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/saleforge/pos/services/internal/merchant/bootstrap"
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

	addr := getEnv("MERCHANT_PORT", ":8081")
	logger.Info("starting merchant service", "addr", addr)
	app, err := bootstrap.New(bootstrap.Config{
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		OtelEndpoint: os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
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

func deriveJWKSURL(addr string) string {
	host := addr
	if strings.HasPrefix(host, ":") {
		host = "localhost" + host
	}
	if !strings.Contains(host, "://") {
		host = "http://" + host
	}
	return strings.TrimRight(host, "/") + "/.well-known/jwks.json"
}
