package handler

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/internal/merchant/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
)

type StaffHandler struct {
	uc usecase.StaffUsecase
}

func NewStaffHandler(uc usecase.StaffUsecase) *StaffHandler {
	return &StaffHandler{uc: uc}
}

type assignStaffReq struct {
	BranchID  string          `json:"branch_id"`
	UserID    string          `json:"user_id"`
	Role      domain.StaffRole `json:"role"`
	IsDefault bool            `json:"is_default"`
}

func (h *StaffHandler) AssignStaff(c echo.Context) error {
	var req assignStaffReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	merchantID := httputil.GetMerchantID(c)
	if merchantID == "" || req.BranchID == "" || req.UserID == "" || req.Role == "" {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}

	result, err := h.uc.AssignStaff(c.Request().Context(), usecase.AssignStaffInput{
		MerchantID: merchantID,
		BranchID:   req.BranchID,
		UserID:     req.UserID,
		Role:       req.Role,
		IsDefault:  req.IsDefault,
	})
	if err != nil {
		switch err {
		case domain.ErrMerchantNotFound, domain.ErrBranchNotFound:
			return httputil.WriteError(c, http.StatusNotFound, err)
		case domain.ErrStaffExists:
			return httputil.WriteError(c, http.StatusConflict, err)
		case domain.ErrInvalidStaff:
			return httputil.WriteError(c, http.StatusBadRequest, err)
		}
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusCreated, result)
}

func (h *StaffHandler) GetStaff(c echo.Context) error {
	id := c.Param("id")
	staff, err := h.uc.GetStaff(c.Request().Context(), id)
	if err != nil {
		if err == domain.ErrStaffNotFound {
			return httputil.WriteError(c, http.StatusNotFound, err)
		}
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, staff)
}

func (h *StaffHandler) ListStaffByBranch(c echo.Context) error {
	branchID := c.Param("branchId")
	staff, err := h.uc.ListStaffByBranch(c.Request().Context(), branchID)
	if err != nil {
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, staff)
}

func (h *StaffHandler) ListStaffByMerchant(c echo.Context) error {
	merchantID := httputil.GetMerchantID(c)
	staff, err := h.uc.ListStaffByMerchant(c.Request().Context(), merchantID)
	if err != nil {
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, staff)
}

type updateStaffReq struct {
	Role   *domain.StaffRole   `json:"role,omitempty"`
	Status *domain.StaffStatus `json:"status,omitempty"`
}

func (h *StaffHandler) UpdateStaff(c echo.Context) error {
	id := c.Param("id")
	var req updateStaffReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	staff, err := h.uc.UpdateStaff(c.Request().Context(), usecase.UpdateStaffInput{
		ID:     id,
		Role:   req.Role,
		Status: req.Status,
	})
	if err != nil {
		if err == domain.ErrStaffNotFound {
			return httputil.WriteError(c, http.StatusNotFound, err)
		}
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, staff)
}

func (h *StaffHandler) RemoveStaff(c echo.Context) error {
	id := c.Param("id")
	if err := h.uc.RemoveStaff(c.Request().Context(), id); err != nil {
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, nil)
}

func (h *StaffHandler) MyStaffAssignments(c echo.Context) error {
	userID, _ := c.Get("user_id").(string)
	merchantID := httputil.GetMerchantID(c)
	if merchantID == "" {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}

	assignments, err := h.uc.GetMyStaffAssignments(c.Request().Context(), userID, merchantID)
	if err != nil {
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, assignments)
}

type setDefaultBranchReq struct {
	BranchID string `json:"branch_id"`
}

func (h *StaffHandler) SetMyDefaultBranch(c echo.Context) error {
	var req setDefaultBranchReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	if req.BranchID == "" {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}

	userID, _ := c.Get("user_id").(string)
	if err := h.uc.SetMyDefaultBranch(c.Request().Context(), userID, req.BranchID); err != nil {
		switch err {
		case domain.ErrBranchNotFound:
			return httputil.WriteError(c, http.StatusNotFound, err)
		}
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, map[string]string{"message": "default branch updated"})
}
