package mock

import (
	"context"
	"restaurant/services/auth/internal/model"
	"sync"
)

type MockRepository struct {
	tokens        map[string]*model.Token
	tokensRWMutex sync.RWMutex
}

func NewMock() *MockRepository {
	return &MockRepository{
		tokens: make(map[string]*model.Token),
	}
}

func (m *MockRepository) SaveRefreshToken(ctx context.Context, token *model.Token) error {
	m.tokensRWMutex.Lock()
	defer m.tokensRWMutex.Unlock()

	return m.saveRefreshToken(ctx, token)
}

func (m *MockRepository) GetRefreshToken(ctx context.Context, tokenStr string) (*model.Token, error) {
	m.tokensRWMutex.RLock()
	defer m.tokensRWMutex.RUnlock()

	token, ok := m.tokens[tokenStr]
	if !ok {
		return nil, model.ErrTokenNotFound
	}
	return token, nil
}

func (m *MockRepository) RevokeRefreshToken(ctx context.Context, tokenStr string) error {
	m.tokensRWMutex.Lock()
	defer m.tokensRWMutex.Unlock()

	return m.revokeRefreshToken(ctx, tokenStr)
}

func (m *MockRepository) revokeRefreshToken(ctx context.Context, tokenStr string) error {
	token, ok := m.tokens[tokenStr]
	if !ok {
		return model.ErrTokenNotFound
	}

	token.Revoked = true
	m.tokens[tokenStr] = token

	return nil
}

func (m *MockRepository) RevokeAndSave(ctx context.Context, oldToken string, newToken *model.Token) error {
	m.tokensRWMutex.Lock()
	defer m.tokensRWMutex.Unlock()

	if err := m.revokeRefreshToken(ctx, oldToken); err != nil {
		return err
	}

	if err := m.saveRefreshToken(ctx, newToken); err != nil {
		return err
	}

	return nil
}

func (m *MockRepository) saveRefreshToken(ctx context.Context, token *model.Token) error {
	if _, ok := m.tokens[token.Token]; ok {
		return model.ErrTokenAlreadyExists
	}

	m.tokens[token.Token] = token
	return nil
}
