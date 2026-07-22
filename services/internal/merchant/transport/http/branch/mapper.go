package branch

import (
	"github.com/saleforge/pos/services/internal/merchant/usecase"
)

func createReqToInput(merchantID int64, req createBranchReq) usecase.CreateBranchParams {
	return usecase.CreateBranchParams{
		MerchantID:     merchantID,
		Name:           req.Name,
		Code:           req.Code,
		Address:        req.Address,
		Phone:          req.Phone,
		OperatingDays:  req.OperatingDays,
		OperatingHours: req.OperatingHours,
	}
}

func updateReqToInput(id int64, req updateBranchReq) usecase.UpdateBranchParams {
	return usecase.UpdateBranchParams{
		ID:             id,
		Name:           req.Name,
		Address:        req.Address,
		Phone:          req.Phone,
		Status:         req.Status,
		OperatingDays:  req.OperatingDays,
		OperatingHours: req.OperatingHours,
	}
}
