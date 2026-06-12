package service

import (
	"context"
	"errors"
	"restaurant/services/user/internal/model"
	"restaurant/services/user/internal/storage/mock"
	"restaurant/services/user/pkg/hasher"
	"testing"
)

func TestCreateUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		userName string
		password string
		setup    func(*UserService)
		wantErr  error
	}{
		{
			name:     "Valid create user",
			userName: "ValidName",
			password: "ValidPassword",
			setup:    func(u *UserService) {},
			wantErr:  nil,
		},
		{
			name:     "Valid create user | already exists",
			userName: "ValidName",
			password: "ValidPassword",
			setup: func(u *UserService) {
				if _, err := u.CreateUser(context.Background(), "ValidName", "ValidPassword"); err != nil {
					t.Errorf("setup: %v", err)
				}
			},
			wantErr: model.ErrUserAlreadyExists,
		},
		{
			name:     "Invalid name",
			userName: ".",
			password: "ValidPassword",
			setup:    func(u *UserService) {},
			wantErr:  model.ErrInvalidName,
		},
		{
			name:     "Invalid password",
			userName: "ValidName",
			password: ".",
			setup:    func(u *UserService) {},
			wantErr:  model.ErrWeakPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := mock.New()
			hasher := hasher.NewMock()
			u := NewUserService(storage, hasher)
			tt.setup(u)

			_, err := u.CreateUser(ctx, tt.userName, tt.password)
			if err != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected %v, got %v", tt.wantErr, err)
				}
			} else {
				if tt.wantErr != nil {
					t.Errorf("not expected error, got %v", err)
				}
			}
		})
	}
}

func TestLoginUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		userName string
		password string
		setup    func(*UserService)
		wantErr  error
	}{
		{
			name:     "Valid credentials",
			userName: "ValidName",
			password: "ValidPassword",
			setup: func(u *UserService) {
				if _, err := u.CreateUser(context.Background(), "ValidName", "ValidPassword"); err != nil {
					t.Errorf("setup: %v", err)
				}
			},
			wantErr: nil,
		},
		{
			name:     "Invalid credentials",
			userName: "ValidName",
			password: "InvalidPassword",
			setup: func(u *UserService) {
				if _, err := u.CreateUser(context.Background(), "ValidName", "ValidPassword"); err != nil {
					t.Errorf("setup: %v", err)
				}
			},
			wantErr: model.ErrInvalidCredentials,
		},
		{
			name:     "Invalid credentials | validate name",
			userName: ".",
			password: "ValidatePassword",
			setup:    func(u *UserService) {},
			wantErr:  model.ErrInvalidName,
		},
		{
			name:     "Invalid credentials | validate password",
			userName: "ValidateName",
			password: ".",
			setup:    func(u *UserService) {},
			wantErr:  model.ErrWeakPassword,
		},
		{
			name:     "User not exists",
			userName: "ValidName",
			password: "ValidPassword",
			setup:    func(u *UserService) {},
			wantErr:  model.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := mock.New()
			hasher := hasher.NewMock()
			u := NewUserService(storage, hasher)
			tt.setup(u)

			_, err := u.LoginUser(ctx, tt.userName, tt.password)
			if err != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected %v, got %v", tt.wantErr, err)
				}
			} else {
				if tt.wantErr != nil {
					t.Errorf("not expected error, got %v", err)
				}
			}
		})
	}
}
