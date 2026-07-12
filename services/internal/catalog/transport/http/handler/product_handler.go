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
	if req.Name == "" || req.SKU == "" || req.CategoryID == 0 {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}

	result, err := h.uc.Create(c.Request().Context(), usecase.CreateProductInput{
		MerchantID:  httputil.GetMerchantID(c),
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
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	result, err := h.uc.GetByID(c.Request().Context(), id)
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
	search := c.QueryParam("search")

	merchantID := httputil.GetMerchantID(c)
	result, err := h.uc.List(c.Request().Context(), merchantID, search, offset, limit)
	if err != nil {
		logger.Error("product.List failed", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
	items := make([]productResponse, len(result.Items))
	for i, p := range result.Items {
		items[i] = toProductResponse(p)
	}
	return httputil.WriteJSON(c, http.StatusOK, paginatedResponse[productResponse]{
		Items: items,
		Meta: paginatedMeta{
			Total:  result.Meta.Total,
			Offset: result.Meta.Offset,
			Limit:  result.Meta.Limit,
		},
	})
}

func (h *ProductHandler) Update(c echo.Context) error {
	var req updateProductReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	result, err := h.uc.Update(c.Request().Context(), usecase.UpdateProductInput{
		ID:          id,
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
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	err = h.uc.Delete(c.Request().Context(), id)
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
