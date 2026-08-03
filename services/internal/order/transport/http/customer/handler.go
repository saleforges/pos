package customer

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/order/domain"
	"github.com/saleforge/pos/services/internal/order/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/logger"
)

type Handler struct {
	uc usecase.CustomerUsecase
}

func NewHandler(uc usecase.CustomerUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Create(c echo.Context) error {
	var req createCustomerReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	merchantID := httputil.GetMerchantID(c)
	result, err := h.uc.Create(c.Request().Context(), usecase.CreateCustomerParams{
		MerchantID: merchantID,
		Name:       req.Name,
		Phone:      req.Phone,
		Address:    req.Address,
		Note:       req.Note,
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
	items, err := h.uc.List(c.Request().Context(), merchantID, c.QueryParam("q"))
	if err != nil {
		logger.Error("customer.List failed", "error", err.Error())
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

	var req updateCustomerReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	merchantID := httputil.GetMerchantID(c)
	result, err := h.uc.Update(c.Request().Context(), usecase.UpdateCustomerParams{
		ID:         id,
		MerchantID: merchantID,
		Name:       req.Name,
		Phone:      req.Phone,
		Address:    req.Address,
		Note:       req.Note,
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
	return httputil.WriteJSON(c, http.StatusOK, map[string]string{"message": "customer deleted"})
}

func mapError(c echo.Context, err error) error {
	switch err {
	case domain.ErrCustomerNotFound:
		return httputil.WriteError(c, http.StatusNotFound, err)
	case domain.ErrInvalidCustomer:
		return httputil.WriteError(c, http.StatusBadRequest, err)
	default:
		logger.Error("customer handler error", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
}
