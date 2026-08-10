package httputil

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const ContextKeyPermissions = "permissions"

// Permission strings must match services/internal/iam/domain/role.go exactly —
// that file is the source of truth for what gets minted into the JWT.
const (
	PermCatalogCreate   = "catalog.create"
	PermCatalogUpdate   = "catalog.update"
	PermCatalogDelete   = "catalog.delete"
	PermInventoryWrite  = "inventory.write"
	PermInventoryAdjust = "inventory.adjust"
)

// RequirePermission blocks the request unless the caller's JWT carries the
// given permission. Must run after an Auth middleware that already set
// ContextKeyPermissions from the verified token's claims.
func RequirePermission(permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			perms, _ := c.Get(ContextKeyPermissions).([]string)
			for _, p := range perms {
				if p == permission {
					return next(c)
				}
			}
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "missing required permission: " + permission,
			})
		}
	}
}
