package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/saleforge/pos/services/internal/order/domain"
	"github.com/saleforge/pos/services/internal/order/port/repository"
)

type Config struct {
	BaseURL string
	APIKey  string
}

type client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func New(cfg Config) repository.PaymentClient {
	return &client{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:  cfg.APIKey,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

type createPaymentReq struct {
	MerchantID int64               `json:"merchantId"`
	OrderID    int64               `json:"orderId"`
	Amount     float64             `json:"amount"`
	BuyerName  string              `json:"buyerName,omitempty"`
	BuyerEmail string              `json:"buyerEmail,omitempty"`
	BuyerPhone string              `json:"buyerPhone,omitempty"`
	Items      []createPaymentItem `json:"items"`
}

type createPaymentItem struct {
	ItemName  string  `json:"itemName"`
	Quantity  float64 `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
}

func (c *client) CreatePayment(ctx context.Context, params repository.CreatePaymentParams) (*repository.PaymentResult, error) {
	items := make([]createPaymentItem, len(params.Items))
	for i, it := range params.Items {
		items[i] = createPaymentItem{ItemName: it.ItemName, Quantity: it.Quantity, UnitPrice: it.UnitPrice}
	}
	body, err := json.Marshal(createPaymentReq{
		MerchantID: params.MerchantID,
		OrderID:    params.OrderID,
		Amount:     params.Amount,
		BuyerName:  params.BuyerName,
		BuyerEmail: params.BuyerEmail,
		BuyerPhone: params.BuyerPhone,
		Items:      items,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/internal/payments", c.baseURL), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, domain.ErrPaymentGatewayUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			switch errResp.Code {
			case "PAY003":
				return nil, domain.ErrPaymentGatewayNotConfigured
			case "PAY004", "PAY005", "PAY008":
				return nil, domain.ErrPaymentGatewayUnavailable
			}
		}
		return nil, domain.ErrPaymentGatewayUnavailable
	}

	var result repository.PaymentResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, domain.ErrPaymentGatewayUnavailable
	}
	return &result, nil
}
