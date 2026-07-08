package httputil

import "github.com/labstack/echo/v4"

type APIResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(c echo.Context, status int, data interface{}) error {
	return c.JSON(status, APIResponse{
		Message: "success",
		Data:    data,
	})
}

func WriteJSON(c echo.Context, status int, data interface{}) error {
	return Success(c, status, data)
}

func WriteError(c echo.Context, status int, err error) error {
	return c.JSON(status, APIResponse{
		Message: err.Error(),
	})
}
