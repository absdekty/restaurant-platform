package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"restaurant/services/user/internal/model"
	"time"
)

type UserStorage interface {
	CreateUser(ctx context.Context, user *model.User) error
	FindByName(ctx context.Context, name string) (*model.User, error)
}

type UserHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) bool
}

type UserService struct {
	storage UserStorage
	hasher  UserHasher
}

func NewUserService(storage UserStorage, hasher UserHasher) *UserService {
	return &UserService{storage: storage, hasher: hasher}
}

func (u *UserService) CreateUser(ctx context.Context, name, password string) (*model.User, error) {
	if err := model.ValidateName(name); err != nil {
		return nil, fmt.Errorf("validate name: %w", err)
	}

	if err := model.ValidatePassword(password); err != nil {
		return nil, fmt.Errorf("validate password: %w", err)
	}

	hashedPassword, err := u.hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		ID:        uuid.New().String(),
		Name:      name,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}

	if err = u.storage.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}

	return user, nil
}

func (u *UserService) LoginUser(ctx context.Context, name, password string) (string, error) {
	if err := model.ValidateName(name); err != nil {
		return "", fmt.Errorf("validate name: %w", err)
	}

	if err := model.ValidatePassword(password); err != nil {
		return "", fmt.Errorf("validate password: %w", err)
	}

	user, err := u.storage.FindByName(ctx, name)
	if err != nil {
		return "", fmt.Errorf("find user: %w", err)
	}

	if ok := u.hasher.Compare(user.Password, password); !ok {
		return "", model.ErrInvalidCredentials
	}

	return user.ID, nil
}
