package stock

type createStockReq struct {
	BranchID      int64 `json:"branchId"`
	ProductItemID int64 `json:"productItemId"`
	Available     int64 `json:"available"`
}

type updateStockReq struct {
	Available int64 `json:"available"`
}

type adjustStockReq struct {
	MerchantID    int64             `json:"merchantId"`
	BranchID      int64             `json:"branchId"`
	ReferenceType string            `json:"referenceType"`
	ReferenceID   int64             `json:"referenceId"`
	Items         []adjustStockItem `json:"items"`
}

type adjustStockItem struct {
	ProductItemID int64 `json:"productItemId"`
	Quantity      int64 `json:"quantity"`
}
