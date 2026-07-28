package merchant

import (
	"encoding/json"

	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/internal/merchant/usecase"
)

func createReqToInput(req createMerchantReq) usecase.CreateMerchantParams {
	params := usecase.CreateMerchantParams{
		Name:      req.Name,
		LegalName: req.LegalName,
		Address:   req.Address,
		Phone:     req.Phone,
		Email:     req.Email,
		TaxID:     req.TaxID,
	}
	if len(req.Settings) > 0 {
		params.Settings = mapToMerchantSettings(req.Settings)
	}
	return params
}

func mapToMerchantSettings(data map[string]interface{}) domain.MerchantSettings {
	var s domain.MerchantSettings
	b, err := json.Marshal(data)
	if err != nil {
		return s
	}
	json.Unmarshal(b, &s)
	return s
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
