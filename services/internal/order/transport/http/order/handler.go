package order

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/order/domain"
	"github.com/saleforge/pos/services/internal/order/transport/http/middleware"
	"github.com/saleforge/pos/services/internal/order/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/logger"
)

type Handler struct {
	uc usecase.OrderUsecase
}

func NewHandler(uc usecase.OrderUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Create(c echo.Context) error {
	var req createOrderReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	dueDate, err := parseDueDate(req.DueDate)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	items := make([]usecase.CreateOrderItemParams, len(req.Items))
	for i, it := range req.Items {
		items[i] = usecase.CreateOrderItemParams{
			ProductItemID: it.ProductItemID,
			ItemName:      it.ItemName,
			UnitPrice:     it.UnitPrice,
			Quantity:      it.Quantity,
		}
	}

	merchantID := httputil.GetMerchantID(c)
	userID := getUserID(c)

	result, err := h.uc.Create(c.Request().Context(), usecase.CreateOrderParams{
		MerchantID: merchantID,
		BranchID:   req.BranchID,
		CreatedBy:  userID,
		CustomerID: req.CustomerID,
		DueDate:    dueDate,
		Note:       req.Note,
		Items:      items,
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

	var branchID *int64
	if v := c.QueryParam("branchId"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
		}
		branchID = &id
	}

	var status *domain.OrderStatus
	if v := c.QueryParam("status"); v != "" {
		s := domain.OrderStatus(v)
		status = &s
	}

	var paymentStatus *domain.PaymentStatus
	if v := c.QueryParam("paymentStatus"); v != "" {
		p := domain.PaymentStatus(v)
		paymentStatus = &p
	}

	items, err := h.uc.List(c.Request().Context(), merchantID, branchID, status, paymentStatus)
	if err != nil {
		logger.Error("order.List failed", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
	return httputil.WriteJSON(c, http.StatusOK, map[string]interface{}{
		"data": items,
	})
}

func (h *Handler) Cancel(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	var req cancelOrderReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	if req.Status != string(domain.OrderStatusCancelled) {
		return httputil.WriteError(c, http.StatusBadRequest, domain.ErrInvalidTransition)
	}

	merchantID := httputil.GetMerchantID(c)
	result, err := h.uc.Cancel(c.Request().Context(), id, merchantID)
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

	var req updateOrderReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		t, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
		}
		dueDate = &t
	}

	merchantID := httputil.GetMerchantID(c)
	result, err := h.uc.Update(c.Request().Context(), usecase.UpdateOrderParams{
		ID:         id,
		MerchantID: merchantID,
		DueDate:    dueDate,
		Note:       req.Note,
	})
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, result)
}

func (h *Handler) AddPayment(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	var req addPaymentReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	paidAt := time.Now().UTC()
	if req.PaidAt != nil {
		t, err := time.Parse(time.RFC3339, *req.PaidAt)
		if err != nil {
			return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
		}
		paidAt = t
	}

	merchantID := httputil.GetMerchantID(c)
	userID := getUserID(c)

	result, err := h.uc.AddPayment(c.Request().Context(), usecase.AddPaymentParams{
		OrderID:    id,
		MerchantID: merchantID,
		CreatedBy:  userID,
		Amount:     req.Amount,
		Method:     domain.PaymentMethod(req.Method),
		PaidAt:     paidAt,
	})
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, result)
}

func getUserID(c echo.Context) int64 {
	id, _ := c.Get(middleware.ContextKeyUserID).(int64)
	return id
}

func mapError(c echo.Context, err error) error {
	switch err {
	case domain.ErrOrderNotFound, domain.ErrCustomerNotFound:
		return httputil.WriteError(c, http.StatusNotFound, err)
	case domain.ErrInvalidOrder, domain.ErrInvalidOrderItem, domain.ErrInvalidPayment:
		return httputil.WriteError(c, http.StatusBadRequest, err)
	case domain.ErrInvalidTransition:
		return httputil.WriteError(c, http.StatusConflict, err)
	case domain.ErrPaymentExceedsTotal, domain.ErrInsufficientStock:
		return httputil.WriteError(c, http.StatusBadRequest, err)
	case domain.ErrOrderStockNotFound:
		return httputil.WriteError(c, http.StatusNotFound, err)
	default:
		logger.Error("order handler error", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
}
