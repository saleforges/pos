package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/merchant/port/repository"
	"github.com/saleforge/pos/services/pkg/httputil"
)

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

func BranchContext(staffRepo repository.StaffRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, ok := c.Get("user_id").(int64)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			merchantID := httputil.GetMerchantID(c)

			var assignments []BranchAssignment

			if merchantID != 0 {
				staffList, err := staffRepo.ListByUserAndMerchant(c.Request().Context(), userID, merchantID)
				if err == nil {
					for _, s := range staffList {
						assignments = append(assignments, BranchAssignment{
							StaffID:   s.ID,
							BranchID:  s.BranchID,
							Role:      string(s.Role),
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
