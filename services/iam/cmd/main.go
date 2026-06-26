package main

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/saleforge/pos/services/iam/internal/bootstrap"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("no .env file loaded: %v", err)
	}

	addr := getEnv("HTTP_ADDR", ":8080")
	jwksURL := deriveJWKSURL(addr)
	log.Printf("IAM service starting on %s", addr)
	log.Printf("JWKS published at %s", jwksURL)
	app, err := bootstrap.New(bootstrap.Config{
		JWTPrivateKeyPEM: getEnv("JWT_PRIVATE_KEY_PEM", ""),
		JWTKeyID:         getEnv("JWT_KEY_ID", "iam-key-1"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
	})
	if err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}
	if err := app.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
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
