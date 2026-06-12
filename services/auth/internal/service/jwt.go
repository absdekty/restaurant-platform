package service

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"restaurant/services/auth/internal/model"
	"time"
)

/* Интерфейс для хранения refresh-token */
type JWTStorage interface {
	SaveRefreshToken(ctx context.Context, token *model.Token) error
	GetRefreshToken(ctx context.Context, token string) (*model.Token, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	RevokeAndSave(ctx context.Context, oldToken string, newToken *model.Token) error
}

/* Структура сервиса */
type JWTService struct {
	secretKey []byte
	tokenTTLA time.Duration
	tokenTTLR time.Duration
	storage   JWTStorage
}

/* Токен-Claims */
type JWTClaims struct {
	jwt.RegisteredClaims
	TokenType string
}

/* Конструктор */
func NewJWT(secretKey string, accessTTL, refreshTTL time.Duration, storage JWTStorage) *JWTService {
	return &JWTService{
		secretKey: []byte(secretKey),
		tokenTTLA: accessTTL,
		tokenTTLR: refreshTTL,
		storage:   storage,
	}
}

/* Генерирует access+refresh токены, сохраняя refresh */
func (j *JWTService) GenerateTokens(ctx context.Context, userID string) (string, string, int32, error) {
	accessToken, refreshToken, err := j.generateTokens(userID)
	if err != nil {
		return "", "", 0, err
	}

	if err := j.saveRefreshToken(ctx, userID, refreshToken); err != nil {
		return "", "", 0, err
	}

	return accessToken, refreshToken, int32(j.tokenTTLR / time.Second), nil
}

/* Валидирует access-token */
func (j *JWTService) ValidateAccessToken(tokenStr string) (string, error) {
	claims := &JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse access token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid access token")
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return "", fmt.Errorf("access token expired")
	}

	if claims.TokenType != "access" {
		return "", fmt.Errorf("invalid token type: expected access, got %s", claims.TokenType)
	}

	return claims.Subject, nil
}

/* Валидирует refresh-token */
func (j *JWTService) ValidateRefreshToken(ctx context.Context, tokenStr string) (string, error) {
	claims := &JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse refresh token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid refresh token")
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return "", fmt.Errorf("refresh token expired")
	}

	if claims.TokenType != "refresh" {
		return "", fmt.Errorf("invalid token type: expected refresh, got %s", claims.TokenType)
	}

	storagedToken, err := j.storage.GetRefreshToken(ctx, tokenStr)
	if err != nil {
		return "", err
	}
	if storagedToken.Revoked {
		return "", model.ErrTokenRevoked
	}

	return claims.Subject, nil
}

/* Отзывает refresh-token */
func (j *JWTService) RevokeRefreshToken(ctx context.Context, token string) error {
	return j.storage.RevokeRefreshToken(ctx, token)
}

/* Обновляет access+refresh токены, валидируя и отзывая старый */
func (j *JWTService) RefreshTokens(ctx context.Context, token string) (string, string, int32, error) {
	userID, err := j.ValidateRefreshToken(ctx, token)
	if err != nil {
		return "", "", 0, err
	}

	accessToken, refreshToken, err := j.generateTokens(userID)
	if err != nil {
		return "", "", 0, err
	}

	tokenToSave := j.newRefreshToken(userID, refreshToken)

	if err := j.storage.RevokeAndSave(ctx, token, tokenToSave); err != nil {
		return "", "", 0, err
	}

	return accessToken, refreshToken, int32(j.tokenTTLR / time.Second), nil
}

/* Генерирует access+refresh токены, не сохраняя refresh */
func (j *JWTService) generateTokens(userID string) (string, string, error) {
	accessToken, err := j.generateAccessToken(userID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := j.generateRefreshToken(userID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

/* Генерирует access-token */
func (j *JWTService) generateAccessToken(userID string) (string, error) {
	claims := &JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenTTLA)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
		TokenType: "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %w", err)
	}

	return tokenString, nil
}

/* Генерирует refresh-token */
func (j *JWTService) generateRefreshToken(userID string) (string, error) {
	claims := &JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenTTLR)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
		TokenType: "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return tokenString, nil
}

/* Сохраняет refresh-token в storage */
func (j *JWTService) saveRefreshToken(ctx context.Context, userID, token string) error {
	tokenToSave := j.newRefreshToken(userID, token)

	if err := j.storage.SaveRefreshToken(ctx, tokenToSave); err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}

	return nil
}

/* Конструктор refresh-token */
func (j *JWTService) newRefreshToken(userID, token string) *model.Token {
	return &model.Token{
		UserID:    userID,
		Token:     token,
		Revoked:   false,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(j.tokenTTLR),
	}
}
