package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	authv3 "restaurant/api/proto/auth/v3"
	"restaurant/pkg/interceptors"
)

type AuthClient struct {
	client authv3.AuthServiceClient
	conn   *grpc.ClientConn
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

func NewAuthClient(creds credentials.TransportCredentials, addr string, config AuthConfig) (*AuthClient, error) {
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
	}, nil
}

func (a *AuthClient) Close() error {
	return a.conn.Close()
}

func (a *AuthClient) GenerateTokens(ctx context.Context, userID string) (string, string, int32, error) {
	resp, err := a.client.GenerateTokens(ctx, &authv3.GenerateTokensRequest{
		UserId: userID,
	})
	if err != nil {
		return "", "", 0, err
	}

	return resp.AccessToken, resp.RefreshToken, resp.RefreshTokenTtl, nil
}
