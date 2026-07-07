package handler

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/domain"
)

func (h *AuthHandler) ListPermissions(c echo.Context) error {
	permissions, err := h.authUsecase.ListPermissions(c.Request().Context())
	if err != nil {
		return writeError(c, http.StatusInternalServerError, err)
	}

	return writeJSON(c, http.StatusOK, permissions)
}

func (h *AuthHandler) CreatePermission(c echo.Context) error {
	var req createPermissionRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return writeError(c, http.StatusBadRequest, errInvalidBody)
	}

	if err := h.authUsecase.CreatePermission(c.Request().Context(), domain.Permission(req.Permission)); err != nil {
		return writeError(c, http.StatusInternalServerError, err)
	}

	return writeJSON(c, http.StatusCreated, map[string]interface{}{
		"permission": req.Permission,
	})
}

func (h *AuthHandler) DeletePermission(c echo.Context) error {
	if err := h.authUsecase.DeletePermission(c.Request().Context(), domain.Permission(c.Param("name"))); err != nil {
		return writeError(c, http.StatusInternalServerError, err)
	}

	return writeJSON(c, http.StatusOK, map[string]string{"message": "permission deleted"})
}
