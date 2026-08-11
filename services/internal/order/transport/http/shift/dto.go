package shift

type openShiftReq struct {
	BranchID     int64   `json:"branchId"`
	StartingCash float64 `json:"startingCash"`
}

type closeShiftReq struct {
	ActualCash float64 `json:"actualCash"`
	Note       string  `json:"note,omitempty"`
}
