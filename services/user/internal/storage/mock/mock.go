package mock

import (
	"context"
	"restaurant/services/user/internal/model"
	"sync"
)

type MockStorage struct {
	users        map[string]*model.User
	usersRWMutex sync.RWMutex
}

func New() *MockStorage {
	return &MockStorage{
		users: make(map[string]*model.User),
	}
}

func (m *MockStorage) CreateUser(ctx context.Context, user *model.User) error {
	m.usersRWMutex.Lock()
	defer m.usersRWMutex.Unlock()

	if _, ok := m.users[user.Name]; ok {
		return model.ErrUserAlreadyExists
	}

	m.users[user.Name] = user
	return nil
}

func (m *MockStorage) FindByName(ctx context.Context, name string) (*model.User, error) {
	m.usersRWMutex.RLock()
	defer m.usersRWMutex.RUnlock()

	if user, ok := m.users[name]; ok {
		return user, nil
	}

	return nil, model.ErrUserNotFound
}
