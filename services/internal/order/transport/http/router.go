package httptransport

import (
	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/order/transport/http/customer"
	"github.com/saleforge/pos/services/internal/order/transport/http/middleware"
	"github.com/saleforge/pos/services/internal/order/transport/http/order"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/jwks"
	"github.com/saleforge/pos/services/pkg/otel"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

func NewRouter(
	orderHandler *order.Handler,
	customerHandler *customer.Handler,
	internalHandler *order.InternalHandler,
	verifier *jwks.Verifier,
	internalKey string,
) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(otelecho.Middleware("order-service"))
	e.Use(otel.LoggingMiddleware())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})
	e.GET("/metrics", func(c echo.Context) error {
		otel.MetricsHandler().ServeHTTP(c.Response(), c.Request())
		return nil
	})

	// Internal service-to-service endpoints (called by payment service).
	internal := e.Group("/internal", middleware.InternalAuth(internalKey))
	internal.GET("/orders/:id", internalHandler.GetOrder)
	internal.POST("/orders/:id/paid", internalHandler.NotifyPaid)

	api := e.Group("/api/v1", middleware.Auth(verifier))

	orderGroup := api.Group("/orders", httputil.MerchantMiddleware())
	orderGroup.POST("", orderHandler.Create)
	orderGroup.GET("", orderHandler.List)
	orderGroup.GET("/:id", orderHandler.GetByID)
	orderGroup.PATCH("/:id", orderHandler.Update)
	orderGroup.PATCH("/:id/status", orderHandler.Cancel)
	orderGroup.POST("/:id/payments", orderHandler.AddPayment)
	api.GET("/reports/sales", orderHandler.SalesReport, httputil.MerchantMiddleware())

	customerGroup := api.Group("/customers", httputil.MerchantMiddleware())
	customerGroup.POST("", customerHandler.Create)
	customerGroup.GET("", customerHandler.List)
	customerGroup.GET("/sync", customerHandler.Sync)
	customerGroup.GET("/:id", customerHandler.GetByID)
	customerGroup.PATCH("/:id", customerHandler.Update)
	customerGroup.DELETE("/:id", customerHandler.Delete)

	return e
}
