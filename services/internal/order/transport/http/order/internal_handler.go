package order

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/order/domain"
	"github.com/saleforge/pos/services/internal/order/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
)

// InternalHandler exposes service-to-service endpoints (guarded by
// X-Internal-Key). Called by the payment service when a gateway payment
// succeeds.
type InternalHandler struct {
	uc usecase.OrderUsecase
}

func NewInternalHandler(uc usecase.OrderUsecase) *InternalHandler {
	return &InternalHandler{uc: uc}
}

type notifyPaidReq struct {
	Amount float64 `json:"amount"`
	Method string  `json:"method"`
}

// NotifyPaid records a gateway-confirmed payment on an order.
func (h *InternalHandler) NotifyPaid(c echo.Context) error {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	var req notifyPaidReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	if req.Amount <= 0 {
		return httputil.WriteError(c, http.StatusBadRequest, domain.ErrInvalidPayment)
	}

	// merchant id comes from the payment service (it owns the transaction).
	var merchantID int64
	if raw := c.Request().Header.Get("X-Merchant-Id"); raw != "" {
		merchantID, _ = strconv.ParseInt(raw, 10, 64)
	}
	if merchantID == 0 {
		// fallback: the internal API key is merchant-agnostic, so require
		// merchant in the header — payment service always sends it.
		return httputil.WriteError(c, http.StatusBadRequest, domain.ErrInvalidPayment)
	}

	err = h.uc.NotifyPaid(c.Request().Context(), usecase.NotifyPaidParams{
		OrderID:    orderID,
		MerchantID: merchantID,
		Amount:     req.Amount,
		Method:     req.Method,
	})
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, map[string]string{"message": "ok"})
}
