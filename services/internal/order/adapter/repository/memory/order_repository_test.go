package memory

import (
	"context"
	"testing"

	"github.com/saleforge/pos/services/internal/order/domain"
)

func TestOrderRepository(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("create and get by id", func(t *testing.T) {
		repo := NewOrderRepository()
		order := &domain.Order{
			MerchantID: 1, BranchID: 1, CreatedBy: 5, Status: domain.OrderStatusCompleted,
			Subtotal: 30000, Total: 30000,
			Items: []domain.OrderItem{
				{ProductItemID: 35, ItemName: "Es Teh", UnitPrice: 15000, Quantity: 2, LineTotal: 30000},
			},
		}
		if err := repo.Create(ctx, order); err != nil {
			t.Fatalf("Create: %v", err)
		}
		if order.ID == 0 {
			t.Error("expected non-zero id")
		}
		if order.Items[0].OrderID != order.ID {
			t.Errorf("expected item orderID %d, got %d", order.ID, order.Items[0].OrderID)
		}
		got, err := repo.GetByID(ctx, order.ID, 1)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Subtotal != 30000 {
			t.Errorf("expected subtotal 30000, got %f", got.Subtotal)
		}
	})

	t.Run("cross-merchant get returns not found", func(t *testing.T) {
		repo := NewOrderRepository()
		repo.Create(ctx, &domain.Order{MerchantID: 1, BranchID: 1, CreatedBy: 5, Status: domain.OrderStatusCompleted})
		_, err := repo.GetByID(ctx, 1, 2)
		if err != domain.ErrOrderNotFound {
			t.Errorf("expected ErrOrderNotFound, got %v", err)
		}
	})

	t.Run("list filters by payment status", func(t *testing.T) {
		repo := NewOrderRepository()
		repo.Create(ctx, &domain.Order{MerchantID: 1, BranchID: 1, CreatedBy: 5, Status: domain.OrderStatusCompleted, Subtotal: 10000, Total: 10000})
		repo.Create(ctx, &domain.Order{MerchantID: 1, BranchID: 1, CreatedBy: 5, Status: domain.OrderStatusCompleted, Subtotal: 20000, Total: 20000})
		repo.AddPayment(ctx, 1, 1, &domain.PaymentRecord{Amount: 5000, Method: domain.PaymentMethodCash, CreatedBy: 5})

		partial := domain.PaymentStatusPartial
		orders, err := repo.List(ctx, 1, nil, nil, &partial)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(orders) != 1 {
			t.Errorf("expected 1 partial order, got %d", len(orders))
		}
	})

	t.Run("add payment accumulates", func(t *testing.T) {
		repo := NewOrderRepository()
		repo.Create(ctx, &domain.Order{MerchantID: 1, BranchID: 1, CreatedBy: 5, Status: domain.OrderStatusCompleted, Subtotal: 10000, Total: 10000})
		repo.AddPayment(ctx, 1, 1, &domain.PaymentRecord{Amount: 4000, Method: domain.PaymentMethodCash, CreatedBy: 5})
		repo.AddPayment(ctx, 1, 1, &domain.PaymentRecord{Amount: 6000, Method: domain.PaymentMethodQRIS, CreatedBy: 5})

		got, _ := repo.GetByID(ctx, 1, 1)
		if got.PaidAmount != 10000 {
			t.Errorf("expected paidAmount 10000, got %f", got.PaidAmount)
		}
		if len(got.Payments) != 2 {
			t.Errorf("expected 2 payments, got %d", len(got.Payments))
		}
	})

	t.Run("update status", func(t *testing.T) {
		repo := NewOrderRepository()
		repo.Create(ctx, &domain.Order{MerchantID: 1, BranchID: 1, CreatedBy: 5, Status: domain.OrderStatusCompleted})
		got, err := repo.UpdateStatus(ctx, 1, 1, domain.OrderStatusCancelled)
		if err != nil {
			t.Fatalf("UpdateStatus: %v", err)
		}
		if got.Status != domain.OrderStatusCancelled {
			t.Errorf("expected cancelled, got %s", got.Status)
		}
	})
}
