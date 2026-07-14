package http

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/catalog/transport/http/handler"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/jwks"
	"github.com/saleforge/pos/services/pkg/otel"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

func authMiddleware(verifier *jwks.Verifier) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := ""
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
					token = parts[1]
				}
			}
			if token == "" {
				if cookie, err := c.Cookie("access_token"); err == nil {
					token = cookie.Value
				}
			}
			if token == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			claims, err := verifier.Verify(token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			}

			c.Set(httputil.ContextKeyMerchantID, claims.MerchantID)
			c.Set(httputil.ContextKeyUserType, claims.UserType)
			return next(c)
		}
	}
}

func NewRouter(
	catHandler *handler.CategoryHandler,
	prodHandler *handler.ProductHandler,
	varHandler *handler.VariantHandler,
	imgHandler *handler.ImageHandler,
	verifier *jwks.Verifier,
) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(otelecho.Middleware("catalog-service"))
	e.Use(otel.LoggingMiddleware())

	e.GET("/metrics", func(c echo.Context) error {
		otel.MetricsHandler().ServeHTTP(c.Response(), c.Request())
		return nil
	})

	api := e.Group("/api/v1", authMiddleware(verifier))

	api.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	cat := api.Group("/categories", httputil.MerchantMiddleware())
	cat.POST("", catHandler.Create)
	cat.GET("", catHandler.List)
	cat.GET("/:id", catHandler.GetByID)
	cat.PUT("/:id", catHandler.Update)
	cat.DELETE("/:id", catHandler.Delete)

	prod := api.Group("/products", httputil.MerchantMiddleware())
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
		img := api.Group("/images", httputil.MerchantMiddleware())
		img.POST("", imgHandler.Upload)
	}

	return e
}
