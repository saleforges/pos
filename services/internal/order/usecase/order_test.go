package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/saleforge/pos/services/internal/order/domain"
)

func TestOrderUsecase_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("success with totals", func(t *testing.T) {
		uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo())

		customerID := int64(1)
		order, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1,
			BranchID:   1,
			CreatedBy:  5,
			CustomerID: &customerID,
			Items: []CreateOrderItemParams{
				{ProductItemID: 35, ItemName: "Es Teh Manis - Large", UnitPrice: 15000, Quantity: 2},
				{ProductItemID: 36, ItemName: "Gula Pasir 250g", UnitPrice: 8000, Quantity: 1},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.ID == 0 {
			t.Error("expected non-zero id")
		}
		if order.Subtotal != 38000 {
			t.Errorf("expected subtotal 38000, got %f", order.Subtotal)
		}
		if order.Total != 38000 {
			t.Errorf("expected total 38000, got %f", order.Total)
		}
		if order.Status != domain.OrderStatusCompleted {
			t.Errorf("expected status completed, got %s", order.Status)
		}
		if order.PaymentStatus != domain.PaymentStatusUnpaid {
			t.Errorf("expected paymentStatus unpaid, got %s", order.PaymentStatus)
		}
		if order.CreatedBy != 5 {
			t.Errorf("expected createdBy 5, got %d", order.CreatedBy)
		}
		if len(order.Items) != 2 {
			t.Errorf("expected 2 items, got %d", len(order.Items))
		}
		if order.Items[0].LineTotal != 30000 {
			t.Errorf("expected lineTotal 30000, got %f", order.Items[0].LineTotal)
		}
	})

	t.Run("empty items returns error", func(t *testing.T) {
		uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo())
		_, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5, Items: []CreateOrderItemParams{},
		})
		if err != domain.ErrInvalidOrder {
			t.Errorf("expected ErrInvalidOrder, got %v", err)
		}
	})

	t.Run("missing branch returns error", func(t *testing.T) {
		uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo())
		_, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 1000, Quantity: 1}},
		})
		if err != domain.ErrInvalidOrder {
			t.Errorf("expected ErrInvalidOrder, got %v", err)
		}
	})

	t.Run("zero quantity item returns error", func(t *testing.T) {
		uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo())
		_, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 1000, Quantity: 0}},
		})
		if err != domain.ErrInvalidOrderItem {
			t.Errorf("expected ErrInvalidOrderItem, got %v", err)
		}
	})

	t.Run("customer from another merchant rejected", func(t *testing.T) {
		repo := &mockCustomerRepo{}
		repo.Create(ctx, &domain.Customer{MerchantID: 2, Name: "Lain"})
		uc := NewOrderUsecase(&mockOrderRepo{}, repo)

		customerID := int64(1)
		_, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5, CustomerID: &customerID,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 1000, Quantity: 1}},
		})
		if err != domain.ErrCustomerNotFound {
			t.Errorf("expected ErrCustomerNotFound for cross-merchant customer, got %v", err)
		}
	})

	t.Run("credit sale without due date defaults to +7 days", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo())

		customerID := int64(1)
		order, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5, CustomerID: &customerID,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 1000, Quantity: 1}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.DueDate == nil {
			t.Fatal("expected default due date for credit sale")
		}
		expected := time.Now().UTC().AddDate(0, 0, 7)
		if order.DueDate.After(expected.Add(time.Hour)) || order.DueDate.Before(expected.Add(-time.Hour)) {
			t.Errorf("expected due date ~+7 days, got %v", order.DueDate)
		}
	})

	t.Run("walk-in sale without due date stays nil", func(t *testing.T) {
		uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo())
		order, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 1000, Quantity: 1}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.DueDate != nil {
			t.Errorf("expected nil due date for walk-in sale, got %v", order.DueDate)
		}
	})

	t.Run("explicit due date preserved", func(t *testing.T) {
		uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo())
		due := time.Now().UTC().AddDate(0, 0, 30)
		customerID := int64(1)
		order, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5, CustomerID: &customerID, DueDate: &due,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 1000, Quantity: 1}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !order.DueDate.Equal(due) {
			t.Errorf("expected explicit due date preserved, got %v", order.DueDate)
		}
	})
}

func TestOrderUsecase_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("updates due date and note", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo())
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 1000, Quantity: 1}},
		})

		due := time.Now().UTC().AddDate(0, 0, 14)
		note := "Perpanjang"
		order, err := uc.Update(ctx, UpdateOrderParams{ID: 1, MerchantID: 1, DueDate: &due, Note: &note})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !order.DueDate.Equal(due) {
			t.Errorf("expected due date %v, got %v", due, order.DueDate)
		}
		if order.Note != "Perpanjang" {
			t.Errorf("expected note 'Perpanjang', got '%s'", order.Note)
		}
	})

	t.Run("partial update keeps other fields", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo())
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5, Note: "Asli",
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 1000, Quantity: 1}},
		})

		note := "Ganti"
		order, err := uc.Update(ctx, UpdateOrderParams{ID: 1, MerchantID: 1, Note: &note})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.Note != "Ganti" {
			t.Errorf("expected note 'Ganti', got '%s'", order.Note)
		}
		if order.DueDate != nil {
			t.Errorf("expected due date untouched, got %v", order.DueDate)
		}
	})

	t.Run("cancelled order cannot update", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo())
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 1000, Quantity: 1}},
		})
		uc.Cancel(ctx, 1, 1)

		note := "Nope"
		_, err := uc.Update(ctx, UpdateOrderParams{ID: 1, MerchantID: 1, Note: &note})
		if err != domain.ErrInvalidTransition {
			t.Errorf("expected ErrInvalidTransition for cancelled order, got %v", err)
		}
	})
}

func TestOrderUsecase_GetByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("returns order", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo())
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 1000, Quantity: 1}},
		})

		order, err := uc.GetByID(ctx, 1, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.ID != 1 {
			t.Errorf("expected id 1, got %d", order.ID)
		}
	})

	t.Run("cross-merchant returns not found", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo())
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 1000, Quantity: 1}},
		})

		_, err := uc.GetByID(ctx, 1, 2)
		if err != domain.ErrOrderNotFound {
			t.Errorf("expected ErrOrderNotFound, got %v", err)
		}
	})
}

func TestOrderUsecase_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("filters by payment status", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo())
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 10000, Quantity: 1}},
		})

		// pay half -> partial
		uc.AddPayment(ctx, AddPaymentParams{OrderID: 1, MerchantID: 1, CreatedBy: 5, Amount: 5000, Method: domain.PaymentMethodCash})

		paid := domain.PaymentStatusPartial
		orders, err := uc.List(ctx, 1, nil, nil, &paid)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(orders) != 1 {
			t.Errorf("expected 1 partial order, got %d", len(orders))
		}
	})
}

func TestOrderUsecase_Cancel(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("cancel completed order", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo())
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 1000, Quantity: 1}},
		})

		order, err := uc.Cancel(ctx, 1, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.Status != domain.OrderStatusCancelled {
			t.Errorf("expected cancelled, got %s", order.Status)
		}
	})

	t.Run("cancel non-existent returns not found", func(t *testing.T) {
		uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo())
		_, err := uc.Cancel(ctx, 999, 1)
		if err != domain.ErrOrderNotFound {
			t.Errorf("expected ErrOrderNotFound, got %v", err)
		}
	})
}

func TestOrderUsecase_AddPayment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("full payment marks paid", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo())
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 10000, Quantity: 1}},
		})

		order, err := uc.AddPayment(ctx, AddPaymentParams{OrderID: 1, MerchantID: 1, CreatedBy: 5, Amount: 10000, Method: domain.PaymentMethodCash})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.PaymentStatus != domain.PaymentStatusPaid {
			t.Errorf("expected paid, got %s", order.PaymentStatus)
		}
		if order.PaidAmount != 10000 {
			t.Errorf("expected paidAmount 10000, got %f", order.PaidAmount)
		}
	})

	t.Run("partial payment marks partial", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo())
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 10000, Quantity: 1}},
		})

		order, err := uc.AddPayment(ctx, AddPaymentParams{OrderID: 1, MerchantID: 1, CreatedBy: 5, Amount: 4000, Method: domain.PaymentMethodCash})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.PaymentStatus != domain.PaymentStatusPartial {
			t.Errorf("expected partial, got %s", order.PaymentStatus)
		}
	})

	t.Run("accumulates partial payments", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo())
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 10000, Quantity: 1}},
		})

		uc.AddPayment(ctx, AddPaymentParams{OrderID: 1, MerchantID: 1, CreatedBy: 5, Amount: 4000, Method: domain.PaymentMethodCash})
		order, err := uc.AddPayment(ctx, AddPaymentParams{OrderID: 1, MerchantID: 1, CreatedBy: 5, Amount: 6000, Method: domain.PaymentMethodQRIS})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.PaymentStatus != domain.PaymentStatusPaid {
			t.Errorf("expected paid after second payment, got %s", order.PaymentStatus)
		}
		if len(order.Payments) != 2 {
			t.Errorf("expected 2 payment records, got %d", len(order.Payments))
		}
	})

	t.Run("payment exceeding balance rejected", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo())
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 10000, Quantity: 1}},
		})

		_, err := uc.AddPayment(ctx, AddPaymentParams{OrderID: 1, MerchantID: 1, CreatedBy: 5, Amount: 15000, Method: domain.PaymentMethodCash})
		if err != domain.ErrPaymentExceedsTotal {
			t.Errorf("expected ErrPaymentExceedsTotal, got %v", err)
		}
	})

	t.Run("zero amount rejected", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo())
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 10000, Quantity: 1}},
		})

		_, err := uc.AddPayment(ctx, AddPaymentParams{OrderID: 1, MerchantID: 1, CreatedBy: 5, Amount: 0, Method: domain.PaymentMethodCash})
		if err != domain.ErrInvalidPayment {
			t.Errorf("expected ErrInvalidPayment, got %v", err)
		}
	})

	t.Run("invalid method rejected", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo())
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 10000, Quantity: 1}},
		})

		_, err := uc.AddPayment(ctx, AddPaymentParams{OrderID: 1, MerchantID: 1, CreatedBy: 5, Amount: 5000, Method: "bitcoin"})
		if err != domain.ErrInvalidPayment {
			t.Errorf("expected ErrInvalidPayment, got %v", err)
		}
	})
}
