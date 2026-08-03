package productcomponent

type createProductComponentReq struct {
	Items []createProductComponentItemReq `json:"items"`
}

type createProductComponentItemReq struct {
	ComponentProductItemID int64   `json:"componentProductItemId"`
	Quantity               float64 `json:"quantity"`
	UnitID                 int64   `json:"unitId"`
	ConversionFactor       float64 `json:"conversionFactor,omitempty"`
}

type updateProductComponentReq struct {
	Items []createProductComponentItemReq `json:"items"`
}
