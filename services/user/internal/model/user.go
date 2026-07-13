package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        string
	Name      string
	Password  string // Already hashed
	CreatedAt time.Time
}

func NewUser(name, hashedPassword string) (*User, error) {
	return &User{
		ID:        uuid.New().String(),
		Name:      name,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}, nil
}

func ValidateName(name string) error {
	if len(name) < 3 || len(name) > 20 {
		return ErrInvalidName
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 || len(password) > 16 {
		return ErrWeakPassword
	}
	return nil
}
