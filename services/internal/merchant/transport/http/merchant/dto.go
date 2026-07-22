package merchant

type createMerchantReq struct {
	Name      string                   `json:"name"`
	LegalName string                   `json:"legalName"`
	Address   string                   `json:"address"`
	Phone     string                   `json:"phone"`
	Email     string                   `json:"email"`
	TaxID     string                   `json:"taxId"`
	Settings  map[string]interface{}   `json:"settings"`
}

type updateMerchantReq struct {
	Name      *string                   `json:"name,omitempty"`
	LegalName *string                   `json:"legalName,omitempty"`
	Address   *string                   `json:"address,omitempty"`
	Phone     *string                   `json:"phone,omitempty"`
	Email     *string                   `json:"email,omitempty"`
	TaxID     *string                   `json:"taxId,omitempty"`
	Status    *string                   `json:"status,omitempty"`
	Settings  *map[string]interface{}   `json:"settings,omitempty"`
}
