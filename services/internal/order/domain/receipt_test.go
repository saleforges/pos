package domain

import (
	"strings"
	"testing"
	"time"
)

func TestBuildReceipt(t *testing.T) {
	order := &Order{
		ID:         42,
		BranchID:   1,
		Status:     OrderStatusCompleted,
		Subtotal:   30000,
		Total:      30000,
		PaidAmount: 50000,
		CreatedAt:  time.Date(2026, 8, 5, 10, 30, 0, 0, time.UTC),
		Items: []OrderItem{
			{ItemName: "Es Teh", Quantity: 2, UnitPrice: 15000, LineTotal: 30000},
		},
		Payments: []PaymentRecord{
			{Method: PaymentMethodCash, Amount: 50000, PaidAt: time.Now()},
		},
	}

	r := BuildReceipt(order)
	if r.OrderID != 42 {
		t.Errorf("expected order id 42, got %d", r.OrderID)
	}
	if r.Change != 20000 {
		t.Errorf("expected change 20000, got %f", r.Change)
	}
	if r.Debt != 0 {
		t.Errorf("expected no debt, got %f", r.Debt)
	}
	if !strings.Contains(r.Text, "#42") {
		t.Error("receipt text missing order number")
	}
	if !strings.Contains(r.Text, "Es Teh") {
		t.Error("receipt text missing item")
	}
	if !strings.Contains(r.Text, "Kembali") {
		t.Error("receipt text missing change")
	}
}

func TestBuildReceiptDebt(t *testing.T) {
	order := &Order{
		ID:         7,
		BranchID:   1,
		Status:     OrderStatusCompleted,
		Total:      30000,
		PaidAmount: 10000,
		CreatedAt:  time.Now(),
		Items:      []OrderItem{{ItemName: "Kopi", Quantity: 1, UnitPrice: 30000, LineTotal: 30000}},
	}
	r := BuildReceipt(order)
	if r.Debt != 20000 {
		t.Errorf("expected debt 20000, got %f", r.Debt)
	}
	if !strings.Contains(r.Text, "Sisa") {
		t.Error("receipt text missing debt line")
	}
}
