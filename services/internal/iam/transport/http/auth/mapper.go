package auth

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/transport/http/common"
)

const (
	accessCookieName  = "access_token"
	refreshCookieName = "refresh_token"
)

func setTokenCookies(c echo.Context, accessToken, refreshToken string, expiresIn int) {
	// Always set Secure=true — in production TLS is terminated at proxy level;
	// checking c.Request().TLS is unreliable when behind a reverse proxy.
	secure := true
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

func toMeResponse(u domain.User) common.MeResponse {
	r := common.MeResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Type:      string(u.Type),
		Status:    string(u.Status),
		Roles:     make([]common.RoleResponse, 0),
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if u.SystemRole != nil {
		r.Roles = append(r.Roles, common.RoleResponse{
			ID:   u.SystemRole.ID,
			Name: u.SystemRole.Name,
		})
	}

	for _, ra := range u.Roles {
		role := common.RoleResponse{
			ID:          ra.Role.ID,
			Name:        ra.Role.Name,
			BranchScope: string(ra.BranchScope),
			IsDefault:   ra.IsDefault,
		}
		if ra.MerchantID != 0 {
			role.Merchant = &common.MerchantDTO{ID: ra.MerchantID, Name: ra.MerchantName}
		}
		if ra.BranchID != nil {
			role.Branch = &common.BranchDTO{
				ID:   *ra.BranchID,
				Name: ra.BranchName,
			}
		}
		r.Roles = append(r.Roles, role)
	}
	return r
}
