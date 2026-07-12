package model

import "errors"

/* Ошибки хранилища */
var (
	ErrTokenAlreadyExists = errors.New("Token already exists")
	ErrTokenNotFound      = errors.New("Token not found")
	ErrTokenExpired       = errors.New("Token expired")
)

/* Ошибки сервиса */
var (
	ErrTokenRevoked = errors.New("Token revoked")
)
