package customer

type createCustomerReq struct {
	Name    string `json:"name"`
	Phone   string `json:"phone,omitempty"`
	Address string `json:"address,omitempty"`
	Note    string `json:"note,omitempty"`
}

type updateCustomerReq struct {
	Name    *string `json:"name,omitempty"`
	Phone   *string `json:"phone,omitempty"`
	Address *string `json:"address,omitempty"`
	Note    *string `json:"note,omitempty"`
}

type setPricesReq struct {
	Items []priceEntryReq `json:"items"`
}

type priceEntryReq struct {
	ProductItemID int64   `json:"productItemId"`
	Price         float64 `json:"price"`
}
