package permission

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/transport/http/common"
	"github.com/saleforge/pos/services/internal/iam/usecase"
)

type Handler struct {
	permService usecase.PermissionUsecase
}

func NewHandler(permService usecase.PermissionUsecase) *Handler {
	return &Handler{permService: permService}
}

func (h *Handler) ListPermissions(c echo.Context) error {
	permissions, err := h.permService.ListPermissions(c.Request().Context())
	if err != nil {
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	return common.WriteJSON(c, http.StatusOK, permissions)
}

func (h *Handler) CreatePermission(c echo.Context) error {
	var req createPermissionRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return common.WriteError(c, http.StatusBadRequest, common.ErrInvalidBody)
	}

	if err := h.permService.CreatePermission(c.Request().Context(), domain.Permission(req.Permission)); err != nil {
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	return common.WriteJSON(c, http.StatusCreated, map[string]interface{}{
		"permission": req.Permission,
	})
}

func (h *Handler) DeletePermission(c echo.Context) error {
	if err := h.permService.DeletePermission(c.Request().Context(), domain.Permission(c.Param("name"))); err != nil {
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	return common.WriteJSON(c, http.StatusOK, nil)
}
