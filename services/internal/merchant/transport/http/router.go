package httptransport

import (
	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/merchant/port"
	"github.com/saleforge/pos/services/internal/merchant/port/repository"
	"github.com/saleforge/pos/services/internal/merchant/transport/http/handler"
	"github.com/saleforge/pos/services/internal/merchant/transport/http/middleware"
	"github.com/saleforge/pos/services/pkg/otel"
)

func NewRouter(merchantHandler *handler.MerchantHandler, branchHandler *handler.BranchHandler, staffHandler *handler.StaffHandler, tokenValidator port.TokenValidator, staffRepo repository.StaffRepository) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Use(otel.MetricsMiddleware())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	e.GET("/metrics", func(c echo.Context) error {
		otel.MetricsHandler().ServeHTTP(c.Response(), c.Request())
		return nil
	})

	auth := e.Group("")
	auth.Use(otel.TracingMiddleware(), otel.LoggingMiddleware(), middleware.Auth(tokenValidator))

	// Merchant
	auth.POST("/api/v1/merchants", merchantHandler.Create)
	auth.GET("/api/v1/merchants", merchantHandler.List)
	auth.GET("/api/v1/merchants/:id", merchantHandler.Get)
	auth.PATCH("/api/v1/merchants/:id", merchantHandler.Update)
	auth.DELETE("/api/v1/merchants/:id", merchantHandler.Delete)

	// Branch (nested under merchant)
	auth.POST("/api/v1/merchants/:merchantId/branches", branchHandler.CreateBranch)
	auth.GET("/api/v1/merchants/:merchantId/branches", branchHandler.ListBranches)
	auth.GET("/api/v1/branches/:id", branchHandler.GetBranch)
	auth.PATCH("/api/v1/branches/:id", branchHandler.UpdateBranch)
	auth.DELETE("/api/v1/branches/:id", branchHandler.DeleteBranch)

	// Staff management
	auth.POST("/api/v1/staff", staffHandler.AssignStaff)
	auth.GET("/api/v1/staff/:id", staffHandler.GetStaff)
	auth.PATCH("/api/v1/staff/:id", staffHandler.UpdateStaff)
	auth.DELETE("/api/v1/staff/:id", staffHandler.RemoveStaff)
	auth.GET("/api/v1/merchants/:merchantId/staff", staffHandler.ListStaffByMerchant)
	auth.GET("/api/v1/branches/:branchId/staff", staffHandler.ListStaffByBranch)

	// Staff branch context (enriched with branch-scoped RBAC)
	branchCtx := auth.Group("")
	branchCtx.Use(middleware.BranchContext(staffRepo))

	branchCtx.GET("/api/v1/me/merchants/:merchantId/assignments", staffHandler.MyStaffAssignments)
	branchCtx.PUT("/api/v1/me/default-branch", staffHandler.SetMyDefaultBranch)

	return e
}


