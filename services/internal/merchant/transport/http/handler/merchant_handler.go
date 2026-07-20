package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/internal/merchant/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
)

type MerchantHandler struct {
	uc usecase.MerchantUsecase
}

func NewMerchantHandler(uc usecase.MerchantUsecase) *MerchantHandler {
	return &MerchantHandler{uc: uc}
}

type createMerchantReq struct {
	Name      string                  `json:"name"`
	LegalName string                  `json:"legal_name"`
	Address   string                  `json:"address"`
	Phone     string                  `json:"phone"`
	Email     string                  `json:"email"`
	TaxID     string                  `json:"tax_id"`
	Settings  domain.MerchantSettings `json:"settings"`
}

func (h *MerchantHandler) Create(c echo.Context) error {
	var req createMerchantReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	if req.Name == "" || req.Email == "" {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}

	result, err := h.uc.CreateMerchant(c.Request().Context(), usecase.CreateMerchantInput{
		Name:      req.Name,
		LegalName: req.LegalName,
		Address:   req.Address,
		Phone:     req.Phone,
		Email:     req.Email,
		TaxID:     req.TaxID,
		Settings:  req.Settings,
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

func (h *MerchantHandler) Get(c echo.Context) error {
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

func (h *MerchantHandler) List(c echo.Context) error {
	offset := parseInt(c.QueryParam("offset"), 0)
	limit := parseInt(c.QueryParam("limit"), 20)

	merchants, err := h.uc.ListMerchants(c.Request().Context(), offset, limit)
	if err != nil {
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, merchants)
}

type updateMerchantReq struct {
	Name      *string                  `json:"name,omitempty"`
	LegalName *string                  `json:"legal_name,omitempty"`
	Address   *string                  `json:"address,omitempty"`
	Phone     *string                  `json:"phone,omitempty"`
	Email     *string                  `json:"email,omitempty"`
	TaxID     *string                  `json:"tax_id,omitempty"`
	Status    *domain.MerchantStatus   `json:"status,omitempty"`
	Settings  *domain.MerchantSettings `json:"settings,omitempty"`
}

func (h *MerchantHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	var req updateMerchantReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	merchant, err := h.uc.UpdateMerchant(c.Request().Context(), usecase.UpdateMerchantInput{
		ID:        id,
		Name:      req.Name,
		LegalName: req.LegalName,
		Address:   req.Address,
		Phone:     req.Phone,
		Email:     req.Email,
		TaxID:     req.TaxID,
		Status:    req.Status,
		Settings:  req.Settings,
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

func (h *MerchantHandler) Delete(c echo.Context) error {
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

func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return defaultVal
		}
		n = n*10 + int(c-'0')
	}
	return n
}
