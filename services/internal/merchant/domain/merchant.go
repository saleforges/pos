package domain

import "time"

type MerchantStatus string

const (
	MerchantStatusActive    MerchantStatus = "active"
	MerchantStatusInactive  MerchantStatus = "inactive"
	MerchantStatusSuspended MerchantStatus = "suspended"
)

type MerchantSettings struct {
	TaxRate            float64 `json:"tax_rate"`
	Currency           string  `json:"currency"`
	Timezone           string  `json:"timezone"`
	ReceiptFooter      string  `json:"receipt_footer,omitempty"`
	ReceiptLogo        string  `json:"receipt_logo,omitempty"`
	OrderPrefix        string  `json:"order_prefix,omitempty"`
	LowStockThreshold  int     `json:"low_stock_threshold"`
}

type Merchant struct {
	ID        int64           `json:"id"`
	Name      string           `json:"name"`
	LegalName string           `json:"legal_name"`
	Address   string           `json:"address"`
	Phone     string           `json:"phone"`
	Email     string           `json:"email"`
	LogoURL   string           `json:"logo_url,omitempty"`
	TaxID     string           `json:"tax_id,omitempty"`
	Status    MerchantStatus   `json:"status"`
	Settings  MerchantSettings `json:"settings"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
