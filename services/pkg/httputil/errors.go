package httputil

import "errors"

var (
	ErrInvalidBody   = errors.New("invalid request body")
	ErrMissingFields = errors.New("missing required fields")
)
