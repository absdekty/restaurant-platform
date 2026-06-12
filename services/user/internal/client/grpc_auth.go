package client

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	authv3 "restaurant/api/proto/auth/v3"
)

type AuthClient struct {
	client authv3.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthClient(addr string) (*AuthClient, error) {
	serviceConfig := `{
		"methodConfig": [
			{
				"name": [{"service": "auth.v3.AuthService"}],
				"retryPolicy": {
					"maxAttempts": 3,
					"initialBackoff": "0.1s",
					"maxBackoff": "1s",
					"backoffMultiplier": 2,
					"retryableStatusCodes": ["UNAVAILABLE"]
				}
			}
		]
	}`

	keepaliveParams := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             1 * time.Second,
		PermitWithoutStream: true,
	}

	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(serviceConfig),
		grpc.WithKeepaliveParams(keepaliveParams))
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
