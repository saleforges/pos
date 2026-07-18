package user

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/transport/http/common"
	"github.com/saleforge/pos/services/internal/iam/usecase"
)

type Handler struct {
	userService usecase.UserUsecase
}

func NewHandler(userService usecase.UserUsecase) *Handler {
	return &Handler{userService: userService}
}

func (h *Handler) ListUsers(c echo.Context) error {
	users, err := h.userService.ListUsers(c.Request().Context(), 0, 100)
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

	result, err := h.userService.Register(c.Request().Context(), usecase.RegisterParams{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Roles:    req.Roles,
	})
	if err != nil {
		if err == domain.ErrPasswordPolicy {
			return common.WriteError(c, http.StatusBadRequest, err)
		}
		if err == domain.ErrUserAlreadyExists || err == domain.ErrInvalidRole {
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
		return common.WriteError(c, http.StatusNotFound, domain.ErrUserNotFound)
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
		if err == domain.ErrUserNotFound {
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
		return common.WriteError(c, http.StatusNotFound, domain.ErrUserNotFound)
	}

	return common.WriteJSON(c, http.StatusOK, nil)
}
