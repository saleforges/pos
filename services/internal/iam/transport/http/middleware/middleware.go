package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
)

const claimsKey = "claims"

func CORSMiddleware(allowedOrigins []string) echo.MiddlewareFunc {
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = struct{}{}
	}
	allowAll := len(originSet) == 0

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			origin := c.Request().Header.Get("Origin")

			if origin == "" || allowAll {
				if origin != "" {
					c.Response().Header().Set("Access-Control-Allow-Origin", origin)
				}
				c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
				if c.Request().Method == http.MethodOptions {
					return c.NoContent(http.StatusNoContent)
				}
				return next(c)
			}

			// Validate against allow-list
			if _, allowed := originSet[origin]; !allowed {
				return c.NoContent(http.StatusForbidden)
			}

			c.Response().Header().Set("Access-Control-Allow-Origin", origin)
			c.Response().Header().Add("Vary", "Origin")
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
			if c.Request().Method == http.MethodOptions {
				return c.NoContent(http.StatusNoContent)
			}
			return next(c)
		}
	}
}

func AuthMiddleware(tokenVerifier func(context.Context, string) (*port.TokenClaims, error)) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := extractToken(c)
			if token == "" {
				return writeError(c, http.StatusUnauthorized, domain.ErrInvalidToken)
			}

			claims, err := tokenVerifier(c.Request().Context(), token)
			if err != nil {
				return writeError(c, http.StatusUnauthorized, err)
			}

			c.Set(claimsKey, claims)
			return next(c)
		}
	}
}

func extractToken(c echo.Context) string {
	// try Authorization header first
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			if parts[1] == "" {
				return ""
			}
			return parts[1]
		}
	}

	// fallback to access_token cookie
	cookie, err := c.Cookie("access_token")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func RBACMiddleware(required domain.Permission, hasPermission func(*port.TokenClaims, domain.Permission) bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := c.Get(claimsKey).(*port.TokenClaims)
			if !ok {
				return writeError(c, http.StatusForbidden, domain.ErrForbidden)
			}

			if !hasPermission(claims, required) {
				return writeError(c, http.StatusForbidden, domain.ErrForbidden)
			}

			return next(c)
		}
	}
}

func BranchAccessMiddleware(branchParam string, hasAccess func(*port.TokenClaims, string) bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := c.Get(claimsKey).(*port.TokenClaims)
			if !ok {
				return writeError(c, http.StatusForbidden, domain.ErrForbidden)
			}

			branchID := c.Param(branchParam)
			if branchID == "" {
				return writeError(c, http.StatusBadRequest, errMissingBranchID)
			}

			if !hasAccess(claims, branchID) {
				return writeError(c, http.StatusForbidden, domain.ErrForbidden)
			}

			return next(c)
		}
	}
}

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
}

type visitor struct {
	count    int
	lastSeen time.Time
}

func newRateLimiter() *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) cleanup() {
	for {
		time.Sleep(1 * time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 1*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) Allow(ip string, limit int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
		return true
	}

	if time.Since(v.lastSeen) > window {
		v.count = 1
		v.lastSeen = time.Now()
		return true
	}

	v.count++
	v.lastSeen = time.Now()

	return v.count <= limit
}

func RateLimitMiddleware(limit int, window time.Duration) echo.MiddlewareFunc {
	rl := newRateLimiter()
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			if !rl.Allow(ip, limit, window) {
				return writeError(c, http.StatusTooManyRequests, rateLimitExceeded)
			}
			return next(c)
		}
	}
}

func writeJSON(c echo.Context, status int, data interface{}) error {
	return c.JSON(status, data)
}

func writeError(c echo.Context, status int, err error) error {
	return writeJSON(c, status, map[string]string{"error": err.Error()})
}
