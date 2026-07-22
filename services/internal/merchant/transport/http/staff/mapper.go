package staff

import "github.com/saleforge/pos/services/internal/merchant/usecase"

func assignReqToInput(merchantID int64, req assignStaffReq) usecase.AssignStaffParams {
	return usecase.AssignStaffParams{
		MerchantID: merchantID,
		BranchID:   req.BranchID,
		UserID:     req.UserID,
		Role:       req.Role,
		IsDefault:  req.IsDefault,
	}
}

func updateReqToInput(id int64, req updateStaffReq) usecase.UpdateStaffParams {
	return usecase.UpdateStaffParams{
		ID:     id,
		Role:   req.Role,
		Status: req.Status,
	}
}
