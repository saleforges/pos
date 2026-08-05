package middleware

import "errors"

var (
	errMissingBranchID = errors.New("missing branch id")
	rateLimitExceeded  = errors.New("rate limit exceeded, try again later")
)
