package user

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")

	ErrUserNotFound = errors.New("user not found")

	ErrInvalidPassword = errors.New("invalid password")

	ErrInvalidRole = errors.New("invalid role")

	ErrCannotChangeOwnRole = errors.New("admins cannot change their own role")
)
