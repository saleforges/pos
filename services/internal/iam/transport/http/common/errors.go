package common

import "errors"

var (
	ErrInvalidBody   = errors.New("invalid request body")
	ErrMissingFields = errors.New("missing required fields")
	ErrUnauthorized  = errors.New("unauthorized")
)
