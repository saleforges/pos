package stock

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
	uc usecase.StockUsecase
}

func NewHandler(uc usecase.StockUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Create(c echo.Context) error {
	var req createStockReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	merchantID := httputil.GetMerchantID(c)

	result, err := h.uc.Create(c.Request().Context(), usecase.CreateStockParams{
		MerchantID:    merchantID,
		BranchID:      req.BranchID,
		ProductItemID: req.ProductItemID,
		Available:     req.Available,
	})
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusCreated, result)
}

func (h *Handler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	merchantID := httputil.GetMerchantID(c)
	result, err := h.uc.GetByID(c.Request().Context(), id, merchantID)
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, result)
}

func (h *Handler) List(c echo.Context) error {
	merchantID := httputil.GetMerchantID(c)

	items, err := h.uc.List(c.Request().Context(), merchantID)
	if err != nil {
		logger.Error("stock.List failed", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}

	return httputil.WriteJSON(c, http.StatusOK, map[string]interface{}{
		"data": items,
	})
}

func (h *Handler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	merchantID := httputil.GetMerchantID(c)

	var req updateStockReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	result, err := h.uc.Update(c.Request().Context(), usecase.UpdateStockParams{
		ID:         id,
		MerchantID: merchantID,
		Available:  req.Available,
	})
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, result)
}

func (h *Handler) Transfer(c echo.Context) error {
	var req struct {
		FromBranchID int64                     `json:"fromBranchId"`
		ToBranchID   int64                     `json:"toBranchId"`
		Items        []usecase.AdjustStockItem `json:"items"`
	}
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, domain.ErrInvalidStock)
	}
	merchantID := httputil.GetMerchantID(c)
	if err := h.uc.Transfer(c.Request().Context(), usecase.TransferStockParams{
		MerchantID:   merchantID,
		FromBranchID: req.FromBranchID,
		ToBranchID:   req.ToBranchID,
		Items:        req.Items,
	}); err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) Produce(c echo.Context) error {
	var req produceStockReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	merchantID := httputil.GetMerchantID(c)

	result, err := h.uc.Produce(c.Request().Context(), usecase.ProduceParams{
		MerchantID:    merchantID,
		BranchID:      req.BranchID,
		ProductItemID: req.ProductItemID,
		Quantity:      req.Quantity,
	})
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusCreated, result)
}

func (h *Handler) Opname(c echo.Context) error {
	var req opnameStockReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	merchantID := httputil.GetMerchantID(c)

	result, err := h.uc.Opname(c.Request().Context(), usecase.OpnameParams{
		MerchantID:     merchantID,
		BranchID:       req.BranchID,
		ProductItemID:  req.ProductItemID,
		ActualQuantity: req.ActualQuantity,
		Reason:         req.Reason,
	})
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, result)
}

func (h *Handler) Sync(c echo.Context) error {
	merchantID := httputil.GetMerchantID(c)

	var branchID int64
	if raw := c.QueryParam("branchId"); raw != "" {
		branchID, _ = strconv.ParseInt(raw, 10, 64)
	}

	var lastSync *time.Time
	if raw := c.QueryParam("lastSync"); raw != "" {
		t, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
		}
		lastSync = &t
	}

	result, err := h.uc.Sync(c.Request().Context(), merchantID, branchID, lastSync)
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, result)
}

func (h *Handler) Deduct(c echo.Context) error {
	var req adjustStockReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	items := make([]usecase.AdjustStockItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = usecase.AdjustStockItem{ProductItemID: it.ProductItemID, Quantity: it.Quantity}
	}

	if err := h.uc.Deduct(c.Request().Context(), usecase.AdjustStockParams{
		MerchantID:    req.MerchantID,
		BranchID:      req.BranchID,
		ReferenceType: req.ReferenceType,
		ReferenceID:   req.ReferenceID,
		Items:         items,
	}); err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, map[string]string{"message": "stock deducted"})
}

func (h *Handler) Restore(c echo.Context) error {
	var req adjustStockReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	items := make([]usecase.AdjustStockItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = usecase.AdjustStockItem{ProductItemID: it.ProductItemID, Quantity: it.Quantity}
	}

	if err := h.uc.Restore(c.Request().Context(), usecase.AdjustStockParams{
		MerchantID:    req.MerchantID,
		BranchID:      req.BranchID,
		ReferenceType: req.ReferenceType,
		ReferenceID:   req.ReferenceID,
		Items:         items,
	}); err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, map[string]string{"message": "stock restored"})
}

func mapError(c echo.Context, err error) error {
	switch err {
	case domain.ErrStockNotFound:
		return httputil.WriteError(c, http.StatusNotFound, err)
	case domain.ErrInvalidStock, domain.ErrNegativeAvailable, domain.ErrInsufficientStock, domain.ErrProductComponentNotFound:
		return httputil.WriteError(c, http.StatusBadRequest, err)
	default:
		logger.Error("stock handler error", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
}
