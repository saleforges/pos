package handler

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	"github.com/saleforge/pos/services/internal/iam/usecase"
	"github.com/saleforge/pos/services/pkg/logger"
)

func (h *AuthHandler) Register(c echo.Context) error {
	var req registerRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return writeError(c, http.StatusBadRequest, errInvalidBody)
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		return writeError(c, http.StatusBadRequest, errMissingFields)
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
		logger.Error("register failed", "error", err.Error())
		return writeError(c, http.StatusInternalServerError, err)
	}

	return writeJSON(c, http.StatusCreated, authResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req loginRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return writeError(c, http.StatusBadRequest, errInvalidBody)
	}

	if req.Username == "" || req.Password == "" {
		return writeError(c, http.StatusBadRequest, errMissingFields)
	}

	result, err := h.authUsecase.Login(c.Request().Context(), usecase.LoginInput{
		Username:  req.Username,
		Password:  req.Password,
		IPAddress: c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	})
	if err != nil {
		if err == domain.ErrInvalidCredentials || err == domain.ErrUserDisabled {
			return writeError(c, http.StatusUnauthorized, err)
		}
		return writeError(c, http.StatusInternalServerError, err)
	}

	return writeJSON(c, http.StatusOK, authResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	})
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	var req refreshRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return writeError(c, http.StatusBadRequest, errInvalidBody)
	}

	if req.RefreshToken == "" {
		return writeError(c, http.StatusBadRequest, errMissingFields)
	}

	result, err := h.authUsecase.RefreshToken(c.Request().Context(), usecase.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
		IPAddress:    c.RealIP(),
		UserAgent:    c.Request().UserAgent(),
	})
	if err != nil {
		if err == domain.ErrInvalidRefreshToken {
			return writeError(c, http.StatusUnauthorized, err)
		}
		if err == domain.ErrUserDisabled {
			return writeError(c, http.StatusForbidden, err)
		}
		return writeError(c, http.StatusInternalServerError, err)
	}

	return writeJSON(c, http.StatusOK, authResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	var req logoutRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return writeError(c, http.StatusBadRequest, errInvalidBody)
	}

	claims := c.Get(claimsKey).(*port.TokenClaims)

	err := h.authUsecase.Logout(c.Request().Context(), usecase.LogoutInput{
		RefreshToken: req.RefreshToken,
		UserID:       claims.UserID,
	})
	if err != nil {
		return writeError(c, http.StatusInternalServerError, err)
	}

	return writeJSON(c, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func (h *AuthHandler) Introspect(c echo.Context) error {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return writeError(c, http.StatusBadRequest, errInvalidBody)
	}

	result, err := h.authUsecase.Introspect(c.Request().Context(), req.Token)
	if err != nil {
		return writeJSON(c, http.StatusOK, map[string]bool{"active": false})
	}

	return writeJSON(c, http.StatusOK, result)
}

func (h *AuthHandler) Me(c echo.Context) error {
	claims, ok := c.Get(claimsKey).(*port.TokenClaims)
	if !ok {
		return writeError(c, http.StatusUnauthorized, errUnauthorized)
	}

	user, err := h.authUsecase.GetUser(c.Request().Context(), claims.UserID)
	if err != nil {
		return writeError(c, http.StatusUnauthorized, errUnauthorized)
	}

	return writeJSON(c, http.StatusOK, toUserResponse(*user))
}
