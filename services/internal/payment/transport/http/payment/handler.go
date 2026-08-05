package payment

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/payment/adapter/client/ipaymu"
	"github.com/saleforge/pos/services/internal/payment/domain"
	"github.com/saleforge/pos/services/internal/payment/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/logger"
)

// Handler exposes the internal payment creation API (called by the order
// service) and the public gateway callback.
type Handler struct {
	uc usecase.PaymentUsecase
	va string
}

func NewHandler(uc usecase.PaymentUsecase, va string) *Handler {
	return &Handler{uc: uc, va: va}
}

type createPaymentReq struct {
	OrderID    int64  `json:"orderId"`
	Method     string `json:"method,omitempty"`
	BuyerName  string `json:"buyerName,omitempty"`
	BuyerEmail string `json:"buyerEmail,omitempty"`
	BuyerPhone string `json:"buyerPhone,omitempty"`
}

func (h *Handler) Create(c echo.Context) error {
	var req createPaymentReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, domain.ErrInvalidPayment)
	}
	if req.OrderID == 0 {
		return httputil.WriteError(c, http.StatusBadRequest, domain.ErrInvalidPayment)
	}

	merchantID := httputil.GetMerchantID(c)
	result, err := h.uc.Create(c.Request().Context(), usecase.CreatePaymentParams{
		MerchantID: merchantID,
		OrderID:    req.OrderID,
		Method:     req.Method,
		BuyerName:  req.BuyerName,
		BuyerEmail: req.BuyerEmail,
		BuyerPhone: req.BuyerPhone,
	})
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, result)
}

// GetByID returns one payment transaction for the merchant (auth required).
func (h *Handler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, domain.ErrInvalidPayment)
	}
	merchantID := httputil.GetMerchantID(c)

	result, err := h.uc.GetByID(c.Request().Context(), id, merchantID)
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, result)
}

// Callback is the public gateway webhook. Security = signature verification.
func (h *Handler) Callback(c echo.Context) error {
	signature := c.Request().Header.Get("X-Signature")

	raw := map[string]interface{}{}
	if err := c.Bind(&raw); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, domain.ErrInvalidCallback)
	}

	if !ipaymu.VerifyCallback(h.va, signature, raw) {
		logger.Warn("ipaymu callback rejected: bad signature")
		return httputil.WriteError(c, http.StatusUnauthorized, domain.ErrInvalidCallback)
	}

	params := usecase.CallbackParams{
		ReferenceID: str(raw, "reference_id"),
		Status:      str(raw, "status"),
		StatusCode:  statusCode(raw),
		Amount:      str(raw, "amount"),
		PaymentRef:  str(raw, "trx_id"),
		Via:         str(raw, "via"),
		Channel:     str(raw, "channel"),
		Signature:   signature,
		MerchantVA:  str(raw, "va"),
	}

	if err := h.uc.HandleCallback(c.Request().Context(), params); err != nil {
		if err == domain.ErrInvalidCallback || err == domain.ErrPaymentNotFound {
			return httputil.WriteError(c, http.StatusBadRequest, err)
		}
		logger.Error("payment callback processing failed", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}

	return httputil.WriteJSON(c, http.StatusOK, map[string]string{"message": "ok"})
}

// GetStaticQR returns the merchant's permanent counter QRIS.
func (h *Handler) GetStaticQR(c echo.Context) error {
	merchantID := httputil.GetMerchantID(c)
	result, err := h.uc.GetStaticQR(c.Request().Context(), merchantID)
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, result)
}

func (h *Handler) UpdateStaticQR(c echo.Context) error {
	var req struct {
		PaymentNo string `json:"paymentNo"`
		QrString  string `json:"qrString"`
		QrImage   string `json:"qrImage"`
	}
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, domain.ErrInvalidPayment)
	}
	merchantID := httputil.GetMerchantID(c)
	qr := &domain.StaticQR{
		MerchantID: merchantID,
		PaymentNo:  req.PaymentNo,
		QrString:   req.QrString,
		QrImage:    req.QrImage,
	}
	if err := h.uc.UpdateStaticQR(c.Request().Context(), qr); err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, qr)
}

func mapError(c echo.Context, err error) error {
	switch err {
	case domain.ErrInvalidPayment:
		return httputil.WriteError(c, http.StatusBadRequest, err)
	case domain.ErrOrderNotPayable, domain.ErrAlreadyPaid:
		return httputil.WriteError(c, http.StatusConflict, err)
	case domain.ErrGatewayNotConfigured:
		return httputil.WriteError(c, http.StatusServiceUnavailable, err)
	case domain.ErrGatewayUnavailable, domain.ErrGatewayError, domain.ErrOrderClientUnavailable:
		return httputil.WriteError(c, http.StatusBadGateway, err)
	case domain.ErrPaymentNotFound:
		return httputil.WriteError(c, http.StatusNotFound, err)
	default:
		logger.Error("payment handler error", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
}
