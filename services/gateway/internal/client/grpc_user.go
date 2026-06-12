package client

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
	"restaurant/services/gateway/internal/model"

	userv1 "restaurant/api/proto/user/v1"
)

type UserClient struct {
	client userv1.UserServiceClient
	conn   *grpc.ClientConn
}

func NewUserClient(creds credentials.TransportCredentials, addr string) (*UserClient, error) {
	serviceConfig := `{
		"methodConfig": [
			{
				"name": [{"service": "user.v1.UserService"}],
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
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultServiceConfig(serviceConfig),
		grpc.WithKeepaliveParams(keepaliveParams))
	if err != nil {
		return nil, err
	}

	return &UserClient{
		client: userv1.NewUserServiceClient(conn),
		conn:   conn,
	}, nil
}

func (a *UserClient) Close() error {
	return a.conn.Close()
}

func (a *UserClient) RegisterUser(ctx context.Context, name, password string) (string, error) {
	resp, err := a.client.RegisterUser(ctx, &userv1.RegisterUserRequest{
		Name:     name,
		Password: password,
	})
	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			return "", model.ErrUserInvalidRegisterDetails
		}

		if status.Code(err) == codes.AlreadyExists {
			return "", model.ErrUserAlreadyExists
		}

		return "", err
	}

	return resp.UserId, nil
}

func (a *UserClient) LoginUser(ctx context.Context, name, password string) (string, string, int32, error) {
	resp, err := a.client.LoginUser(ctx, &userv1.LoginUserRequest{
		Name:     name,
		Password: password,
	})
	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			return "", "", 0, model.ErrUserInvalidCredentials
		}

		if status.Code(err) == codes.NotFound {
			return "", "", 0, model.ErrUserNotFound
		}

		return "", "", 0, err
	}

	return resp.AccessToken, resp.RefreshToken, resp.RefreshTokenTtl, nil
}
