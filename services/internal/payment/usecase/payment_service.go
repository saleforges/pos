package usecase

import (
	"context"
	"fmt"
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
	// Fetch the order snapshot to validate state and size the payment.
	order, err := uc.orderClient.GetOrder(ctx, params.OrderID, params.MerchantID)
	if err != nil {
		return nil, err
	}
	if order.Status != "completed" {
		return nil, domain.ErrOrderNotPayable
	}
	remaining := order.Total - order.PaidAmount
	if remaining <= 0 {
		return nil, domain.ErrAlreadyPaid
	}

	now := time.Now().UTC()
	payment := &domain.PaymentTransaction{
		MerchantID: params.MerchantID,
		BranchID:   order.BranchID,
		OrderID:    params.OrderID,
		Gateway:    "ipaymu",
		Status:     domain.PaymentStatusPending,
		Amount:     remaining,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := payment.Validate(); err != nil {
		return nil, err
	}
	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	items := make([]repository.CreatePaymentItem, len(order.Items))
	for i, it := range order.Items {
		items[i] = repository.CreatePaymentItem{ItemName: it.ItemName, Quantity: it.Quantity, UnitPrice: it.UnitPrice}
	}

	result, err := uc.gateway.CreatePayment(ctx, repository.CreatePaymentParams{
		ReferenceID: strconv.FormatInt(params.OrderID, 10),
		Method:      params.Method,
		BuyerName:   params.BuyerName,
		BuyerEmail:  params.BuyerEmail,
		BuyerPhone:  params.BuyerPhone,
		Items:       items,
	})
	if err != nil {
		return nil, err
	}

	payment.PaymentURL = result.PaymentURL
	payment.PaymentNo = result.PaymentNo
	payment.QrString = result.QrString
	payment.QrImage = result.QrImage
	payment.ExpiredAt = result.ExpiredAt
	payment.SessionID = result.SessionID
	payment.PaymentRef = result.PaymentRef
	payment.UpdatedAt = time.Now().UTC()
	if err := uc.paymentRepo.UpdateDetails(ctx, payment.ID, payment); err != nil {
		return nil, err
	}
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
	// recorded this payment ref, it's a duplicate — no-op.
	if payment.PaymentRef != "" && payment.PaymentRef == params.PaymentRef && payment.Status == domain.PaymentStatusPaid {
		return nil
	}

	amount, err := strconv.ParseFloat(params.Amount, 64)
	if err != nil || amount <= 0 {
		amount = payment.Amount
	}
	if amount <= 0 {
		return domain.ErrInvalidCallback
	}

	if err := uc.paymentRepo.MarkPaid(ctx, payment.ID, params.PaymentRef); err != nil {
		return err
	}

	return uc.orderClient.NotifyPaid(ctx, orderID, payment.MerchantID, amount, gatewayMethod(params.Via, params.Channel))
}

func (uc *paymentUsecase) ConfirmCash(ctx context.Context, params CreatePaymentParams) (*domain.PaymentTransaction, error) {
	order, err := uc.orderClient.GetOrder(ctx, params.OrderID, params.MerchantID)
	if err != nil {
		return nil, err
	}
	if order.Status != "completed" {
		return nil, domain.ErrOrderNotPayable
	}
	remaining := order.Total - order.PaidAmount
	if remaining <= 0 {
		return nil, domain.ErrAlreadyPaid
	}
	amount := float64(params.Amount)
	if amount <= 0 || amount > remaining {
		amount = remaining
	}

	now := time.Now().UTC()
	payment := &domain.PaymentTransaction{
		MerchantID: params.MerchantID,
		BranchID:   order.BranchID,
		OrderID:    params.OrderID,
		Gateway:    "cash",
		Status:     domain.PaymentStatusPaid,
		Amount:     amount,
		PaymentRef: fmt.Sprintf("cash-%d", now.Unix()),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := payment.Validate(); err != nil {
		return nil, err
	}
	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}
	if err := uc.orderClient.NotifyPaid(ctx, order.ID, params.MerchantID, amount, "cash"); err != nil {
		return nil, err
	}
	return payment, nil
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

func (uc *paymentUsecase) GetByID(ctx context.Context, id int64, merchantID int64) (*domain.PaymentTransaction, error) {
	payment, err := uc.paymentRepo.GetByID(ctx, id, merchantID)
	if err != nil {
		return nil, err
	}
	if payment.Status == domain.PaymentStatusPending && payment.ExpiredAt != "" {
		// iPaymu sends expiry in WIB (UTC+7); parse as fixed zone so the
		// comparison is correct regardless of server timezone.
		jakarta := time.FixedZone("WIB", 7*60*60)
		if t, perr := time.ParseInLocation("2006-01-02 15:04:05", payment.ExpiredAt, jakarta); perr == nil && time.Now().After(t) {
			_ = uc.paymentRepo.MarkExpired(ctx, payment.ID)
			payment.Status = domain.PaymentStatusExpired
		}
	}
	return payment, nil
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

func (uc *paymentUsecase) GetStaticQR(ctx context.Context, merchantID, branchID int64) (*domain.StaticQR, error) {
	return uc.paymentRepo.GetStaticQR(ctx, merchantID, branchID)
}

func (uc *paymentUsecase) UpdateStaticQR(ctx context.Context, qr *domain.StaticQR) error {
	if qr == nil || qr.MerchantID == 0 || qr.QrImage == "" {
		return domain.ErrInvalidPayment
	}
	return uc.paymentRepo.UpsertStaticQR(ctx, qr)
}

func (uc *paymentUsecase) Sync(ctx context.Context, merchantID, branchID int64, lastSync *time.Time) (*PaymentSyncResult, error) {
	payments, err := uc.paymentRepo.SyncByBranch(ctx, merchantID, branchID, lastSync)
	if err != nil {
		return nil, err
	}
	if payments == nil {
		payments = []domain.PaymentTransaction{}
	}
	return &PaymentSyncResult{
		Payments:  payments,
		SyncToken: time.Now().UTC().Format(time.RFC3339Nano),
	}, nil
}

var _ PaymentUsecase = (*paymentUsecase)(nil)
