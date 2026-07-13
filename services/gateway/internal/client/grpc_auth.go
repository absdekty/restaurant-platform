package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"restaurant/pkg/interceptors"
	"restaurant/pkg/models"
	"restaurant/services/gateway/internal/model"

	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"

	authv3 "restaurant/api/proto/auth/v3"
)

type AuthClient struct {
	client authv3.AuthServiceClient
	conn   *grpc.ClientConn
	cb     CircuitBreaker
}

type AuthConfig struct {
	RetryMaxAttempts       int
	RetryInitialBackoff    string
	RetryMaxBackoff        string
	RetryBackoffMultiplier float64

	KeepaliveTime          time.Duration
	KeepaliveTimeout       time.Duration
	KeepalivePermitWithout bool
}

func NewAuthClient(creds credentials.TransportCredentials, addr string, config AuthConfig, cb CircuitBreaker) (*AuthClient, error) {
	serviceConfig := fmt.Sprintf(`{
		"methodConfig": [{
			"name": [{"service": "auth.v3.AuthService"}],
			"retryPolicy": {
				"maxAttempts": %d,
				"initialBackoff": "%s",
				"maxBackoff": "%s",
				"backoffMultiplier": %.1f,
				"retryableStatusCodes": ["UNAVAILABLE"]
			}
		}]
	}`,
		config.RetryMaxAttempts,
		config.RetryInitialBackoff,
		config.RetryMaxBackoff,
		config.RetryBackoffMultiplier,
	)

	keepaliveParams := keepalive.ClientParameters{
		Time:                config.KeepaliveTime,
		Timeout:             config.KeepaliveTimeout,
		PermitWithoutStream: config.KeepalivePermitWithout,
	}

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultServiceConfig(serviceConfig),
		grpc.WithKeepaliveParams(keepaliveParams),
		grpc.WithUnaryInterceptor(interceptors.TraceClient()))
	if err != nil {
		return nil, err
	}

	return &AuthClient{
		client: authv3.NewAuthServiceClient(conn),
		conn:   conn,
		cb:     cb,
	}, nil
}

func (a *AuthClient) Close() error {
	return a.conn.Close()
}

func (a *AuthClient) ValidateToken(ctx context.Context, token string) (string, error) {
	result, err := a.cb.Execute(func() (interface{}, error) {
		return a.client.ValidateToken(ctx, &authv3.ValidateTokenRequest{
			AccessToken: token,
		})
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			return "", models.ErrServiceUnavailable
		}

		if status.Code(err) == codes.Unavailable {
			return "", models.ErrServiceUnavailable
		}

		if status.Code(err) == codes.Unauthenticated ||
			status.Code(err) == codes.PermissionDenied ||
			status.Code(err) == codes.NotFound {
			return "", model.ErrUnauthorized
		}
		return "", err
	}

	resp := result.(*authv3.ValidateTokenResponse)
	return resp.UserId, nil
}

func (a *AuthClient) GenerateTokens(ctx context.Context, token string) (string, string, int32, error) {
	result, err := a.cb.Execute(func() (interface{}, error) {
		return a.client.GenerateTokens(ctx, &authv3.GenerateTokensRequest{
			UserId: token,
		})
	})

	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			return "", "", 0, models.ErrServiceUnavailable
		}

		if status.Code(err) == codes.Unavailable {
			return "", "", 0, models.ErrServiceUnavailable
		}

		return "", "", 0, err
	}

	resp := result.(*authv3.GenerateTokensResponse)
	return resp.AccessToken, resp.RefreshToken, resp.RefreshTokenTtl, nil
}

func (a *AuthClient) RefreshTokens(ctx context.Context, token string) (string, string, int32, error) {
	result, err := a.cb.Execute(func() (interface{}, error) {
		return a.client.RefreshTokens(ctx, &authv3.RefreshTokensRequest{
			RefreshToken: token,
		})
	})

	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			return "", "", 0, models.ErrServiceUnavailable
		}

		if status.Code(err) == codes.Unavailable {
			return "", "", 0, models.ErrServiceUnavailable
		}

		if status.Code(err) == codes.PermissionDenied ||
			status.Code(err) == codes.Unauthenticated {
			return "", "", 0, model.ErrInvalidToken
		}
		return "", "", 0, err
	}

	resp := result.(*authv3.RefreshTokensResponse)
	return resp.AccessToken, resp.RefreshToken, resp.RefreshTokenTtl, nil
}

func (a *AuthClient) RevokeRefreshToken(ctx context.Context, token string) error {
	_, err := a.cb.Execute(func() (interface{}, error) {
		return a.client.RevokeRefreshToken(ctx, &authv3.RevokeRefreshTokenRequest{
			RefreshToken: token,
		})
	})

	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			return models.ErrServiceUnavailable
		}

		if status.Code(err) == codes.Unavailable {
			return models.ErrServiceUnavailable
		}

		if status.Code(err) == codes.Unauthenticated {
			return model.ErrTokenNotFound
		}
		return err
	}

	return nil
}
