package handler

import "errors"

var (
	errInvalidBody   = errors.New("invalid request body")
	errMissingFields = errors.New("missing required fields")
	errUnauthorized  = errors.New("unauthorized")
)
