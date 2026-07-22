package merchant

import (
	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/internal/merchant/usecase"
)

func createReqToInput(req createMerchantReq, settings domain.MerchantSettings) usecase.CreateMerchantParams {
	return usecase.CreateMerchantParams{
		Name:      req.Name,
		LegalName: req.LegalName,
		Address:   req.Address,
		Phone:     req.Phone,
		Email:     req.Email,
		TaxID:     req.TaxID,
		Settings:  settings,
	}
}

func updateReqToInput(id int64, req updateMerchantReq) usecase.UpdateMerchantParams {
	return usecase.UpdateMerchantParams{
		ID:        id,
		Name:      req.Name,
		LegalName: req.LegalName,
		Address:   req.Address,
		Phone:     req.Phone,
		Email:     req.Email,
		TaxID:     req.TaxID,
		Status:    nil,
	}
}
