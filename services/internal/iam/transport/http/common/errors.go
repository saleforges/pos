package common

import "errors"

var (
	ErrInvalidBody      = errors.New("invalid request body")
	ErrMissingFields    = errors.New("missing required fields")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrInvalidID        = errors.New("invalid id parameter")
)
