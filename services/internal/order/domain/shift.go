package domain

import "time"

type ShiftStatus string

const (
	ShiftStatusOpen   ShiftStatus = "open"
	ShiftStatusClosed ShiftStatus = "closed"
)

type Shift struct {
	ID           int64       `json:"id"`
	MerchantID   int64       `json:"merchantId"`
	BranchID     int64       `json:"branchId"`
	OpenedBy     int64       `json:"openedBy"`
	ClosedBy     *int64      `json:"closedBy,omitempty"`
	Status       ShiftStatus `json:"status"`
	StartingCash float64     `json:"startingCash"`
	ExpectedCash *float64    `json:"expectedCash,omitempty"`
	ActualCash   *float64    `json:"actualCash,omitempty"`
	Variance     *float64    `json:"variance,omitempty"`
	Note         string      `json:"note,omitempty"`
	OpenedAt     time.Time   `json:"openedAt"`
	ClosedAt     *time.Time  `json:"closedAt,omitempty"`
}

func (s *Shift) Validate() error {
	if s.MerchantID == 0 {
		return ErrInvalidShift
	}
	if s.BranchID == 0 {
		return ErrInvalidShift
	}
	if s.StartingCash < 0 {
		return ErrInvalidShift
	}
	return nil
}
