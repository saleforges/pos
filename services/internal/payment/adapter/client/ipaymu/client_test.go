package ipaymu

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/saleforge/pos/services/internal/payment/port/repository"
)

func TestClient_CreatePayment(t *testing.T) {
	t.Parallel()

	// Real iPaymu sandbox response shape: Data is an OBJECT (not array).
	t.Run("parses object Data response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("signature") == "" {
				t.Error("expected signature header")
			}
			if r.Header.Get("va") == "" {
				t.Error("expected va header")
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"Status":200,"Success":true,"Message":"Success","Data":{"SessionID":"e7954619-191e-4cb8-a5ec-8b2109168bb5","Url":"https://sandbox-payment.ipaymu.com/#/e7954619"}}`))
		}))
		defer srv.Close()

		c := New(Config{BaseURL: srv.URL, VA: "0000001326904469", APIKey: "SANDBOXKEY"})
		result, err := c.CreatePayment(context.Background(), repository.CreatePaymentParams{
			ReferenceID: "21",
			Items:       []repository.CreatePaymentItem{{ItemName: "Es Teh", Quantity: 1, UnitPrice: 15000}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.SessionID != "e7954619-191e-4cb8-a5ec-8b2109168bb5" {
			t.Errorf("unexpected session id: %s", result.SessionID)
		}
		if result.PaymentURL == "" {
			t.Error("expected payment url")
		}
	})

	t.Run("gateway error surfaces message", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"Status":500,"Success":false,"Message":"Invalid signature","Data":null}`))
		}))
		defer srv.Close()

		c := New(Config{BaseURL: srv.URL, VA: "va", APIKey: "key"})
		_, err := c.CreatePayment(context.Background(), repository.CreatePaymentParams{})
		if err == nil {
			t.Fatal("expected error for failed gateway response")
		}
	})

	t.Run("not configured", func(t *testing.T) {
		c := New(Config{})
		_, err := c.CreatePayment(context.Background(), repository.CreatePaymentParams{})
		if err == nil {
			t.Fatal("expected not-configured error")
		}
	})
}

// Ensure the request body serializes the fields iPaymu expects.
func TestClient_RequestShape(t *testing.T) {
	t.Parallel()

	var got map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"Status":200,"Success":true,"Message":"Success","Data":{"SessionID":"s","Url":"u"}}`))
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL, VA: "va", APIKey: "key"})
	_, err := c.CreatePayment(context.Background(), repository.CreatePaymentParams{
		ReferenceID: "21",
		Items:       []repository.CreatePaymentItem{{ItemName: "Es Teh", Quantity: 2, UnitPrice: 15000}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["referenceId"] != "21" {
		t.Errorf("expected referenceId 21, got %v", got["referenceId"])
	}
	products := got["product"].([]interface{})
	if products[0] != "Es Teh" {
		t.Errorf("expected product name, got %v", products[0])
	}
	prices := got["price"].([]interface{})
	if prices[0] != float64(30000) {
		t.Errorf("expected price 30000 (unit*qty), got %v", prices[0])
	}
}
