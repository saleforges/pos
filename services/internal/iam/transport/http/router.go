package httptransport

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	"github.com/saleforge/pos/services/internal/iam/transport/http/handler"
	"github.com/saleforge/pos/services/internal/iam/transport/http/middleware"
	"github.com/saleforge/pos/services/internal/iam/usecase"
	"github.com/saleforge/pos/services/pkg/otel"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

func NewRouter(
	authHandler *handler.AuthHandler,
	authUsecase usecase.AuthService,
	jwksProvider port.JWKSProvider,
) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Pre(middleware.CORSMiddleware())

	e.Use(otelecho.Middleware("iam-service"))
	e.Use(otel.LoggingMiddleware())

	e.GET("/metrics", func(c echo.Context) error {
		otel.MetricsHandler().ServeHTTP(c.Response(), c.Request())
		return nil
	})

	e.GET("/.well-known/jwks.json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, jwksProvider.JWKS())
	})

	api := e.Group("/api/v1")

	loginRateLimit := middleware.RateLimitMiddleware(5, 1*time.Minute)
	refreshRateLimit := middleware.RateLimitMiddleware(20, 1*time.Minute)

	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login, loginRateLimit)
	api.POST("/auth/refresh", authHandler.Refresh, refreshRateLimit)
	api.POST("/auth/introspect", authHandler.Introspect)

	protected := middleware.AuthMiddleware(authUsecase.ValidateToken)

	api.POST("/auth/logout", authHandler.Logout, protected)
	api.GET("/auth/me", authHandler.Me, protected)

	userManage := middleware.RBACMiddleware(domain.UserList, authUsecase.HasPermission)
	api.GET("/users", authHandler.ListUsers, protected, userManage)
	api.POST("/users", authHandler.CreateUser, protected,
		middleware.RBACMiddleware(domain.UserCreate, authUsecase.HasPermission))

	api.GET("/users/:id", authHandler.GetUser, protected,
		middleware.RBACMiddleware(domain.UserRead, authUsecase.HasPermission))
	api.PATCH("/users/:id", authHandler.UpdateUser, protected,
		middleware.RBACMiddleware(domain.UserUpdate, authUsecase.HasPermission))
	api.DELETE("/users/:id", authHandler.DeleteUser, protected,
		middleware.RBACMiddleware(domain.UserDelete, authUsecase.HasPermission))

	api.POST("/users/:id/roles", authHandler.AssignRole, protected,
		middleware.RBACMiddleware(domain.RoleAssign, authUsecase.HasPermission))
	api.DELETE("/users/:id/roles/:roleId", authHandler.RemoveRole, protected,
		middleware.RBACMiddleware(domain.RoleAssign, authUsecase.HasPermission))

	roleManage := middleware.RBACMiddleware(domain.RoleRead, authUsecase.HasPermission)
	api.GET("/roles", authHandler.ListRoles, protected, roleManage)
	api.POST("/roles", authHandler.CreateRole, protected,
		middleware.RBACMiddleware(domain.RoleCreate, authUsecase.HasPermission))
	api.GET("/roles/:id", authHandler.GetRole, protected, roleManage)
	api.PATCH("/roles/:id", authHandler.UpdateRole, protected,
		middleware.RBACMiddleware(domain.RoleUpdate, authUsecase.HasPermission))
	api.DELETE("/roles/:id", authHandler.DeleteRole, protected,
		middleware.RBACMiddleware(domain.RoleDelete, authUsecase.HasPermission))

	api.POST("/roles/:id/permissions", authHandler.AssignPermission, protected,
		middleware.RBACMiddleware(domain.PermissionAssign, authUsecase.HasPermission))
	api.DELETE("/roles/:id/permissions/:permissionId", authHandler.RemovePermission, protected,
		middleware.RBACMiddleware(domain.PermissionAssign, authUsecase.HasPermission))

	permManage := middleware.RBACMiddleware(domain.PermissionRead, authUsecase.HasPermission)
	api.GET("/permissions", authHandler.ListPermissions, protected, permManage)
	api.POST("/permissions", authHandler.CreatePermission, protected,
		middleware.RBACMiddleware(domain.PermissionCreate, authUsecase.HasPermission))
	api.DELETE("/permissions/:name", authHandler.DeletePermission, protected,
		middleware.RBACMiddleware(domain.PermissionDelete, authUsecase.HasPermission))

	return e
}
