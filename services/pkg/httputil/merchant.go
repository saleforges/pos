package httputil

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	ContextKeyMerchantID = "merchant_id"
	ContextKeyUserType   = "user_type"
)

func MerchantMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			merchantID := c.Request().Header.Get("X-Merchant-Id")
			userType, _ := c.Get(ContextKeyUserType).(string)

			// Platform user: header optional
			if userType == "platform" {
				if merchantID != "" {
					c.Set(ContextKeyMerchantID, merchantID)
				}
				return next(c)
			}

			// Merchant user or unknown: header required
			if merchantID == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "X-Merchant-Id header is required",
				})
			}
			c.Set(ContextKeyMerchantID, merchantID)
			return next(c)
		}
	}
}

func GetMerchantID(c echo.Context) string {
	id, _ := c.Get(ContextKeyMerchantID).(string)
	return id
}
