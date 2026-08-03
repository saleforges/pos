package usecase

import (
	"context"
	"strconv"
	"time"

	"github.com/saleforge/pos/services/internal/payment/domain"
	"github.com/saleforge/pos/services/internal/payment/port/repository"
)

type paymentUsecase struct {
	paymentRepo repository.PaymentRepository
	gateway     repository.GatewayClient
	orderClient repository.OrderClient
}

func NewPaymentUsecase(paymentRepo repository.PaymentRepository, gateway repository.GatewayClient, orderClient repository.OrderClient) PaymentUsecase {
	return &paymentUsecase{paymentRepo: paymentRepo, gateway: gateway, orderClient: orderClient}
}

// Create opens a gateway payment for an order and records the transaction.
func (uc *paymentUsecase) Create(ctx context.Context, params CreatePaymentParams) (*domain.PaymentTransaction, error) {
	now := time.Now().UTC()
	payment := &domain.PaymentTransaction{
		MerchantID: params.MerchantID,
		OrderID:    params.OrderID,
		Gateway:    "ipaymu",
		Status:     domain.PaymentStatusPending,
		Amount:     params.Amount,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := payment.Validate(); err != nil {
		return nil, err
	}
	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	items := make([]repository.CreatePaymentItem, len(params.Items))
	for i, it := range params.Items {
		items[i] = repository.CreatePaymentItem{ItemName: it.ItemName, Quantity: it.Quantity, UnitPrice: it.UnitPrice}
	}

	result, err := uc.gateway.CreatePayment(ctx, repository.CreatePaymentParams{
		ReferenceID: strconv.FormatInt(params.OrderID, 10),
		BuyerName:   params.BuyerName,
		BuyerEmail:  params.BuyerEmail,
		BuyerPhone:  params.BuyerPhone,
		Items:       items,
	})
	if err != nil {
		return nil, err
	}

	// Persist the gateway payment URL back onto the transaction.
	if err := uc.paymentRepo.UpdatePaymentURL(ctx, payment.ID, result.PaymentURL, result.SessionID); err != nil {
		return nil, err
	}
	payment.PaymentURL = result.PaymentURL
	payment.SessionID = result.SessionID
	return payment, nil
}

// HandleCallback processes a gateway webhook. Idempotent: duplicates are
// detected via gateway_ref uniqueness; already-settled payments are no-ops.
func (uc *paymentUsecase) HandleCallback(ctx context.Context, params CallbackParams) error {
	// Locate the transaction by order reference (reference_id = order id).
	orderID, err := strconv.ParseInt(params.ReferenceID, 10, 64)
	if err != nil {
		return domain.ErrInvalidCallback
	}

	payment, err := uc.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return domain.ErrPaymentNotFound
	}

	// Only successful payments settle the order.
	if params.StatusCode != 1 {
		return nil
	}

	// Idempotency: gateway retries callbacks until 200. If we already
	// recorded this gateway ref, it's a duplicate — no-op.
	if payment.GatewayRef != "" && payment.GatewayRef == params.GatewayRef && payment.Status == domain.PaymentStatusPaid {
		return nil
	}

	amount, err := strconv.ParseFloat(params.Amount, 64)
	if err != nil || amount <= 0 {
		return domain.ErrInvalidCallback
	}

	if err := uc.paymentRepo.MarkPaid(ctx, payment.ID, params.GatewayRef); err != nil {
		return err
	}

	return uc.orderClient.NotifyPaid(ctx, orderID, payment.MerchantID, amount, gatewayMethod(params.Via, params.Channel))
}

// gatewayMethod maps gateway 'via' values to our payment method enum.
func gatewayMethod(via, channel string) string {
	switch via {
	case "qris":
		return "qris"
	case "cod":
		return "cash"
	case "va", "ew", "ewallet", "e-wallet":
		return "transfer"
	default:
		return "transfer"
	}
}

func (uc *paymentUsecase) GetByOrderID(ctx context.Context, orderID int64, merchantID int64) (*domain.PaymentTransaction, error) {
	payment, err := uc.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if payment.MerchantID != merchantID {
		return nil, domain.ErrPaymentNotFound
	}
	return payment, nil
}

var _ PaymentUsecase = (*paymentUsecase)(nil)
