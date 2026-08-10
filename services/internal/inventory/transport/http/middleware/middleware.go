package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/jwks"
)

const (
	ContextKeyUserID     = "user_id"
	ContextKeyUserType   = "user_type"
	ContextKeyMerchantID = "merchant_id"
)

func Auth(verifier *jwks.Verifier) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := extractToken(c)
			if token == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
			}

			claims, err := verifier.Verify(token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			}

			c.Set(ContextKeyUserID, claims.UserID)
			c.Set(ContextKeyUserType, claims.UserType)
			c.Set(ContextKeyMerchantID, claims.MerchantID)
			c.Set(httputil.ContextKeyPermissions, claims.Permissions)

			return next(c)
		}
	}
}

func extractToken(c echo.Context) string {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}
	cookie, err := c.Cookie("access_token")
	if err != nil {
		return ""
	}
	return cookie.Value
}
