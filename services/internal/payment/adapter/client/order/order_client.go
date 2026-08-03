package order

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/saleforge/pos/services/internal/payment/domain"
	"github.com/saleforge/pos/services/internal/payment/port/repository"
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

func New(cfg Config) repository.OrderClient {
	return &client{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type notifyPaidReq struct {
	OrderID    int64   `json:"orderId"`
	MerchantID int64   `json:"merchantId"`
	Amount     float64 `json:"amount"`
	Method     string  `json:"method"`
}

func (c *client) NotifyPaid(ctx context.Context, orderID, merchantID int64, amount float64, method string) error {
	body, err := json.Marshal(notifyPaidReq{
		OrderID:    orderID,
		MerchantID: merchantID,
		Amount:     amount,
		Method:     method,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/internal/orders/%d/paid", c.baseURL, orderID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Key", c.apiKey)
	req.Header.Set("X-Merchant-Id", strconv.FormatInt(merchantID, 10))

	resp, err := c.http.Do(req)
	if err != nil {
		return domain.ErrOrderClientUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	var errResp struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
		switch errResp.Code {
		case "ORD001":
			return domain.ErrPaymentNotFound
		}
	}
	return domain.ErrOrderClientUnavailable
}
