package middleware

import "errors"

var (
	errInvalidBody     = errors.New("invalid request body")
	errMissingFields   = errors.New("missing required fields")
	errMissingBranchID = errors.New("missing branch id")
	rateLimitExceeded  = errors.New("rate limit exceeded")
)
