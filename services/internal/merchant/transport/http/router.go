package httptransport

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/merchant/port/repository"
	"github.com/saleforge/pos/services/internal/merchant/transport/http/handler"
	"github.com/saleforge/pos/services/internal/merchant/transport/http/middleware"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/jwks"
	"github.com/saleforge/pos/services/pkg/otel"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

// staffAssignmentAdapter wraps repository.StaffRepository to implement middleware.StaffAssignmentProvider.
type staffAssignmentAdapter struct {
	repo repository.StaffRepository
}

func (a *staffAssignmentAdapter) ListByUserAndMerchant(ctx context.Context, userID, merchantID int64) ([]middleware.StaffAssignment, error) {
	staffList, err := a.repo.ListByUserAndMerchant(ctx, userID, merchantID)
	if err != nil {
		return nil, err
	}
	result := make([]middleware.StaffAssignment, len(staffList))
	for i, s := range staffList {
		result[i] = middleware.StaffAssignment{
			ID:        s.ID,
			BranchID:  s.BranchID,
			Role:      string(s.Role),
			IsDefault: s.IsDefault,
		}
	}
	return result, nil
}

func NewRouter(merchantHandler *handler.MerchantHandler, branchHandler *handler.BranchHandler, staffHandler *handler.StaffHandler, verifier *jwks.Verifier, staffRepo repository.StaffRepository) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(otelecho.Middleware("merchant-service"))
	e.Use(otel.LoggingMiddleware())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	e.GET("/metrics", func(c echo.Context) error {
		otel.MetricsHandler().ServeHTTP(c.Response(), c.Request())
		return nil
	})

	auth := e.Group("")
	auth.Use(middleware.Auth(verifier))

	// Merchant
	auth.POST("/api/v1/merchants", merchantHandler.Create)
	auth.GET("/api/v1/merchants", merchantHandler.List)
	auth.GET("/api/v1/merchants/:id", merchantHandler.Get)
	auth.PATCH("/api/v1/merchants/:id", merchantHandler.Update)
	auth.DELETE("/api/v1/merchants/:id", merchantHandler.Delete)

	// Branch — merchant context from header
	branchGroup := auth.Group("/api/v1/branches", httputil.MerchantMiddleware())
	branchGroup.POST("", branchHandler.CreateBranch)
	branchGroup.GET("", branchHandler.ListBranches)
	branchGroup.GET("/:id", branchHandler.GetBranch)
	branchGroup.PATCH("/:id", branchHandler.UpdateBranch)
	branchGroup.DELETE("/:id", branchHandler.DeleteBranch)

	// Staff — merchant context from header
	staffGroup := auth.Group("/api/v1/staff", httputil.MerchantMiddleware())
	staffGroup.POST("", staffHandler.AssignStaff)
	staffGroup.GET("", staffHandler.ListStaffByMerchant)
	staffGroup.GET("/:id", staffHandler.GetStaff)
	staffGroup.PATCH("/:id", staffHandler.UpdateStaff)
	staffGroup.DELETE("/:id", staffHandler.RemoveStaff)

	auth.GET("/api/v1/branches/:branchId/staff", staffHandler.ListStaffByBranch)

	// Staff assignments — merchant context from header
	assignGroup := auth.Group("/api/v1/me", httputil.MerchantMiddleware())
	assignGroup.Use(middleware.BranchContext(&staffAssignmentAdapter{repo: staffRepo}))
	assignGroup.GET("/assignments", staffHandler.MyStaffAssignments)
	assignGroup.PUT("/default-branch", staffHandler.SetMyDefaultBranch)

	return e
}


