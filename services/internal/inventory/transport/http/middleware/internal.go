package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/labstack/echo/v4"
)

// InternalAuth guards service-to-service endpoints with a shared key.
// These routes are not exposed through Caddy — only reachable inside the
// docker network by other services.
func InternalAuth(key string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			provided := c.Request().Header.Get("X-Internal-Key")
			if key == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(key)) != 1 {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid internal key"})
			}
			return next(c)
		}
	}
}
