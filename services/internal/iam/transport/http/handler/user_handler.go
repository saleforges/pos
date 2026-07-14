package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/usecase"
)

func toUserResponse(u domain.User) userResponse {
	r := userResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Type:      string(u.Type),
		Status:    string(u.Status),
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if u.SystemRole != nil {
		r.Role = u.SystemRole.Name
	}
	return r
}

func toMeResponse(u domain.User) meResponse {
	r := meResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Type:      string(u.Type),
		Status:    string(u.Status),
		Roles:     make([]roleResponse, 0),
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if u.SystemRole != nil {
		r.Roles = append(r.Roles, roleResponse{
			ID:   u.SystemRole.ID,
			Name: u.SystemRole.Name,
		})
	}

	for _, ra := range u.Roles {
		role := roleResponse{
			ID:          ra.Role.ID,
			Name:        ra.Role.Name,
			BranchScope: string(ra.BranchScope),
			IsDefault:   ra.IsDefault,
		}
		if ra.MerchantID != 0 {
			role.Merchant = &merchantDTO{ID: ra.MerchantID, Name: ra.MerchantName}
		}
		if ra.BranchID != nil {
			role.Branch = &branchDTO{
				ID:   *ra.BranchID,
				Name: ra.BranchName,
			}
		}
		r.Roles = append(r.Roles, role)
	}
	return r
}

func (h *AuthHandler) ListUsers(c echo.Context) error {
	users, err := h.authUsecase.ListUsers(c.Request().Context(), 0, 100)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, err)
	}

	result := make([]userResponse, 0, len(users))
	for _, u := range users {
		result = append(result, toUserResponse(u))
	}

	return writeJSON(c, http.StatusOK, result)
}

func (h *AuthHandler) CreateUser(c echo.Context) error {
	var req createUserRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return writeError(c, http.StatusBadRequest, errInvalidBody)
	}

	result, err := h.authUsecase.Register(c.Request().Context(), usecase.RegisterInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Roles:    req.Roles,
	})
	if err != nil {
		if err == domain.ErrPasswordPolicy {
			return writeError(c, http.StatusBadRequest, err)
		}
		if err == domain.ErrUserAlreadyExists || err == domain.ErrInvalidRole {
			return writeError(c, http.StatusConflict, err)
		}
		return writeError(c, http.StatusInternalServerError, err)
	}

	return writeJSON(c, http.StatusCreated, authResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	})
}

func (h *AuthHandler) GetUser(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return writeError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	user, err := h.authUsecase.GetUser(c.Request().Context(), id)
	if err != nil {
		return writeError(c, http.StatusNotFound, domain.ErrUserNotFound)
	}

	return writeJSON(c, http.StatusOK, toUserResponse(*user))
}

func (h *AuthHandler) UpdateUser(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return writeError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	var req updateUserRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return writeError(c, http.StatusBadRequest, errInvalidBody)
	}

	var status *domain.UserStatus
	if req.Status != nil {
		s := domain.UserStatus(*req.Status)
		status = &s
	}

	user, err := h.authUsecase.UpdateUser(c.Request().Context(), usecase.UpdateUserInput{
		ID:       id,
		Username: req.Username,
		Email:    req.Email,
		Status:   status,
	})
	if err != nil {
		if err == domain.ErrUserNotFound {
			return writeError(c, http.StatusNotFound, err)
		}
		return writeError(c, http.StatusInternalServerError, err)
	}

	return writeJSON(c, http.StatusOK, toUserResponse(*user))
}

func (h *AuthHandler) DeleteUser(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return writeError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	if err := h.authUsecase.DeleteUser(c.Request().Context(), id); err != nil {
		return writeError(c, http.StatusNotFound, domain.ErrUserNotFound)
	}

	return writeJSON(c, http.StatusOK, nil)
}
