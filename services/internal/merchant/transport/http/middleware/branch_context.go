package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/merchant/port/repository"
	"github.com/saleforge/pos/services/pkg/httputil"
)

const (
	ContextKeyStaffBranches  = "staff_branches"
	ContextKeyDefaultBranch  = "default_branch"
	ContextKeyCurrentStaffID = "current_staff_id"
)

type BranchAssignment struct {
	StaffID   string `json:"staff_id"`
	BranchID  string `json:"branch_id"`
	Role      string `json:"role"`
	IsDefault bool   `json:"is_default"`
}

func BranchContext(staffRepo repository.StaffRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, ok := c.Get("user_id").(string)
			if !ok || userID == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			merchantID := httputil.GetMerchantID(c)

			var assignments []BranchAssignment
			var defaultBranch string

			if merchantID != "" {
				staffList, err := staffRepo.ListByUserAndMerchant(c.Request().Context(), userID, merchantID)
				if err == nil {
					for _, s := range staffList {
						assignments = append(assignments, BranchAssignment{
							StaffID:   s.ID,
							BranchID:  s.BranchID,
							Role:      string(s.Role),
							IsDefault: s.IsDefault,
						})
						if s.IsDefault {
							defaultBranch = s.BranchID
						}
					}
				}
			}

			c.Set(ContextKeyStaffBranches, assignments)
			c.Set(ContextKeyDefaultBranch, defaultBranch)

			return next(c)
		}
	}
}
