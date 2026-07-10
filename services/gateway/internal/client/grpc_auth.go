package client

import (
	"context"
	"time"

	"restaurant/pkg/interceptors"
	"restaurant/services/gateway/internal/model"

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
}

func NewAuthClient(creds credentials.TransportCredentials, addr string) (*AuthClient, error) {
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

func (a *AuthClient) ValidateToken(ctx context.Context, token string) (string, error) {
	resp, err := a.client.ValidateToken(ctx, &authv3.ValidateTokenRequest{
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

func (a *AuthClient) RefreshTokens(ctx context.Context, token string) (string, string, int32, error) {
	resp, err := a.client.RefreshTokens(ctx, &authv3.RefreshTokensRequest{
		RefreshToken: token,
	})
	if err != nil {
		if status.Code(err) == codes.PermissionDenied ||
			status.Code(err) == codes.Unauthenticated {
			return "", "", 0, model.ErrInvalidToken
		}
		return "", "", 0, err
	}

	return resp.AccessToken, resp.RefreshToken, resp.RefreshTokenTtl, nil
}

func (a *AuthClient) RevokeRefreshToken(ctx context.Context, token string) error {
	_, err := a.client.RevokeRefreshToken(ctx, &authv3.RevokeRefreshTokenRequest{
		RefreshToken: token,
	})
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return model.ErrTokenNotFound
		}
		return err
	}

	return nil
}
