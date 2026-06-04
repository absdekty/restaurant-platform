package model

import "errors"

/* Ошибки хранилища */
var (
	TokenAlreadyExists = errors.New("Token already exist")
	ErrTokenNotFound   = errors.New("Token not found")
)

/* Ошибки сервиса */
var (
	ErrTokenRevoked = errors.New("Token revoked")
)
