package httptransport

import (
	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/catalog/transport/http/category"
	"github.com/saleforge/pos/services/internal/catalog/transport/http/image"
	"github.com/saleforge/pos/services/internal/catalog/transport/http/middleware"
	"github.com/saleforge/pos/services/internal/catalog/transport/http/product"
	productitem "github.com/saleforge/pos/services/internal/catalog/transport/http/product_item"
	"github.com/saleforge/pos/services/internal/catalog/transport/http/sync"
	"github.com/saleforge/pos/services/internal/catalog/transport/http/unit"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/jwks"
	"github.com/saleforge/pos/services/pkg/otel"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

func NewRouter(
	productHandler *product.Handler,
	itemHandler *productitem.Handler,
	catHandler *category.Handler,
	unitHandler *unit.Handler,
	imgHandler *image.Handler,
	syncHandler *sync.Handler,
	verifier *jwks.Verifier,
) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(otelecho.Middleware("catalog-service"))
	e.Use(otel.LoggingMiddleware())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})
	e.GET("/metrics", func(c echo.Context) error {
		otel.MetricsHandler().ServeHTTP(c.Response(), c.Request())
		return nil
	})

	api := e.Group("/api/v1", middleware.Auth(verifier))

	api.GET("/units", unitHandler.List)

	cat := api.Group("/categories", httputil.MerchantMiddleware())
	cat.POST("", catHandler.Create)
	cat.GET("", catHandler.List)
	cat.GET("/:id", catHandler.Get)
	cat.PATCH("/:id", catHandler.Update)
	cat.PATCH("/:id/restore", catHandler.Restore)
	cat.DELETE("/:id", catHandler.Delete)

	prod := api.Group("/products", httputil.MerchantMiddleware())
	prod.POST("", productHandler.Create)
	prod.POST("/bulk", productHandler.BulkCreate)
	prod.PATCH("/bulk/:id", productHandler.BulkUpdate)
	prod.PATCH("/:id/restore", productHandler.Restore)
	prod.GET("", productHandler.List)
	prod.GET("/:id", productHandler.Get)
	prod.PATCH("/:id", productHandler.Update)
	prod.DELETE("/:id", productHandler.Delete)

	// Product items nested under product
	prod.POST("/:productId/items", itemHandler.Create)
	prod.GET("/:productId/items", itemHandler.ListByProduct)

	// Standalone product-item endpoints
	api.GET("/product-items", itemHandler.ListByMerchant)
	api.GET("/product-items/:id", itemHandler.GetByID)
	api.PATCH("/product-items/:id", itemHandler.Update)
	api.PATCH("/product-items/:id/restore", itemHandler.Restore)
	api.DELETE("/product-items/:id", itemHandler.Delete)
	api.PUT("/product-items/:id/branch-price", itemHandler.SetBranchPrice)
	api.DELETE("/product-items/:id/branch-price", itemHandler.DeleteBranchPrice)

	// Image upload
	if imgHandler != nil {
		img := api.Group("/images", httputil.MerchantMiddleware())
		img.POST("", imgHandler.Upload)
	}

	// Mobile sync endpoint
	api.POST("/sync", syncHandler.Sync, httputil.MerchantMiddleware())

	return e
}
