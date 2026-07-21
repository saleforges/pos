package role

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	"github.com/saleforge/pos/services/internal/iam/transport/http/common"
	"github.com/saleforge/pos/services/internal/iam/usecase"
)

type Handler struct {
	roleService usecase.RoleUsecase
}

func NewHandler(roleService usecase.RoleUsecase) *Handler {
	return &Handler{roleService: roleService}
}

func parseID(c echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func (h *Handler) ListRoles(c echo.Context) error {
	var mid *int64
	if claims, ok := c.Get(common.ClaimsKey).(*port.TokenClaims); ok && claims.MerchantID > 0 {
		mid = &claims.MerchantID
	}
	roles, err := h.roleService.ListRoles(c.Request().Context(), mid)
	if err != nil {
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	return common.WriteJSON(c, http.StatusOK, roles)
}

func (h *Handler) CreateRole(c echo.Context) error {
	var req createRoleRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return common.WriteError(c, http.StatusBadRequest, common.ErrInvalidBody)
	}

	perms := make([]domain.Permission, len(req.Permissions))
	for i, p := range req.Permissions {
		perms[i] = domain.Permission(p)
	}

	role, err := h.roleService.CreateRole(c.Request().Context(), usecase.CreateRoleParams{
		Name:        req.Name,
		Description: req.Description,
		Permissions: perms,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRole) {
			return common.WriteError(c, http.StatusConflict, err)
		}
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	return common.WriteJSON(c, http.StatusCreated, role)
}

func (h *Handler) GetRole(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return common.WriteError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	role, err := h.roleService.GetRole(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRole) {
			return common.WriteError(c, http.StatusNotFound, domain.ErrInvalidRole)
		}
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	return common.WriteJSON(c, http.StatusOK, role)
}

func (h *Handler) UpdateRole(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return common.WriteError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	var req updateRoleRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return common.WriteError(c, http.StatusBadRequest, common.ErrInvalidBody)
	}

	role, err := h.roleService.UpdateRole(c.Request().Context(), usecase.UpdateRoleParams{
		ID:          id,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRole) {
			return common.WriteError(c, http.StatusNotFound, err)
		}
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	return common.WriteJSON(c, http.StatusOK, role)
}

func (h *Handler) DeleteRole(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return common.WriteError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	if err := h.roleService.DeleteRole(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrInvalidRole) {
			return common.WriteError(c, http.StatusNotFound, domain.ErrInvalidRole)
		}
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	return common.WriteJSON(c, http.StatusOK, nil)
}

func (h *Handler) AssignRole(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return common.WriteError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	var req assignRoleRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return common.WriteError(c, http.StatusBadRequest, common.ErrInvalidBody)
	}

	if err := h.roleService.AssignRole(c.Request().Context(), id, req.Role); err != nil {
		if errors.Is(err, domain.ErrInvalidRole) || errors.Is(err, domain.ErrUserNotFound) {
			return common.WriteError(c, http.StatusNotFound, err)
		}
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	return common.WriteJSON(c, http.StatusOK, nil)
}

func (h *Handler) RemoveRole(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return common.WriteError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	if err := h.roleService.RemoveRole(c.Request().Context(), id, c.Param("roleId")); err != nil {
		return common.WriteError(c, http.StatusNotFound, err)
	}

	return common.WriteJSON(c, http.StatusOK, nil)
}

func (h *Handler) AssignPermission(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return common.WriteError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	var req assignPermissionRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return common.WriteError(c, http.StatusBadRequest, common.ErrInvalidBody)
	}

	if err := h.roleService.AssignPermission(c.Request().Context(), id, domain.Permission(req.Permission)); err != nil {
		if errors.Is(err, domain.ErrInvalidRole) {
			return common.WriteError(c, http.StatusNotFound, err)
		}
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	return common.WriteJSON(c, http.StatusOK, nil)
}

func (h *Handler) RemovePermission(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return common.WriteError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	if err := h.roleService.RemovePermission(c.Request().Context(), id, domain.Permission(c.Param("permissionId"))); err != nil {
		if errors.Is(err, domain.ErrInvalidRole) {
			return common.WriteError(c, http.StatusNotFound, err)
		}
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	return common.WriteJSON(c, http.StatusOK, nil)
}
