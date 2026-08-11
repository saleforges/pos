package productitem

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/logger"
)

func parseProductItemStatus(s *string) *domain.ProductItemStatus {
	if s == nil {
		return nil
	}
	v := domain.ProductItemStatus(*s)
	return &v
}

type Handler struct {
	uc usecase.ProductItemUsecase
}

func NewHandler(uc usecase.ProductItemUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Create(c echo.Context) error {
	productID, err := strconv.ParseInt(c.Param("productId"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	var req createProductItemReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	if req.Name == "" {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}

	var unitID *int64
	if req.UnitID != nil {
		uid := int64(*req.UnitID)
		unitID = &uid
	}

	currency := req.Currency
	if currency == "" {
		currency = "IDR"
	}

	merchantID := httputil.GetMerchantID(c)

	result, err := h.uc.Create(c.Request().Context(), usecase.CreateProductItemParams{
		MerchantID:     merchantID,
		ProductID:      productID,
		Name:           req.Name,
		SKU:            req.SKU,
		UnitID:         unitID,
		PriceAmount:    req.Price,
		PriceCurrency:  currency,
		TrackInventory: req.TrackInventory,
		ImageURL:       req.ImageURL,
	})
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusCreated, result)
}

func (h *Handler) SetBranchPrice(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	var req struct {
		BranchID int64   `json:"branchId"`
		Amount   float64 `json:"amount"`
		Currency string  `json:"currency"`
	}
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	merchantID := httputil.GetMerchantID(c)
	if err := h.uc.SetBranchPrice(c.Request().Context(), id, req.BranchID, merchantID, req.Amount, req.Currency); err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) DeleteBranchPrice(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	branchID, _ := strconv.ParseInt(c.QueryParam("branchId"), 10, 64)
	merchantID := httputil.GetMerchantID(c)
	if err := h.uc.DeleteBranchPrice(c.Request().Context(), id, branchID, merchantID); err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) GetBranchPrice(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	branchID, _ := strconv.ParseInt(c.QueryParam("branchId"), 10, 64)
	merchantID := httputil.GetMerchantID(c)
	price, err := h.uc.GetBranchPrice(c.Request().Context(), id, branchID, merchantID)
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, price)
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

func (h *Handler) ListByProduct(c echo.Context) error {
	productID, err := strconv.ParseInt(c.Param("productId"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	merchantID := httputil.GetMerchantID(c)
	items, err := h.uc.ListByProduct(c.Request().Context(), productID, merchantID)
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, items)
}

func (h *Handler) ListByMerchant(c echo.Context) error {
	merchantID := httputil.GetMerchantID(c)

	items, err := h.uc.ListByMerchant(c.Request().Context(), merchantID)
	if err != nil {
		logger.Error("product_item.ListByMerchant failed", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}

	type productBrief struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	type unitBrief struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}
	type priceBrief struct {
		Amount   float64 `json:"amount"`
		Currency string  `json:"currency"`
	}
	type itemResp struct {
		ID             int64        `json:"id"`
		Name           string       `json:"name"`
		SKU            string       `json:"sku,omitempty"`
		Product        productBrief `json:"product"`
		Unit           *unitBrief   `json:"unit,omitempty"`
		Price          priceBrief   `json:"price"`
		TrackInventory bool         `json:"trackInventory"`
		Status         string       `json:"status"`
	}

	result := make([]itemResp, 0, len(items))
	for _, it := range items {
		resp := itemResp{
			ID:   it.ID,
			Name: it.Name,
			SKU:  it.SKU,
			Price: priceBrief{
				Amount:   it.Price.Amount,
				Currency: it.Price.Currency,
			},
			TrackInventory: it.TrackInventory,
			Status:         string(it.Status),
		}
		result = append(result, resp)
	}

	return httputil.WriteJSON(c, http.StatusOK, map[string]interface{}{
		"data": result,
	})
}

func (h *Handler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	merchantID := httputil.GetMerchantID(c)

	var req updateProductItemReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	var unitID *int64
	if req.UnitID != nil {
		uid := int64(*req.UnitID)
		unitID = &uid
	}

	result, err := h.uc.Update(c.Request().Context(), usecase.UpdateProductItemParams{
		ID:             id,
		MerchantID:     merchantID,
		Name:           req.Name,
		SKU:            req.SKU,
		UnitID:         unitID,
		PriceAmount:    req.Price,
		PriceCurrency:  req.Currency,
		TrackInventory: req.TrackInventory,
		ImageURL:       req.ImageURL,
		Status:         parseProductItemStatus(req.Status),
	})
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, result)
}

func (h *Handler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	merchantID := httputil.GetMerchantID(c)
	if err := h.uc.Delete(c.Request().Context(), id, merchantID); err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, map[string]string{"message": "item deleted"})
}

func (h *Handler) Restore(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	merchantID := httputil.GetMerchantID(c)
	result, err := h.uc.Restore(c.Request().Context(), id, merchantID)
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, result)
}

func mapError(c echo.Context, err error) error {
	switch err {
	case domain.ErrProductItemNotFound, domain.ErrProductNotFound:
		return httputil.WriteError(c, http.StatusNotFound, err)
	case domain.ErrInvalidProductItem, domain.ErrUnitNotFound:
		return httputil.WriteError(c, http.StatusBadRequest, err)
	case domain.ErrSKUDuplicate:
		return httputil.WriteError(c, http.StatusConflict, err)
	default:
		logger.Error("product item handler error", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
}
