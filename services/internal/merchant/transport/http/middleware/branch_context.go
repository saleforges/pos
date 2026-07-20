package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/pkg/httputil"
)

// StaffAssignmentProvider is a minimal interface to decouple middleware from the repository layer.
type StaffAssignmentProvider interface {
	ListByUserAndMerchant(ctx context.Context, userID, merchantID int64) ([]StaffAssignment, error)
}

// StaffAssignment mirrors the branch/staff assignment data needed by the middleware.
type StaffAssignment struct {
	ID        int64
	BranchID  int64
	Role      string
	IsDefault bool
}

const (
	ContextKeyStaffBranches  = "staff_branches"
	ContextKeyCurrentStaffID = "current_staff_id"
)

type BranchAssignment struct {
	StaffID   int64  `json:"staff_id"`
	BranchID  int64  `json:"branch_id"`
	Role      string `json:"role"`
	IsDefault bool   `json:"is_default"`
}

func BranchContext(provider StaffAssignmentProvider) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, ok := c.Get("user_id").(int64)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			merchantID := httputil.GetMerchantID(c)

			var assignments []BranchAssignment

			if merchantID != 0 {
				staffList, err := provider.ListByUserAndMerchant(c.Request().Context(), userID, merchantID)
				if err != nil {
					log.Printf("[WARN] BranchContext: failed to list staff assignments for user %d, merchant %d: %v", userID, merchantID, err)
				} else {
					for _, s := range staffList {
						assignments = append(assignments, BranchAssignment{
							StaffID:   s.ID,
							BranchID:  s.BranchID,
							Role:      s.Role,
							IsDefault: s.IsDefault,
						})
					}
				}
			}

			c.Set(ContextKeyStaffBranches, assignments)

			return next(c)
		}
	}
}
