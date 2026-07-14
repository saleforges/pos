package main

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
	minioadapter "github.com/saleforge/pos/services/internal/catalog/adapter/storage/minio"
	"github.com/saleforge/pos/services/internal/catalog/bootstrap"
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

	addr := getEnv("CATALOG_PORT", ":8082")
	logger.Info("Catalog service starting", "addr", addr)

	app, err := bootstrap.New(bootstrap.Config{
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		IAMBaseURL:   os.Getenv("IAM_BASE_URL"),
		OtelEndpoint: os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		Minio: minioadapter.Config{
			Endpoint:  os.Getenv("MINIO_ENDPOINT"),
			AccessKey: os.Getenv("MINIO_ACCESS_KEY"),
			SecretKey: os.Getenv("MINIO_SECRET_KEY"),
			Bucket:    getEnv("MINIO_BUCKET", "catalog-dev"),
			PublicURL: getEnv("MINIO_PUBLIC_URL", "https://minio.saleforges.com"),
			UseSSL:    os.Getenv("MINIO_USE_SSL") == "true",
		},
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
