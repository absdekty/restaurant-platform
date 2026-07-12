package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

type RedisClient struct {
	*redis.Client
}

func NewRedis(cfg *RedisConfig) (*RedisClient, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf("addr cant be empty")
	}

	if cfg.PoolSize == 0 {
		cfg.PoolSize = 10
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &RedisClient{Client: rdb}, nil
}
