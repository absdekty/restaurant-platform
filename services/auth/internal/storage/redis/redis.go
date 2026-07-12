package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"restaurant/services/auth/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
)

type Storage struct {
	client *redis.Client
	prefix string
}

func New(client *redis.Client, prefix string) *Storage {
	return &Storage{
		client: client,
		prefix: prefix + ":",
	}
}

func (s *Storage) Close() error {
	return s.client.Close()
}

func (s *Storage) SaveRefreshToken(ctx context.Context, token *model.Token) error {
	key := s.getKey(token.Token)

	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	ttl := time.Until(token.ExpiresAt)
	if ttl <= 0 {
		return model.ErrTokenExpired
	}

	ok, err := s.client.SetNX(ctx, key, data, ttl).Result()
	if err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}
	if !ok {
		return model.ErrTokenAlreadyExists
	}

	return nil
}

func (s *Storage) GetRefreshToken(ctx context.Context, token string) (*model.Token, error) {
	key := s.getKey(token)

	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, model.ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	t := &model.Token{}
	if err := json.Unmarshal([]byte(data), t); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	if t.Revoked {
		return nil, model.ErrTokenRevoked
	}

	return t, nil
}

func (s *Storage) RevokeRefreshToken(ctx context.Context, token string) error {
	key := s.getKey(token)

	deleted, err := s.client.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	if deleted == 0 {
		return model.ErrTokenNotFound
	}

	return nil
}

func (s *Storage) RevokeAndSave(ctx context.Context, oldToken string, newToken *model.Token) error {
	pipe := s.client.Pipeline()

	oldKey := s.getKey(oldToken)
	newKey := s.getKey(newToken.Token)

	exists, err := s.client.Exists(ctx, oldKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check old token: %w", err)
	}
	if exists == 0 {
		return model.ErrTokenNotFound
	}

	if oldKey == newKey {
		return model.ErrTokenAlreadyExists
	}

	data, err := json.Marshal(newToken)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	ttl := time.Until(newToken.ExpiresAt)
	if ttl <= 0 {
		return model.ErrTokenExpired
	}

	delCmd := pipe.Del(ctx, oldKey)
	setCmd := pipe.SetNX(ctx, newKey, data, ttl)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to revoke and save: %w", err)
	}

	deleted, err := delCmd.Result()
	if err != nil {
		return fmt.Errorf("failed to delete old token: %w", err)
	}
	if deleted == 0 {
		return model.ErrTokenNotFound
	}

	ok, err := setCmd.Result()
	if err != nil {
		return fmt.Errorf("failed to save new token: %w", err)
	}
	if !ok {
		return model.ErrTokenAlreadyExists
	}

	return nil
}

func (s *Storage) getKey(token string) string {
	return s.prefix + "refresh:" + token
}
