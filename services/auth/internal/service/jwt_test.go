package service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"restaurant/services/auth/internal/storage/mock"
	"testing"
	"time"
)

func TestJWTService(t *testing.T) {
	ctx := context.Background()

	mockStorage := mock.NewMock()
	service := NewJWT("secretkey", 15*time.Minute, 7*24*time.Hour, mockStorage)
	userID := "userid"

	t.Run("Генерация токенов", func(t *testing.T) {
		accessToken, refreshToken, _, err := service.GenerateTokens(ctx, userID)
		require.NoError(t, err)

		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		assert.NotEqual(t, accessToken, refreshToken)
	})

	t.Run("Валидация сгенерированных токенов", func(t *testing.T) {
		accessToken, refreshToken, _, _ := service.GenerateTokens(ctx, userID)

		userIDToken, err := service.ValidateAccessToken(accessToken)
		assert.NoError(t, err)
		assert.Equal(t, userID, userIDToken)

		userIDToken, err = service.ValidateRefreshToken(ctx, refreshToken)
		assert.NoError(t, err)
		assert.Equal(t, userID, userIDToken)
	})

	t.Run("Валидация сгенерированных, подмененных друг другом токенов", func(t *testing.T) {
		accessToken, refreshToken, _, _ := service.GenerateTokens(ctx, userID)

		_, err := service.ValidateRefreshToken(ctx, accessToken)
		assert.Error(t, err)

		_, err = service.ValidateAccessToken(refreshToken)
		assert.Error(t, err)
	})

	t.Run("RefreshTokens с обоими токенами", func(t *testing.T) {
		accessToken, refreshToken, _, _ := service.GenerateTokens(ctx, userID)

		_, _, _, err := service.RefreshTokens(ctx, accessToken)
		assert.Error(t, err)

		_, _, _, err = service.RefreshTokens(ctx, refreshToken)
		assert.NoError(t, err)
	})

	t.Run("Валидация RefreshTokens", func(t *testing.T) {
		accessToken, refreshToken, _, _ := service.GenerateTokens(ctx, userID)
		accessToken, refreshToken, _, _ = service.RefreshTokens(ctx, refreshToken)

		userIDToken, err := service.ValidateAccessToken(accessToken)
		assert.NoError(t, err)
		assert.Equal(t, userID, userIDToken)

		userIDToken, err = service.ValidateRefreshToken(ctx, refreshToken)
		assert.NoError(t, err)
		assert.Equal(t, userID, userIDToken)
	})

	t.Run("Отзыв valid/invalid токенов", func(t *testing.T) {
		err := service.RevokeRefreshToken(ctx, "invalid")
		assert.Error(t, err)

		accessToken, refreshToken, _, _ := service.GenerateTokens(ctx, userID)

		err = service.RevokeRefreshToken(ctx, accessToken)
		assert.Error(t, err)

		err = service.RevokeRefreshToken(ctx, refreshToken)
		assert.NoError(t, err)
	})

	t.Run("RefreshTokens с отозванным токеном", func(t *testing.T) {
		_, refreshToken, _, _ := service.GenerateTokens(ctx, userID)

		err := service.RevokeRefreshToken(ctx, refreshToken)
		assert.NoError(t, err)

		_, _, _, err = service.RefreshTokens(ctx, refreshToken)
		assert.Error(t, err)
	})
}
