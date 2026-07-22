package branch

import "github.com/saleforge/pos/services/internal/merchant/domain"

type createBranchReq struct {
	Name           string                  `json:"name"`
	Code           string                  `json:"code"`
	Address        string                  `json:"address"`
	Phone          string                  `json:"phone"`
	OperatingDays  []string                `json:"operatingDays"`
	OperatingHours *domain.OperatingHours  `json:"operatingHours"`
}

type updateBranchReq struct {
	Name           *string                 `json:"name,omitempty"`
	Address        *string                 `json:"address,omitempty"`
	Phone          *string                 `json:"phone,omitempty"`
	Status         *domain.BranchStatus    `json:"status,omitempty"`
	OperatingDays  []string                `json:"operatingDays,omitempty"`
	OperatingHours *domain.OperatingHours  `json:"operatingHours,omitempty"`
}
