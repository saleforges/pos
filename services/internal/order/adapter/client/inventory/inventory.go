package inventory

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

func New(cfg Config) repository.InventoryClient {
	return &client{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:  cfg.APIKey,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type adjustRequest struct {
	MerchantID    int64               `json:"merchantId"`
	BranchID      int64               `json:"branchId"`
	ReferenceType string              `json:"referenceType"`
	ReferenceID   int64               `json:"referenceId"`
	Items         []adjustRequestItem `json:"items"`
}

type adjustRequestItem struct {
	ProductItemID int64 `json:"productItemId"`
	Quantity      int64 `json:"quantity"`
}

func (c *client) DeductStock(ctx context.Context, merchantID, branchID int64, referenceType string, referenceID int64, items []repository.StockAdjustmentItem) error {
	return c.adjust(ctx, "deduct", merchantID, branchID, referenceType, referenceID, items)
}

func (c *client) RestoreStock(ctx context.Context, merchantID, branchID int64, referenceType string, referenceID int64, items []repository.StockAdjustmentItem) error {
	return c.adjust(ctx, "restore", merchantID, branchID, referenceType, referenceID, items)
}

func (c *client) adjust(ctx context.Context, action string, merchantID, branchID int64, referenceType string, referenceID int64, items []repository.StockAdjustmentItem) error {
	reqItems := make([]adjustRequestItem, len(items))
	for i, it := range items {
		reqItems[i] = adjustRequestItem{ProductItemID: it.ProductItemID, Quantity: it.Quantity}
	}
	body, err := json.Marshal(adjustRequest{
		MerchantID:    merchantID,
		BranchID:      branchID,
		ReferenceType: referenceType,
		ReferenceID:   referenceID,
		Items:         reqItems,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/internal/stocks/%s", c.baseURL, action), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return domain.ErrInternal
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	// Map inventory error codes to order domain errors.
	var errResp struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
		switch errResp.Code {
		case "INV010":
			return domain.ErrInsufficientStock
		case "INV001":
			return domain.ErrOrderStockNotFound
		}
	}
	return domain.ErrInternal
}
