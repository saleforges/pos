package handler

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/iam/internal/domain"
	"github.com/saleforge/pos/services/iam/internal/usecase"
)

func toUserResponse(u domain.User) userResponse {
	return userResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Roles:     u.Roles,
		Status:    string(u.Status),
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
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

	return writeJSON(c, http.StatusCreated, toUserResponse(result.User))
}

func (h *AuthHandler) GetUser(c echo.Context) error {
	user, err := h.authUsecase.GetUser(c.Request().Context(), c.Param("id"))
	if err != nil {
		return writeError(c, http.StatusNotFound, domain.ErrUserNotFound)
	}

	return writeJSON(c, http.StatusOK, toUserResponse(*user))
}

func (h *AuthHandler) UpdateUser(c echo.Context) error {
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
		ID:       c.Param("id"),
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
	if err := h.authUsecase.DeleteUser(c.Request().Context(), c.Param("id")); err != nil {
		return writeError(c, http.StatusNotFound, domain.ErrUserNotFound)
	}

	return writeJSON(c, http.StatusOK, map[string]string{"message": "user deleted"})
}
