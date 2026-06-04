package model

import "errors"

/* Ошибки сервиса */
var (
	ErrTokenRevoked = errors.New("Token revoked")
)
