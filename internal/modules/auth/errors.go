package auth

import "errors"

var (
	// Login Errors
	ErrInvalidCredentials = errors.New("invalid email or password")

	ErrInactiveUser = errors.New("user account is inactive")

	ErrInvalidToken = errors.New("invalid token")

	ErrMissingToken = errors.New("authorization token is required")

	ErrExpiredToken = errors.New("token has expired")
)