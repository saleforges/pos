package category

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
	uc usecase.CategoryUsecase
}

func NewHandler(uc usecase.CategoryUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Create(c echo.Context) error {
	var req createCategoryReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	if req.Name == "" {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}

	merchantID := httputil.GetMerchantID(c)
	result, err := h.uc.Create(c.Request().Context(), usecase.CreateCategoryParams{
		MerchantID: merchantID,
		Name:       req.Name,
		ParentID:   req.ParentID,
	})
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusCreated, result)
}

func (h *Handler) List(c echo.Context) error {
	merchantID := httputil.GetMerchantID(c)
	items, err := h.uc.ListByMerchant(c.Request().Context(), merchantID)
	if err != nil {
		logger.Error("category.List failed", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
	return httputil.WriteJSON(c, http.StatusOK, items)
}

func (h *Handler) Get(c echo.Context) error {
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

func (h *Handler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	merchantID := httputil.GetMerchantID(c)

	var req updateCategoryReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	result, err := h.uc.Update(c.Request().Context(), usecase.UpdateCategoryParams{
		ID:         id,
		MerchantID: merchantID,
		Name:       req.Name,
		ParentID:   req.ParentID,
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
	return httputil.WriteJSON(c, http.StatusOK, map[string]string{"message": "category deleted"})
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
	case domain.ErrCategoryNotFound:
		return httputil.WriteError(c, http.StatusNotFound, err)
	case domain.ErrInvalidCategory:
		return httputil.WriteError(c, http.StatusBadRequest, err)
	default:
		logger.Error("category handler error", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
}
