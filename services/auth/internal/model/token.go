package model

import "time"

type Token struct {
	UserID    string
	Token     string
	Revoked   bool
	CreatedAt time.Time
	ExpiresAt time.Time
}
