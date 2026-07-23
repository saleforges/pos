package sellableitem

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/logger"
)

type Handler struct {
	uc usecase.SellableItemUsecase
}

func NewHandler(uc usecase.SellableItemUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Create(c echo.Context) error {
	productID, err := strconv.ParseInt(c.Param("productId"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	var req createSellableItemReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	if req.Name == "" || req.UnitID == 0 {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}

	result, err := h.uc.Create(c.Request().Context(), usecase.CreateSellableItemParams{
		ProductID:      productID,
		Name:           req.Name,
		UnitID:         req.UnitID,
		Price:          req.Price,
		TrackInventory: req.TrackInventory,
		ImageURL:       req.ImageURL,
	})
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusCreated, result)
}

func (h *Handler) ListByProduct(c echo.Context) error {
	productID, err := strconv.ParseInt(c.Param("productId"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	items, err := h.uc.ListByProduct(c.Request().Context(), productID)
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, items)
}

func (h *Handler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	var req updateSellableItemReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	result, err := h.uc.Update(c.Request().Context(), usecase.UpdateSellableItemParams{
		ID:             id,
		Name:           req.Name,
		UnitID:         req.UnitID,
		Price:          req.Price,
		TrackInventory: req.TrackInventory,
		ImageURL:       req.ImageURL,
		Status:         req.Status,
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
	if err := h.uc.Delete(c.Request().Context(), id); err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, map[string]string{"message": "item deleted"})
}

func (h *Handler) Restore(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	result, err := h.uc.Restore(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, result)
}

func mapError(c echo.Context, err error) error {
	switch err {
	case domain.ErrSellableItemNotFound, domain.ErrProductNotFound:
		return httputil.WriteError(c, http.StatusNotFound, err)
	case domain.ErrInvalidSellableItem, domain.ErrUnitNotFound:
		return httputil.WriteError(c, http.StatusBadRequest, err)
	default:
		logger.Error("sellable item handler error", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
}
