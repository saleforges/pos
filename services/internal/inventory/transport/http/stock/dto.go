package stock

type createStockReq struct {
	BranchID      int64 `json:"branchId"`
	ProductItemID int64 `json:"productItemId"`
	Available     int64 `json:"available"`
}

type updateStockReq struct {
	Available int64 `json:"available"`
}
