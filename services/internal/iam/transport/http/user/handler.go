package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/transport/http/common"
	"github.com/saleforge/pos/services/internal/iam/usecase"
)

type Handler struct {
	authService usecase.AuthUsecase
	userService usecase.UserUsecase
}

func NewHandler(authService usecase.AuthUsecase, userService usecase.UserUsecase) *Handler {
	return &Handler{authService: authService, userService: userService}
}

func (h *Handler) ListUsers(c echo.Context) error {
	offset, limit := 0, 100
	if o, err := strconv.Atoi(c.QueryParam("offset")); err == nil && o >= 0 {
		offset = o
	}
	if l, err := strconv.Atoi(c.QueryParam("limit")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}
	users, err := h.userService.ListUsers(c.Request().Context(), offset, limit)
	if err != nil {
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	result := make([]common.UserResponse, 0, len(users))
	for _, u := range users {
		result = append(result, toUserResponse(u))
	}

	return common.WriteJSON(c, http.StatusOK, result)
}

func (h *Handler) CreateUser(c echo.Context) error {
	var req createUserRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return common.WriteError(c, http.StatusBadRequest, common.ErrInvalidBody)
	}

	result, err := h.authService.Register(c.Request().Context(), usecase.RegisterParams{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Roles:    req.Roles,
	})
	if err != nil {
		if errors.Is(err, domain.ErrPasswordPolicy) {
			return common.WriteError(c, http.StatusBadRequest, err)
		}
		if errors.Is(err, domain.ErrUserAlreadyExists) || errors.Is(err, domain.ErrInvalidRole) {
			return common.WriteError(c, http.StatusConflict, err)
		}
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	return common.WriteJSON(c, http.StatusCreated, authResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	})
}

func (h *Handler) GetUser(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return common.WriteError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	user, err := h.userService.GetUser(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return common.WriteError(c, http.StatusNotFound, domain.ErrUserNotFound)
		}
		return common.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}

	return common.WriteJSON(c, http.StatusOK, toUserResponse(*user))
}

func (h *Handler) UpdateUser(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return common.WriteError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	var req updateUserRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return common.WriteError(c, http.StatusBadRequest, common.ErrInvalidBody)
	}

	var status *domain.UserStatus
	if req.Status != nil {
		s := domain.UserStatus(*req.Status)
		status = &s
	}

	user, err := h.userService.UpdateUser(c.Request().Context(), usecase.UpdateUserParams{
		ID:       id,
		Username: req.Username,
		Email:    req.Email,
		Status:   status,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return common.WriteError(c, http.StatusNotFound, err)
		}
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	return common.WriteJSON(c, http.StatusOK, toUserResponse(*user))
}

func (h *Handler) DeleteUser(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return common.WriteError(c, http.StatusBadRequest, domain.ErrInvalidRole)
	}
	if err := h.userService.DeleteUser(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return common.WriteError(c, http.StatusNotFound, domain.ErrUserNotFound)
		}
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	return common.WriteJSON(c, http.StatusOK, nil)
}
