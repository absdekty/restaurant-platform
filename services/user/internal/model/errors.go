package model

import "errors"

/* Ошибки обьектов */
var (
	ErrInvalidName  = errors.New("Имя должно быть от 3 до 20 символов")
	ErrWeakPassword = errors.New("Пароль должен быть от 8 до 16 символов")
)

/* Ошибки репозитория */
var (
	ErrUserAlreadyExists = errors.New("Пользователь с таким id или name уже существует")
	ErrUserNotFound      = errors.New("Пользователь не существует")
)

/* Ошибки сервиса */
var (
	ErrInvalidCredentials = errors.New("Неверные данные для входа")
)
