package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
)

const claimsKey = "claims"

type AuthHandler struct {
	authUsecase usecase.AuthService
}

func NewAuthHandler(authUsecase usecase.AuthService) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

func writeJSON(c echo.Context, status int, data interface{}) error {
	return httputil.Success(c, status, data)
}

func writeError(c echo.Context, status int, err error) error {
	return httputil.WriteError(c, status, err)
}
