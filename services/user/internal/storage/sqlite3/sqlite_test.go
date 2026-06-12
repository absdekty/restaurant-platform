package sqlite3

import (
	"context"
	"restaurant/services/user/internal/model"
	"testing"
)

func TestCreateUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		user    *model.User
		setup   func(*Storage)
		wantErr error
	}{
		{
			name:    "User not exists",
			user:    &model.User{ID: "userid"},
			setup:   func(storage *Storage) {},
			wantErr: nil,
		},
		{
			name: "User with current ID already exists",
			user: &model.User{ID: "userid", Name: "name"},
			setup: func(storage *Storage) {
				if err := storage.CreateUser(context.Background(), &model.User{ID: "userid"}); err != nil {
					t.Errorf("setup: %v", err)
				}
			},
			wantErr: model.ErrUserAlreadyExists,
		},
		{
			name: "User with current name already exists",
			user: &model.User{ID: "userid", Name: "name"},
			setup: func(storage *Storage) {
				if err := storage.CreateUser(context.Background(), &model.User{ID: "userid1", Name: "name"}); err != nil {
					t.Errorf("setup: %v", err)
				}
			},
			wantErr: model.ErrUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := New(":memory:")
			if err != nil {
				t.Skipf("cannot start db: %v", err)
			}

			t.Cleanup(func() {
				if err := storage.Close(); err != nil {
					t.Errorf("failed to close db: %v", err)
				}
			})

			tt.setup(storage)

			err = storage.CreateUser(ctx, tt.user)
			if tt.wantErr != nil {
				if tt.wantErr != err {
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

	tests := []struct {
		name    string
		user    string
		setup   func(*Storage)
		wantErr error
	}{
		{
			name: "User exists",
			user: "name",
			setup: func(storage *Storage) {
				if err := storage.CreateUser(context.Background(), &model.User{Name: "name"}); err != nil {
					t.Errorf("setup: %v", err)
				}
			},
			wantErr: nil,
		},
		{
			name:    "User not exists",
			user:    "name",
			setup:   func(storage *Storage) {},
			wantErr: model.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := New(":memory:")
			if err != nil {
				t.Skipf("cannot start db: %v", err)
			}

			t.Cleanup(func() {
				if err := storage.Close(); err != nil {
					t.Errorf("failed to close db: %v", err)
				}
			})

			tt.setup(storage)

			_, err = storage.FindByName(ctx, tt.user)
			if tt.wantErr != nil {
				if tt.wantErr != err {
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
