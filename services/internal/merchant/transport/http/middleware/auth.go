package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/merchant/port"
)

const (
	ContextKeyUserID       = "user_id"
	ContextKeyRoles        = "roles"
	ContextKeyPermissions  = "permissions"
	ContextKeyMerchantID   = "merchant_id"
	ContextKeyMerchantRole = "merchant_role"
)

func Auth(tokenValidator port.TokenValidator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid authorization header format"})
			}

			claims, err := tokenValidator.Validate(parts[1])
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			}

			c.Set(ContextKeyUserID, claims.UserID)
			c.Set(ContextKeyRoles, claims.Roles)
			c.Set(ContextKeyPermissions, claims.Permissions)
			c.Set(ContextKeyMerchantID, claims.MerchantID)
			c.Set(ContextKeyMerchantRole, claims.MerchantRole)

			return next(c)
		}
	}
}
