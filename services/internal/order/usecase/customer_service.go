package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/order/domain"
	"github.com/saleforge/pos/services/internal/order/port/repository"
)

type customerUsecase struct {
	repo repository.CustomerRepository
}

func NewCustomerUsecase(repo repository.CustomerRepository) CustomerUsecase {
	return &customerUsecase{repo: repo}
}

func (uc *customerUsecase) Create(ctx context.Context, params CreateCustomerParams) (*domain.Customer, error) {
	now := time.Now().UTC()
	customer := &domain.Customer{
		MerchantID: params.MerchantID,
		Name:       params.Name,
		Phone:      params.Phone,
		Address:    params.Address,
		Note:       params.Note,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := customer.Validate(); err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, customer); err != nil {
		return nil, err
	}
	return customer, nil
}

func (uc *customerUsecase) GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Customer, error) {
	return uc.repo.GetByID(ctx, id, merchantID)
}

func (uc *customerUsecase) List(ctx context.Context, merchantID int64, search string) ([]domain.Customer, error) {
	return uc.repo.List(ctx, merchantID, search)
}

func (uc *customerUsecase) Update(ctx context.Context, params UpdateCustomerParams) (*domain.Customer, error) {
	customer, err := uc.repo.GetByID(ctx, params.ID, params.MerchantID)
	if err != nil {
		return nil, err
	}
	if params.Name != nil {
		if *params.Name == "" {
			return nil, domain.ErrInvalidCustomer
		}
		customer.Name = *params.Name
	}
	if params.Phone != nil {
		customer.Phone = *params.Phone
	}
	if params.Address != nil {
		customer.Address = *params.Address
	}
	if params.Note != nil {
		customer.Note = *params.Note
	}
	customer.UpdatedAt = time.Now().UTC()

	if err := uc.repo.Update(ctx, customer); err != nil {
		return nil, err
	}
	return customer, nil
}

func (uc *customerUsecase) Delete(ctx context.Context, id int64, merchantID int64) error {
	return uc.repo.Delete(ctx, id, merchantID)
}

func (uc *customerUsecase) Sync(ctx context.Context, merchantID int64, lastSync *time.Time) (*CustomerSyncResult, error) {
	customers, err := uc.repo.ListChangedSince(ctx, merchantID, lastSync)
	if err != nil {
		return nil, err
	}
	if customers == nil {
		customers = []domain.Customer{}
	}
	return &CustomerSyncResult{
		Customers: customers,
		SyncToken: time.Now().UTC().Format(time.RFC3339Nano),
	}, nil
}

var _ CustomerUsecase = (*customerUsecase)(nil)
