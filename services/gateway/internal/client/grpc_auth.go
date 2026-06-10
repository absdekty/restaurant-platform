package client

import (
	"context"
	"time"

	"restaurant/services/gateway/internal/model"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"

	authv2 "restaurant/api/proto/auth/v2"
)

type AuthClient struct {
	client authv2.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthClient(addr string) (*AuthClient, error) {
	serviceConfig := `{
		"methodConfig": [
			{
				"name": [{"service": "auth.v2.AuthService"}],
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
		client: authv2.NewAuthServiceClient(conn),
		conn:   conn,
	}, nil
}

func (a *AuthClient) Close() error {
	return a.conn.Close()
}

func (a *AuthClient) ValidateToken(ctx context.Context, token string) (string, error) {
	resp, err := a.client.ValidateToken(ctx, &authv2.ValidateTokenRequest{
		AccessToken: token,
	})
	if err != nil {
		if status.Code(err) == codes.Unauthenticated ||
			status.Code(err) == codes.PermissionDenied ||
			status.Code(err) == codes.NotFound {
			return "", model.ErrUnauthorized
		}
		return "", err
	}

	return resp.UserId, nil
}
