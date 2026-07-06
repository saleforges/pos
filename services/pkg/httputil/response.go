package httputil

import "github.com/labstack/echo/v4"

func WriteJSON(c echo.Context, status int, data interface{}) error {
	return c.JSON(status, data)
}

func WriteError(c echo.Context, status int, err error) error {
	return WriteJSON(c, status, map[string]string{"error": err.Error()})
}
