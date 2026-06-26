package httptransport

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/iam/internal/domain"
	"github.com/saleforge/pos/services/iam/internal/port"
	"github.com/saleforge/pos/services/iam/internal/transport/http/handler"
	"github.com/saleforge/pos/services/iam/internal/transport/http/middleware"
	"github.com/saleforge/pos/services/iam/internal/usecase"
)

func NewRouter(
	authHandler *handler.AuthHandler,
	authUsecase usecase.AuthService,
	jwksProvider port.JWKSProvider,
) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Pre(middleware.CORSMiddleware())

	loginRateLimit := middleware.RateLimitMiddleware(5, 1*time.Minute)
	refreshRateLimit := middleware.RateLimitMiddleware(20, 1*time.Minute)

	e.POST("/api/v1/auth/register", authHandler.Register)
	e.POST("/api/v1/auth/login", authHandler.Login, loginRateLimit)
	e.POST("/api/v1/auth/refresh", authHandler.Refresh, refreshRateLimit)
	e.POST("/api/v1/auth/introspect", authHandler.Introspect)

	protected := middleware.AuthMiddleware(authUsecase.ValidateToken)

	e.POST("/api/v1/auth/logout", authHandler.Logout, protected)
	e.GET("/api/v1/auth/me", authHandler.Me, protected)

	userManage := middleware.RBACMiddleware(domain.UserList, authUsecase.HasPermission)
	e.GET("/api/v1/users", authHandler.ListUsers, protected, userManage)
	e.POST("/api/v1/users", authHandler.CreateUser, protected,
		middleware.RBACMiddleware(domain.UserCreate, authUsecase.HasPermission))

	e.GET("/api/v1/users/:id", authHandler.GetUser, protected,
		middleware.RBACMiddleware(domain.UserRead, authUsecase.HasPermission))
	e.PATCH("/api/v1/users/:id", authHandler.UpdateUser, protected,
		middleware.RBACMiddleware(domain.UserUpdate, authUsecase.HasPermission))
	e.DELETE("/api/v1/users/:id", authHandler.DeleteUser, protected,
		middleware.RBACMiddleware(domain.UserDelete, authUsecase.HasPermission))

	e.POST("/api/v1/users/:id/roles", authHandler.AssignRole, protected,
		middleware.RBACMiddleware(domain.RoleAssign, authUsecase.HasPermission))
	e.DELETE("/api/v1/users/:id/roles/:roleId", authHandler.RemoveRole, protected,
		middleware.RBACMiddleware(domain.RoleAssign, authUsecase.HasPermission))

	roleManage := middleware.RBACMiddleware(domain.RoleRead, authUsecase.HasPermission)
	e.GET("/api/v1/roles", authHandler.ListRoles, protected, roleManage)
	e.POST("/api/v1/roles", authHandler.CreateRole, protected,
		middleware.RBACMiddleware(domain.RoleCreate, authUsecase.HasPermission))
	e.GET("/api/v1/roles/:name", authHandler.GetRole, protected, roleManage)
	e.PATCH("/api/v1/roles/:name", authHandler.UpdateRole, protected,
		middleware.RBACMiddleware(domain.RoleUpdate, authUsecase.HasPermission))
	e.DELETE("/api/v1/roles/:name", authHandler.DeleteRole, protected,
		middleware.RBACMiddleware(domain.RoleDelete, authUsecase.HasPermission))

	e.POST("/api/v1/roles/:name/permissions", authHandler.AssignPermission, protected,
		middleware.RBACMiddleware(domain.PermissionAssign, authUsecase.HasPermission))
	e.DELETE("/api/v1/roles/:name/permissions/:permissionId", authHandler.RemovePermission, protected,
		middleware.RBACMiddleware(domain.PermissionAssign, authUsecase.HasPermission))

	permManage := middleware.RBACMiddleware(domain.PermissionRead, authUsecase.HasPermission)
	e.GET("/api/v1/permissions", authHandler.ListPermissions, protected, permManage)
	e.POST("/api/v1/permissions", authHandler.CreatePermission, protected,
		middleware.RBACMiddleware(domain.PermissionCreate, authUsecase.HasPermission))
	e.DELETE("/api/v1/permissions/:name", authHandler.DeletePermission, protected,
		middleware.RBACMiddleware(domain.PermissionDelete, authUsecase.HasPermission))

	e.GET("/.well-known/jwks.json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, jwksProvider.JWKS())
	})

	return e
}
