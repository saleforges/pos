package httptransport

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	httpaudit "github.com/saleforge/pos/services/internal/iam/transport/http/audit"
	httpauth "github.com/saleforge/pos/services/internal/iam/transport/http/auth"
	"github.com/saleforge/pos/services/internal/iam/transport/http/middleware"
	httpperm "github.com/saleforge/pos/services/internal/iam/transport/http/permission"
	httprole "github.com/saleforge/pos/services/internal/iam/transport/http/role"
	httpuser "github.com/saleforge/pos/services/internal/iam/transport/http/user"
	"github.com/saleforge/pos/services/internal/iam/usecase"
	"github.com/saleforge/pos/services/pkg/otel"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

type RateLimitConfig struct {
	LoginLimit   int
	LoginWindow  time.Duration
	RefreshLimit int
}

func NewRouter(
	authHandler *httpauth.Handler,
	userHandler *httpuser.Handler,
	roleHandler *httprole.Handler,
	permHandler *httpperm.Handler,
	auditHandler *httpaudit.Handler,
	authService usecase.AuthUsecase,
	hasPermissionFunc func(*port.TokenClaims, domain.Permission) bool,
	jwksProvider port.JWKSProvider,
	allowedOrigins []string,
	rateLimit RateLimitConfig,
) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Pre(middleware.CORSMiddleware(allowedOrigins))

	e.Use(otelecho.Middleware("iam-service"))
	e.Use(otel.LoggingMiddleware())
	e.Use(echomw.BodyLimit("2MB"))

	e.GET("/metrics", func(c echo.Context) error {
		otel.MetricsHandler().ServeHTTP(c.Response(), c.Request())
		return nil
	})

	e.GET("/.well-known/jwks.json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, jwksProvider.JWKS())
	})

	api := e.Group("/api/v1")

	loginRateLimit := middleware.RateLimitMiddleware(rateLimit.LoginLimit, rateLimit.LoginWindow, middleware.LoginKey)
	refreshRateLimit := middleware.RateLimitMiddleware(rateLimit.RefreshLimit, 1*time.Minute, middleware.ClientIPKey)

	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login, loginRateLimit)
	api.POST("/auth/refresh", authHandler.Refresh, refreshRateLimit)
	api.POST("/auth/introspect", authHandler.Introspect)

	protected := middleware.AuthMiddleware(authService.ValidateToken)

	api.POST("/auth/logout", authHandler.Logout, protected)
	api.POST("/auth/switch-context", authHandler.SwitchContext, protected)
	api.PUT("/auth/me/default-role", authHandler.SetDefaultRole, protected)
	api.GET("/auth/me", authHandler.Me, protected)

	userManage := middleware.RBACMiddleware(domain.UserList, hasPermissionFunc)
	api.GET("/users", userHandler.ListUsers, protected, userManage)
	api.POST("/users", userHandler.CreateUser, protected,
		middleware.RBACMiddleware(domain.UserCreate, hasPermissionFunc))

	api.GET("/users/:id", userHandler.GetUser, protected,
		middleware.RBACMiddleware(domain.UserRead, hasPermissionFunc))
	api.PATCH("/users/:id", userHandler.UpdateUser, protected,
		middleware.RBACMiddleware(domain.UserUpdate, hasPermissionFunc))
	api.DELETE("/users/:id", userHandler.DeleteUser, protected,
		middleware.RBACMiddleware(domain.UserDelete, hasPermissionFunc))

	api.POST("/users/:id/roles", roleHandler.AssignRole, protected,
		middleware.RBACMiddleware(domain.RoleAssign, hasPermissionFunc))
	api.DELETE("/users/:id/roles/:roleId", roleHandler.RemoveRole, protected,
		middleware.RBACMiddleware(domain.RoleAssign, hasPermissionFunc))

	roleManage := middleware.RBACMiddleware(domain.RoleRead, hasPermissionFunc)
	api.GET("/roles", roleHandler.ListRoles, protected, roleManage)
	api.POST("/roles", roleHandler.CreateRole, protected,
		middleware.RBACMiddleware(domain.RoleCreate, hasPermissionFunc))
	api.GET("/roles/:id", roleHandler.GetRole, protected, roleManage)
	api.PATCH("/roles/:id", roleHandler.UpdateRole, protected,
		middleware.RBACMiddleware(domain.RoleUpdate, hasPermissionFunc))
	api.DELETE("/roles/:id", roleHandler.DeleteRole, protected,
		middleware.RBACMiddleware(domain.RoleDelete, hasPermissionFunc))

	api.POST("/roles/:id/permissions", roleHandler.AssignPermission, protected,
		middleware.RBACMiddleware(domain.PermissionAssign, hasPermissionFunc))
	api.DELETE("/roles/:id/permissions/:permissionId", roleHandler.RemovePermission, protected,
		middleware.RBACMiddleware(domain.PermissionAssign, hasPermissionFunc))

	api.GET("/audit/logins", auditHandler.ListLogins, protected,
		middleware.RBACMiddleware(domain.AuditView, hasPermissionFunc))

	permManage := middleware.RBACMiddleware(domain.PermissionRead, hasPermissionFunc)
	api.GET("/permissions", permHandler.ListPermissions, protected, permManage)
	api.POST("/permissions", permHandler.CreatePermission, protected,
		middleware.RBACMiddleware(domain.PermissionCreate, hasPermissionFunc))
	api.DELETE("/permissions/:name", permHandler.DeletePermission, protected,
		middleware.RBACMiddleware(domain.PermissionDelete, hasPermissionFunc))

	return e
}
