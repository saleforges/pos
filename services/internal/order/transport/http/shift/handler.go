package shift

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/order/domain"
	"github.com/saleforge/pos/services/internal/order/transport/http/middleware"
	"github.com/saleforge/pos/services/internal/order/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/logger"
)

type Handler struct {
	uc usecase.ShiftUsecase
}

func NewHandler(uc usecase.ShiftUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Open(c echo.Context) error {
	var req openShiftReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	merchantID := httputil.GetMerchantID(c)
	result, err := h.uc.Open(c.Request().Context(), usecase.OpenShiftParams{
		MerchantID:   merchantID,
		BranchID:     req.BranchID,
		OpenedBy:     getUserID(c),
		StartingCash: req.StartingCash,
	})
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusCreated, result)
}

func (h *Handler) Close(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	var req closeShiftReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	merchantID := httputil.GetMerchantID(c)
	result, err := h.uc.Close(c.Request().Context(), usecase.CloseShiftParams{
		ShiftID:    id,
		MerchantID: merchantID,
		ClosedBy:   getUserID(c),
		ActualCash: req.ActualCash,
		Note:       req.Note,
	})
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, result)
}

func (h *Handler) GetActive(c echo.Context) error {
	merchantID := httputil.GetMerchantID(c)
	var branchID int64
	if raw := c.QueryParam("branchId"); raw != "" {
		branchID, _ = strconv.ParseInt(raw, 10, 64)
	}

	result, err := h.uc.GetActive(c.Request().Context(), merchantID, branchID)
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, result)
}

func (h *Handler) List(c echo.Context) error {
	merchantID := httputil.GetMerchantID(c)
	var branchID int64
	if raw := c.QueryParam("branchId"); raw != "" {
		branchID, _ = strconv.ParseInt(raw, 10, 64)
	}

	items, err := h.uc.List(c.Request().Context(), merchantID, branchID)
	if err != nil {
		logger.Error("shift.List failed", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
	return httputil.WriteJSON(c, http.StatusOK, map[string]interface{}{
		"data": items,
	})
}

func getUserID(c echo.Context) int64 {
	id, _ := c.Get(middleware.ContextKeyUserID).(int64)
	return id
}

func mapError(c echo.Context, err error) error {
	switch err {
	case domain.ErrShiftNotFound:
		return httputil.WriteError(c, http.StatusNotFound, err)
	case domain.ErrInvalidShift:
		return httputil.WriteError(c, http.StatusBadRequest, err)
	case domain.ErrShiftAlreadyOpen, domain.ErrShiftAlreadyClosed:
		return httputil.WriteError(c, http.StatusConflict, err)
	default:
		logger.Error("shift handler error", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
}
