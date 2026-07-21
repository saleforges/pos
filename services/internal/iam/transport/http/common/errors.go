package common

import (
	"errors"
	"regexp"
)

var (
	ErrInvalidBody      = errors.New("invalid request body")
	ErrMissingFields    = errors.New("missing required fields")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrInvalidID        = errors.New("invalid id parameter")
)

// EmailRegex validates basic email format. ponytail: upgrade to RFC 5322 if stricter
// validation is needed; keep here as single source of truth for transport-layer checks.
var EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
