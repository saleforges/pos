package httputil

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/pkg/pagination"
)

type APIResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ErrorResponse struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
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

// WriteError writes a JSON error response.
// If the error message contains "CODE: " prefix it is split into code+message.
// Example: "MCH001: merchant not found" → {"code":"MCH001","message":"merchant not found"}
func WriteError(c echo.Context, status int, err error) error {
	msg := err.Error()
	code := extractCode(&msg)
	return c.JSON(status, ErrorResponse{Code: code, Message: msg})
}

// WritePaginated writes a paginated list response.
func WritePaginated(c echo.Context, status int, data interface{}, meta pagination.Metadata) error {
	return c.JSON(status, map[string]interface{}{
		"message":    "success",
		"data":       data,
		"pagination": meta,
	})
}

// ParsePageParams extracts pagination params from request query.
func ParsePageParams(c echo.Context) pagination.Params {
	return pagination.Parse(c.Request())
}

// extractCode looks for "CODE: " prefix and extracts it, modifying msg in place.
func extractCode(msg *string) string {
	idx := strings.Index(*msg, ": ")
	if idx > 0 && len((*msg)[:idx]) > 0 && !strings.Contains((*msg)[:idx], " ") {
		code := (*msg)[:idx]
		*msg = (*msg)[idx+2:]
		return code
	}
	return ""
}
