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

type CategoryHandler struct {
	uc usecase.CategoryUsecase
}

func NewCategoryHandler(uc usecase.CategoryUsecase) *CategoryHandler {
	return &CategoryHandler{uc: uc}
}

func (h *CategoryHandler) Create(c echo.Context) error {
	var req createCategoryReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	if req.Name == "" || req.Slug == "" {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}

	result, err := h.uc.Create(c.Request().Context(), usecase.CreateCategoryInput{
		MerchantID:  c.Param("merchantID"),
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		ParentID:    req.ParentID,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		return mapCategoryError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusCreated, toCategoryResponse(*result))
}

func (h *CategoryHandler) GetByID(c echo.Context) error {
	result, err := h.uc.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return mapCategoryError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, toCategoryResponse(*result))
}

func (h *CategoryHandler) List(c echo.Context) error {
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	merchantID := c.Param("merchantID")
	result, err := h.uc.List(c.Request().Context(), merchantID, offset, limit)
	if err != nil {
		logger.Error("category.List failed", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
	resp := make([]categoryResponse, len(result))
	for i, cat := range result {
		resp[i] = toCategoryResponse(cat)
	}
	return httputil.WriteJSON(c, http.StatusOK, resp)
}

func (h *CategoryHandler) Update(c echo.Context) error {
	var req updateCategoryReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	result, err := h.uc.Update(c.Request().Context(), usecase.UpdateCategoryInput{
		ID:          c.Param("id"),
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		ParentID:    req.ParentID,
		SortOrder:   req.SortOrder,
		Status:      req.Status,
	})
	if err != nil {
		return mapCategoryError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, toCategoryResponse(*result))
}

func (h *CategoryHandler) Delete(c echo.Context) error {
	err := h.uc.Delete(c.Request().Context(), c.Param("id"))
	if err != nil {
		return mapCategoryError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, nil)
}

func mapCategoryError(c echo.Context, err error) error {
	switch err {
	case domain.ErrCategoryNotFound:
		return httputil.WriteError(c, http.StatusNotFound, err)
	case domain.ErrCategoryExists:
		return httputil.WriteError(c, http.StatusConflict, err)
	default:
		logger.Error("category handler error", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
}
