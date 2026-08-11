package movement

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/inventory/domain"
	"github.com/saleforge/pos/services/internal/inventory/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/logger"
)

type Handler struct {
	uc usecase.StockMovementUsecase
}

func NewHandler(uc usecase.StockMovementUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) List(c echo.Context) error {
	merchantID := httputil.GetMerchantID(c)

	var branchID int64
	if raw := c.QueryParam("branchId"); raw != "" {
		branchID, _ = strconv.ParseInt(raw, 10, 64)
	}

	var productItemID *int64
	if raw := c.QueryParam("productItemId"); raw != "" {
		if v, err := strconv.ParseInt(raw, 10, 64); err == nil {
			productItemID = &v
		}
	}

	var from, to *time.Time
	if raw := c.QueryParam("from"); raw != "" {
		t, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
		}
		from = &t
	}
	if raw := c.QueryParam("to"); raw != "" {
		t, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
		}
		to = &t
	}

	result, err := h.uc.List(c.Request().Context(), usecase.ListMovementsParams{
		MerchantID:    merchantID,
		BranchID:      branchID,
		ProductItemID: productItemID,
		From:          from,
		To:            to,
	})
	if err != nil {
		if err == domain.ErrInvalidStock {
			return httputil.WriteError(c, http.StatusBadRequest, err)
		}
		logger.Error("movement.List failed", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
	return httputil.WriteJSON(c, http.StatusOK, map[string]interface{}{
		"data": result,
	})
}
