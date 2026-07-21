package auth

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	"github.com/saleforge/pos/services/internal/iam/transport/http/common"
	"github.com/saleforge/pos/services/internal/iam/usecase"
	"github.com/saleforge/pos/services/pkg/logger"
)

type Handler struct {
	authService   usecase.AuthUsecase
	userService   usecase.UserUsecase
	secureCookies bool
}

func NewHandler(authService usecase.AuthUsecase, userService usecase.UserUsecase, secureCookies bool) *Handler {
	return &Handler{authService: authService, userService: userService, secureCookies: secureCookies}
}

func (h *Handler) Register(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return common.WriteError(c, http.StatusBadRequest, common.ErrInvalidBody)
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		return common.WriteError(c, http.StatusBadRequest, common.ErrMissingFields)
	}

	result, err := h.authService.Register(c.Request().Context(), usecase.RegisterParams{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		IPAddress: c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	})
	if err != nil {
		if errors.Is(err, domain.ErrPasswordPolicy) {
			return common.WriteError(c, http.StatusBadRequest, err)
		}
		if errors.Is(err, domain.ErrUserAlreadyExists) || errors.Is(err, domain.ErrInvalidRole) {
			return common.WriteError(c, http.StatusConflict, err)
		}
		logger.Error("register failed", "error", err.Error())
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	setTokenCookies(c, result.AccessToken, result.RefreshToken, result.ExpiresIn, h.secureCookies)

	return common.WriteJSON(c, http.StatusCreated, authResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	})
}

func (h *Handler) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return common.WriteError(c, http.StatusBadRequest, common.ErrInvalidBody)
	}

	if req.Username == "" || req.Password == "" {
		return common.WriteError(c, http.StatusBadRequest, common.ErrMissingFields)
	}

	result, err := h.authService.Login(c.Request().Context(), usecase.LoginParams{
		Username:  req.Username,
		Password:  req.Password,
		IPAddress: c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) || errors.Is(err, domain.ErrUserDisabled) {
			return common.WriteError(c, http.StatusUnauthorized, err)
		}
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	setTokenCookies(c, result.AccessToken, result.RefreshToken, result.ExpiresIn, h.secureCookies)

	return common.WriteJSON(c, http.StatusOK, authResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	})
}

func (h *Handler) Refresh(c echo.Context) error {
	var req refreshRequest
	if err := c.Bind(&req); err != nil {
		return common.WriteError(c, http.StatusBadRequest, common.ErrInvalidBody)
	}

	refreshToken := req.RefreshToken
	if refreshToken == "" {
		cookie, err := c.Cookie(refreshCookieName)
		if err == nil && cookie.Value != "" {
			refreshToken = cookie.Value
		}
	}

	if refreshToken == "" {
		return common.WriteError(c, http.StatusBadRequest, common.ErrMissingFields)
	}

	result, err := h.authService.RefreshToken(c.Request().Context(), usecase.RefreshTokenParams{
		RefreshToken: refreshToken,
		IPAddress:    c.RealIP(),
		UserAgent:    c.Request().UserAgent(),
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRefreshToken) {
			clearTokenCookies(c, h.secureCookies)
			return common.WriteError(c, http.StatusUnauthorized, err)
		}
		if errors.Is(err, domain.ErrUserDisabled) {
			return common.WriteError(c, http.StatusForbidden, err)
		}
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	setTokenCookies(c, result.AccessToken, result.RefreshToken, result.ExpiresIn, h.secureCookies)

	return common.WriteJSON(c, http.StatusOK, authResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	})
}

func (h *Handler) Logout(c echo.Context) error {
	claims := c.Get(common.ClaimsKey).(*port.TokenClaims)

	err := h.authService.Logout(c.Request().Context(), usecase.LogoutParams{
		SessionID: claims.SessionID,
	})
	if err != nil {
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	clearTokenCookies(c, h.secureCookies)
	return common.WriteJSON(c, http.StatusOK, nil)
}

func (h *Handler) SwitchContext(c echo.Context) error {
	var req switchContextRequest
	if err := c.Bind(&req); err != nil {
		return common.WriteError(c, http.StatusBadRequest, common.ErrInvalidBody)
	}

	claims := c.Get(common.ClaimsKey).(*port.TokenClaims)

	result, err := h.authService.SwitchContext(c.Request().Context(), claims.SessionID, req.UserRoleID)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) || errors.Is(err, domain.ErrForbidden) {
			return common.WriteError(c, http.StatusForbidden, err)
		}
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	setTokenCookies(c, result.AccessToken, "", result.ExpiresIn, h.secureCookies)

	return common.WriteJSON(c, http.StatusOK, authResponse{
		AccessToken: result.AccessToken,
		ExpiresIn:   result.ExpiresIn,
	})
}

func (h *Handler) SetDefaultRole(c echo.Context) error {
	var req setDefaultRoleRequest
	if err := c.Bind(&req); err != nil {
		return common.WriteError(c, http.StatusBadRequest, common.ErrInvalidBody)
	}

	claims := c.Get(common.ClaimsKey).(*port.TokenClaims)

	if err := h.authService.SetDefaultRole(c.Request().Context(), claims.UserID, req.RoleID); err != nil {
		return common.WriteError(c, http.StatusInternalServerError, err)
	}

	return common.WriteJSON(c, http.StatusOK, nil)
}

func (h *Handler) Introspect(c echo.Context) error {
	var req introspectRequest
	if err := c.Bind(&req); err != nil {
		return common.WriteError(c, http.StatusBadRequest, common.ErrInvalidBody)
	}

	result, err := h.authService.Introspect(c.Request().Context(), req.Token)
	if err != nil {
		return common.WriteJSON(c, http.StatusOK, map[string]bool{"active": false})
	}

	return common.WriteJSON(c, http.StatusOK, result)
}

func (h *Handler) Me(c echo.Context) error {
	claims, ok := c.Get(common.ClaimsKey).(*port.TokenClaims)
	if !ok {
		return common.WriteError(c, http.StatusUnauthorized, common.ErrUnauthorized)
	}

	user, err := h.userService.GetUser(c.Request().Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return common.WriteError(c, http.StatusNotFound, err)
		}
		logger.Error("me: get user failed", "error", err.Error())
		return common.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}

	return common.WriteJSON(c, http.StatusOK, toMeResponse(*user))
}
