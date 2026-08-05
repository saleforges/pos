package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/saleforge/pos/services/internal/iam/bootstrap"
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

	addr := getEnv("HTTP_ADDR", ":8080")
	jwksURL := deriveJWKSURL(addr)
	logger.Info("IAM service starting", "addr", addr, "jwks_url", jwksURL)

	originsStr := os.Getenv("CORS_ALLOWED_ORIGINS")
	var allowedOrigins []string
	if originsStr != "" {
		for _, o := range strings.Split(originsStr, ",") {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	app, err := bootstrap.New(bootstrap.Config{
		JWTPrivateKeyPEM:  getEnv("JWT_PRIVATE_KEY_PEM", ""),
		JWTKeyID:          getEnv("JWT_KEY_ID", "iam-key-1"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		OtelEndpoint:      os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		RedisAddr:         os.Getenv("REDIS_ADDR"),
		TokenHasherSecret: getEnv("TOKEN_HASHER_SECRET", ""),
		SecureCookies:     os.Getenv("SECURE_COOKIES") != "false",
		AllowedOrigins:    allowedOrigins,
		LoginRateLimit:    getEnvInt("LOGIN_RATE_LIMIT", 5),
		LoginRateWindow:   time.Duration(getEnvInt("LOGIN_RATE_WINDOW", 60)) * time.Second,
		RefreshRateLimit:  getEnvInt("REFRESH_RATE_LIMIT", 20),
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

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
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
