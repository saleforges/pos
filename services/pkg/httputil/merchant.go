package httputil

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

const (
	ContextKeyMerchantID = "merchant_id"
	ContextKeyUserType   = "user_type"
)

func MerchantMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			merchantIDStr := c.Request().Header.Get("X-Merchant-Id")
			userType, _ := c.Get(ContextKeyUserType).(string)

			// Platform user: header optional
			if userType == "platform" {
				if merchantIDStr != "" {
					if id, err := strconv.ParseInt(merchantIDStr, 10, 64); err == nil {
						c.Set(ContextKeyMerchantID, id)
					}
				}
				return next(c)
			}

			// Merchant user or unknown: header required
			if merchantIDStr == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "X-Merchant-Id header is required",
				})
			}
			id, err := strconv.ParseInt(merchantIDStr, 10, 64)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "invalid X-Merchant-Id header",
				})
			}
			c.Set(ContextKeyMerchantID, id)
			return next(c)
		}
	}
}

func GetMerchantID(c echo.Context) int64 {
	id, _ := c.Get(ContextKeyMerchantID).(int64)
	return id
}
