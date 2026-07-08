package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/logger"
)

type ProductHandler struct {
	uc usecase.ProductUsecase
}

func NewProductHandler(uc usecase.ProductUsecase) *ProductHandler {
	return &ProductHandler{uc: uc}
}

func (h *ProductHandler) Create(c echo.Context) error {
	var req createProductReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	if req.Name == "" || req.SKU == "" || req.CategoryID == "" {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}

	result, err := h.uc.Create(c.Request().Context(), usecase.CreateProductInput{
		MerchantID:  c.Param("merchantID"),
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		SKU:         req.SKU,
		Barcode:     req.Barcode,
		Description: req.Description,
		Price:       req.Price,
		Cost:        req.Cost,
		TaxRate:     req.TaxRate,
		Unit:        req.Unit,
		ImageURL:    req.ImageURL,
	})
	if err != nil {
		return mapProductError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusCreated, toProductResponse(*result))
}

func (h *ProductHandler) GetByID(c echo.Context) error {
	result, err := h.uc.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return mapProductError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, toProductResponse(*result))
}

func (h *ProductHandler) List(c echo.Context) error {
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	merchantID := c.Param("merchantID")
	result, err := h.uc.List(c.Request().Context(), merchantID, offset, limit)
	if err != nil {
		logger.Error("product.List failed", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
	resp := make([]productResponse, len(result))
	for i, p := range result {
		resp[i] = toProductResponse(p)
	}
	return httputil.WriteJSON(c, http.StatusOK, resp)
}

func (h *ProductHandler) Update(c echo.Context) error {
	var req updateProductReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	result, err := h.uc.Update(c.Request().Context(), usecase.UpdateProductInput{
		ID:          c.Param("id"),
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		SKU:         req.SKU,
		Barcode:     req.Barcode,
		Description: req.Description,
		Price:       req.Price,
		Cost:        req.Cost,
		TaxRate:     req.TaxRate,
		Unit:        req.Unit,
		ImageURL:    req.ImageURL,
		Status:      req.Status,
	})
	if err != nil {
		return mapProductError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, toProductResponse(*result))
}

func (h *ProductHandler) Delete(c echo.Context) error {
	err := h.uc.Delete(c.Request().Context(), c.Param("id"))
	if err != nil {
		return mapProductError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, nil)
}

func mapProductError(c echo.Context, err error) error {
	switch err {
	case domain.ErrProductNotFound:
		return httputil.WriteError(c, http.StatusNotFound, err)
	case domain.ErrCategoryNotFound:
		return httputil.WriteError(c, http.StatusBadRequest, err)
	case domain.ErrSkuExists, domain.ErrBarcodeExists:
		return httputil.WriteError(c, http.StatusConflict, err)
	default:
		logger.Error("product handler error", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
}
