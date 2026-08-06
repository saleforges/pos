package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/order/domain"
)

type CustomerUsecase interface {
	Create(ctx context.Context, params CreateCustomerParams) (*domain.Customer, error)
	GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Customer, error)
	List(ctx context.Context, merchantID int64, search string) ([]domain.Customer, error)
	Update(ctx context.Context, params UpdateCustomerParams) (*domain.Customer, error)
	Delete(ctx context.Context, id int64, merchantID int64) error
	Sync(ctx context.Context, merchantID int64, lastSync *time.Time) (*CustomerSyncResult, error)
	SetPrices(ctx context.Context, params SetCustomerPricesParams) error
	GetPrices(ctx context.Context, merchantID, customerID int64) ([]domain.CustomerPrice, error)
}

type CustomerSyncResult struct {
	Customers []domain.Customer      `json:"customers"`
	Prices    []domain.CustomerPrice `json:"prices"`
	SyncToken string                 `json:"syncToken"`
}

type PriceEntry struct {
	ProductItemID int64   `json:"productItemId"`
	Price         float64 `json:"price"`
}

type SetCustomerPricesParams struct {
	MerchantID int64
	CustomerID int64
	Prices     []PriceEntry
}

type CreateCustomerParams struct {
	MerchantID int64
	Name       string
	Phone      string
	Address    string
	Note       string
}

type UpdateCustomerParams struct {
	ID         int64
	MerchantID int64
	Name       *string
	Phone      *string
	Address    *string
	Note       *string
}
