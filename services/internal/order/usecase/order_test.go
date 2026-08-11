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
		uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo(), &mockInventoryClient{})

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

	t.Run("discount reduces total but not subtotal", func(t *testing.T) {
		uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo(), &mockInventoryClient{})
		order, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5, Discount: 5000,
			Items: []CreateOrderItemParams{{ProductItemID: 35, ItemName: "Es Teh", UnitPrice: 15000, Quantity: 2}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.Subtotal != 30000 {
			t.Errorf("expected subtotal 30000, got %f", order.Subtotal)
		}
		if order.Discount != 5000 {
			t.Errorf("expected discount 5000, got %f", order.Discount)
		}
		if order.Total != 25000 {
			t.Errorf("expected total 25000, got %f", order.Total)
		}
	})

	t.Run("negative discount rejected", func(t *testing.T) {
		uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo(), &mockInventoryClient{})
		_, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5, Discount: -1000,
			Items: []CreateOrderItemParams{{ProductItemID: 35, ItemName: "Es Teh", UnitPrice: 15000, Quantity: 1}},
		})
		if err != domain.ErrInvalidOrder {
			t.Errorf("expected ErrInvalidOrder, got %v", err)
		}
	})

	t.Run("discount exceeding subtotal rejected", func(t *testing.T) {
		uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo(), &mockInventoryClient{})
		_, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5, Discount: 999999,
			Items: []CreateOrderItemParams{{ProductItemID: 35, ItemName: "Es Teh", UnitPrice: 15000, Quantity: 1}},
		})
		if err != domain.ErrInvalidOrder {
			t.Errorf("expected ErrInvalidOrder, got %v", err)
		}
	})

	t.Run("empty items returns error", func(t *testing.T) {
		uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo(), &mockInventoryClient{})
		_, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5, Items: []CreateOrderItemParams{},
		})
		if err != domain.ErrInvalidOrder {
			t.Errorf("expected ErrInvalidOrder, got %v", err)
		}
	})

	t.Run("missing branch returns error", func(t *testing.T) {
		uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo(), &mockInventoryClient{})
		_, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 1000, Quantity: 1}},
		})
		if err != domain.ErrInvalidOrder {
			t.Errorf("expected ErrInvalidOrder, got %v", err)
		}
	})

	t.Run("zero quantity item returns error", func(t *testing.T) {
		uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo(), &mockInventoryClient{})
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
		uc := NewOrderUsecase(&mockOrderRepo{}, repo, &mockInventoryClient{})

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
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), &mockInventoryClient{})

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
		uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo(), &mockInventoryClient{})
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
		uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo(), &mockInventoryClient{})
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
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), &mockInventoryClient{})
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
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), &mockInventoryClient{})
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
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), &mockInventoryClient{})
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 1000, Quantity: 1}},
		})
		uc.Cancel(ctx, CancelOrderParams{OrderID: 1, MerchantID: 1})

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
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), &mockInventoryClient{})
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
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), &mockInventoryClient{})
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
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), &mockInventoryClient{})
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 10000, Quantity: 1}},
		})

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
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), &mockInventoryClient{})
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "X", UnitPrice: 1000, Quantity: 1}},
		})

		order, err := uc.Cancel(ctx, CancelOrderParams{OrderID: 1, MerchantID: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.Status != domain.OrderStatusCancelled {
			t.Errorf("expected cancelled, got %s", order.Status)
		}
	})

	t.Run("cancel non-existent returns not found", func(t *testing.T) {
		uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo(), &mockInventoryClient{})
		_, err := uc.Cancel(ctx, CancelOrderParams{OrderID: 999, MerchantID: 1})
		if err != domain.ErrOrderNotFound {
			t.Errorf("expected ErrOrderNotFound, got %v", err)
		}
	})

	t.Run("cancelling a paid order reverses the payment", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), &mockInventoryClient{})
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "Es Teh", UnitPrice: 15000, Quantity: 1}},
		})
		if _, err := uc.AddPayment(ctx, AddPaymentParams{OrderID: 1, MerchantID: 1, CreatedBy: 5, Amount: 15000, Method: domain.PaymentMethodCash}); err != nil {
			t.Fatalf("unexpected error paying: %v", err)
		}

		order, err := uc.Cancel(ctx, CancelOrderParams{OrderID: 1, MerchantID: 1, CancelledBy: 6})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.PaidAmount != 0 {
			t.Errorf("expected paidAmount reversed to 0, got %f", order.PaidAmount)
		}
		if order.PaymentStatus != domain.PaymentStatusUnpaid {
			t.Errorf("expected paymentStatus unpaid after reversal, got %s", order.PaymentStatus)
		}

		fresh, _ := uc.GetByID(ctx, 1, 1)
		var sawRefund bool
		for _, p := range fresh.Payments {
			if p.Method == "refund" && p.Amount == -15000 {
				sawRefund = true
			}
		}
		if !sawRefund {
			t.Errorf("expected a -15000 refund payment record, got %+v", fresh.Payments)
		}
	})

	t.Run("cancelling a partially paid order reverses only what was paid", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), &mockInventoryClient{})
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "Es Teh", UnitPrice: 30000, Quantity: 1}},
		})
		uc.AddPayment(ctx, AddPaymentParams{OrderID: 1, MerchantID: 1, CreatedBy: 5, Amount: 10000, Method: domain.PaymentMethodCash})

		order, err := uc.Cancel(ctx, CancelOrderParams{OrderID: 1, MerchantID: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.PaidAmount != 0 {
			t.Errorf("expected paidAmount reversed to 0, got %f", order.PaidAmount)
		}
	})

	t.Run("cancelling an unpaid order adds no refund record", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), &mockInventoryClient{})
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "Es Teh", UnitPrice: 15000, Quantity: 1}},
		})

		if _, err := uc.Cancel(ctx, CancelOrderParams{OrderID: 1, MerchantID: 1}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		fresh, _ := uc.GetByID(ctx, 1, 1)
		if len(fresh.Payments) != 0 {
			t.Errorf("expected no payment records for a never-paid order, got %+v", fresh.Payments)
		}
	})

	t.Run("cancelling an already-cancelled order is rejected", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), &mockInventoryClient{})
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 1, ItemName: "Es Teh", UnitPrice: 15000, Quantity: 1}},
		})
		uc.Cancel(ctx, CancelOrderParams{OrderID: 1, MerchantID: 1})

		_, err := uc.Cancel(ctx, CancelOrderParams{OrderID: 1, MerchantID: 1})
		if err != domain.ErrInvalidTransition {
			t.Errorf("expected ErrInvalidTransition, got %v", err)
		}
	})
}

func TestOrderUsecase_AddPayment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("full payment marks paid", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), &mockInventoryClient{})
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
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), &mockInventoryClient{})
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
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), &mockInventoryClient{})
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
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), &mockInventoryClient{})
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
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), &mockInventoryClient{})
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
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), &mockInventoryClient{})
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

func TestOrderUsecase_StockDeduction(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("create completed order deducts stock", func(t *testing.T) {
		inv := &mockInventoryClient{}
		uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo(), inv)

		_, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{
				{ProductItemID: 35, ItemName: "Es Teh", UnitPrice: 15000, Quantity: 2},
				{ProductItemID: 36, ItemName: "Gula", UnitPrice: 8000, Quantity: 1},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(inv.deducted) != 2 {
			t.Fatalf("expected 2 deducted items, got %d", len(inv.deducted))
		}
		if inv.deducted[0].ProductItemID != 35 || inv.deducted[0].Quantity != 2 {
			t.Errorf("unexpected first deducted item: %+v", inv.deducted[0])
		}
	})

	t.Run("insufficient stock cancels order and returns error", func(t *testing.T) {
		repo := &mockOrderRepo{}
		inv := &mockInventoryClient{deductErr: domain.ErrInsufficientStock}
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), inv)

		_, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 35, ItemName: "Es Teh", UnitPrice: 15000, Quantity: 2}},
		})
		if err != domain.ErrInsufficientStock {
			t.Fatalf("expected ErrInsufficientStock, got %v", err)
		}
		order, _ := repo.GetByID(ctx, 1, 1)
		if order == nil || order.Status != domain.OrderStatusCancelled {
			t.Errorf("expected order auto-cancelled, got %+v", order)
		}
	})

	t.Run("cancel restores stock", func(t *testing.T) {
		inv := &mockInventoryClient{}
		uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo(), inv)
		uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5,
			Items: []CreateOrderItemParams{{ProductItemID: 35, ItemName: "Es Teh", UnitPrice: 15000, Quantity: 2}},
		})

		_, err := uc.Cancel(ctx, CancelOrderParams{OrderID: 1, MerchantID: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(inv.restored) != 1 {
			t.Fatalf("expected 1 restored item, got %d", len(inv.restored))
		}
		if inv.restored[0].ProductItemID != 35 || inv.restored[0].Quantity != 2 {
			t.Errorf("unexpected restored item: %+v", inv.restored[0])
		}
	})
}

func TestOrderUsecase_Idempotency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	newUC := func() (OrderUsecase, *mockOrderRepo, *mockInventoryClient) {
		repo := &mockOrderRepo{}
		inv := &mockInventoryClient{}
		uc := NewOrderUsecase(repo, newMockCustomerRepo(), inv)
		return uc, repo, inv
	}

	t.Run("same clientOrderId returns existing order", func(t *testing.T) {
		uc, _, inv := newUC()

		first, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5, ClientOrderID: "uuid-abc-123",
			Items: []CreateOrderItemParams{{ProductItemID: 35, ItemName: "Es Teh", UnitPrice: 15000, Quantity: 2}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		second, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5, ClientOrderID: "uuid-abc-123",
			Items: []CreateOrderItemParams{{ProductItemID: 35, ItemName: "Es Teh", UnitPrice: 15000, Quantity: 2}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if second.ID != first.ID {
			t.Errorf("expected same order id %d, got %d", first.ID, second.ID)
		}
		if len(inv.deducted) != 1 {
			t.Errorf("expected 1 deduction, got %d", len(inv.deducted))
		}
	})

	t.Run("different clientOrderId creates new order", func(t *testing.T) {
		uc, _, _ := newUC()

		first, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5, ClientOrderID: "uuid-1",
			Items: []CreateOrderItemParams{{ProductItemID: 35, ItemName: "Es Teh", UnitPrice: 15000, Quantity: 1}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		second, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5, ClientOrderID: "uuid-2",
			Items: []CreateOrderItemParams{{ProductItemID: 35, ItemName: "Es Teh", UnitPrice: 15000, Quantity: 1}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if second.ID == first.ID {
			t.Errorf("expected different order ids, both %d", first.ID)
		}
	})

	t.Run("clientOrderId scoped to merchant", func(t *testing.T) {
		uc, _, _ := newUC()

		first, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 1, BranchID: 1, CreatedBy: 5, ClientOrderID: "uuid-shared",
			Items: []CreateOrderItemParams{{ProductItemID: 35, ItemName: "Es Teh", UnitPrice: 15000, Quantity: 1}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		second, err := uc.Create(ctx, CreateOrderParams{
			MerchantID: 2, BranchID: 1, CreatedBy: 5, ClientOrderID: "uuid-shared",
			Items: []CreateOrderItemParams{{ProductItemID: 35, ItemName: "Es Teh", UnitPrice: 15000, Quantity: 1}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if second.ID == first.ID {
			t.Errorf("expected different order ids across merchants")
		}
	})
}

func TestOrderUsecase_SyncOrders(t *testing.T) {
	ctx := context.Background()
	repo := &mockOrderRepo{}
	uc := NewOrderUsecase(repo, newMockCustomerRepo(), &mockInventoryClient{})

	order := &domain.Order{ID: 1, MerchantID: 1, BranchID: 1, Status: domain.OrderStatusCompleted, Total: 15000, PaidAmount: 15000, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := repo.Create(ctx, order); err != nil {
		t.Fatalf("create: %v", err)
	}

	res, err := uc.SyncOrders(ctx, 1, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(res.Orders))
	}
	if res.Orders[0].PaymentStatus != domain.PaymentStatusPaid {
		t.Errorf("expected paid status, got %s", res.Orders[0].PaymentStatus)
	}
	if res.SyncToken == "" {
		t.Error("expected sync token")
	}

	future := time.Now().Add(time.Hour)
	res2, err := uc.SyncOrders(ctx, 1, 0, &future)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res2.Orders) != 0 {
		t.Errorf("expected 0 orders after future lastSync, got %d", len(res2.Orders))
	}
}

func TestOrderUsecase_Create_AppliesCustomerCustomPrice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := newMockCustomerRepo()
	repo.prices = map[int64]float64{35: 12000}

	uc := NewOrderUsecase(&mockOrderRepo{}, repo, &mockInventoryClient{})

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
	if order.Items[0].UnitPrice != 12000 {
		t.Errorf("expected custom price 12000 for item 35, got %f", order.Items[0].UnitPrice)
	}
	if order.Items[0].LineTotal != 24000 {
		t.Errorf("expected line total 24000, got %f", order.Items[0].LineTotal)
	}
	if order.Items[1].UnitPrice != 8000 {
		t.Errorf("expected default price 8000 for item 36, got %f", order.Items[1].UnitPrice)
	}
	if order.Subtotal != 32000 {
		t.Errorf("expected subtotal 32000, got %f", order.Subtotal)
	}
}

func TestOrderUsecase_Create_NoCustomerKeepsClientPrice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	uc := NewOrderUsecase(&mockOrderRepo{}, newMockCustomerRepo(), &mockInventoryClient{})

	order, err := uc.Create(ctx, CreateOrderParams{
		MerchantID: 1,
		BranchID:   1,
		CreatedBy:  5,
		Items: []CreateOrderItemParams{
			{ProductItemID: 35, ItemName: "Es Teh Manis - Large", UnitPrice: 15000, Quantity: 2},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.Items[0].UnitPrice != 15000 {
		t.Errorf("expected client price 15000, got %f", order.Items[0].UnitPrice)
	}
	if order.Subtotal != 30000 {
		t.Errorf("expected subtotal 30000, got %f", order.Subtotal)
	}
}
