package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/saleforge/pos/services/internal/payment/domain"
	"github.com/saleforge/pos/services/internal/payment/port/repository"
)

func TestPaymentUsecase_GetByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := newMockPaymentRepo()
	uc := NewPaymentUsecase(repo, &mockGateway{}, &mockOrderClient{})
	uc.Create(ctx, CreatePaymentParams{
		MerchantID: 1, OrderID: 5,
	})

	t.Run("returns own payment", func(t *testing.T) {
		p, err := uc.GetByID(ctx, 1, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.OrderID != 5 {
			t.Errorf("expected order 5, got %d", p.OrderID)
		}
	})

	t.Run("other merchant cannot read", func(t *testing.T) {
		_, err := uc.GetByID(ctx, 1, 99)
		if err != domain.ErrPaymentNotFound {
			t.Errorf("expected ErrPaymentNotFound, got %v", err)
		}
	})

	t.Run("missing payment", func(t *testing.T) {
		_, err := uc.GetByID(ctx, 999, 1)
		if err != domain.ErrPaymentNotFound {
			t.Errorf("expected ErrPaymentNotFound, got %v", err)
		}
	})
}

func TestPaymentUsecase_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("creates payment with gateway URL", func(t *testing.T) {
		repo := newMockPaymentRepo()
		gw := &mockGateway{}
		uc := NewPaymentUsecase(repo, gw, &mockOrderClient{})

		result, err := uc.Create(ctx, CreatePaymentParams{
			MerchantID: 1, OrderID: 5,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !gw.called {
			t.Error("expected gateway to be called")
		}
		if result.PaymentURL == "" {
			t.Error("expected payment url")
		}
		if result.Status != domain.PaymentStatusPending {
			t.Errorf("expected pending status, got %s", result.Status)
		}
	})

	t.Run("gateway not configured", func(t *testing.T) {
		uc := NewPaymentUsecase(newMockPaymentRepo(), &mockGateway{err: domain.ErrGatewayNotConfigured}, &mockOrderClient{})
		_, err := uc.Create(ctx, CreatePaymentParams{
			MerchantID: 1, OrderID: 5,
		})
		if err != domain.ErrGatewayNotConfigured {
			t.Errorf("expected ErrGatewayNotConfigured, got %v", err)
		}
	})

	t.Run("invalid amount rejected", func(t *testing.T) {
		uc := NewPaymentUsecase(newMockPaymentRepo(), &mockGateway{}, &mockOrderClient{})
		_, err := uc.Create(ctx, CreatePaymentParams{MerchantID: 1, OrderID: 5})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("cancelled order rejected", func(t *testing.T) {
		uc := NewPaymentUsecase(newMockPaymentRepo(), &mockGateway{}, &mockOrderClient{
			order: &repository.OrderInfo{ID: 5, MerchantID: 1, Status: "cancelled", Total: 30000},
		})
		_, err := uc.Create(ctx, CreatePaymentParams{MerchantID: 1, OrderID: 5})
		if err != domain.ErrOrderNotPayable {
			t.Errorf("expected ErrOrderNotPayable, got %v", err)
		}
	})

	t.Run("already paid order rejected", func(t *testing.T) {
		uc := NewPaymentUsecase(newMockPaymentRepo(), &mockGateway{}, &mockOrderClient{
			order: &repository.OrderInfo{ID: 5, MerchantID: 1, Status: "completed", Total: 30000, PaidAmount: 30000},
		})
		_, err := uc.Create(ctx, CreatePaymentParams{MerchantID: 1, OrderID: 5})
		if err != domain.ErrAlreadyPaid {
			t.Errorf("expected ErrAlreadyPaid, got %v", err)
		}
	})

	t.Run("amount is remaining balance", func(t *testing.T) {
		repo := newMockPaymentRepo()
		uc := NewPaymentUsecase(repo, &mockGateway{}, &mockOrderClient{
			order: &repository.OrderInfo{ID: 5, MerchantID: 1, Status: "completed", Total: 30000, PaidAmount: 10000},
		})
		result, err := uc.Create(ctx, CreatePaymentParams{MerchantID: 1, OrderID: 5})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Amount != 20000 {
			t.Errorf("expected amount 20000 (remaining), got %v", result.Amount)
		}
	})
}

func TestPaymentUsecase_HandleCallback(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	newSetup := func() (PaymentUsecase, *mockPaymentRepo, *mockOrderClient) {
		repo := newMockPaymentRepo()
		orderClient := &mockOrderClient{}
		uc := NewPaymentUsecase(repo, &mockGateway{}, orderClient)
		uc.Create(ctx, CreatePaymentParams{
			MerchantID: 1, OrderID: 5,
		})
		return uc, repo, orderClient
	}

	t.Run("successful callback notifies order", func(t *testing.T) {
		uc, _, orderClient := newSetup()
		err := uc.HandleCallback(ctx, CallbackParams{
			ReferenceID: "5", Status: "berhasil", StatusCode: 1,
			Amount: "30000", Via: "qris", PaymentRef: "TRX123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !orderClient.notified {
			t.Error("expected order service to be notified")
		}
	})

	t.Run("duplicate callback is idempotent", func(t *testing.T) {
		uc, repo, orderClient := newSetup()
		err := uc.HandleCallback(ctx, CallbackParams{
			ReferenceID: "5", StatusCode: 1, Amount: "30000", Via: "qris", PaymentRef: "TRX123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		notifiedCount := 1
		if orderClient.notified {
			notifiedCount++
		}
		// second callback with same gateway ref → no-op
		err = uc.HandleCallback(ctx, CallbackParams{
			ReferenceID: "5", StatusCode: 1, Amount: "30000", Via: "qris", PaymentRef: "TRX123",
		})
		if err != nil {
			t.Fatalf("unexpected error on duplicate: %v", err)
		}
		_ = notifiedCount
		payments, _ := repo.GetByOrderID(ctx, 5)
		if payments == nil || payments.Status != domain.PaymentStatusPaid {
			t.Errorf("expected payment paid, got %+v", payments)
		}
	})

	t.Run("pending callback ignored", func(t *testing.T) {
		uc, _, orderClient := newSetup()
		err := uc.HandleCallback(ctx, CallbackParams{
			ReferenceID: "5", StatusCode: 0, Amount: "30000", Via: "qris",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if orderClient.notified {
			t.Error("pending callback should not notify order")
		}
	})

	t.Run("unknown order rejected", func(t *testing.T) {
		uc, _, _ := newSetup()
		err := uc.HandleCallback(ctx, CallbackParams{
			ReferenceID: "999", StatusCode: 1, Amount: "1000",
		})
		if err != domain.ErrPaymentNotFound {
			t.Errorf("expected ErrPaymentNotFound, got %v", err)
		}
	})
}

func TestPaymentUsecase_StaticQR(t *testing.T) {
	ctx := context.Background()
	repo := newMockPaymentRepo()
	uc := NewPaymentUsecase(repo, &mockGateway{}, &mockOrderClient{})

	qr := &domain.StaticQR{MerchantID: 1, PaymentNo: "STATIC-1", QrString: "QR", QrImage: "https://img/qr.png"}
	if err := uc.UpdateStaticQR(ctx, qr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := uc.GetStaticQR(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.QrImage != qr.QrImage || got.PaymentNo != qr.PaymentNo {
		t.Errorf("mismatch: got %+v want %+v", got, qr)
	}

	if _, err := uc.GetStaticQR(ctx, 99); err != domain.ErrPaymentNotFound {
		t.Errorf("expected not found, got %v", err)
	}

	if err := uc.UpdateStaticQR(ctx, &domain.StaticQR{MerchantID: 1}); err != domain.ErrInvalidPayment {
		t.Errorf("expected invalid, got %v", err)
	}
}

func TestPaymentUsecase_Sync(t *testing.T) {
	ctx := context.Background()
	repo := newMockPaymentRepo()
	uc := NewPaymentUsecase(repo, &mockGateway{}, &mockOrderClient{})

	repo.Create(ctx, &domain.PaymentTransaction{ID: 1, MerchantID: 1, OrderID: 10, Gateway: "ipaymu", Status: domain.PaymentStatusPending, Amount: 15000, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	repo.Create(ctx, &domain.PaymentTransaction{ID: 2, MerchantID: 2, OrderID: 20, Gateway: "ipaymu", Status: domain.PaymentStatusPaid, Amount: 5000, CreatedAt: time.Now(), UpdatedAt: time.Now()})

	res, err := uc.Sync(ctx, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Payments) != 1 || res.Payments[0].OrderID != 10 {
		t.Errorf("expected 1 payment for merchant 1, got %+v", res.Payments)
	}
	if res.SyncToken == "" {
		t.Error("expected sync token")
	}

	future := time.Now().Add(time.Hour)
	res2, err := uc.Sync(ctx, 1, &future)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res2.Payments) != 0 {
		t.Errorf("expected 0 payments after future lastSync, got %d", len(res2.Payments))
	}
}

func TestPaymentUsecase_GetByIDExpired(t *testing.T) {
	ctx := context.Background()
	repo := newMockPaymentRepo()
	uc := NewPaymentUsecase(repo, &mockGateway{}, &mockOrderClient{})

	repo.Create(ctx, &domain.PaymentTransaction{
		MerchantID: 1, OrderID: 30, Gateway: "ipaymu",
		Status: domain.PaymentStatusPending, Amount: 10000,
		ExpiredAt: time.Now().In(time.FixedZone("WIB", 7*60*60)).Add(-time.Hour).Format("2006-01-02 15:04:05"),
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	paymentID := repo.seq

	got, err := uc.GetByID(ctx, paymentID, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != domain.PaymentStatusExpired {
		t.Errorf("expected expired, got %s", got.Status)
	}

	fresh := repo.payments[paymentID]
	if fresh.Status != domain.PaymentStatusExpired {
		t.Error("expected repo state also expired")
	}
}
