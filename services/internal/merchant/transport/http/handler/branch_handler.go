package handler

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/internal/merchant/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
)

type BranchHandler struct {
	uc usecase.BranchUsecase
}

func NewBranchHandler(uc usecase.BranchUsecase) *BranchHandler {
	return &BranchHandler{uc: uc}
}

type createBranchReq struct {
	MerchantID     string                 `json:"merchant_id"`
	Name           string                 `json:"name"`
	Code           string                 `json:"code"`
	Address        string                 `json:"address"`
	Phone          string                 `json:"phone"`
	OperatingDays  []string               `json:"operating_days"`
	OperatingHours *domain.OperatingHours `json:"operating_hours"`
}

func (h *BranchHandler) CreateBranch(c echo.Context) error {
	var req createBranchReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	if req.Name == "" || req.MerchantID == "" || req.Code == "" {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}

	result, err := h.uc.CreateBranch(c.Request().Context(), usecase.CreateBranchInput{
		MerchantID:     req.MerchantID,
		Name:           req.Name,
		Code:           req.Code,
		Address:        req.Address,
		Phone:          req.Phone,
		OperatingDays:  req.OperatingDays,
		OperatingHours: req.OperatingHours,
	})
	if err != nil {
		if err == domain.ErrMerchantNotFound {
			return httputil.WriteError(c, http.StatusNotFound, err)
		}
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusCreated, result)
}

func (h *BranchHandler) GetBranch(c echo.Context) error {
	id := c.Param("id")
	branch, err := h.uc.GetBranch(c.Request().Context(), id)
	if err != nil {
		if err == domain.ErrBranchNotFound {
			return httputil.WriteError(c, http.StatusNotFound, err)
		}
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, branch)
}

func (h *BranchHandler) ListBranches(c echo.Context) error {
	merchantID := c.Param("merchantId")
	branches, err := h.uc.ListBranches(c.Request().Context(), merchantID)
	if err != nil {
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, branches)
}

type updateBranchReq struct {
	Name           *string                `json:"name,omitempty"`
	Address        *string                `json:"address,omitempty"`
	Phone          *string                `json:"phone,omitempty"`
	Status         *domain.BranchStatus   `json:"status,omitempty"`
	OperatingDays  []string               `json:"operating_days,omitempty"`
	OperatingHours *domain.OperatingHours `json:"operating_hours,omitempty"`
}

func (h *BranchHandler) UpdateBranch(c echo.Context) error {
	id := c.Param("id")
	var req updateBranchReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	branch, err := h.uc.UpdateBranch(c.Request().Context(), usecase.UpdateBranchInput{
		ID:             id,
		Name:           req.Name,
		Address:        req.Address,
		Phone:          req.Phone,
		Status:         req.Status,
		OperatingDays:  req.OperatingDays,
		OperatingHours: req.OperatingHours,
	})
	if err != nil {
		if err == domain.ErrBranchNotFound {
			return httputil.WriteError(c, http.StatusNotFound, err)
		}
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, branch)
}

func (h *BranchHandler) DeleteBranch(c echo.Context) error {
	id := c.Param("id")
	if err := h.uc.DeleteBranch(c.Request().Context(), id); err != nil {
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, map[string]string{"message": "branch deleted"})
}
