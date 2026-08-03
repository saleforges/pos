package ipaymu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/saleforge/pos/services/internal/payment/domain"
	"github.com/saleforge/pos/services/internal/payment/port/repository"
)

type Config struct {
	BaseURL   string // https://sandbox.ipaymu.com or https://my.ipaymu.com
	VA        string
	APIKey    string
	ReturnURL string
	CancelURL string
	NotifyURL string
}

type client struct {
	cfg  Config
	http *http.Client
}

func New(cfg Config) repository.GatewayClient {
	return &client{cfg: cfg, http: &http.Client{Timeout: 15 * time.Second}}
}

type createPaymentReq struct {
	Product     []string `json:"product"`
	Qty         []int    `json:"qty"`
	Price       []int64  `json:"price"`
	ReturnURL   string   `json:"returnUrl"`
	CancelURL   string   `json:"cancelUrl"`
	NotifyURL   string   `json:"notifyUrl"`
	ReferenceID string   `json:"referenceId"`
	BuyerName   string   `json:"buyerName,omitempty"`
	BuyerEmail  string   `json:"buyerEmail,omitempty"`
	BuyerPhone  string   `json:"buyerPhone,omitempty"`
}

// CreatePayment requests a payment URL from iPaymu for an order.
func (c *client) CreatePayment(ctx context.Context, params repository.CreatePaymentParams) (*repository.PaymentResult, error) {
	if c.cfg.VA == "" || c.cfg.APIKey == "" {
		return nil, domain.ErrGatewayNotConfigured
	}

	products := make([]string, len(params.Items))
	qty := make([]int, len(params.Items))
	prices := make([]int64, len(params.Items))
	for i, it := range params.Items {
		products[i] = it.ItemName
		qty[i] = int(it.Quantity)
		prices[i] = int64(it.UnitPrice * it.Quantity)
	}

	body, err := json.Marshal(createPaymentReq{
		Product:     products,
		Qty:         qty,
		Price:       prices,
		ReturnURL:   c.cfg.ReturnURL,
		CancelURL:   c.cfg.CancelURL,
		NotifyURL:   c.cfg.NotifyURL,
		ReferenceID: params.ReferenceID,
		BuyerName:   params.BuyerName,
		BuyerEmail:  params.BuyerEmail,
		BuyerPhone:  params.BuyerPhone,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/api/v2/payment", c.cfg.BaseURL), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("va", c.cfg.VA)
	req.Header.Set("signature", SignRequest("POST", c.cfg.VA, c.cfg.APIKey, body))
	req.Header.Set("timestamp", Timestamp())

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, domain.ErrGatewayUnavailable
	}
	defer resp.Body.Close()

	var envelope struct {
		Status  int    `json:"Status"`
		Success bool   `json:"Success"`
		Message string `json:"Message"`
		Data    []struct {
			SessionID string `json:"sessionId"`
			URL       string `json:"url"`
		} `json:"Data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, domain.ErrGatewayUnavailable
	}
	if !envelope.Success || len(envelope.Data) == 0 {
		return nil, fmt.Errorf("%w: %s", domain.ErrGatewayError, envelope.Message)
	}

	return &repository.PaymentResult{
		SessionID:  envelope.Data[0].SessionID,
		PaymentURL: envelope.Data[0].URL,
	}, nil
}
