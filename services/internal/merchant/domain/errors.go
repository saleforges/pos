package domain

import "errors"

var (
	ErrMerchantNotFound    = errors.New("MCH001: merchant not found")
	ErrMerchantExists      = errors.New("MCH002: merchant already exists")
	ErrBranchNotFound      = errors.New("MCH003: branch not found")
	ErrBranchExists        = errors.New("MCH004: branch code already exists")
	ErrInvalidMerchant     = errors.New("MCH005: invalid merchant data")
	ErrInvalidBranch       = errors.New("MCH006: invalid branch data")
	ErrStaffNotFound       = errors.New("MCH007: staff member not found")
	ErrStaffExists         = errors.New("MCH008: staff already assigned to this branch")
	ErrInvalidStaff        = errors.New("MCH009: invalid staff data")
	ErrInvalidToken        = errors.New("MCH010: invalid or expired token")
	ErrForbidden           = errors.New("MCH011: insufficient permissions")
	ErrInternal            = errors.New("MCH500: internal error")
)
