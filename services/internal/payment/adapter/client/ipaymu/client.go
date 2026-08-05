package ipaymu

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
	Product       []string `json:"product"`
	Qty           []int    `json:"qty"`
	Price         []int64  `json:"price"`
	ReturnURL     string   `json:"returnUrl"`
	CancelURL     string   `json:"cancelUrl"`
	NotifyURL     string   `json:"notifyUrl"`
	ReferenceID   string   `json:"referenceId"`
	PaymentMethod string   `json:"paymentMethod,omitempty"`
	BuyerName     string   `json:"buyerName,omitempty"`
	BuyerEmail    string   `json:"buyerEmail,omitempty"`
	BuyerPhone    string   `json:"buyerPhone,omitempty"`
}

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
		Product:       products,
		Qty:           qty,
		Price:         prices,
		ReturnURL:     c.cfg.ReturnURL,
		CancelURL:     c.cfg.CancelURL,
		NotifyURL:     c.cfg.NotifyURL,
		ReferenceID:   params.ReferenceID,
		PaymentMethod: params.Method,
		BuyerName:     params.BuyerName,
		BuyerEmail:    params.BuyerEmail,
		BuyerPhone:    params.BuyerPhone,
	})
	if err != nil {
		return nil, err
	}

	// Direct mode (paymentMethod set) hits /payment/direct and returns the
	// QRIS/VA payload; otherwise the hosted page returns a redirect URL.
	endpoint := fmt.Sprintf("%s/api/v2/payment", c.cfg.BaseURL)
	if params.Method != "" {
		endpoint = fmt.Sprintf("%s/api/v2/payment/direct", c.cfg.BaseURL)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
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
		Data    struct {
			SessionID     string `json:"SessionID"`
			URL           string `json:"Url"`
			TransactionID int64  `json:"TransactionId"`
			Via           string `json:"Via"`
			PaymentNo     string `json:"PaymentNo"`
			QrString      string `json:"QrString"`
			QrImage       string `json:"QrImage"`
			Expired       string `json:"Expired"`
		} `json:"Data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, domain.ErrGatewayUnavailable
	}
	if !envelope.Success {
		return nil, fmt.Errorf("%w: %s", domain.ErrGatewayError, envelope.Message)
	}

	result := &repository.PaymentResult{
		SessionID:  envelope.Data.SessionID,
		PaymentURL: envelope.Data.URL,
		PaymentRef: strconv.FormatInt(envelope.Data.TransactionID, 10),
		Via:        envelope.Data.Via,
		PaymentNo:  envelope.Data.PaymentNo,
		QrString:   envelope.Data.QrString,
		QrImage:    envelope.Data.QrImage,
		ExpiredAt:  envelope.Data.Expired,
	}
	if params.Method != "" && result.QrImage == "" && result.PaymentNo == "" {
		return nil, fmt.Errorf("%w: no payment details returned", domain.ErrGatewayError)
	}
	if params.Method == "" && result.PaymentURL == "" {
		return nil, fmt.Errorf("%w: no payment url returned", domain.ErrGatewayError)
	}
	return result, nil
}
