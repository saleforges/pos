package common

import (
	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/pkg/httputil"
)

const ClaimsKey = "claims"

func WriteJSON(c echo.Context, status int, data interface{}) error {
	return httputil.Success(c, status, data)
}

func WriteError(c echo.Context, status int, err error) error {
	return httputil.WriteError(c, status, err)
}
