package domain

import "time"

type MerchantStatus string

const (
	MerchantStatusActive    MerchantStatus = "active"
	MerchantStatusInactive  MerchantStatus = "inactive"
	MerchantStatusSuspended MerchantStatus = "suspended"
)

type MerchantSettings struct {
	TaxRate            float64 `json:"taxRate"`
	Currency           string  `json:"currency"`
	Timezone           string  `json:"timezone"`
	ReceiptFooter      string  `json:"receiptFooter,omitempty"`
	ReceiptLogo        string  `json:"receiptLogo,omitempty"`
	OrderPrefix        string  `json:"orderPrefix,omitempty"`
	LowStockThreshold  int     `json:"lowStockThreshold"`
}

type Merchant struct {
	ID        int64           `json:"id"`
	Name      string           `json:"name"`
	LegalName string           `json:"legalName"`
	Address   string           `json:"address"`
	Phone     string           `json:"phone"`
	Email     string           `json:"email"`
	LogoURL   string           `json:"logoUrl,omitempty"`
	TaxID     string           `json:"taxId,omitempty"`
	Status    MerchantStatus   `json:"status"`
	Settings  MerchantSettings `json:"settings"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}
