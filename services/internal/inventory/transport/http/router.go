package httptransport

import (
	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/inventory/transport/http/middleware"
	productcomponent "github.com/saleforge/pos/services/internal/inventory/transport/http/product_component"
	"github.com/saleforge/pos/services/internal/inventory/transport/http/stock"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/jwks"
	"github.com/saleforge/pos/services/pkg/otel"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

func NewRouter(
	stockHandler *stock.Handler,
	componentHandler *productcomponent.Handler,
	verifier *jwks.Verifier,
	internalKey string,
) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(otelecho.Middleware("inventory-service"))
	e.Use(otel.LoggingMiddleware())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})
	e.GET("/metrics", func(c echo.Context) error {
		otel.MetricsHandler().ServeHTTP(c.Response(), c.Request())
		return nil
	})

	api := e.Group("/api/v1", middleware.Auth(verifier))

	requireWrite := httputil.RequirePermission(httputil.PermInventoryWrite)
	requireAdjust := httputil.RequirePermission(httputil.PermInventoryAdjust)

	// Stock endpoints
	stockGroup := api.Group("/stocks", httputil.MerchantMiddleware())
	stockGroup.POST("", stockHandler.Create, requireWrite)
	stockGroup.GET("", stockHandler.List)
	stockGroup.POST("/transfer", stockHandler.Transfer, requireWrite)
	stockGroup.GET("/sync", stockHandler.Sync)
	stockGroup.GET("/:id", stockHandler.GetByID)
	stockGroup.PUT("/:id", stockHandler.Update, requireAdjust)

	// Product component endpoints
	componentGroup := api.Group("/product-items", httputil.MerchantMiddleware())
	componentGroup.GET("/:productItemId/components", componentHandler.GetByProductItem)
	componentGroup.PUT("/:productItemId/components", componentHandler.CreateOrUpdate, requireWrite)

	// Internal service-to-service endpoints (not exposed via Caddy)
	internal := e.Group("/internal", middleware.InternalAuth(internalKey))
	internal.POST("/stocks/deduct", stockHandler.Deduct)
	internal.POST("/stocks/restore", stockHandler.Restore)

	return e
}
