package main

import (
	"context"
	"fmt"
	"os"

	catalog "github.com/saleforge/pos/services/internal/catalog/adapter/repository/postgres"
	iam "github.com/saleforge/pos/services/internal/iam/adapter/repository/postgres"
	merchant "github.com/saleforge/pos/services/internal/merchant/adapter/repository/postgres"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://devuser:***@localhost:5432/devdb?sslmode=disable"
	}

	fmt.Println("Running IAM migrations...")
	if err := iam.RunMigrations(dsn); err != nil {
		fmt.Fprintf(os.Stderr, "IAM migration failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("IAM migrations done")

	fmt.Println("Running Merchant migrations...")
	if err := merchant.RunMigrations(dsn); err != nil {
		fmt.Fprintf(os.Stderr, "Merchant migration failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Merchant migrations done")

	fmt.Println("Running Catalog migrations...")
	if err := catalog.RunMigrations(dsn); err != nil {
		fmt.Fprintf(os.Stderr, "Catalog migration failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Catalog migrations done")

	fmt.Println("Seeding IAM data...")
	pool, err := iam.Connect(context.Background(), dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect failed: %v\n", err)
		os.Exit(1)
	}
	if err := iam.SeedData(context.Background(), pool); err != nil {
		fmt.Fprintf(os.Stderr, "seed failed: %v\n", err)
		os.Exit(1)
	}
	pool.Close()
	fmt.Println("IAM seed done")

	fmt.Println("Seeding Catalog data...")
	catPool, err := catalog.Connect(context.Background(), dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "catalog connect failed: %v\n", err)
		os.Exit(1)
	}
	if err := catalog.SeedData(context.Background(), catPool); err != nil {
		fmt.Fprintf(os.Stderr, "catalog seed failed: %v\n", err)
		os.Exit(1)
	}
	catPool.Close()
	fmt.Println("Catalog seed done")
}
