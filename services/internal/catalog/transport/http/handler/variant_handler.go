package handler

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/logger"
)

type VariantHandler struct {
	uc usecase.VariantUsecase
}

func NewVariantHandler(uc usecase.VariantUsecase) *VariantHandler {
	return &VariantHandler{uc: uc}
}

func (h *VariantHandler) Create(c echo.Context) error {
	var req createVariantReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	if req.Name == "" || req.SKU == "" {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}

	result, err := h.uc.Create(c.Request().Context(), usecase.CreateVariantInput{
		ProductID: c.Param("productID"),
		Name:      req.Name,
		SKU:       req.SKU,
		Barcode:   req.Barcode,
		Price:     req.Price,
		Cost:      req.Cost,
		ImageURL:  req.ImageURL,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		return mapVariantError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusCreated, toVariantResponse(*result))
}

func (h *VariantHandler) ListByProduct(c echo.Context) error {
	result, err := h.uc.ListByProduct(c.Request().Context(), c.Param("productID"))
	if err != nil {
		return mapVariantError(c, err)
	}
	resp := make([]variantResponse, len(result))
	for i, v := range result {
		resp[i] = toVariantResponse(v)
	}
	return httputil.WriteJSON(c, http.StatusOK, resp)
}

func (h *VariantHandler) Update(c echo.Context) error {
	var req updateVariantReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	result, err := h.uc.Update(c.Request().Context(), usecase.UpdateVariantInput{
		ID:        c.Param("id"),
		Name:      req.Name,
		SKU:       req.SKU,
		Barcode:   req.Barcode,
		Price:     req.Price,
		Cost:      req.Cost,
		ImageURL:  req.ImageURL,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		return mapVariantError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, toVariantResponse(*result))
}

func (h *VariantHandler) Delete(c echo.Context) error {
	err := h.uc.Delete(c.Request().Context(), c.Param("id"))
	if err != nil {
		return mapVariantError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, nil)
}

func mapVariantError(c echo.Context, err error) error {
	switch err {
	case domain.ErrVariantNotFound:
		return httputil.WriteError(c, http.StatusNotFound, err)
	case domain.ErrProductNotFound:
		return httputil.WriteError(c, http.StatusBadRequest, err)
	default:
		logger.Error("variant handler error", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
}
