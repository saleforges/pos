package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/catalog/transport/http/handler"
	"github.com/saleforge/pos/services/pkg/otel"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

func NewRouter(
	catHandler *handler.CategoryHandler,
	prodHandler *handler.ProductHandler,
	varHandler *handler.VariantHandler,
	imgHandler *handler.ImageHandler,
) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(otelecho.Middleware("catalog-service"))
	e.Use(otel.LoggingMiddleware())

	e.GET("/metrics", func(c echo.Context) error {
		otel.MetricsHandler().ServeHTTP(c.Response(), c.Request())
		return nil
	})

	api := e.Group("/api/v1")

	api.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	cat := api.Group("/merchants/:merchantID/categories")
	cat.POST("", catHandler.Create)
	cat.GET("", catHandler.List)
	cat.GET("/:id", catHandler.GetByID)
	cat.PUT("/:id", catHandler.Update)
	cat.DELETE("/:id", catHandler.Delete)

	prod := api.Group("/merchants/:merchantID/products")
	prod.POST("", prodHandler.Create)
	prod.GET("", prodHandler.List)
	prod.GET("/:id", prodHandler.GetByID)
	prod.PUT("/:id", prodHandler.Update)
	prod.DELETE("/:id", prodHandler.Delete)

	prod.POST("/:productID/variants", varHandler.Create)
	prod.GET("/:productID/variants", varHandler.ListByProduct)
	prod.PUT("/variants/:id", varHandler.Update)
	prod.DELETE("/variants/:id", varHandler.Delete)

	if imgHandler != nil {
		img := api.Group("/merchants/:merchantID/images")
		img.POST("", imgHandler.Upload)
	}

	return e
}
