package model

import "errors"

/* Ошибки auth-service */
var (
	ErrUnauthorized  = errors.New("Unauthozired")
	ErrInvalidToken  = errors.New("Token revoked or expired")
	ErrTokenNotFound = errors.New("Token not found")
)

/* Ошибки user-service */
var (
	ErrUserInvalidRegisterDetails = errors.New("Invalid register details")
	ErrUserAlreadyExists          = errors.New("User with this name already exists")

	ErrUserInvalidCredentials = errors.New("Invalid credentials")
	ErrUserNotFound           = errors.New("User not found")
)
