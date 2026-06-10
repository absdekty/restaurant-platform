package sqlite3

import (
	"context"
	"restaurant/services/auth/internal/model"
	"testing"
)

func TestSaveRefreshToken(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		token   *model.Token
		setup   func(*Storage)
		wantErr error
	}{
		{
			name:    "Valid token",
			token:   &model.Token{Token: "token"},
			setup:   func(storage *Storage) {},
			wantErr: nil,
		},
		{
			name:  "Token already exists",
			token: &model.Token{Token: "token"},
			setup: func(storage *Storage) {
				if err := storage.SaveRefreshToken(context.Background(), &model.Token{Token: "token"}); err != nil {
					t.Errorf("setup: %v", err)
				}
			},
			wantErr: model.ErrTokenAlreadyExists,
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

			err = storage.SaveRefreshToken(ctx, tt.token)
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

func TestGetRefreshToken(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		token   string
		setup   func(*Storage)
		wantErr error
	}{
		{
			name:  "Token exists",
			token: "token",
			setup: func(storage *Storage) {
				if err := storage.SaveRefreshToken(context.Background(), &model.Token{Token: "token"}); err != nil {
					t.Errorf("setup: %v", err)
				}
			},
			wantErr: nil,
		},
		{
			name:    "Token not exists",
			token:   "token",
			setup:   func(storage *Storage) {},
			wantErr: model.ErrTokenNotFound,
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

			_, err = storage.GetRefreshToken(ctx, tt.token)
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

func TestRevokeRefreshToken(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		token   string
		setup   func(*Storage)
		wantErr error
	}{
		{
			name:  "Token exists",
			token: "token",
			setup: func(storage *Storage) {
				if err := storage.SaveRefreshToken(context.Background(), &model.Token{Token: "token"}); err != nil {
					t.Errorf("setup: %v", err)
				}
			},
			wantErr: nil,
		},
		{
			name:    "Token not exists",
			token:   "token",
			setup:   func(storage *Storage) {},
			wantErr: model.ErrTokenNotFound,
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

			err = storage.RevokeRefreshToken(ctx, tt.token)
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

func TestRevokeAndSave(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		tokenStr string
		token    *model.Token
		setup    func(*Storage)
		wantErr  error
	}{
		{
			name:     "Valid revoke & save",
			tokenStr: "token",
			token:    &model.Token{Token: "token1"},
			setup: func(storage *Storage) {
				if err := storage.SaveRefreshToken(context.Background(), &model.Token{Token: "token"}); err != nil {
					t.Errorf("setup: %v", err)
				}
			},
			wantErr: nil,
		},
		{
			name:     "OldToken not exists",
			tokenStr: "token",
			token:    &model.Token{Token: "token1"},
			setup:    func(storage *Storage) {},
			wantErr:  model.ErrTokenNotFound,
		},
		{
			name:     "Revoke & save the same token",
			tokenStr: "token",
			token:    &model.Token{Token: "token"},
			setup: func(storage *Storage) {
				if err := storage.SaveRefreshToken(context.Background(), &model.Token{Token: "token"}); err != nil {
					t.Errorf("setup: %v", err)
				}
			},
			wantErr: model.ErrTokenAlreadyExists,
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

			err = storage.RevokeAndSave(ctx, tt.tokenStr, tt.token)
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
