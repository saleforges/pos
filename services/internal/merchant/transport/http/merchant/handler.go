package merchant

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/internal/merchant/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
)

type Handler struct {
	uc usecase.MerchantUsecase
}

func NewHandler(uc usecase.MerchantUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Create(c echo.Context) error {
	var req createMerchantReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	if req.Name == "" || req.Email == "" {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}

	result, err := h.uc.CreateMerchant(c.Request().Context(), usecase.CreateMerchantParams{
		Name:      req.Name,
		LegalName: req.LegalName,
		Address:   req.Address,
		Phone:     req.Phone,
		Email:     req.Email,
		TaxID:     req.TaxID,
	})
	if err != nil {
		switch err {
		case domain.ErrMerchantExists:
			return httputil.WriteError(c, http.StatusConflict, err)
		case domain.ErrInvalidMerchant:
			return httputil.WriteError(c, http.StatusBadRequest, err)
		}
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}

	return httputil.WriteJSON(c, http.StatusCreated, result)
}

func (h *Handler) Get(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	merchant, err := h.uc.GetMerchant(c.Request().Context(), id)
	if err != nil {
		if err == domain.ErrMerchantNotFound {
			return httputil.WriteError(c, http.StatusNotFound, err)
		}
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, merchant)
}

func (h *Handler) List(c echo.Context) error {
	p := httputil.ParsePageParams(c)
	data, meta, err := h.uc.ListMerchants(c.Request().Context(), p)
	if err != nil {
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WritePaginated(c, http.StatusOK, data, *meta)
}

func (h *Handler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	var req updateMerchantReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	merchant, err := h.uc.UpdateMerchant(c.Request().Context(), usecase.UpdateMerchantParams{
		ID:        id,
		Name:      req.Name,
		LegalName: req.LegalName,
		Address:   req.Address,
		Phone:     req.Phone,
		Email:     req.Email,
		TaxID:     req.TaxID,
	})
	if err != nil {
		switch err {
		case domain.ErrMerchantNotFound:
			return httputil.WriteError(c, http.StatusNotFound, err)
		case domain.ErrMerchantExists:
			return httputil.WriteError(c, http.StatusConflict, err)
		case domain.ErrInvalidMerchant:
			return httputil.WriteError(c, http.StatusBadRequest, err)
		}
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, merchant)
}

func (h *Handler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	if err := h.uc.DeleteMerchant(c.Request().Context(), id); err != nil {
		switch err {
		case domain.ErrMerchantNotFound:
			return httputil.WriteError(c, http.StatusNotFound, err)
		case domain.ErrMerchantHasBranches:
			return httputil.WriteError(c, http.StatusConflict, err)
		}
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, map[string]string{"message": "merchant deleted"})
}
