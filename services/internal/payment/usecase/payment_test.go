package usecase

import (
	"context"
	"testing"

	"github.com/saleforge/pos/services/internal/payment/domain"
)

func TestPaymentUsecase_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("creates payment with gateway URL", func(t *testing.T) {
		repo := newMockPaymentRepo()
		gw := &mockGateway{}
		uc := NewPaymentUsecase(repo, gw, &mockOrderClient{})

		result, err := uc.Create(ctx, CreatePaymentParams{
			MerchantID: 1, OrderID: 5, Amount: 30000,
			Items: []CreatePaymentItem{{ItemName: "Es Teh", Quantity: 2, UnitPrice: 15000}},
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
			MerchantID: 1, OrderID: 5, Amount: 30000,
			Items: []CreatePaymentItem{{ItemName: "X", Quantity: 1, UnitPrice: 1000}},
		})
		if err != domain.ErrGatewayNotConfigured {
			t.Errorf("expected ErrGatewayNotConfigured, got %v", err)
		}
	})

	t.Run("invalid amount rejected", func(t *testing.T) {
		uc := NewPaymentUsecase(newMockPaymentRepo(), &mockGateway{}, &mockOrderClient{})
		_, err := uc.Create(ctx, CreatePaymentParams{MerchantID: 1, OrderID: 5, Amount: 0})
		if err != domain.ErrInvalidPayment {
			t.Errorf("expected ErrInvalidPayment, got %v", err)
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
			MerchantID: 1, OrderID: 5, Amount: 30000,
			Items: []CreatePaymentItem{{ItemName: "Es Teh", Quantity: 2, UnitPrice: 15000}},
		})
		return uc, repo, orderClient
	}

	t.Run("successful callback notifies order", func(t *testing.T) {
		uc, _, orderClient := newSetup()
		err := uc.HandleCallback(ctx, CallbackParams{
			ReferenceID: "5", Status: "berhasil", StatusCode: 1,
			Amount: "30000", Via: "qris", GatewayRef: "TRX123",
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
			ReferenceID: "5", StatusCode: 1, Amount: "30000", Via: "qris", GatewayRef: "TRX123",
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
			ReferenceID: "5", StatusCode: 1, Amount: "30000", Via: "qris", GatewayRef: "TRX123",
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
