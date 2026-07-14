package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	"github.com/saleforge/pos/services/internal/iam/usecase"
	"github.com/saleforge/pos/services/pkg/logger"
)

const (
	accessCookieName  = "access_token"
	refreshCookieName = "refresh_token"
)

func setTokenCookies(c echo.Context, accessToken, refreshToken string, expiresIn int) {
	secure := c.Request().TLS != nil
	accessMaxAge := expiresIn
	refreshMaxAge := 30 * 24 * 3600 // 30 days

	c.SetCookie(&http.Cookie{
		Name:     accessCookieName,
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   accessMaxAge,
	})
	c.SetCookie(&http.Cookie{
		Name:     refreshCookieName,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   refreshMaxAge,
	})
}

func clearTokenCookies(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     accessCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
	c.SetCookie(&http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req registerRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return writeError(c, http.StatusBadRequest, errInvalidBody)
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		return writeError(c, http.StatusBadRequest, errMissingFields)
	}

	result, err := h.authUsecase.Register(c.Request().Context(), usecase.RegisterInput{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		IPAddress: c.RealIP(),
		UserAgent: c.Request().UserAgent(),
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

	setTokenCookies(c, result.AccessToken, result.RefreshToken, result.ExpiresIn)

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

	setTokenCookies(c, result.AccessToken, result.RefreshToken, result.ExpiresIn)

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

	refreshToken := req.RefreshToken
	if refreshToken == "" {
		cookie, err := c.Cookie(refreshCookieName)
		if err == nil && cookie.Value != "" {
			refreshToken = cookie.Value
		}
	}

	if refreshToken == "" {
		return writeError(c, http.StatusBadRequest, errMissingFields)
	}

	result, err := h.authUsecase.RefreshToken(c.Request().Context(), usecase.RefreshTokenInput{
		RefreshToken: refreshToken,
		IPAddress:    c.RealIP(),
		UserAgent:    c.Request().UserAgent(),
	})
	if err != nil {
		if err == domain.ErrInvalidRefreshToken {
			clearTokenCookies(c)
			return writeError(c, http.StatusUnauthorized, err)
		}
		if err == domain.ErrUserDisabled {
			return writeError(c, http.StatusForbidden, err)
		}
		return writeError(c, http.StatusInternalServerError, err)
	}

	setTokenCookies(c, result.AccessToken, result.RefreshToken, result.ExpiresIn)

	return writeJSON(c, http.StatusOK, authResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	claims := c.Get(claimsKey).(*port.TokenClaims)

	err := h.authUsecase.Logout(c.Request().Context(), usecase.LogoutInput{
		SessionID: claims.SessionID,
	})
	if err != nil {
		return writeError(c, http.StatusInternalServerError, err)
	}

	clearTokenCookies(c)
	return writeJSON(c, http.StatusOK, nil)
}

func (h *AuthHandler) SwitchContext(c echo.Context) error {
	var req struct {
		UserRoleID int64 `json:"userRoleId"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return writeError(c, http.StatusBadRequest, errInvalidBody)
	}

	claims := c.Get(claimsKey).(*port.TokenClaims)

	result, err := h.authUsecase.SwitchContext(c.Request().Context(), claims.SessionID, req.UserRoleID)
	if err != nil {
		if err == domain.ErrSessionNotFound || err == domain.ErrForbidden {
			return writeError(c, http.StatusForbidden, err)
		}
		return writeError(c, http.StatusInternalServerError, err)
	}

	setTokenCookies(c, result.AccessToken, "", result.ExpiresIn)

	return writeJSON(c, http.StatusOK, authResponse{
		AccessToken: result.AccessToken,
		ExpiresIn:   result.ExpiresIn,
	})
}

func (h *AuthHandler) SetDefaultRole(c echo.Context) error {
	var req struct {
		RoleID int64 `json:"userRoleId"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return writeError(c, http.StatusBadRequest, errInvalidBody)
	}

	claims := c.Get(claimsKey).(*port.TokenClaims)

	if err := h.authUsecase.SetDefaultRole(c.Request().Context(), claims.UserID, req.RoleID); err != nil {
		return writeError(c, http.StatusInternalServerError, err)
	}

	return writeJSON(c, http.StatusOK, nil)
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

	return writeJSON(c, http.StatusOK, toMeResponse(*user))
}
