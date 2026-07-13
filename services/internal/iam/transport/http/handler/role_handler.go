package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/usecase"
)

func (h *AuthHandler) ListRoles(c echo.Context) error {
	var mid *int64
	if s := c.Request().Header.Get("X-Merchant-Id"); s != "" {
		if id, err := strconv.ParseInt(s, 10, 64); err == nil {
			mid = &id
		}
	}
	roles, err := h.authUsecase.ListRoles(c.Request().Context(), mid)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, err)
	}

	return writeJSON(c, http.StatusOK, roles)
}

func (h *AuthHandler) CreateRole(c echo.Context) error {
	var req createRoleRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return writeError(c, http.StatusBadRequest, errInvalidBody)
	}

	perms := make([]domain.Permission, len(req.Permissions))
	for i, p := range req.Permissions {
		perms[i] = domain.Permission(p)
	}

	role, err := h.authUsecase.CreateRole(c.Request().Context(), usecase.CreateRoleInput{
		Name:        req.Name,
		Description: req.Description,
		Permissions: perms,
	})
	if err != nil {
		if err == domain.ErrInvalidRole {
			return writeError(c, http.StatusConflict, err)
		}
		return writeError(c, http.StatusInternalServerError, err)
	}

	return writeJSON(c, http.StatusCreated, role)
}

func parseID(c echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func (h *AuthHandler) GetRole(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	role, err := h.authUsecase.GetRole(c.Request().Context(), id)
	if err != nil {
		return writeError(c, http.StatusNotFound, domain.ErrInvalidRole)
	}

	return writeJSON(c, http.StatusOK, role)
}

func (h *AuthHandler) UpdateRole(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	var req updateRoleRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return writeError(c, http.StatusBadRequest, errInvalidBody)
	}

	role, err := h.authUsecase.UpdateRole(c.Request().Context(), usecase.UpdateRoleInput{
		ID:          id,
		Description: req.Description,
	})
	if err != nil {
		if err == domain.ErrInvalidRole {
			return writeError(c, http.StatusNotFound, err)
		}
		return writeError(c, http.StatusInternalServerError, err)
	}

	return writeJSON(c, http.StatusOK, role)
}

func (h *AuthHandler) DeleteRole(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	if err := h.authUsecase.DeleteRole(c.Request().Context(), id); err != nil {
		return writeError(c, http.StatusConflict, domain.ErrInvalidRole)
	}

	return writeJSON(c, http.StatusOK, nil)
}

func (h *AuthHandler) AssignRole(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return writeError(c, http.StatusBadRequest, errInvalidBody)
	}

	if err := h.authUsecase.AssignRole(c.Request().Context(), id, req.Role); err != nil {
		if err == domain.ErrInvalidRole || err == domain.ErrUserNotFound {
			return writeError(c, http.StatusNotFound, err)
		}
		return writeError(c, http.StatusInternalServerError, err)
	}

	return writeJSON(c, http.StatusOK, nil)
}

func (h *AuthHandler) RemoveRole(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	if err := h.authUsecase.RemoveRole(c.Request().Context(), id, c.Param("roleId")); err != nil {
		return writeError(c, http.StatusNotFound, err)
	}

	return writeJSON(c, http.StatusOK, nil)
}

func (h *AuthHandler) AssignPermission(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	var req struct {
		Permission string `json:"permission"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return writeError(c, http.StatusBadRequest, errInvalidBody)
	}

	if err := h.authUsecase.AssignPermission(c.Request().Context(), id, domain.Permission(req.Permission)); err != nil {
		if err == domain.ErrInvalidRole {
			return writeError(c, http.StatusNotFound, err)
		}
		return writeError(c, http.StatusInternalServerError, err)
	}

	return writeJSON(c, http.StatusOK, nil)
}

func (h *AuthHandler) RemovePermission(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	if err := h.authUsecase.RemovePermission(c.Request().Context(), id, domain.Permission(c.Param("permissionId"))); err != nil {
		if err == domain.ErrInvalidRole {
			return writeError(c, http.StatusNotFound, err)
		}
		return writeError(c, http.StatusInternalServerError, err)
	}

	return writeJSON(c, http.StatusOK, nil)
}
