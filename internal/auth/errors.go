package auth

import "errors"

var (
	ErrInvalidRequest     = errors.New("invalid request")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email taken")
	ErrEmailUnverified    = errors.New("email unverified")
	ErrAccountLocked      = errors.New("account locked")
	ErrAccountDisabled    = errors.New("account disabled")
	ErrInvalidToken       = errors.New("invalid token")
	ErrPasswordReused     = errors.New("password reused")
	ErrUnauthenticated    = errors.New("unauthenticated")
	ErrCSRF               = errors.New("csrf failed")
)
