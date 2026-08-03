package productcomponent

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/inventory/domain"
	"github.com/saleforge/pos/services/internal/inventory/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/logger"
)

type Handler struct {
	uc usecase.ProductComponentUsecase
}

func NewHandler(uc usecase.ProductComponentUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) GetByProductItem(c echo.Context) error {
	productItemID, err := strconv.ParseInt(c.Param("productItemId"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	merchantID := httputil.GetMerchantID(c)
	result, err := h.uc.GetByProductItem(c.Request().Context(), productItemID, merchantID)
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, result)
}

func (h *Handler) CreateOrUpdate(c echo.Context) error {
	productItemID, err := strconv.ParseInt(c.Param("productItemId"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	merchantID := httputil.GetMerchantID(c)

	var req updateProductComponentReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	items := make([]usecase.CreateProductComponentItemParams, len(req.Items))
	for i, item := range req.Items {
		// Default conversion factor is 1 (quantity already in stock unit).
		factor := item.ConversionFactor
		if factor <= 0 {
			factor = 1
		}
		items[i] = usecase.CreateProductComponentItemParams{
			ComponentProductItemID: item.ComponentProductItemID,
			Quantity:               item.Quantity,
			UnitID:                 item.UnitID,
			ConversionFactor:       factor,
		}
	}

	// Try to update existing component first
	existing, err := h.uc.GetByProductItem(c.Request().Context(), productItemID, merchantID)
	if err == nil && existing != nil {
		// Update existing component
		result, err := h.uc.Update(c.Request().Context(), usecase.UpdateProductComponentParams{
			MerchantID:    merchantID,
			ProductItemID: productItemID,
			Items:         items,
		})
		if err != nil {
			return mapError(c, err)
		}
		return httputil.WriteJSON(c, http.StatusOK, result)
	}

	// Create new component
	result, err := h.uc.Create(c.Request().Context(), usecase.CreateProductComponentParams{
		MerchantID:    merchantID,
		ProductItemID: productItemID,
		Items:         items,
	})
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusCreated, result)
}

func mapError(c echo.Context, err error) error {
	switch err {
	case domain.ErrProductComponentNotFound:
		return httputil.WriteError(c, http.StatusNotFound, err)
	case domain.ErrInvalidProductComponent, domain.ErrInvalidProductComponentItem, domain.ErrNoComponentItems:
		return httputil.WriteError(c, http.StatusBadRequest, err)
	case domain.ErrComponentAlreadyExists:
		return httputil.WriteError(c, http.StatusConflict, err)
	default:
		logger.Error("product_component handler error", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
}
