package domain

import "errors"

var (
	ErrUserNotFound        = errors.New("AUTH006: user not found")
	ErrInvalidCredentials  = errors.New("AUTH001: invalid credentials")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserDisabled        = errors.New("AUTH002: user disabled")
	ErrInvalidRole         = errors.New("invalid role")
	ErrForbidden           = errors.New("AUTH005: permission denied")
	ErrInvalidToken        = errors.New("AUTH003: access token expired")
	ErrInvalidRefreshToken = errors.New("AUTH004: invalid refresh token")
	ErrInternal            = errors.New("internal server error")
	ErrPasswordPolicy      = errors.New("password does not meet policy requirements")
)
