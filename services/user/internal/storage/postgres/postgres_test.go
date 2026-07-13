package postgres

import (
	"context"
	"errors"
	"testing"

	"restaurant/services/user/internal/model"

	"github.com/google/uuid"
)

func setup(t *testing.T) *Storage {
	t.Helper()

	storage, err := New(Config{
		Addr:     "localhost:5432",
		User:     "restaurant",
		Password: "restaurant",
		Name:     "users_db",
		SSLMode:  "disable",
	})
	if err != nil {
		t.Skipf("cannot connect to postgres: %v", err)
	}

	if _, err := storage.ExecContext(context.Background(), "DELETE FROM users"); err != nil {
		t.Fatalf("failed to clean table: %v", err)
	}

	t.Cleanup(func() {
		storage.Close()
	})

	return storage
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()

	userID1 := uuid.Must(uuid.NewV7()).String()
	userID2 := uuid.Must(uuid.NewV7()).String()
	userName := "name_" + uuid.New().String()[:8]

	tests := []struct {
		name    string
		user    *model.User
		setup   func(*Storage)
		wantErr error
	}{
		{
			name:    "User not exists",
			user:    &model.User{ID: userID1, Name: userName + "_1"},
			setup:   func(s *Storage) {},
			wantErr: nil,
		},
		{
			name: "User with same ID already exists",
			user: &model.User{ID: userID1, Name: userName + "_2"},
			setup: func(s *Storage) {
				s.CreateUser(ctx, &model.User{ID: userID1, Name: userName + "_1"})
			},
			wantErr: model.ErrUserAlreadyExists,
		},
		{
			name: "User with same name already exists",
			user: &model.User{ID: userID2, Name: userName + "_1"},
			setup: func(s *Storage) {
				s.CreateUser(ctx, &model.User{ID: userID1, Name: userName + "_1"})
			},
			wantErr: model.ErrUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := setup(t)
			tt.setup(storage)

			err := storage.CreateUser(ctx, tt.user)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestFindByName(t *testing.T) {
	ctx := context.Background()

	userID := uuid.Must(uuid.NewV7()).String()
	userName := "name_" + uuid.New().String()[:8]

	tests := []struct {
		name    string
		user    string
		setup   func(*Storage)
		wantErr error
	}{
		{
			name: "User exists",
			user: userName,
			setup: func(s *Storage) {
				s.CreateUser(ctx, &model.User{ID: userID, Name: userName})
			},
			wantErr: nil,
		},
		{
			name:    "User not exists",
			user:    "notfound",
			setup:   func(s *Storage) {},
			wantErr: model.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := setup(t)
			tt.setup(storage)

			_, err := storage.FindByName(ctx, tt.user)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}
