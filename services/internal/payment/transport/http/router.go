package httptransport

import (
	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/payment/transport/http/middleware"
	"github.com/saleforge/pos/services/internal/payment/transport/http/payment"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/jwks"
	"github.com/saleforge/pos/services/pkg/otel"
)

func NewRouter(paymentHandler *payment.Handler, verifier *jwks.Verifier, internalKey string) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(otel.LoggingMiddleware())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})
	e.GET("/metrics", func(c echo.Context) error {
		otel.MetricsHandler().ServeHTTP(c.Response(), c.Request())
		return nil
	})

	// Gateway callback — public (gateway calls it), security via signature.
	e.POST("/api/v1/payments/ipaymu/callback", paymentHandler.Callback)

	// Merchant-facing API (JWT auth).
	api := e.Group("/api/v1", middleware.Auth(verifier))
	paymentGroup := api.Group("/payments", httputil.MerchantMiddleware())
	paymentGroup.GET("/:id", paymentHandler.GetByID)

	// Internal service-to-service endpoints.
	internal := e.Group("/internal", middleware.InternalAuth(internalKey))
	internal.POST("/payments", paymentHandler.Create)

	return e
}
