package staff

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/internal/merchant/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
)

type Handler struct {
	uc usecase.StaffUsecase
}

func NewHandler(uc usecase.StaffUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) AssignStaff(c echo.Context) error {
	var req assignStaffReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	merchantID := httputil.GetMerchantID(c)
	if req.BranchID == 0 || req.UserID == 0 || req.Role == "" {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}

	result, err := h.uc.AssignStaff(c.Request().Context(), usecase.AssignStaffParams{
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

func (h *Handler) GetStaff(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	staff, err := h.uc.GetStaff(c.Request().Context(), id)
	if err != nil {
		if err == domain.ErrStaffNotFound {
			return httputil.WriteError(c, http.StatusNotFound, err)
		}
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, staff)
}

func (h *Handler) ListStaffByBranch(c echo.Context) error {
	branchID, err := strconv.ParseInt(c.Param("branchId"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	p := httputil.ParsePageParams(c)
	data, meta, err := h.uc.ListStaffByBranch(c.Request().Context(), branchID, p)
	if err != nil {
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WritePaginated(c, http.StatusOK, data, *meta)
}

func (h *Handler) ListStaffByMerchant(c echo.Context) error {
	merchantID := httputil.GetMerchantID(c)
	p := httputil.ParsePageParams(c)
	data, meta, err := h.uc.ListStaffByMerchant(c.Request().Context(), merchantID, p)
	if err != nil {
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WritePaginated(c, http.StatusOK, data, *meta)
}

func (h *Handler) UpdateStaff(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	var req updateStaffReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	staff, err := h.uc.UpdateStaff(c.Request().Context(), usecase.UpdateStaffParams{
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

func (h *Handler) RemoveStaff(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	if err := h.uc.RemoveStaff(c.Request().Context(), id); err != nil {
		if err == domain.ErrStaffNotFound {
			return httputil.WriteError(c, http.StatusNotFound, err)
		}
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, nil)
}

func (h *Handler) MyStaffAssignments(c echo.Context) error {
	userID, ok := c.Get("user_id").(int64)
	if !ok {
		return httputil.WriteError(c, http.StatusUnauthorized, httputil.ErrMissingFields)
	}
	merchantID := httputil.GetMerchantID(c)
	if merchantID == 0 {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}

	assignments, err := h.uc.GetMyStaffAssignments(c.Request().Context(), userID, merchantID)
	if err != nil {
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, assignments)
}

func (h *Handler) SetMyDefaultBranch(c echo.Context) error {
	var req setDefaultBranchReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	if req.BranchID == 0 {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}

	userID, ok := c.Get("user_id").(int64)
	if !ok {
		return httputil.WriteError(c, http.StatusUnauthorized, httputil.ErrMissingFields)
	}
	if err := h.uc.SetMyDefaultBranch(c.Request().Context(), userID, req.BranchID); err != nil {
		switch err {
		case domain.ErrBranchNotFound:
			return httputil.WriteError(c, http.StatusNotFound, err)
		}
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, map[string]string{"message": "default branch updated"})
}
