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
	verifier *jwks.Verifier,
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

	api := e.Group("/api/v1", middleware.Auth(verifier))

	orderGroup := api.Group("/orders", httputil.MerchantMiddleware())
	orderGroup.POST("", orderHandler.Create)
	orderGroup.GET("", orderHandler.List)
	orderGroup.GET("/:id", orderHandler.GetByID)
	orderGroup.PATCH("/:id/status", orderHandler.Cancel)
	orderGroup.PATCH("/:id/due-date", orderHandler.UpdateDueDate)
	orderGroup.POST("/:id/payments", orderHandler.AddPayment)

	customerGroup := api.Group("/customers", httputil.MerchantMiddleware())
	customerGroup.POST("", customerHandler.Create)
	customerGroup.GET("", customerHandler.List)
	customerGroup.GET("/:id", customerHandler.GetByID)
	customerGroup.PATCH("/:id", customerHandler.Update)
	customerGroup.DELETE("/:id", customerHandler.Delete)

	return e
}
