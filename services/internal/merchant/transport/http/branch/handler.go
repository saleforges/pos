package branch

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
	uc usecase.BranchUsecase
}

func NewHandler(uc usecase.BranchUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) CreateBranch(c echo.Context) error {
	var req createBranchReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	merchantID := httputil.GetMerchantID(c)
	if req.Name == "" || req.Code == "" {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}

	result, err := h.uc.CreateBranch(c.Request().Context(), usecase.CreateBranchParams{
		MerchantID:     merchantID,
		Name:           req.Name,
		Code:           req.Code,
		Address:        req.Address,
		Phone:          req.Phone,
		OperatingDays:  req.OperatingDays,
		OperatingHours: req.OperatingHours,
	})
	if err != nil {
		switch err {
		case domain.ErrMerchantNotFound:
			return httputil.WriteError(c, http.StatusNotFound, err)
		case domain.ErrBranchExists:
			return httputil.WriteError(c, http.StatusConflict, err)
		case domain.ErrInvalidBranch:
			return httputil.WriteError(c, http.StatusBadRequest, err)
		}
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusCreated, result)
}

func (h *Handler) GetBranch(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	branch, err := h.uc.GetBranch(c.Request().Context(), id)
	if err != nil {
		if err == domain.ErrBranchNotFound {
			return httputil.WriteError(c, http.StatusNotFound, err)
		}
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, branch)
}

func (h *Handler) ListBranches(c echo.Context) error {
	merchantID := httputil.GetMerchantID(c)
	p := httputil.ParsePageParams(c)
	data, meta, err := h.uc.ListBranches(c.Request().Context(), merchantID, p)
	if err != nil {
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WritePaginated(c, http.StatusOK, data, *meta)
}

func (h *Handler) UpdateBranch(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	var req updateBranchReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	branch, err := h.uc.UpdateBranch(c.Request().Context(), usecase.UpdateBranchParams{
		ID:             id,
		Name:           req.Name,
		Address:        req.Address,
		Phone:          req.Phone,
		Status:         req.Status,
		OperatingDays:  req.OperatingDays,
		OperatingHours: req.OperatingHours,
	})
	if err != nil {
		switch err {
		case domain.ErrBranchNotFound:
			return httputil.WriteError(c, http.StatusNotFound, err)
		case domain.ErrBranchExists:
			return httputil.WriteError(c, http.StatusConflict, err)
		case domain.ErrInvalidBranch:
			return httputil.WriteError(c, http.StatusBadRequest, err)
		}
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, branch)
}

func (h *Handler) DeleteBranch(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	if err := h.uc.DeleteBranch(c.Request().Context(), id); err != nil {
		switch err {
		case domain.ErrBranchNotFound:
			return httputil.WriteError(c, http.StatusNotFound, err)
		case domain.ErrBranchHasStaff:
			return httputil.WriteError(c, http.StatusConflict, err)
		}
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, map[string]string{"message": "branch deleted"})
}
